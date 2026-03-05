package main

import (
	"time"

	"github.com/spf13/cobra"

	"sumi/internal/wrap"
)

func wrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "wrap",
		Short:  "Process wrappers with crash recovery",
		Hidden: true,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "hyprland",
			Short: "Run Hyprland with crash recovery (for greetd)",
			RunE: func(cmd *cobra.Command, args []string) error {
				return wrap.RunWithRecovery(3, 60*time.Second)
			},
		},
	)

	return cmd
}
