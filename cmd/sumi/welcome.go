package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"sumi/internal/ui"
)

func welcomeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "welcome",
		Short: "First-run welcome wizard",
		Long:  "A multi-page TUI showing keybinds, wallpaper picker, and hardware profile summary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := ui.NewWelcomeWizard()
			p := tea.NewProgram(m, tea.WithAltScreen())
			_, err := p.Run()
			return silenceQuit(err)
		},
	}
}
