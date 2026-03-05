package theme

import "github.com/charmbracelet/lipgloss"

var (
	// BoxStyle is a rounded-border card used for the header and output pane.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Surface2).
			Padding(0, 2)

	// SectionStyle marks a logical group of installer steps.
	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Text)

	// StepStyle is a step that is in progress.
	StepStyle = lipgloss.NewStyle().
			Foreground(Blue)

	// OkStyle marks a completed step.
	OkStyle = lipgloss.NewStyle().
		Foreground(Green)

	// FailStyle marks a failed step.
	FailStyle = lipgloss.NewStyle().
			Foreground(Red)

	// WarnStyle marks a warning.
	WarnStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	// SubtextStyle is used for secondary, dimmed text.
	SubtextStyle = lipgloss.NewStyle().
			Foreground(Subtext)

	// LogBorderStyle is the border around the scrollable output pane.
	LogBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Overlay).
			Padding(0, 1)
)

// Section renders a section header: "  ── name"
func Section(name string) string {
	return SectionStyle.Render("  ── " + name)
}

// Ok renders a success line: "  ✓  text"
func Ok(text string) string {
	return OkStyle.Render("  ✓  " + text)
}

// Fail renders a failure line: "  ✗  text"
func Fail(text string) string {
	return FailStyle.Render("  ✗  " + text)
}

// Step renders an in-progress line: "  ·  text"
func Step(text string) string {
	return StepStyle.Render("  ·  " + text)
}

// Warn renders a warning line: "  !  text"
func Warn(text string) string {
	return WarnStyle.Render("  !  " + text)
}
