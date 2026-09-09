package cmd

import (
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// run initializes and runs the TUI.
func run(cmd *cobra.Command, args []string) error {
	var (
		err     error
		model   tea.Model
		program *tea.Program
	)

	// Initialize the model with the default state.
	model = Model{
		// Our to-do list is a grocery list.
		Choices: []string{
			"https://google.com",
			"https://kaiju.uds.dev",
		},

		// This map which indicates which choices are currently selected. We're
		// using the map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		Selected: make(map[int]struct{}),
		Progress: progress.New(progress.WithDefaultBlend()),
	}

	// Use the model to create a new Bubble Tea program and run it.
	program = tea.NewProgram(model)
	_, err = program.Run()
	if err != nil {
		return err
	}

	return nil
}
