package cmd

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Choices      []string         // Items on the list.
	Cursor       int              // The item our cursor is pointing at.
	Selected     map[int]struct{} // The items selected.
	ServerStatus int
	ServerError  string
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var ok bool

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:
		// What key was pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return model, tea.Quit

		// The "up" and "k" keys move the cursor up.
		case "up", "k":
			if model.Cursor > 0 {
				model.Cursor--
			}

		// The "down" and "j" keys move the cursor down.
		case "down", "j":
			if model.Cursor < len(model.Choices)-1 {
				model.Cursor++
			}

		// The enter key and space bar toggle the selected state for the
		// item that the cursor is pointing at.
		case "enter", "space":
			_, ok = model.Selected[model.Cursor]
			if ok {
				delete(model.Selected, model.Cursor)
			} else {
				model.Selected[model.Cursor] = struct{}{}
			}

			return model, func() tea.Msg {
				return checkServer("http://google.com")
			}
		}

	case statusMsg:
		model.ServerStatus = int(msg)

	case errMsg:
		model.ServerError = msg.Error()
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return model, nil
}

func (model Model) View() tea.View {
	var (
		checked string
		choice  string
		cursor  string
		index   int
		ok      bool
		ui      string
	)

	ui = "What role were you assigned for the exercise?\n\n"

	// Iterate over our choices.
	for index, choice = range model.Choices {

		// Is the cursor pointing at this choice?
		cursor = " " // no cursor.
		if model.Cursor == index {
			cursor = ">" // cursor!
		}

		// Was this choice selected?
		checked = " " // This means the choice was not selected.
		if _, ok = model.Selected[index]; ok {
			checked = "x" // This means the choice was selected.
		}

		// Render the row.
		ui += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// Display server status or error if available.
	if model.ServerError != "" {
		ui += fmt.Sprintf("\nError: %s\n", model.ServerError)
	} else if model.ServerStatus != 0 {
		ui += fmt.Sprintf("\nServer status: %d\n", model.ServerStatus)
	}

	// The footer.
	ui += "\nPress q to quit.\n"

	// Send the UI for rendering
	return tea.NewView(ui)
}
