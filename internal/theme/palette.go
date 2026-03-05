package theme

import "github.com/charmbracelet/lipgloss"

// Monochrome palette — the sumi identity is the wallpaper, not a fixed palette.
// These are structural TUI colors. Accent comes from wallpaper extraction at runtime.
const (
	// Backgrounds
	Base     = lipgloss.Color("#0a0a0a") // deepest background
	Surface0 = lipgloss.Color("#1a1a1a") // cards, panels
	Surface1 = lipgloss.Color("#2a2a2a") // hover, selection bg
	Surface2 = lipgloss.Color("#3a3a3a") // borders

	// Text
	Text    = lipgloss.Color("#d4d4d4") // primary text
	Subtext = lipgloss.Color("#8a8a8a") // secondary, muted
	Dim     = lipgloss.Color("#6a6a6a") // tertiary
	Overlay = lipgloss.Color("#4a4a4a") // subtle structural elements

	// Status — the only color in the TUI
	Green  = lipgloss.Color("#9ece6a") // success
	Red    = lipgloss.Color("#f7768e") // error
	Yellow = lipgloss.Color("#e0af68") // warning
	Blue   = lipgloss.Color("#7aa2f7") // info, in-progress
)
