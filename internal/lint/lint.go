package lint

import (
	"fmt"
	"os"
	"path/filepath"
)

// Severity levels for lint issues.
const (
	SevError   = "error"
	SevWarning = "warning"
)

// Issue represents a single lint finding.
type Issue struct {
	File     string
	Line     int
	Message  string
	Severity string
	Fix      func() error // optional auto-fix
}

// RunAll validates all managed configs and returns issues.
func RunAll(sumiDir, home string) []Issue {
	var issues []Issue
	issues = append(issues, CheckHyprland(filepath.Join(home, ".config/hypr"))...)
	issues = append(issues, CheckFoot(filepath.Join(home, ".config/foot/foot.ini"))...)
	issues = append(issues, CheckWaybar(filepath.Join(home, ".config/waybar"))...)
	return issues
}

// PrintIssues displays issues to stdout.
func PrintIssues(issues []Issue) int {
	errors := 0
	warnings := 0
	for _, issue := range issues {
		prefix := "warning"
		if issue.Severity == SevError {
			prefix = "error"
			errors++
		} else {
			warnings++
		}
		if issue.Line > 0 {
			fmt.Printf("  %s: %s:%d: %s\n", prefix, issue.File, issue.Line, issue.Message)
		} else {
			fmt.Printf("  %s: %s: %s\n", prefix, issue.File, issue.Message)
		}
	}

	if errors > 0 {
		return 2
	}
	if warnings > 0 {
		return 1
	}
	return 0
}

// FixIssues applies auto-fixes for issues that have them.
func FixIssues(issues []Issue) int {
	fixed := 0
	for _, issue := range issues {
		if issue.Fix != nil {
			if err := issue.Fix(); err == nil {
				fixed++
				fmt.Printf("  fixed: %s: %s\n", issue.File, issue.Message)
			}
		}
	}
	return fixed
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
