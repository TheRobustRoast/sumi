package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sumi/internal/theme"
)

// controlEntry is a single item in the control center.
type controlEntry struct {
	key   string
	label string
	desc  string
	cmd   []string
	tui   bool // if true, use tea.ExecProcess
}

func (e controlEntry) Title() string       { return e.label }
func (e controlEntry) Description() string { return e.desc }
func (e controlEntry) FilterValue() string { return e.key + " " + e.label + " " + e.desc }

func controlEntries() []list.Item {
	home := os.Getenv("HOME")
	return []list.Item{
		// ── TUI apps (interactive, return to control center after exit) ──
		controlEntry{"net", "impala", "wifi & network manager", []string{"impala"}, true},
		controlEntry{"bt", "bluetuith", "bluetooth manager", []string{"bluetuith"}, true},
		controlEntry{"vol", "pulsemixer", "audio mixer", []string{"pulsemixer"}, true},
		controlEntry{"mon", "btop", "system monitor", []string{"btop"}, true},
		controlEntry{"files", "yazi", "file manager", []string{"yazi"}, true},
		controlEntry{"disk", "ncdu", "disk usage analyzer", []string{"ncdu", home}, true},
		controlEntry{"git", "lazygit", "git TUI", []string{"lazygit"}, true},
		controlEntry{"viz", "cava", "audio visualizer", []string{"cava"}, true},
		controlEntry{"bw", "bandwhich", "bandwidth monitor", []string{"sudo", "bandwhich"}, true},
		controlEntry{"proc", "procs", "process viewer", []string{"procs"}, true},
		controlEntry{"tmux", "tmux-session", "attach/create tmux session", []string{"tmux", "new-session", "-A", "-s", "main"}, true},

		// ── sumi TUI commands (interactive, full-screen) ──
		controlEntry{"wall", "wallpaper", "pick wallpaper", []string{"sumi", "wallpaper", "pick"}, false},
		controlEntry{"clip", "clipboard", "clipboard history", []string{"sumi", "clipboard"}, false},
		controlEntry{"emoji", "emoji-picker", "emoji & symbols", []string{"sumi", "emoji"}, false},
		controlEntry{"keys", "keybinds", "keybind cheatsheet", []string{"sumi", "keys"}, false},
		controlEntry{"proj", "project-launch", "open project in tmux", []string{"sumi", "project"}, false},
		controlEntry{"wt", "git-worktree", "manage git worktrees", []string{"sumi", "worktree", "list"}, false},
		controlEntry{"note", "scratch-note", "quick notes (markdown)", []string{"sumi", "note"}, false},
		controlEntry{"notif", "notifications", "notification center", []string{"sumi", "notify"}, false},
		controlEntry{"theme", "theme-toggle", "dark/light mode switch", []string{"sumi", "theme"}, false},
		controlEntry{"power", "power-menu", "shutdown/reboot/suspend", []string{"sumi", "power", "menu"}, false},

		// ── sumi quick actions (run and return) ──
		controlEntry{"pwr", "power-profile", "cycle power profile", []string{"sumi", "power", "profile"}, false},
		controlEntry{"game", "gaming-mode", "toggle gaming mode", []string{"sumi", "mode", "gaming"}, false},
		controlEntry{"focus", "focus-mode", "toggle DND / focus", []string{"sumi", "mode", "focus"}, false},
		controlEntry{"disp", "monitor-detect", "detect/configure monitors", []string{"sumi", "monitor", "detect"}, false},
		controlEntry{"sess", "session-save", "save window layout", []string{"sumi", "session", "save"}, false},
		controlEntry{"dots", "dotsync", "backup/restore configs", []string{"sumi", "sync", "snapshot"}, false},
		controlEntry{"clean", "cleanup", "trim clipboard & old files", []string{"sumi", "cleanup"}, false},
		controlEntry{"lint", "lint", "validate configs", []string{"sumi", "lint"}, false},
		controlEntry{"lock", "hyprlock", "lock screen", []string{"hyprlock"}, false},
	}
}

// ControlCenter is the Bubble Tea model for the control center.
type ControlCenter struct {
	list   list.Model
	choice *controlEntry
	err    error
	width  int
	height int
}

// NewControlCenter creates the control center TUI.
func NewControlCenter() ControlCenter {
	items := controlEntries()

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#ffffff")).
		BorderLeftForeground(theme.Blue)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(theme.Subtext).
		BorderLeftForeground(theme.Blue)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(theme.Text)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(theme.Dim)

	l := list.New(items, delegate, 60, 30)
	l.Title = "sumi :: control"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(theme.Text).
		Bold(true).
		Padding(0, 1)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.SetShowStatusBar(false)

	return ControlCenter{list: l}
}

func (m ControlCenter) Init() tea.Cmd {
	return nil
}

func (m ControlCenter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Don't intercept keys when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(controlEntry); ok {
				m.choice = &item
				if item.tui {
					c := exec.Command(item.cmd[0], item.cmd[1:]...) //nolint:gosec
					return m, tea.ExecProcess(c, func(err error) tea.Msg {
						return controlDoneMsg{err: err}
					})
				}
				// Non-TUI: run and quit
				c := exec.Command(item.cmd[0], item.cmd[1:]...) //nolint:gosec
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				c.Stdin = os.Stdin
				if err := c.Run(); err != nil {
					m.err = err
				}
				return m, tea.Quit
			}
		}
	case controlDoneMsg:
		// TUI app exited, return to list
		m.choice = nil
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ControlCenter) View() string {
	if m.err != nil {
		return fmt.Sprintf("%s\n%s\n",
			theme.Fail("Launch failed: "+m.err.Error()),
			theme.SubtextStyle.Render("Press q to exit."),
		)
	}
	return m.list.View()
}

type controlDoneMsg struct {
	err error
}
