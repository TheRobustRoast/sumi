package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sumi/internal/model"
	"sumi/internal/runner"
	"sumi/internal/steps"
	"sumi/internal/theme"
)

// StepRunnerConfig configures a generic step-runner TUI.
type StepRunnerConfig struct {
	Title    string // e.g. "doctor", "uninstall"
	Subtitle string // e.g. "health checks"
	Steps    []model.Step
	Ctx      model.InstallCtx
	// DoneMsg is the text shown on success. Empty = default.
	DoneMsg string
}

// StepRunner is a reusable Bubble Tea model for running a list of steps.
type StepRunner struct {
	cfg StepRunnerConfig

	cur    int
	stream *runner.Stream
	log    []string

	vp   viewport.Model
	spin spinner.Model

	done  bool
	err   error
	width int
	ready bool
}

// NewStepRunner creates a step-runner TUI from the given config.
func NewStepRunner(cfg StepRunnerConfig) StepRunner {
	return StepRunner{
		cfg:  cfg,
		spin: NewSpinner(),
	}
}

func (m StepRunner) Init() tea.Cmd {
	return func() tea.Msg { return model.NextStepMsg{} }
}

func (m StepRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		const logHeight = 8
		if !m.ready {
			m.vp = NewViewport(msg.Width, logHeight)
			m.ready = true
		} else {
			m.vp.Width = msg.Width - 6
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" || msg.String() == "esc" {
			return m, tea.Quit
		}

	case model.NextStepMsg:
		if m.cur >= len(m.cfg.Steps) {
			m.done = true
			return m, nil
		}
		newCur, stream, cmd := RunCurrentStep(m.cfg.Steps, m.cur, m.cfg.Ctx)
		m.cur = newCur
		m.log = nil
		if stream != nil {
			m.stream = stream
			return m, tea.Batch(cmd, m.spin.Tick)
		}
		if cmd != nil {
			return m, cmd
		}
		return m, nil

	case runner.OutputMsg:
		m.log = append(m.log, msg.Line)
		m.vp.SetContent(strings.Join(m.log, "\n"))
		m.vp.GotoBottom()
		if m.stream != nil {
			return m, m.stream.Next()
		}

	case runner.DoneMsg:
		m.stream = nil
		if msg.Err != nil {
			m.cfg.Steps[m.cur].Status = model.StepFailed
			m.err = msg.Err
			m.cur++
			// Doctor doesn't quit on failure — continues to next check
			return m, func() tea.Msg { return model.NextStepMsg{} }
		}
		m.cfg.Steps[m.cur].Status = model.StepDone
		m.cur++
		return m, func() tea.Msg { return model.NextStepMsg{} }

	case model.GoResultMsg:
		if len(msg.Lines) > 0 {
			m.log = msg.Lines
			m.vp.SetContent(strings.Join(m.log, "\n"))
			m.vp.GotoBottom()
		}
		if msg.Err != nil {
			m.cfg.Steps[m.cur].Status = model.StepFailed
			m.err = msg.Err
			m.cur++
			return m, func() tea.Msg { return model.NextStepMsg{} }
		}
		m.cfg.Steps[m.cur].Status = model.StepDone
		m.cur++
		return m, func() tea.Msg { return model.NextStepMsg{} }

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}

	var vpCmd tea.Cmd
	m.vp, vpCmd = m.vp.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m StepRunner) View() string {
	if !m.ready {
		return "  Starting...\n"
	}

	var b strings.Builder

	title := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.Text).Bold(true).Render("sumi"),
		lipgloss.NewStyle().Foreground(theme.Overlay).Render(" :: "),
		lipgloss.NewStyle().Foreground(theme.Subtext).Render(m.cfg.Title),
	)
	if m.cfg.Subtitle != "" {
		b.WriteString(theme.BoxStyle.Render(title + "\n" + theme.SubtextStyle.Render(m.cfg.Subtitle)))
	} else {
		b.WriteString(theme.BoxStyle.Render(title))
	}
	b.WriteString("\n")

	// Progress bar
	completed, total := CountProgress(m.cfg.Steps)
	if completed > 0 || m.done {
		b.WriteString(RenderProgress(completed, total, m.width))
		b.WriteString("\n")
	}

	RenderSteps(&b, m.cfg.Steps, m.spin)

	if len(m.log) > 0 || m.stream != nil {
		b.WriteString("\n")
		b.WriteString(theme.LogBorderStyle.Width(m.vp.Width + 2).Render(m.vp.View()))
		b.WriteString("\n")
	}

	if m.done {
		passed, failed, skipped := countSteps(m.cfg.Steps)
		b.WriteString("\n")
		summary := theme.Ok(strings.TrimSpace(
			fmt.Sprintf("%d passed", passed),
		))
		if failed > 0 {
			summary += "  " + theme.Fail(fmt.Sprintf("%d failed", failed))
		}
		if skipped > 0 {
			summary += "  " + theme.SubtextStyle.Render(fmt.Sprintf("%d skipped", skipped))
		}
		b.WriteString(summary + "\n")

		// Show fix suggestions for failed steps
		if failed > 0 {
			for _, s := range m.cfg.Steps {
				if s.Status == model.StepFailed {
					if fix := steps.DoctorFix(s.Name); fix != "" {
						b.WriteString(theme.Warn("Fix "+s.Name+": ") + theme.SubtextStyle.Render(fix) + "\n")
					}
				}
			}
		}

		doneMsg := m.cfg.DoneMsg
		if doneMsg == "" {
			doneMsg = "Done"
		}
		if failed == 0 {
			b.WriteString(theme.Ok(doneMsg) + "\n")
		}
		b.WriteString(theme.SubtextStyle.Render("\n  Press q to exit.") + "\n")
	}

	return b.String()
}

func countSteps(steps []model.Step) (passed, failed, skipped int) {
	for _, s := range steps {
		switch s.Status {
		case model.StepDone:
			passed++
		case model.StepFailed:
			failed++
		case model.StepSkipped:
			skipped++
		}
	}
	return
}

