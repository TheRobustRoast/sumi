package main

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"sumi/internal/checkpoint"
	"sumi/internal/theme"
	"sumi/internal/ui"
)

func installCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the sumi rice",
		Long:  "Symlink configs, install packages, enable services, and configure the desktop.\nAutomatically resumes from the last completed step on re-run.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}

			if force {
				// Clear any existing checkpoint so all steps run fresh
				checkpoint.Clear()
			}

			m := ui.New(sumiDir)
			p := tea.NewProgram(m, tea.WithAltScreen())
			final, err := p.Run()
			if err != nil {
				return silenceQuit(err)
			}

			if installer, ok := final.(ui.Installer); ok && installer.RebootRequested {
				fmt.Println(theme.Ok("rebooting..."))
				return exec.Command("sudo", "reboot").Run()
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force all steps to run (ignore checkpoint)")
	return cmd
}
