package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sumi/internal/checkpoint"
	"sumi/internal/hardware"
	"sumi/internal/model"
	"sumi/internal/runner"
	"sumi/internal/steps"
	"sumi/internal/theme"
)

// failAction represents the user's choice on step failure.
type failAction int

const (
	failWaiting failAction = iota
	failRetry
	failSkip
	failQuit
)

// Installer is the bubbletea model for the sumi rice installer.
type Installer struct {
	sumiDir string
	user    string
	home    string
	hw      *hardware.Profile

	steps []model.Step
	cur   int

	stream *runner.Stream
	log    []string

	vp   viewport.Model
	spin spinner.Model

	done            bool
	err             error
	failChoice      failAction
	RebootRequested bool

	ckpt *checkpoint.State

	width  int
	height int
	ready  bool
}

// New creates a ready-to-run Installer for the sumi repo at sumiDir.
func New(sumiDir string) Installer {
	hw := hardware.Detect()

	m := Installer{
		sumiDir: sumiDir,
		user:    os.Getenv("USER"),
		home:    os.Getenv("HOME"),
		hw:      hw,
		spin:    NewSpinner(),
	}
	m.steps = m.buildSteps()

	// Check for existing checkpoint and skip completed steps
	if ckpt := checkpoint.Load(); ckpt != nil && len(ckpt.CompletedSteps) > 0 {
		m.ckpt = ckpt
		// Mark completed steps
		for i := range m.steps {
			if m.steps[i].ID != "" && checkpoint.IsCompleted(ckpt, m.steps[i].ID) {
				m.steps[i].Status = model.StepSkipped
			}
		}
		ids := make([]string, len(m.steps))
		for i, s := range m.steps {
			ids[i] = s.ID
		}
		m.cur = checkpoint.ResumeIndex(ckpt, ids)
	} else {
		m.ckpt = &checkpoint.State{}
	}

	return m
}

func (m *Installer) buildSteps() []model.Step {
	var all []model.Step
	all = append(all, steps.Preflight()...)
	all = append(all, steps.Keyring()...)
	all = append(all, steps.Packages(m.hw)...)
	all = append(all, steps.Dotfiles())
	all = append(all, steps.Greetd())
	all = append(all, steps.Plymouth())
	all = append(all, steps.Dirs())
	all = append(all, steps.Services(m.hw))
	all = append(all, steps.HardwareTweaks(m.hw))
	all = append(all, steps.SetShell())
	all = append(all, steps.DisableConflictingDMs())
	all = append(all, steps.Migrations())
	return all
}

// ── bubbletea interface ───────────────────────────────────────────────────────

func (m Installer) Init() tea.Cmd {
	return func() tea.Msg { return model.NextStepMsg{} }
}

func (m Installer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		const logHeight = 8
		if !m.ready {
			m.vp = NewViewport(msg.Width, logHeight)
			m.ready = true
		} else {
			m.vp.Width = msg.Width - 6
		}

	case tea.KeyMsg:
		// Handle failure choice
		if m.err != nil && m.failChoice == failWaiting {
			switch msg.String() {
			case "r", "R":
				m.failChoice = failRetry
				m.err = nil
				m.steps[m.cur].Status = model.StepPending
				return m, func() tea.Msg { return model.NextStepMsg{} }
			case "s", "S":
				m.failChoice = failSkip
				m.err = nil
				m.steps[m.cur].Status = model.StepSkipped
				if m.steps[m.cur].ID != "" {
					checkpoint.MarkCompleted(m.ckpt, m.steps[m.cur].ID) //nolint:errcheck
				}
				m.cur++
				m.failChoice = failWaiting
				return m, func() tea.Msg { return model.NextStepMsg{} }
			case "q", "Q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "y", "Y":
			if m.done {
				m.RebootRequested = true
				return m, tea.Quit
			}
		case "n", "N", "enter", "esc":
			if m.done {
				return m, tea.Quit
			}
		}

	case model.NextStepMsg:
		if m.cur >= len(m.steps) {
			m.done = true
			checkpoint.Clear()
			return m, nil
		}
		ctx := model.InstallCtx{SumiDir: m.sumiDir, Home: m.home, User: m.user}
		newCur, stream, cmd := RunCurrentStep(m.steps, m.cur, ctx)
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
			m.steps[m.cur].Status = model.StepFailed
			m.err = msg.Err
			m.failChoice = failWaiting
			if m.steps[m.cur].ID != "" {
				checkpoint.MarkFailed(m.ckpt, m.steps[m.cur].ID) //nolint:errcheck
			}
			return m, nil
		}
		m.steps[m.cur].Status = model.StepDone
		if m.steps[m.cur].ID != "" {
			checkpoint.MarkCompleted(m.ckpt, m.steps[m.cur].ID) //nolint:errcheck
		}
		m.cur++
		return m, func() tea.Msg { return model.NextStepMsg{} }

	case model.GoResultMsg:
		if len(msg.Lines) > 0 {
			m.log = msg.Lines
			m.vp.SetContent(strings.Join(m.log, "\n"))
			m.vp.GotoBottom()
		}
		if msg.Err != nil {
			m.steps[m.cur].Status = model.StepFailed
			m.err = msg.Err
			m.failChoice = failWaiting
			if m.steps[m.cur].ID != "" {
				checkpoint.MarkFailed(m.ckpt, m.steps[m.cur].ID) //nolint:errcheck
			}
			return m, nil
		}
		m.steps[m.cur].Status = model.StepDone
		if m.steps[m.cur].ID != "" {
			checkpoint.MarkCompleted(m.ckpt, m.steps[m.cur].ID) //nolint:errcheck
		}
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

func (m Installer) View() string {
	if !m.ready {
		return "  Starting...\n"
	}

	var b strings.Builder

	// Header
	title := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(theme.Text).Bold(true).Render("sumi"),
		lipgloss.NewStyle().Foreground(theme.Overlay).Render(" :: "),
		lipgloss.NewStyle().Foreground(theme.Subtext).Render("install"),
	)
	subtitle := theme.SubtextStyle.Render("Hyprland rice  ·  " + m.hw.Name)
	b.WriteString(theme.BoxStyle.Render(title + "\n" + subtitle))
	b.WriteString("\n")

	// Progress bar
	completed, total := CountProgress(m.steps)
	if completed > 0 || m.done {
		b.WriteString(RenderProgress(completed, total, m.width))
		b.WriteString("\n")
	}

	// Step list
	RenderSteps(&b, m.steps, m.spin)

	// Output log viewport
	if len(m.log) > 0 || m.stream != nil {
		b.WriteString("\n")
		b.WriteString(theme.LogBorderStyle.Width(m.vp.Width + 2).Render(m.vp.View()))
		b.WriteString("\n")
	}

	// Error with retry/skip/quit
	if m.err != nil {
		b.WriteString("\n" + theme.Fail("Error: "+m.err.Error()) + "\n")
		if m.cur < len(m.steps) {
			if diag := steps.DiagnoseError(m.steps[m.cur].Name, m.err); diag != "" {
				b.WriteString(theme.Warn("Hint: "+diag) + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(theme.StepStyle.Render("  [r]") + theme.SubtextStyle.Render(" Retry  "))
		b.WriteString(theme.WarnStyle.Render("[s]") + theme.SubtextStyle.Render(" Skip  "))
		b.WriteString(theme.FailStyle.Render("[q]") + theme.SubtextStyle.Render(" Quit"))
		b.WriteString("\n")
	}

	// Done screen
	if m.done {
		b.WriteString("\n")
		body := lipgloss.JoinVertical(lipgloss.Left,
			theme.Ok("sumi installed"),
			"",
			theme.SubtextStyle.Render("  1.  Drop wallpapers in ~/Pictures/Wallpapers/"),
			theme.SubtextStyle.Render("  2.  Reboot → LUKS unlock → autologin → Hyprland"),
			theme.SubtextStyle.Render("  3.  SUPER+X  control center    SUPER+/  keybinds"),
			"",
			theme.SubtextStyle.Render("  To update: ")+lipgloss.NewStyle().Foreground(theme.Blue).Render("sumi update"),
		)
		b.WriteString(theme.BoxStyle.BorderForeground(theme.Green).Render(body))
		b.WriteString("\n\n")
		b.WriteString(theme.StepStyle.Render("  Reboot now? "))
		b.WriteString(theme.OkStyle.Render("[y]"))
		b.WriteString(theme.SubtextStyle.Render("/[N]  "))
		b.WriteString("\n")
	}

	return b.String()
}
