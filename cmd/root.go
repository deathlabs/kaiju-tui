package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
