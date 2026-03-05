package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"sumi/internal/ui"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current sumi state (wallpaper, theme, hardware, services)",
		Long:  "Display a dashboard of current wallpaper, color theme, hardware profile, and service status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}
			info := ui.GatherStatus(sumiDir)
			fmt.Println(ui.RenderStatus(info))
			return nil
		},
	}
}
