package main

import (
	"github.com/spf13/cobra"

	"sumi/internal/ui"
)

func configCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Edit sumi settings interactively",
		Long:  "Opens an interactive form to edit ~/.config/sumi/config.toml. All values have defaults — the file is created on first save.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunConfigEditor()
		},
	}
}
