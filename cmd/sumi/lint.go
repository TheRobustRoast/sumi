package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/lint"
	"sumi/internal/theme"
)

func lintCmd() *cobra.Command {
	var fixFlag bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate all managed configs",
		Long:  "Check Hyprland, foot, waybar, and other configs for deprecated options, missing files, and syntax errors.\nExit code: 0 = clean, 1 = warnings, 2 = errors.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}
			home := os.Getenv("HOME")

			fmt.Println(theme.Step("validating configs..."))
			issues := lint.RunAll(sumiDir, home)

			if len(issues) == 0 {
				fmt.Println(theme.Ok("all configs valid"))
				return nil
			}

			code := lint.PrintIssues(issues)

			// Count fixable issues
			fixable := 0
			for _, i := range issues {
				if i.Fix != nil {
					fixable++
				}
			}

			if fixFlag {
				fixed := lint.FixIssues(issues)
				if fixed > 0 {
					fmt.Println(theme.Ok(fmt.Sprintf("fixed %d issue(s)", fixed)))
				}
				// Re-check after fixes
				issues = lint.RunAll(sumiDir, home)
				if len(issues) == 0 {
					fmt.Println(theme.Ok("all configs valid after fixes"))
					return nil
				}
			} else if fixable > 0 {
				// Interactive: offer to fix
				var doFix bool
				huh.NewConfirm().
					Title(fmt.Sprintf("Auto-fix %d fixable issue(s)?", fixable)).
					Value(&doFix).
					Run() //nolint:errcheck
				if doFix {
					fixed := lint.FixIssues(issues)
					fmt.Println(theme.Ok(fmt.Sprintf("fixed %d issue(s)", fixed)))
					issues = lint.RunAll(sumiDir, home)
					if len(issues) == 0 {
						fmt.Println(theme.Ok("all configs valid"))
						return nil
					}
				}
			}

			if code == 2 {
				fmt.Println(theme.Fail(fmt.Sprintf("%d issue(s) remain", len(issues))))
			} else {
				fmt.Println(theme.Warn(fmt.Sprintf("%d warning(s) remain", len(issues))))
			}
			os.Exit(code)
			return nil //nolint:govet // unreachable but required by cobra RunE signature
		},
	}
	cmd.Flags().BoolVar(&fixFlag, "fix", false, "Auto-fix known issues")
	return cmd
}
