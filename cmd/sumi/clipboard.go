package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/theme"
)

func clipboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clipboard",
		Aliases: []string{"clip"},
		Short:   "Clipboard history picker",
		Long:    "Browse and paste from clipboard history (cliphist). Use 'wipe' to clear all entries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := exec.LookPath("cliphist"); err != nil {
				return fmt.Errorf("cliphist not found — install with: pacman -S cliphist")
			}
			out, err := exec.Command("cliphist", "list").Output()
			if err != nil {
				return fmt.Errorf("cliphist list: %w", err)
			}

			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
				fmt.Println(theme.Warn("clipboard history is empty"))
				return nil
			}

			// Limit to 50 most recent for TUI performance
			if len(lines) > 50 {
				lines = lines[:50]
			}

			var opts []huh.Option[string]
			for _, line := range lines {
				display := line
				if len(display) > 80 {
					display = display[:80] + "..."
				}
				opts = append(opts, huh.NewOption(display, line))
			}

			var selected string
			err = huh.NewSelect[string]().
				Title("sumi :: clipboard").
				Options(opts...).
				Value(&selected).
				Run()
			if err != nil {
				return silenceQuit(err)
			}

			// Decode and copy
			decode := exec.Command("cliphist", "decode")
			decode.Stdin = strings.NewReader(selected)
			decoded, err := decode.Output()
			if err != nil {
				return fmt.Errorf("cliphist decode: %w", err)
			}

			copy := exec.Command("wl-copy")
			copy.Stdin = strings.NewReader(string(decoded))
			return copy.Run()
		},
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "wipe",
			Short: "Clear clipboard history",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := requireTool("cliphist", "pacman -S cliphist"); err != nil {
					return err
				}
				if err := exec.Command("cliphist", "wipe").Run(); err != nil {
					return fmt.Errorf("cliphist wipe: %w", err)
				}
				exec.Command("notify-send", "-t", "2000", "[ clipboard ]", "history cleared").Run() //nolint:errcheck
				fmt.Println(theme.Ok("clipboard history cleared"))
				return nil
			},
		},
	)

	return cmd
}
