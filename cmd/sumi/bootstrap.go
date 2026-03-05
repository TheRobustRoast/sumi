package main

import (
	"github.com/spf13/cobra"

	"sumi/internal/ui"
)

func bootstrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Install Arch Linux from the live ISO",
		Long:  "Partitions, encrypts, installs, and configures Arch Linux, then stages the sumi rice for first boot.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}
			return ui.RunBootstrap(sumiDir)
		},
	}
}
