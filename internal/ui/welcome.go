package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sumi/internal/hardware"
	"sumi/internal/theme"
)

// WelcomeWizard is a multi-page first-run TUI.
type WelcomeWizard struct {
	page       int
	pages      int
	hw         *hardware.Profile
	wallpapers []string
	wpCursor   int
	width      int
	height     int
}

// NewWelcomeWizard creates the first-run welcome wizard.
func NewWelcomeWizard() WelcomeWizard {
	hw := hardware.Detect()
	home := os.Getenv("HOME")
	wpDir := filepath.Join(home, "Pictures/Wallpapers")
	var wallpapers []string
	if entries, err := os.ReadDir(wpDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") ||
				strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".webp") {
				wallpapers = append(wallpapers, e.Name())
			}
		}
	}
	return WelcomeWizard{
		page:       0,
		pages:      4,
		hw:         hw,
		wallpapers: wallpapers,
	}
}

func (m WelcomeWizard) Init() tea.Cmd { return nil }

func (m WelcomeWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", "right", "l", "n":
			if m.page < m.pages-1 {
				m.page++
			} else {
				return m, tea.Quit
			}
		case "left", "h", "p":
			if m.page > 0 {
				m.page--
			}
		case "j", "down":
			if m.page == 2 && m.wpCursor < len(m.wallpapers)-1 {
				m.wpCursor++
			}
		case "k", "up":
			if m.page == 2 && m.wpCursor > 0 {
				m.wpCursor--
			}
		}
	}
	return m, nil
}

func (m WelcomeWizard) View() string {
	var b strings.Builder

	// Page indicator
	indicator := theme.SubtextStyle.Render(fmt.Sprintf("  %d/%d", m.page+1, m.pages))

	switch m.page {
	case 0:
		b.WriteString(m.renderWelcome())
	case 1:
		b.WriteString(m.renderKeybinds())
	case 2:
		b.WriteString(m.renderWallpapers())
	case 3:
		b.WriteString(m.renderHardware())
	}

	b.WriteString("\n")
	b.WriteString(indicator + "  ")
	if m.page < m.pages-1 {
		b.WriteString(theme.StepStyle.Render("[Enter]") + theme.SubtextStyle.Render(" next  "))
	} else {
		b.WriteString(theme.OkStyle.Render("[Enter]") + theme.SubtextStyle.Render(" done  "))
	}
	if m.page > 0 {
		b.WriteString(theme.SubtextStyle.Render("[←] back  "))
	}
	b.WriteString(theme.SubtextStyle.Render("[q] quit"))
	b.WriteString("\n")

	return b.String()
}

func (m WelcomeWizard) renderWelcome() string {
	titleStyle := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	logo := titleStyle.Render(`
    ┌─────────────────────────────┐
    │                             │
    │     s u m i                 │
    │                             │
    │     Hyprland rice for       │
    │     Arch Linux              │
    │                             │
    └─────────────────────────────┘`)

	body := lipgloss.JoinVertical(lipgloss.Left,
		logo,
		"",
		theme.SubtextStyle.Render("  Welcome to sumi."),
		theme.SubtextStyle.Render("  This wizard will get you started."),
		"",
		theme.SubtextStyle.Render("  Your wallpaper IS your brand identity —"),
		theme.SubtextStyle.Render("  colors are extracted at runtime."),
	)
	return body
}

func (m WelcomeWizard) renderKeybinds() string {
	titleStyle := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	keyStyle := lipgloss.NewStyle().Foreground(theme.Blue).Width(22)
	descStyle := lipgloss.NewStyle().Foreground(theme.Subtext)

	keybinds := [][2]string{
		{"SUPER + Enter", "Terminal (foot)"},
		{"SUPER + X", "Control center"},
		{"SUPER + /", "Keybind cheatsheet"},
		{"SUPER + D", "App launcher (fuzzel)"},
		{"SUPER + E", "File manager (yazi)"},
		{"SUPER + Q", "Close window"},
		{"SUPER + V", "Clipboard history"},
		{"SUPER + Print", "Screenshot"},
		{"SUPER + L", "Lock screen"},
		{"SUPER + 1-9", "Switch workspace"},
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Top 10 Keybinds"))
	b.WriteString("\n\n")
	for _, kb := range keybinds {
		b.WriteString("  " + keyStyle.Render(kb[0]) + descStyle.Render(kb[1]) + "\n")
	}
	return b.String()
}

func (m WelcomeWizard) renderWallpapers() string {
	titleStyle := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Wallpapers"))
	b.WriteString("\n\n")

	if len(m.wallpapers) == 0 {
		b.WriteString(theme.SubtextStyle.Render("  No wallpapers found in ~/Pictures/Wallpapers/") + "\n")
		b.WriteString(theme.SubtextStyle.Render("  Add .png, .jpg, or .webp files to get started.") + "\n")
		b.WriteString("\n")
		b.WriteString(theme.SubtextStyle.Render("  Then run: sumi theme") + "\n")
	} else {
		b.WriteString(theme.SubtextStyle.Render(fmt.Sprintf("  Found %d wallpaper(s):", len(m.wallpapers))) + "\n\n")
		visible := 10
		start := 0
		if m.wpCursor >= visible {
			start = m.wpCursor - visible + 1
		}
		end := start + visible
		if end > len(m.wallpapers) {
			end = len(m.wallpapers)
		}
		for i := start; i < end; i++ {
			prefix := "  "
			style := theme.SubtextStyle
			if i == m.wpCursor {
				prefix = theme.StepStyle.Render("> ")
				style = lipgloss.NewStyle().Foreground(theme.Text)
			}
			b.WriteString("  " + prefix + style.Render(m.wallpapers[i]) + "\n")
		}
		b.WriteString("\n")
		b.WriteString(theme.SubtextStyle.Render("  Set a wallpaper after setup: sumi theme") + "\n")
	}
	return b.String()
}

func (m WelcomeWizard) renderHardware() string {
	titleStyle := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Dim).Width(16)
	valueStyle := lipgloss.NewStyle().Foreground(theme.Text)

	row := func(label, value string) string {
		return "  " + labelStyle.Render(label) + valueStyle.Render(value)
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Hardware Profile"))
	b.WriteString("\n\n")
	b.WriteString(row("Profile", m.hw.Name) + "\n")
	b.WriteString(row("ID", m.hw.ID) + "\n")

	if len(m.hw.Packages) > 0 {
		b.WriteString(row("Packages", strings.Join(m.hw.Packages, ", ")) + "\n")
	}
	if len(m.hw.AURPackages) > 0 {
		b.WriteString(row("AUR", strings.Join(m.hw.AURPackages, ", ")) + "\n")
	}
	if len(m.hw.Services) > 0 {
		b.WriteString(row("Services", strings.Join(m.hw.Services, ", ")) + "\n")
	}
	if len(m.hw.KernelParams) > 0 {
		b.WriteString(row("Kernel", strings.Join(m.hw.KernelParams, " ")) + "\n")
	}
	if len(m.hw.Modules) > 0 {
		b.WriteString(row("Modules", strings.Join(m.hw.Modules, ", ")) + "\n")
	}
	hasTweaks := "none"
	if m.hw.Tweaks != nil {
		hasTweaks = "yes"
	}
	b.WriteString(row("Tweaks", hasTweaks) + "\n")

	b.WriteString("\n")
	b.WriteString(theme.SubtextStyle.Render("  Override in: sumi config → Hardware profile") + "\n")
	b.WriteString("\n")
	b.WriteString(theme.Ok("Setup complete") + "\n")
	b.WriteString(theme.SubtextStyle.Render("  Run 'sumi doctor' to verify your installation.") + "\n")

	return b.String()
}
