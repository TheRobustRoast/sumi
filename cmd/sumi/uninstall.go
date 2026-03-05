package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"sumi/internal/model"
	"sumi/internal/steps"
	"sumi/internal/theme"
	"sumi/internal/ui"
)

func uninstallCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the sumi rice and restore defaults",
		Long:  "Removes all sumi configs, services, and cache. Packages are NOT uninstalled.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Println(theme.Warn("This will remove ALL sumi configuration files."))
				fmt.Println(theme.Step("Your wallpapers, screenshots, and recordings will be kept."))
				fmt.Print("\nType 'yes' to continue: ")
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println(theme.Step("aborted"))
					return nil
				}
			}

			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}

			m := ui.NewStepRunner(ui.StepRunnerConfig{
				Title:    "uninstall",
				Subtitle: "removing sumi configs",
				Steps:    steps.UninstallSteps(),
				Ctx: model.InstallCtx{
					SumiDir: sumiDir,
					Home:    os.Getenv("HOME"),
					User:    os.Getenv("USER"),
				},
				DoneMsg: "sumi uninstalled",
			})

			p := tea.NewProgram(m, tea.WithAltScreen())
			_, err = p.Run()
			return silenceQuit(err)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
