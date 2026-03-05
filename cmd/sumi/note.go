package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"sumi/internal/dev"
	"sumi/internal/theme"
)

func noteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "Quick notes in markdown",
		Long: "Open today's daily note, create new notes, or search with ripgrep.\nNotes are stored in ~/Documents/notes/ as markdown files.",
		Example: `  sumi note              # open today's note
  sumi note new design   # create a new note
  sumi note search TODO  # search all notes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default: open today's daily note
			return dev.OpenToday()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "new [title]",
			Short: "Create a new note",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				title := args[0]
				return dev.NewNote(title)
			},
		},
		&cobra.Command{
			Use:   "search <query>",
			Short: "Search notes with ripgrep",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return dev.SearchNotes(args[0])
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List recent notes",
			RunE: func(cmd *cobra.Command, args []string) error {
				files, err := dev.ListRecent(20)
				if err != nil {
					return err
				}
				if len(files) == 0 {
					fmt.Println(theme.Warn("no notes found"))
					return nil
				}
				for _, f := range files {
					fmt.Println(theme.Step(f))
				}
				return nil
			},
		},
	)

	return cmd
}
