package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "kaiju-tui",
		Short:   "A Terminal User Interface (TUI) for Kaiju.",
		Version: version,
	}
	version string
)

func Execute() {
	var err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
