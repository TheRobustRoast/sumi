package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/sync"
	"sumi/internal/theme"
)

func syncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Dotfile backup and restore",
		Long: "Create named snapshots of your dotfiles, compare them with diffs, and restore previous states.\nSnapshots are stored in ~/.local/share/sumi/snapshots/.",
		Example: `  sumi sync snapshot before-update
  sumi sync diff before-update
  sumi sync restore before-update
  sumi sync export`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Interactive sync menu when no subcommand given
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Dotfile sync").
						Options(
							huh.NewOption("Create snapshot", "snapshot"),
							huh.NewOption("Restore from snapshot", "restore"),
							huh.NewOption("Diff against snapshot", "diff"),
							huh.NewOption("List snapshots", "list"),
							huh.NewOption("Export as tar.gz", "export"),
						).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			// Re-dispatch to the subcommand
			switch choice {
			case "snapshot":
				return sync.Snapshot(os.Getenv("HOME"), "")
			case "list":
				snaps, _ := sync.List(os.Getenv("HOME"))
				if len(snaps) == 0 {
					fmt.Println(theme.Warn("no snapshots"))
					return nil
				}
				for _, s := range snaps {
					fmt.Println(theme.Step(fmt.Sprintf("%s  (%s)", s.Name, s.Time.Format("2006-01-02 15:04"))))
				}
				return nil
			case "restore":
				home := os.Getenv("HOME")
				snaps, _ := sync.List(home)
				if len(snaps) == 0 {
					fmt.Println(theme.Warn("no snapshots — run 'sumi sync snapshot' first"))
					return nil
				}
				var opts []huh.Option[string]
				for _, s := range snaps {
					opts = append(opts, huh.NewOption(
						fmt.Sprintf("%s (%s)", s.Name, s.Time.Format("2006-01-02 15:04")), s.Name))
				}
				var selected string
				huh.NewSelect[string]().Title("Restore from").Options(opts...).Value(&selected).Run() //nolint:errcheck
				if selected == "" {
					return nil
				}
				return sync.Restore(home, selected)
			case "diff":
				home := os.Getenv("HOME")
				snaps, _ := sync.List(home)
				if len(snaps) == 0 {
					fmt.Println(theme.Warn("no snapshots"))
					return nil
				}
				entries, err := sync.Diff(home, snaps[len(snaps)-1].Name)
				if err != nil {
					return err
				}
				if len(entries) == 0 {
					fmt.Println(theme.Ok("no changes"))
					return nil
				}
				for _, e := range entries {
					switch e.Status {
					case "modified":
						fmt.Println(theme.WarnStyle.Render("  M  " + e.Path))
					case "deleted":
						fmt.Println(theme.FailStyle.Render("  D  " + e.Path))
					case "new":
						fmt.Println(theme.OkStyle.Render("  A  " + e.Path))
					}
				}
				return nil
			case "export":
				home := os.Getenv("HOME")
				out := filepath.Join(home, "sumi-config-"+time.Now().Format("20060102")+".tar.gz")
				if err := sync.Export(home, out); err != nil {
					return err
				}
				fmt.Println(theme.Ok("exported: " + out))
				return nil
			}
			return nil
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "snapshot [name]",
			Short: "Backup current dotfiles to a snapshot",
			RunE: func(cmd *cobra.Command, args []string) error {
				name := ""
				if len(args) > 0 {
					name = args[0]
				}
				return sync.Snapshot(os.Getenv("HOME"), name)
			},
		},
		&cobra.Command{
			Use:   "restore [name]",
			Short: "Restore from a snapshot",
			RunE: func(cmd *cobra.Command, args []string) error {
				home := os.Getenv("HOME")
				if len(args) > 0 {
					return sync.Restore(home, args[0])
				}
				// Interactive selection
				snaps, _ := sync.List(home)
				if len(snaps) == 0 {
					fmt.Println(theme.Warn("no snapshots — run 'sumi sync snapshot' first"))
					return nil
				}
				var opts []huh.Option[string]
				for _, s := range snaps {
					opts = append(opts, huh.NewOption(
						fmt.Sprintf("%s (%s)", s.Name, s.Time.Format("2006-01-02 15:04")), s.Name))
				}
				var selected string
				err := huh.NewSelect[string]().
					Title("Restore from snapshot").
					Options(opts...).
					Value(&selected).
					Run()
				if err != nil {
					return silenceQuit(err)
				}

				var confirm bool
				huh.NewConfirm().
					Title("Overwrite current configs with snapshot?").
					Value(&confirm).
					Run() //nolint:errcheck
				if !confirm {
					return nil
				}
				return sync.Restore(home, selected)
			},
		},
		&cobra.Command{
			Use:   "diff [name]",
			Short: "Show differences between snapshot and current config",
			RunE: func(cmd *cobra.Command, args []string) error {
				home := os.Getenv("HOME")
				name := ""
				if len(args) > 0 {
					name = args[0]
				} else {
					snaps, _ := sync.List(home)
					if len(snaps) == 0 {
						fmt.Println(theme.Warn("no snapshots found"))
						return nil
					}
					name = snaps[len(snaps)-1].Name
				}
				entries, err := sync.Diff(home, name)
				if err != nil {
					return err
				}
				if len(entries) == 0 {
					fmt.Println(theme.Ok("no changes since snapshot " + name))
					return nil
				}
				for _, e := range entries {
					switch e.Status {
					case "modified":
						fmt.Println(theme.WarnStyle.Render("  M  " + e.Path))
					case "deleted":
						fmt.Println(theme.FailStyle.Render("  D  " + e.Path))
					case "new":
						fmt.Println(theme.OkStyle.Render("  A  " + e.Path))
					}
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List available snapshots",
			RunE: func(cmd *cobra.Command, args []string) error {
				snaps, _ := sync.List(os.Getenv("HOME"))
				if len(snaps) == 0 {
					fmt.Println(theme.Warn("no snapshots"))
					return nil
				}
				for _, s := range snaps {
					fmt.Println(theme.Step(fmt.Sprintf("%s  (%s)", s.Name, s.Time.Format("2006-01-02 15:04"))))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "export",
			Short: "Export current config as tar.gz",
			RunE: func(cmd *cobra.Command, args []string) error {
				home := os.Getenv("HOME")
				out := filepath.Join(home, "sumi-config-"+time.Now().Format("20060102")+".tar.gz")
				if err := sync.Export(home, out); err != nil {
					return err
				}
				fmt.Println(theme.Ok("exported: " + out))
				return nil
			},
		},
	)

	return cmd
}
