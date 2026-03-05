package ui

import (
	"fmt"

	"sumi/internal/model"
	"sumi/internal/theme"
)

// RenderProgress returns a progress bar string like:
// ▓▓▓▓▓▓▓▓░░░░░░░░  60%  (8/13)
func RenderProgress(completed, total, width int) string {
	if total == 0 {
		return ""
	}

	pct := completed * 100 / total

	// Reserve space for " 100%  (99/99)" = ~15 chars + 2 padding
	labelWidth := len(fmt.Sprintf("  %d%%  (%d/%d)", pct, completed, total))
	barWidth := width - labelWidth - 4 // 2 padding each side
	if barWidth < 10 {
		barWidth = 10
	}

	filled := barWidth * completed / total
	empty := barWidth - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += theme.OkStyle.Render("▓")
	}
	for i := 0; i < empty; i++ {
		bar += theme.SubtextStyle.Render("░")
	}

	label := theme.SubtextStyle.Render(fmt.Sprintf("  %d%%  (%d/%d)", pct, completed, total))
	return "  " + bar + label
}

// CountProgress counts completed and total non-pending steps.
func CountProgress(steps []model.Step) (completed, total int) {
	total = len(steps)
	for _, s := range steps {
		switch s.Status {
		case model.StepDone, model.StepSkipped, model.StepFailed:
			completed++
		}
	}
	return
}
