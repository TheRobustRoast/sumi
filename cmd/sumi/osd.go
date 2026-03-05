package main

import (
	"github.com/spf13/cobra"

	"sumi/internal/osd"
)

func osdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "osd",
		Short: "On-screen display notifications (volume, brightness, mic)",
		Long:  "Adjust volume, brightness, or mic and show a dunst notification with the current level.\nBound to media keys via Hyprland keybinds.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "volume-up",
			Short: "Raise volume and show OSD",
			RunE:  func(cmd *cobra.Command, args []string) error { return osd.ShowVolume("up") },
		},
		&cobra.Command{
			Use:   "volume-down",
			Short: "Lower volume and show OSD",
			RunE:  func(cmd *cobra.Command, args []string) error { return osd.ShowVolume("down") },
		},
		&cobra.Command{
			Use:   "volume-mute",
			Short: "Toggle mute and show OSD",
			RunE:  func(cmd *cobra.Command, args []string) error { return osd.ShowVolume("mute") },
		},
		&cobra.Command{
			Use:   "mic-mute",
			Short: "Toggle mic mute and show OSD",
			RunE:  func(cmd *cobra.Command, args []string) error { return osd.ShowMic() },
		},
		&cobra.Command{
			Use:   "brightness-up",
			Short: "Raise brightness and show OSD",
			RunE:  func(cmd *cobra.Command, args []string) error { return osd.ShowBrightness("up") },
		},
		&cobra.Command{
			Use:   "brightness-down",
			Short: "Lower brightness and show OSD",
			RunE:  func(cmd *cobra.Command, args []string) error { return osd.ShowBrightness("down") },
		},
	)

	return cmd
}
