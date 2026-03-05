package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/dev"
	"sumi/internal/theme"
)

func projectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [path]",
		Short: "Open a project in tmux (interactive picker if no path)",
		Long: "Open a project directory in a tmux session with editor, shell, and run windows.\nWithout arguments, shows an interactive picker using zoxide frecency.",
		Example: `  sumi project              # interactive picker
  sumi project ~/Projects/sumi`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return dev.OpenInTmux(args[0])
			}

			projects, err := dev.ListProjects()
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				fmt.Println(theme.Warn("no projects found"))
				fmt.Println(theme.Step("projects detected via .git, go.mod, Cargo.toml, package.json"))
				fmt.Println(theme.Step("cd into project directories to build zoxide's database"))
				return nil
			}

			var opts []huh.Option[string]
			for _, p := range projects {
				opts = append(opts, huh.NewOption(p, p))
			}

			var selected string
			err = huh.NewSelect[string]().
				Title("sumi :: project launcher").
				Options(opts...).
				Value(&selected).
				Run()
			if err != nil {
				return silenceQuit(err)
			}

			return dev.OpenInTmux(selected)
		},
	}
	return cmd
}
