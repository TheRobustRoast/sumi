package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckWaybar validates waybar configuration.
func CheckWaybar(waybarDir string) []Issue {
	var issues []Issue

	configPath := filepath.Join(waybarDir, "config.jsonc")
	if !fileExists(configPath) {
		configPath = filepath.Join(waybarDir, "config")
		if !fileExists(configPath) {
			return issues
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return issues
	}

	file := filepath.Base(configPath)
	content := string(data)

	// Check custom module scripts exist and are executable
	for i, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		// Look for "exec" fields pointing to scripts
		if strings.Contains(trimmed, "\"exec\":") {
			// Extract the script path
			start := strings.Index(trimmed, "\"exec\":")
			if start < 0 {
				continue
			}
			rest := trimmed[start+7:]
			rest = strings.TrimSpace(rest)
			rest = strings.Trim(rest, "\"")
			rest = strings.TrimSuffix(rest, ",")
			rest = strings.Trim(rest, "\"")

			// Get just the command (before any arguments)
			parts := strings.Fields(rest)
			if len(parts) == 0 {
				continue
			}
			script := parts[0]

			// Expand ~
			if strings.HasPrefix(script, "~") {
				script = os.Getenv("HOME") + script[1:]
			}

			// Only check file-path scripts (not inline shell)
			if strings.HasPrefix(script, "/") || strings.HasPrefix(script, os.Getenv("HOME")) {
				if !fileExists(script) {
					issues = append(issues, Issue{
						File:     file,
						Line:     i + 1,
						Message:  "script not found: " + script,
						Severity: SevWarning,
					})
				} else {
					info, _ := os.Stat(script)
					if info != nil && info.Mode()&0o111 == 0 {
						issues = append(issues, Issue{
							File:     file,
							Line:     i + 1,
							Message:  "script not executable: " + script,
							Severity: SevWarning,
						})
					}
				}
			}
		}
	}

	// Validate JSON syntax with jq if available
	if _, err := exec.LookPath("jq"); err == nil {
		// Strip comments for JSON validation
		if err := exec.Command("jq", ".", configPath).Run(); err != nil {
			issues = append(issues, Issue{
				File:     file,
				Message:  "invalid JSON syntax",
				Severity: SevError,
			})
		}
	}

	return issues
}
