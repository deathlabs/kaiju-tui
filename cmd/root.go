package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/deathlabs/kaiju-tui/internal/auth"
	"github.com/deathlabs/kaiju-tui/internal/tui"
	"github.com/spf13/cobra"
)

// run initializes and runs the TUI.
func run(cmd *cobra.Command, args []string) error {
	var (
		authConfig   auth.KeycloakConfig
		err          error
		screenRouter tui.ScreenRouter
		program      *tea.Program
	)

	authConfig = auth.KeycloakConfig{
		BaseURL:  "https://sso.uds.dev",
		Realm:    "uds",
		ClientID: "kaiju-tui",
	}

	screenRouter = tui.NewScreenRouter("")

	program = tea.NewProgram(
		tui.NewTUI(authConfig, screenRouter),
	)

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
