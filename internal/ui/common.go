package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/model"
	"sumi/internal/runner"
	"sumi/internal/theme"
)

// RenderSteps writes the step list (only non-pending steps) to a string builder.
func RenderSteps(b *strings.Builder, steps []model.Step, spin spinner.Model) {
	var lastSection string
	for _, s := range steps {
		if s.Status == model.StepPending {
			continue
		}
		if s.Section != "" && s.Section != lastSection {
			lastSection = s.Section
			b.WriteString("\n" + theme.Section(s.Section) + "\n\n")
		}
		switch s.Status {
		case model.StepRunning:
			b.WriteString(theme.StepStyle.Render("  ·  "+s.Name+"...  ") + spin.View() + "\n")
		case model.StepDone:
			b.WriteString(theme.Ok(s.Name) + "\n")
		case model.StepSkipped:
			b.WriteString(theme.SubtextStyle.Render("  ·  "+s.Name+" (already done)") + "\n")
		case model.StepFailed:
			b.WriteString(theme.Fail(s.Name) + "\n")
		}
	}
}

// RunCurrentStep executes the current step, returning the new cursor, stream, and tea.Cmd.
func RunCurrentStep(steps []model.Step, cur int, ctx model.InstallCtx) (int, *runner.Stream, tea.Cmd) {
	if cur >= len(steps) {
		return cur, nil, nil
	}
	s := steps[cur]

	if s.Skip != nil && s.Skip(ctx) {
		steps[cur].Status = model.StepSkipped
		return cur + 1, nil, func() tea.Msg { return model.NextStepMsg{} }
	}

	steps[cur].Status = model.StepRunning

	if s.RunGo != nil {
		fn := s.RunGo
		return cur, nil, func() tea.Msg {
			lines, err := fn(ctx)
			return model.GoResultMsg{Lines: lines, Err: err}
		}
	}
	if s.RunStream != nil {
		stream, cmd := s.RunStream(ctx)
		return cur, stream, cmd
	}

	// No-op step
	steps[cur].Status = model.StepDone
	return cur + 1, nil, func() tea.Msg { return model.NextStepMsg{} }
}

// NewSpinner creates a spinner configured for sumi step runners.
func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.StepStyle
	return s
}

// NewViewport creates a viewport configured for the log pane.
func NewViewport(width, height int) viewport.Model {
	return viewport.New(width-6, height)
}
