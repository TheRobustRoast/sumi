package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/mode"
	"sumi/internal/theme"
	"sumi/internal/waybar"
)

func modeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mode",
		Short: "Toggle gaming, focus, and other modes",
		Long: "Gaming mode disables blur/animations and sets performance profile.\nFocus mode pauses notifications and hides waybar.",
		Example: `  sumi mode gaming   # toggle gaming mode
  sumi mode focus    # toggle focus/DND
  sumi mode reset    # disable all modes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Mode toggle").
						Options(
							huh.NewOption("Gaming mode", "gaming"),
							huh.NewOption("Focus/DND mode", "focus"),
							huh.NewOption("Reset all modes", "reset"),
						).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			home := os.Getenv("HOME")
			switch choice {
			case "gaming":
				on, err := mode.ToggleGaming(home)
				if err != nil {
					return err
				}
				if on {
					fmt.Println(theme.Ok("gaming mode ON"))
				} else {
					fmt.Println(theme.Step("gaming mode OFF"))
				}
			case "focus":
				on, err := mode.ToggleFocus(home)
				if err != nil {
					return err
				}
				if on {
					fmt.Println(theme.Ok("focus mode ON"))
				} else {
					fmt.Println(theme.Step("focus mode OFF"))
				}
			case "reset":
				if mode.IsGaming(home) {
					mode.ToggleGaming(home) //nolint:errcheck
				}
				if mode.IsFocus(home) {
					mode.ToggleFocus(home) //nolint:errcheck
				}
				fmt.Println(theme.Ok("all modes disabled"))
			}
			return nil
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "gaming",
			Short: "Toggle gaming mode (max perf, no visual overhead)",
			RunE: func(cmd *cobra.Command, args []string) error {
				on, err := mode.ToggleGaming(os.Getenv("HOME"))
				if err != nil {
					return err
				}
				if on {
					fmt.Println(theme.Ok("gaming mode ON"))
				} else {
					fmt.Println(theme.Step("gaming mode OFF"))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "focus",
			Short: "Toggle focus/DND mode",
			RunE: func(cmd *cobra.Command, args []string) error {
				on, err := mode.ToggleFocus(os.Getenv("HOME"))
				if err != nil {
					return err
				}
				if on {
					fmt.Println(theme.Ok("focus mode ON"))
				} else {
					fmt.Println(theme.Step("focus mode OFF"))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "reset",
			Short: "Disable all modes",
			RunE: func(cmd *cobra.Command, args []string) error {
				home := os.Getenv("HOME")
				if mode.IsGaming(home) {
					mode.ToggleGaming(home) //nolint:errcheck
				}
				if mode.IsFocus(home) {
					mode.ToggleFocus(home) //nolint:errcheck
				}
				fmt.Println(theme.Ok("all modes disabled"))
				return nil
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show current mode state as JSON (for waybar)",
			Run: func(cmd *cobra.Command, args []string) {
				home := os.Getenv("HOME")
				gaming := mode.IsGaming(home)

				// Gaming mode module
				if gaming {
					waybar.PrintJSON(waybar.Module{Text: "[GAME]", Class: "active"})
				} else {
					waybar.PrintJSON(waybar.Module{})
				}
				// Focus status is exposed via the focus-status subcommand
			},
		},
		&cobra.Command{
			Use:   "gaming-status",
			Short: "Gaming mode status (waybar JSON output)",
			Run: func(cmd *cobra.Command, args []string) {
				if mode.IsGaming(os.Getenv("HOME")) {
					waybar.PrintJSON(waybar.Module{Text: "[GAME]", Class: "active"})
				} else {
					waybar.PrintJSON(waybar.Module{})
				}
			},
		},
		&cobra.Command{
			Use:   "focus-status",
			Short: "Focus mode status (waybar JSON output)",
			Run: func(cmd *cobra.Command, args []string) {
				if mode.IsFocus(os.Getenv("HOME")) {
					waybar.PrintJSON(waybar.Module{Text: "[DND]", Class: "active"})
				} else {
					waybar.PrintJSON(waybar.Module{})
				}
			},
		},
	)

	return cmd
}
