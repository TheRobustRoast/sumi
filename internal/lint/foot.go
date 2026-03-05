package lint

import (
	"os"
	"path/filepath"
	"strings"
)

// CheckFoot validates foot terminal config.
func CheckFoot(footIni string) []Issue {
	var issues []Issue

	if !fileExists(footIni) {
		return issues
	}

	data, err := os.ReadFile(footIni)
	if err != nil {
		return issues
	}

	file := filepath.Base(footIni)
	inTweakSection := false

	for i, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)

		// Track sections
		if strings.HasPrefix(trimmed, "[") {
			inTweakSection = strings.HasPrefix(trimmed, "[tweak]")
			if inTweakSection {
				issues = append(issues, Issue{
					File:     file,
					Line:     i + 1,
					Message:  "removed section: [tweak] is no longer valid in foot",
					Severity: SevWarning,
				})
			}
		}

		// Check dpi-aware=auto
		if strings.HasPrefix(trimmed, "dpi-aware") && strings.Contains(trimmed, "auto") {
			issues = append(issues, Issue{
				File:     file,
				Line:     i + 1,
				Message:  "invalid: dpi-aware only accepts yes/no (not auto)",
				Severity: SevError,
				Fix: func() error {
					content := strings.ReplaceAll(string(data), "dpi-aware=auto", "dpi-aware=yes")
					content = strings.ReplaceAll(content, "dpi-aware = auto", "dpi-aware = yes")
					return os.WriteFile(footIni, []byte(content), 0o644)
				},
			})
		}

		// Check removed options
		for _, removed := range []string{"bold-text-uses-bright-colors", "url.protocols", "bell.command="} {
			if strings.HasPrefix(trimmed, removed) {
				issues = append(issues, Issue{
					File:     file,
					Line:     i + 1,
					Message:  "removed option: " + removed,
					Severity: SevWarning,
				})
			}
		}
	}

	return issues
}
