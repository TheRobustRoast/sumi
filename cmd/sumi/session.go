package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/session"
	"sumi/internal/theme"
)

func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Save and restore window layouts",
		Long: "Save the current Hyprland window layout (positions, workspaces, floating state) and restore it later.",
		Example: `  sumi session save work
  sumi session restore work
  sumi session list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Session manager").
						Options(
							huh.NewOption("Save current layout", "save"),
							huh.NewOption("Restore saved layout", "restore"),
							huh.NewOption("List saved sessions", "list"),
						).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			home := os.Getenv("HOME")
			switch choice {
			case "save":
				if err := session.Save(home, ""); err != nil {
					return err
				}
				fmt.Println(theme.Ok("session saved"))
			case "restore":
				names, err := session.List(home)
				if err != nil {
					return err
				}
				if len(names) == 0 {
					fmt.Println(theme.Warn("no saved sessions"))
					return nil
				}
				var opts []huh.Option[string]
				for _, n := range names {
					opts = append(opts, huh.NewOption(n, n))
				}
				var selected string
				huh.NewSelect[string]().Title("Restore session").Options(opts...).Value(&selected).Run() //nolint:errcheck
				if selected == "" {
					return nil
				}
				return session.Restore(home, selected)
			case "list":
				names, err := session.List(home)
				if err != nil {
					return err
				}
				if len(names) == 0 {
					fmt.Println(theme.Warn("no saved sessions"))
					return nil
				}
				for _, n := range names {
					fmt.Println(theme.Step(n))
				}
			}
			return nil
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "save [name]",
			Short: "Save current window layout",
			RunE: func(cmd *cobra.Command, args []string) error {
				name := ""
				if len(args) > 0 {
					name = args[0]
				}
				return session.Save(os.Getenv("HOME"), name)
			},
		},
		&cobra.Command{
			Use:   "restore <name>",
			Short: "Restore saved window layout",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return session.Restore(os.Getenv("HOME"), args[0])
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List saved sessions",
			RunE: func(cmd *cobra.Command, args []string) error {
				names, err := session.List(os.Getenv("HOME"))
				if err != nil {
					return err
				}
				if len(names) == 0 {
					fmt.Println(theme.Warn("no saved sessions"))
					return nil
				}
				for _, n := range names {
					fmt.Println(theme.Step(n))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "delete <name>",
			Short: "Delete a saved session",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return session.Delete(os.Getenv("HOME"), args[0])
			},
		},
	)

	return cmd
}
