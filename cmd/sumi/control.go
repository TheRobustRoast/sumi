package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"sumi/internal/ui"
)

func controlCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "control",
		Short: "Open the control center (quick launcher)",
		Long:  "A fuzzy-filterable list of system tools, launched by SUPER+X.",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := ui.NewControlCenter()
			p := tea.NewProgram(m, tea.WithAltScreen())
			_, err := p.Run()
			return silenceQuit(err)
		},
	}
}
