package cmd

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/auth"
	"github.com/deathlabs/kaiju-tui/internal/tui"
	"github.com/spf13/cobra"
)

// run initializes and runs the TUI.
func run(cmd *cobra.Command, args []string) error {
	var (
		app         tui.App
		authCfg     auth.KeycloakConfig
		err         error
		program     *tea.Program
		serverModel tui.Model
	)

	authCfg = auth.KeycloakConfig{
		BaseURL:  "https://sso.uds.dev",
		Realm:    "uds",
		ClientID: "kaiju-tui",
	}

	serverModel = tui.Model{
		Choices: []string{
			"https://kaiju.uds.dev/api/v1/docs",
		},
		// The keys are the indexes of the `choices` slice above.
		Selected: make(map[int]struct{}),
		Progress: progress.New(progress.WithDefaultBlend()),
	}

	app = tui.NewApp(authCfg, serverModel)

	// Use the app model to create a new Bubble Tea program and run it.
	program = tea.NewProgram(app)
	_, err = program.Run()
	if err != nil {
		return err
	}

	return nil
}

var (
	rootCmd = &cobra.Command{
		Use:     "kaiju-tui",
		Short:   "A Terminal User Interface (TUI) for Kaiju.",
		Version: version,
		RunE:    run,
	}
	version string
)

func Execute() {
	var err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
