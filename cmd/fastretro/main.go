package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/client"
	"github.com/helmedeiros/fastretro-cli/internal/tui"
	"github.com/spf13/cobra"
)

var serverURL string

var rootCmd = &cobra.Command{
	Use:   "fastretro",
	Short: "Terminal client for fastRetro retrospective sessions",
}

var joinCmd = &cobra.Command{
	Use:   "join [room-code-or-url]",
	Short: "Join a retrospective session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.Connect(args[0], serverURL)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer c.Close()

		p := tea.NewProgram(tui.NewModel(c), tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() {
	joinCmd.Flags().StringVarP(&serverURL, "server", "s", "http://localhost:5173", "Server URL")
	rootCmd.AddCommand(joinCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
