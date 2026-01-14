package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/client"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
	"github.com/helmedeiros/fastretro-cli/internal/tui"
	"github.com/spf13/cobra"
)

var serverURL string

func baseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fastretro")
}

var rootCmd = &cobra.Command{
	Use:   "fastretro",
	Short: "Terminal client for fastRetro retrospective sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := storage.NewJSONRegistryRepo(baseDir())
		entries, err := reg.List()
		if err != nil {
			return err
		}

		// Resolve selected team
		selectedID, err := reg.SelectedTeamID()
		if err != nil {
			return err
		}
		var entry domain.TeamEntry

		if selectedID != "" {
			for _, e := range entries {
				if e.ID == selectedID {
					entry = e
					break
				}
			}
		}

		// If no team selected but teams exist, pick first
		if entry.ID == "" && len(entries) > 0 {
			entry = entries[0]
			if err := reg.SetSelectedTeamID(entry.ID); err != nil {
				return err
			}
		}

		// Create shell — if no team, it starts in team selector mode
		shell := tui.NewShellModel(reg, entry, serverURL)
		if entry.ID == "" {
			shell.StartInTeamSelect()
		}

		p := tea.NewProgram(shell, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
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

		model := tui.NewModel(c)

		// Restore persisted identity for this room
		reg := storage.NewJSONRegistryRepo(baseDir())
		if saved := reg.LoadIdentity(c.RoomCode); saved != "" {
			model.SetParticipantID(saved)
			if err := c.ClaimIdentity(saved); err != nil {
				return fmt.Errorf("failed to claim identity: %w", err)
			}
		}

		p := tea.NewProgram(model, tea.WithAltScreen())
		result, err := p.Run()

		// Persist identity for next reconnect
		if m, ok := result.(tui.Model); ok && m.ParticipantID() != "" {
			_ = reg.SaveIdentity(c.RoomCode, m.ParticipantID())
		}

		return err
	},
}

// --- team commands ---

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage teams",
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := storage.NewJSONRegistryRepo(baseDir())
		entries, err := reg.List()
		if err != nil {
			return err
		}
		selectedID, _ := reg.SelectedTeamID()
		if len(entries) == 0 {
			fmt.Println("No teams. Create one with: fastretro team create <name>")
			return nil
		}
		for _, e := range entries {
			marker := "  "
			if e.ID == selectedID {
				marker = "* "
			}
			fmt.Printf("%s%s  (created %s)\n", marker, e.Name, e.CreatedAt)
		}
		return nil
	},
}

var teamCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := storage.NewJSONRegistryRepo(baseDir())
		entries, err := reg.List()
		if err != nil {
			return err
		}
		id := fmt.Sprintf("t-%d", time.Now().UnixMilli())
		entries, err = domain.AddTeamEntry(entries, id, args[0], time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
		if err := reg.Save(entries); err != nil {
			return err
		}
		// Auto-select if first team
		if len(entries) == 1 {
			reg.SetSelectedTeamID(id)
		}
		fmt.Printf("Created team %q\n", args[0])
		return nil
	},
}

var teamSelectCmd = &cobra.Command{
	Use:   "select [name]",
	Short: "Select the active team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := storage.NewJSONRegistryRepo(baseDir())
		entries, err := reg.List()
		if err != nil {
			return err
		}
		entry, ok := domain.FindTeamEntryByName(entries, args[0])
		if !ok {
			return fmt.Errorf("team %q not found", args[0])
		}
		if err := reg.SetSelectedTeamID(entry.ID); err != nil {
			return err
		}
		fmt.Printf("Selected team %q\n", entry.Name)
		return nil
	},
}

var teamDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := storage.NewJSONRegistryRepo(baseDir())
		entries, err := reg.List()
		if err != nil {
			return err
		}
		entry, ok := domain.FindTeamEntryByName(entries, args[0])
		if !ok {
			return fmt.Errorf("team %q not found", args[0])
		}
		entries = domain.RemoveTeamEntry(entries, entry.ID)
		if err := reg.Save(entries); err != nil {
			return err
		}
		// Clear selection if deleted team was selected
		selectedID, _ := reg.SelectedTeamID()
		if selectedID == entry.ID {
			reg.SetSelectedTeamID("")
		}
		// Remove team data directory
		teamDir := reg.TeamDir(entry.ID)
		os.RemoveAll(teamDir)
		fmt.Printf("Deleted team %q\n", entry.Name)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:5173", "Server URL")
	teamCmd.AddCommand(teamListCmd, teamCreateCmd, teamSelectCmd, teamDeleteCmd)
	rootCmd.AddCommand(joinCmd, teamCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
