package screens

import (
	"fmt"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/messages"
	"github.com/deathlabs/kaiju-tui/internal/tools"
)

type FacilitatorScreen struct {
	Choices      []string         // Items on the list.
	Cursor       int              // The item our cursor is pointing at.
	Selected     map[int]struct{} // The items selected.
	ServerStatus int
	ServerError  string
	Progress     progress.Model
	Checking     bool
	Token        string // Bearer token from the device flow.
}

func (screen FacilitatorScreen) Init() tea.Cmd {
	return nil
}

func (screen FacilitatorScreen) Update(msg tea.Msg) (FacilitatorScreen, tea.Cmd) {
	var cmd tea.Cmd

	// Handle incoming messages by Go type.
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:
		// What key was pressed?
		switch msg.String() {

		// These keys exit the program.
		case "ctrl+c", "q":
			return screen, tea.Quit

			// The "up" and "k" keys move the cursor up.
		case "up", "k":
			screen.ServerStatus = 0
			screen.ServerError = ""

			if screen.Cursor > 0 {
				screen.Cursor--
			}

			// The "down" and "j" keys move the cursor down.
		case "down", "j":
			screen.ServerStatus = 0
			screen.ServerError = ""

			if screen.Cursor < len(screen.Choices)-1 {
				screen.Cursor++
			}

		// These keys toggle the selected state for current item.
		case "x", "enter", "space":
			// Reset the server status and error before checking the server.
			screen.ServerStatus = 0
			screen.ServerError = ""
			screen.Selected[screen.Cursor] = struct{}{}
			screen.Checking = true

			progressCmd := screen.Progress.SetPercent(0)

			// Return the updated model and selected command.
			return screen, tea.Batch(
				tools.CheckServerCmd(screen.Choices[screen.Cursor], screen.Token),
				tools.TickCmd(),
				progressCmd,
			)
		default:
			return screen, nil
		}

	// Is it a status message?
	case messages.StatusMessage:
		// Add the server status to the model (the application's state).
		screen.ServerStatus = int(msg)
		screen.ServerError = ""
		screen.Selected = make(map[int]struct{})
		screen.Checking = false

	// Is it an error message?
	case messages.ErrorMessage:
		// Add the error to the model (the application's state).
		screen.ServerError = msg.Message.Error()
		screen.ServerStatus = 0
		screen.Selected = make(map[int]struct{})
		screen.Checking = false

	case messages.TickMessage:
		if !screen.Checking {
			return screen, nil
		}

		if screen.Progress.Percent() < 0.9 {
			cmd = screen.Progress.IncrPercent(0.05)
		}

		return screen, tea.Batch(tools.TickCmd(), cmd)

	case progress.FrameMsg:
		screen.Progress, cmd = screen.Progress.Update(msg)
		return screen, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return screen, nil
}

func (screen FacilitatorScreen) View() tea.View {
	var (
		checkbox string
		choice   string
		cursor   string
		index    int
		selected bool
		ui       string
	)

	ui = "Facilitator Screen\n\n"

	// Iterate over our choices.
	for index, choice = range screen.Choices {

		// Reset the cursor and checked indicators for this row.
		cursor = " "
		checkbox = " "
		_, selected = screen.Selected[index]

		// Is the cursor pointing at this choice?
		if screen.Cursor == index {
			// Highlight the cursor for the current choice.
			cursor = ">"
		}

		// Was this choice selected?
		if selected {
			// If so, mark it with an "x".
			checkbox = "x" // This means the choice was selected.
		}

		// Render the row.
		ui += fmt.Sprintf("%s [%s] %s\n", cursor, checkbox, choice)
	}

	// Display server status or error if there is one.
	if screen.Checking {
		ui += fmt.Sprintf(
			"\nChecking %s...\n%s\n",
			screen.Choices[screen.Cursor],
			screen.Progress.View(),
		)
	} else if screen.ServerError != "" {
		ui += fmt.Sprintf("\nError: %s\n", screen.ServerError)
	} else if screen.ServerStatus != 0 {
		ui += fmt.Sprintf(
			"\nServer status: %d (%s)\n",
			screen.ServerStatus,
			screen.Choices[screen.Cursor],
		)
	}

	// The footer.
	ui += "\nMove: ↑/↓ | Select: Enter | Facilitator: f | Participant: p | Quit: q\n"

	// Return the UI so it can be rendered.
	return tea.NewView(ui)
}
