package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/theme"
)

func notifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Notification center (dunst history)",
		Long:  "Browse, dismiss, and manage dunst notification history.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "Browse notification history",
			RunE:  runNotifyList,
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Clear all notification history",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := requireTool("dunstctl", "pacman -S dunst"); err != nil {
					return err
				}
				if err := exec.Command("dunstctl", "history-clear").Run(); err != nil {
					return fmt.Errorf("dunstctl history-clear: %w", err)
				}
				fmt.Println(theme.Ok("notification history cleared"))
				return nil
			},
		},
		&cobra.Command{
			Use:   "dismiss",
			Short: "Dismiss all visible notifications",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := requireTool("dunstctl", "pacman -S dunst"); err != nil {
					return err
				}
				if err := exec.Command("dunstctl", "close-all").Run(); err != nil {
					return fmt.Errorf("dunstctl close-all: %w", err)
				}
				fmt.Println(theme.Ok("all notifications dismissed"))
				return nil
			},
		},
		&cobra.Command{
			Use:   "pause",
			Short: "Toggle do-not-disturb",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := requireTool("dunstctl", "pacman -S dunst"); err != nil {
					return err
				}
				if err := exec.Command("dunstctl", "set-paused", "toggle").Run(); err != nil {
					return fmt.Errorf("dunstctl set-paused: %w", err)
				}
				return nil
			},
		},
	)

	// Bare `sumi notify` runs the list
	cmd.RunE = runNotifyList
	return cmd
}

func runNotifyList(cmd *cobra.Command, args []string) error {
	if err := requireTool("dunstctl", "pacman -S dunst"); err != nil {
		return err
	}
	out, err := exec.Command("dunstctl", "history").Output()
	if err != nil {
		return fmt.Errorf("dunstctl history: %w", err)
	}

	var history struct {
		Data [][]struct {
			AppName struct{ Data string } `json:"appname"`
			Summary struct{ Data string } `json:"summary"`
			Body    struct{ Data string } `json:"body"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &history); err != nil {
		return fmt.Errorf("parse history: %w", err)
	}

	if len(history.Data) == 0 || len(history.Data[0]) == 0 {
		fmt.Println(theme.Warn("no notification history"))
		return nil
	}

	entries := history.Data[0]
	var options []huh.Option[string]
	for i, e := range entries {
		label := e.AppName.Data
		if label == "" {
			label = "notification"
		}
		label += " │ " + e.Summary.Data
		if e.Body.Data != "" {
			label += " │ " + e.Body.Data
		}
		// Truncate long labels
		if len(label) > 80 {
			label = label[:77] + "..."
		}
		options = append(options, huh.NewOption(label, fmt.Sprintf("%d", i)))
	}

	// Add action options
	options = append(options,
		huh.NewOption("── dismiss all ──", "dismiss"),
		huh.NewOption("── clear history ──", "clear"),
	)

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Notifications (%d)", len(entries))).
				Options(options...).
				Value(&choice),
		),
	)
	if err := form.Run(); err != nil {
		return nil
	}

	switch choice {
	case "dismiss":
		exec.Command("dunstctl", "close-all").Run() //nolint:errcheck
		fmt.Println(theme.Ok("all dismissed"))
	case "clear":
		exec.Command("dunstctl", "history-clear").Run() //nolint:errcheck
		fmt.Println(theme.Ok("history cleared"))
	}

	return nil
}
