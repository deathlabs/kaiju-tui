package cmd

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type Model struct {
	Choices      []string         // Items on the list.
	Cursor       int              // The item our cursor is pointing at.
	Selected     map[int]struct{} // The items selected.
	ServerStatus int
	ServerError  string
	Progress     progress.Model
	Checking     bool
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Handle incoming messages by Go type.
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:
		// What key was pressed?
		switch msg.String() {

		// These keys exit the program.
		case "ctrl+c", "q":
			return model, tea.Quit

			// The "up" and "k" keys move the cursor up.
		case "up", "k":
			model.ServerStatus = 0
			model.ServerError = ""

			if model.Cursor > 0 {
				model.Cursor--
			}

			// The "down" and "j" keys move the cursor down.
		case "down", "j":
			model.ServerStatus = 0
			model.ServerError = ""

			if model.Cursor < len(model.Choices)-1 {
				model.Cursor++
			}

		// These keys toggle the selected state for current item.
		case "x", "enter", "space":
			// Reset the server status and error before checking the server.
			model.ServerStatus = 0
			model.ServerError = ""
			model.Selected[model.Cursor] = struct{}{}
			model.Checking = true

			progressCmd := model.Progress.SetPercent(0)

			// Return the updated model and selected command.
			return model, tea.Batch(
				checkServerCmd(model.Choices[model.Cursor]),
				tickCmd(),
				progressCmd,
			)
		}

	// Is it a status message?
	case statusMsg:
		// Add the server status to the model (the application's state).
		model.ServerStatus = int(msg)
		model.ServerError = ""
		model.Selected = make(map[int]struct{})
		model.Checking = false

	// Is it an error message?
	case errMsg:
		// Add the error to the model (the application's state).
		model.ServerError = msg.Error()
		model.ServerStatus = 0
		model.Selected = make(map[int]struct{})
		model.Checking = false

	case tickMsg:
		if !model.Checking {
			return model, nil
		}

		var cmd tea.Cmd

		if model.Progress.Percent() < 0.9 {
			cmd = model.Progress.IncrPercent(0.05)
		}

		return model, tea.Batch(tickCmd(), cmd)

	case progress.FrameMsg:
		var cmd tea.Cmd

		model.Progress, cmd = model.Progress.Update(msg)

		return model, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return model, nil
}

func (model Model) View() tea.View {
	var (
		checkbox string
		choice   string
		cursor   string
		index    int
		selected bool
		ui       string
	)

	ui = "What server do you want to check?\n\n"

	// Iterate over our choices.
	for index, choice = range model.Choices {

		// Reset the cursor and checked indicators for this row.
		cursor = " "
		checkbox = " "
		_, selected = model.Selected[index]

		// Is the cursor pointing at this choice?
		if model.Cursor == index {
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
	if model.Checking {
		ui += fmt.Sprintf(
			"\nChecking %s...\n%s\n",
			model.Choices[model.Cursor],
			model.Progress.View(),
		)
	} else if model.ServerError != "" {
		ui += fmt.Sprintf("\nError: %s\n", model.ServerError)
	} else if model.ServerStatus != 0 {
		ui += fmt.Sprintf(
			"\nServer status: %d (%s)\n",
			model.ServerStatus,
			model.Choices[model.Cursor],
		)
	}

	// The footer.
	ui += "\nPress q to quit.\n"

	// Return the UI so it can be rendered.
	return tea.NewView(ui)
}
