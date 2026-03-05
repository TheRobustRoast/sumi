package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"sumi/internal/theme"
)

func updateCmd() *cobra.Command {
	var autoYes bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Pull latest changes, rebuild, and re-install",
		Long:  "Fetches the latest sumi repo, shows a colored diff, and re-runs install steps.\nOnly changed steps are re-applied unless --force is used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}

			// 1. Fetch remote
			fmt.Println(theme.Step("Fetching latest sumi..."))
			fetch := exec.Command("git", "-C", sumiDir, "fetch")
			fetch.Stdout = os.Stdout
			fetch.Stderr = os.Stderr
			if err := fetch.Run(); err != nil {
				return fmt.Errorf("git fetch: %w", err)
			}

			// 2. Show what changed
			diffOut, _ := exec.Command("git", "-C", sumiDir, "diff", "--stat", "HEAD..origin/main").Output()
			diffStat := strings.TrimSpace(string(diffOut))
			if diffStat != "" {
				fmt.Println()
				fmt.Println(theme.Section("Changes"))
				fmt.Println()
				renderColoredDiff(diffStat)
				fmt.Println()
			} else {
				fmt.Println(theme.Ok("Already up to date"))
				return nil
			}

			// 3. Confirm before pulling
			if !autoYes {
				fmt.Print(theme.StepStyle.Render("  Apply update? ") +
					theme.OkStyle.Render("[y]") +
					theme.SubtextStyle.Render("/[N] "))
				var resp string
				fmt.Scanln(&resp)
				if resp != "y" && resp != "Y" {
					fmt.Println(theme.SubtextStyle.Render("  Aborted."))
					return nil
				}
			}

			// 4. Git pull
			fmt.Println(theme.Step("Pulling latest sumi..."))
			pull := exec.Command("git", "-C", sumiDir, "pull", "--ff-only")
			pull.Stdout = os.Stdout
			pull.Stderr = os.Stderr
			if err := pull.Run(); err != nil {
				// Fall back to rebase if ff-only fails
				fmt.Println(theme.Warn("Fast-forward failed, trying rebase..."))
				pull = exec.Command("git", "-C", sumiDir, "pull", "--rebase", "--autostash")
				pull.Stdout = os.Stdout
				pull.Stderr = os.Stderr
				if err := pull.Run(); err != nil {
					return fmt.Errorf("git pull: %w", err)
				}
			}
			fmt.Println(theme.Ok("Pulled latest changes"))

			// 5. Show post-pull summary
			logOut, _ := exec.Command("git", "-C", sumiDir, "log", "--oneline", "HEAD~5..HEAD").Output()
			if logSummary := strings.TrimSpace(string(logOut)); logSummary != "" {
				fmt.Println()
				fmt.Println(theme.Section("Recent commits"))
				fmt.Println()
				for _, line := range strings.Split(logSummary, "\n") {
					parts := strings.SplitN(line, " ", 2)
					if len(parts) == 2 {
						hash := lipgloss.NewStyle().Foreground(theme.Blue).Render(parts[0])
						fmt.Println("    " + hash + " " + parts[1])
					} else {
						fmt.Println("    " + line)
					}
				}
				fmt.Println()
			}

			// 6. Rebuild binary
			fmt.Println(theme.Step("Rebuilding sumi..."))
			build := exec.Command("go", "build", "-o", filepath.Join(sumiDir, "sumi"), "./cmd/sumi")
			build.Dir = sumiDir
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			if err := build.Run(); err != nil {
				return fmt.Errorf("go build: %w", err)
			}
			fmt.Println(theme.Ok("Binary rebuilt"))

			// 7. Re-exec sumi install with the new binary
			fmt.Println(theme.Step("Running installer..."))
			newBin := filepath.Join(sumiDir, "sumi")
			install := exec.Command(newBin, "install")
			install.Stdout = os.Stdout
			install.Stderr = os.Stderr
			install.Stdin = os.Stdin
			if err := install.Run(); err != nil {
				return fmt.Errorf("sumi install: %w", err)
			}

			// 8. Reload Hyprland if running
			if err := exec.Command("hyprctl", "reload").Run(); err == nil {
				fmt.Println(theme.Ok("Hyprland reloaded"))
			}

			// 9. Daemon-reload user services
			if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err == nil {
				fmt.Println(theme.Ok("User services reloaded"))
			}

			fmt.Println()
			fmt.Println(theme.Ok("sumi updated"))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

// renderColoredDiff prints git diff --stat output with green/red coloring.
func renderColoredDiff(diffStat string) {
	addStyle := lipgloss.NewStyle().Foreground(theme.Green)
	delStyle := lipgloss.NewStyle().Foreground(theme.Red)
	fileStyle := lipgloss.NewStyle().Foreground(theme.Text)

	scanner := bufio.NewScanner(strings.NewReader(diffStat))
	for scanner.Scan() {
		line := scanner.Text()
		// Last line is summary "N files changed, N insertions(+), N deletions(-)"
		if strings.Contains(line, "files changed") || strings.Contains(line, "file changed") {
			colored := line
			if idx := strings.Index(colored, "insertion"); idx > 0 {
				// Find the number before "insertion"
				colored = strings.Replace(colored, "+)", addStyle.Render("+")+lipgloss.NewStyle().Foreground(theme.Subtext).Render(")"), 1)
			}
			if idx := strings.Index(colored, "deletion"); idx > 0 {
				colored = strings.Replace(colored, "-)", delStyle.Render("-")+lipgloss.NewStyle().Foreground(theme.Subtext).Render(")"), 1)
			}
			fmt.Println("    " + colored)
			continue
		}

		// File lines: " path/to/file | N ++--"
		if idx := strings.Index(line, "|"); idx > 0 {
			file := strings.TrimSpace(line[:idx])
			rest := line[idx:]

			// Color the + and - in the bar
			var colored strings.Builder
			for _, c := range rest {
				switch c {
				case '+':
					colored.WriteString(addStyle.Render("+"))
				case '-':
					colored.WriteString(delStyle.Render("-"))
				default:
					colored.WriteRune(c)
				}
			}
			fmt.Println("    " + fileStyle.Render(file) + " " + colored.String())
		} else {
			fmt.Println("    " + line)
		}
	}
}
