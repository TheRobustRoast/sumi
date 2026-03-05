package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sumi/internal/session"
)

func monitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Detect and configure external monitors",
		Long:  "Detect connected monitors via hyprctl and configure positioning and workspace assignment.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "detect",
			Short: "Detect connected monitors and configure",
			RunE: func(cmd *cobra.Command, args []string) error {
				return session.ConfigureMonitors(os.Getenv("HOME"))
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show current monitor configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				status, err := session.MonitorStatus()
				if err != nil {
					return err
				}
				fmt.Println(status)
				return nil
			},
		},
	)

	return cmd
}
