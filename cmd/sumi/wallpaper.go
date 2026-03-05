package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/wallpaper"
)

func wallpaperCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wallpaper",
		Aliases: []string{"wp"},
		Short:   "Manage wallpaper and color pipeline",
		Long: "Set wallpapers and regenerate the system color scheme via wallust.\nColors propagate to foot, waybar, fuzzel, dunst, hyprlock, cava, and tmux.",
		Example: `  sumi wallpaper apply ~/Pictures/Wallpapers/mountain.jpg
  sumi wallpaper random
  sumi wallpaper pick`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Wallpaper").
						Options(
							huh.NewOption("Pick from wallpapers", "pick"),
							huh.NewOption("Random wallpaper", "random"),
							huh.NewOption("Show current", "current"),
						).
						Value(&choice),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			home := os.Getenv("HOME")
			switch choice {
			case "pick":
				images := wallpaper.ListWallpapers(home)
				if len(images) == 0 {
					return fmt.Errorf("no wallpapers found — add images to ~/Pictures/Wallpapers/")
				}
				current := wallpaper.CurrentWallpaper(home)
				var options []huh.Option[string]
				for _, img := range images {
					name := filepath.Base(img)
					label := name
					if img == current {
						label = name + " (current)"
					}
					options = append(options, huh.NewOption(label, img))
				}
				var selected string
				huh.NewSelect[string]().Title("Pick wallpaper").Options(options...).Value(&selected).Run() //nolint:errcheck
				if selected == "" {
					return nil
				}
				return wallpaper.Apply(selected, wallpaper.ApplyOpts{Home: home})
			case "random":
				return wallpaper.Random(home)
			case "current":
				wp := wallpaper.CurrentWallpaper(home)
				if wp == "" {
					fmt.Println("none")
				} else {
					fmt.Println(wp)
				}
			}
			return nil
		},
	}

	var noReload, noWallust bool

	applyCmd := &cobra.Command{
		Use:   "apply <path>",
		Short: "Set wallpaper and regenerate colors",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wallpaper.Apply(args[0], wallpaper.ApplyOpts{
				NoReload:  noReload,
				NoWallust: noWallust,
			})
		},
	}
	applyCmd.Flags().BoolVar(&noReload, "no-reload", false, "Skip waybar/dunst restart")
	applyCmd.Flags().BoolVar(&noWallust, "no-wallust", false, "Skip color extraction")

	randomCmd := &cobra.Command{
		Use:   "random",
		Short: "Apply a random wallpaper",
		RunE: func(cmd *cobra.Command, args []string) error {
			return wallpaper.Random(os.Getenv("HOME"))
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Load wallpaper on login (autostart)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return wallpaper.Init(os.Getenv("HOME"))
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Print current wallpaper path",
		Run: func(cmd *cobra.Command, args []string) {
			wp := wallpaper.CurrentWallpaper(os.Getenv("HOME"))
			if wp == "" {
				fmt.Println("none")
			} else {
				fmt.Println(wp)
			}
		},
	}

	pickCmd := &cobra.Command{
		Use:   "pick",
		Short: "Interactive wallpaper picker",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := os.Getenv("HOME")
			images := wallpaper.ListWallpapers(home)
			if len(images) == 0 {
				return fmt.Errorf("no wallpapers found — add images to ~/Pictures/Wallpapers/")
			}

			current := wallpaper.CurrentWallpaper(home)

			var options []huh.Option[string]
			for _, img := range images {
				name := filepath.Base(img)
				label := name
				if img == current {
					label = name + " (current)"
				}
				options = append(options, huh.NewOption(label, img))
			}

			var selected string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Pick wallpaper").
						Description(fmt.Sprintf("%d wallpapers · colors auto-adapt", len(images))).
						Options(options...).
						Value(&selected),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			}
			if selected == "" {
				return nil
			}
			return wallpaper.Apply(selected, wallpaper.ApplyOpts{Home: home})
		},
	}

	cmd.AddCommand(applyCmd, randomCmd, initCmd, currentCmd, pickCmd)
	return cmd
}
