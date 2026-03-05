package lint

import (
	"os"
	"path/filepath"
	"strings"
)

// CheckHyprland validates Hyprland config files.
func CheckHyprland(hyprDir string) []Issue {
	var issues []Issue

	// Check main config exists
	mainConf := filepath.Join(hyprDir, "hyprland.conf")
	if !fileExists(mainConf) {
		issues = append(issues, Issue{
			File:     mainConf,
			Message:  "hyprland.conf not found",
			Severity: SevError,
		})
		return issues
	}

	// Check sourced files exist
	data, err := os.ReadFile(mainConf)
	if err != nil {
		return issues
	}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "source = ") || strings.HasPrefix(line, "source=") {
			path := strings.TrimPrefix(line, "source = ")
			path = strings.TrimPrefix(path, "source=")
			path = strings.TrimSpace(path)
			// Expand ~ to $HOME
			if strings.HasPrefix(path, "~") {
				path = os.Getenv("HOME") + path[1:]
			}
			if !fileExists(path) && !strings.Contains(path, "*") {
				issues = append(issues, Issue{
					File:     mainConf,
					Line:     i + 1,
					Message:  "sourced file not found: " + path,
					Severity: SevWarning,
				})
			}
		}
	}

	// Check conf.d files for deprecated syntax
	confD := filepath.Join(hyprDir, "conf.d")
	entries, err := os.ReadDir(confD)
	if err != nil {
		return issues
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".conf") {
			continue
		}
		filePath := filepath.Join(confD, e.Name())
		issues = append(issues, checkHyprFile(filePath)...)
	}

	return issues
}

func checkHyprFile(path string) []Issue {
	var issues []Issue
	data, err := os.ReadFile(path)
	if err != nil {
		return issues
	}

	for i, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)

		// Check for deprecated windowrulev2
		if strings.HasPrefix(trimmed, "windowrulev2") {
			issues = append(issues, Issue{
				File:     filepath.Base(path),
				Line:     i + 1,
				Message:  "deprecated: windowrulev2 → use windowrule (Hyprland 0.50+)",
				Severity: SevWarning,
			})
		}

		// Check for old gesture syntax
		if strings.HasPrefix(trimmed, "workspace_swipe") {
			issues = append(issues, Issue{
				File:     filepath.Base(path),
				Line:     i + 1,
				Message:  "deprecated: workspace_swipe → use gesture (Hyprland 0.50+)",
				Severity: SevWarning,
			})
		}

		// Check for deprecated first_launch_animation
		if strings.HasPrefix(trimmed, "first_launch_animation") {
			issues = append(issues, Issue{
				File:     filepath.Base(path),
				Line:     i + 1,
				Message:  "deprecated: first_launch_animation → use animation = monitorAdded, 0",
				Severity: SevWarning,
			})
		}
	}

	return issues
}
