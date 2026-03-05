package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"sumi/internal/model"
	"sumi/internal/steps"
	"sumi/internal/ui"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and configuration",
		Long:  "Runs diagnostic checks on packages, services, configs, and the wallpaper pipeline.\nReports issues with actionable fix suggestions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}

			m := ui.NewStepRunner(ui.StepRunnerConfig{
				Title:    "doctor",
				Subtitle: "health checks",
				Steps:    steps.DoctorChecks(),
				Ctx: model.InstallCtx{
					SumiDir: sumiDir,
					Home:    os.Getenv("HOME"),
					User:    os.Getenv("USER"),
				},
				DoneMsg: "All checks complete",
			})

			p := tea.NewProgram(m, tea.WithAltScreen())
			_, err = p.Run()
			return silenceQuit(err)
		},
	}
}
