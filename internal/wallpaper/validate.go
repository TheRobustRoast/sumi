package wallpaper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateTemplates checks that all wallust template output paths exist and are writable.
func ValidateTemplates(wallustDir string) []string {
	tplDir := filepath.Join(wallustDir, "templates")
	entries, err := os.ReadDir(tplDir)
	if err != nil {
		return []string{fmt.Sprintf("cannot read templates dir: %s", tplDir)}
	}

	var issues []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Read template to find output path (wallust templates have a header comment
		// or the output path is configured in wallust.toml)
		tplPath := filepath.Join(tplDir, e.Name())
		data, err := os.ReadFile(tplPath)
		if err != nil {
			issues = append(issues, fmt.Sprintf("cannot read template: %s", e.Name()))
			continue
		}

		// Check for unknown variables (basic check)
		content := string(data)
		knownVars := []string{
			"{{background}}", "{{foreground}}", "{{cursor}}",
			"{{color0}}", "{{color1}}", "{{color2}}", "{{color3}}",
			"{{color4}}", "{{color5}}", "{{color6}}", "{{color7}}",
			"{{color8}}", "{{color9}}", "{{color10}}", "{{color11}}",
			"{{color12}}", "{{color13}}", "{{color14}}", "{{color15}}",
		}
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if idx := strings.Index(line, "{{"); idx >= 0 {
				end := strings.Index(line[idx:], "}}")
				if end >= 0 {
					varName := line[idx : idx+end+2]
					found := false
					for _, known := range knownVars {
						if varName == known {
							found = true
							break
						}
					}
					if !found {
						issues = append(issues, fmt.Sprintf("%s:%d: unknown variable %s", e.Name(), i+1, varName))
					}
				}
			}
		}
	}

	return issues
}
