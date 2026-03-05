package main

import (
	"github.com/spf13/cobra"

	"sumi/internal/ui"
)

func themeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "theme",
		Short: "Manage wallpaper and color theme",
		Long:  "Interactive form for picking wallpapers, toggling dark/light mode, and managing the color pipeline.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}
			return ui.RunThemePicker(sumiDir)
		},
	}
}
