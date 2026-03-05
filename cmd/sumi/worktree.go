package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/dev"
	"sumi/internal/theme"
)

func worktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Git worktree manager",
		Long: "List, create, and remove git worktrees. Worktrees are created as sibling directories next to the repo root.",
		Example: `  sumi worktree list
  sumi worktree create feature-branch
  sumi worktree remove ../repo-feature-branch`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List existing worktrees",
			RunE: func(cmd *cobra.Command, args []string) error {
				wts, err := dev.ListWorktrees()
				if err != nil {
					return err
				}
				if len(wts) == 0 {
					fmt.Println(theme.Warn("no worktrees found"))
					return nil
				}
				for _, wt := range wts {
					fmt.Println(theme.Step(wt.Path + "  [" + wt.Branch + "]"))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "create <branch>",
			Short: "Create a new worktree",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := dev.CreateWorktree(args[0]); err != nil {
					return err
				}
				fmt.Println(theme.Ok("worktree created: " + args[0]))
				return nil
			},
		},
		&cobra.Command{
			Use:   "remove [path]",
			Short: "Remove a worktree",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) > 0 {
					return dev.RemoveWorktree(args[0])
				}
				// Interactive selection
				wts, err := dev.ListWorktrees()
				if err != nil {
					return err
				}
				if len(wts) < 2 {
					fmt.Println(theme.Warn("no secondary worktrees to remove"))
					return nil
				}
				var opts []huh.Option[string]
				for _, wt := range wts[1:] { // skip main worktree
					opts = append(opts, huh.NewOption(
						fmt.Sprintf("%s [%s]", wt.Path, wt.Branch), wt.Path))
				}
				var selected string
				err = huh.NewSelect[string]().
					Title("Remove worktree").
					Options(opts...).
					Value(&selected).
					Run()
				if err != nil {
					return silenceQuit(err)
				}
				return dev.RemoveWorktree(selected)
			},
		},
		&cobra.Command{
			Use:   "prune",
			Short: "Clean up stale worktree references",
			RunE: func(cmd *cobra.Command, args []string) error {
				return dev.PruneWorktrees()
			},
		},
	)

	return cmd
}
