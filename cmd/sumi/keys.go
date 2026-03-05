package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"sumi/internal/keybinds"
	"sumi/internal/theme"
)

func keysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keys",
		Short: "Interactive keybind explorer",
		Long:  "Browse all Hyprland keybinds with fuzzy search, grouped by section.\nParsed live from keybinds.conf.",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := os.Getenv("HOME")
			confPath := filepath.Join(home, ".config/hypr/conf.d/keybinds.conf")

			if _, err := os.Stat(confPath); err != nil {
				return fmt.Errorf("keybinds.conf not found at %s — run 'sumi install' first", confPath)
			}
			binds, err := keybinds.ParseKeybindsConf(confPath)
			if err != nil {
				return fmt.Errorf("parse keybinds.conf: %w", err)
			}

			var items []list.Item
			for _, b := range binds {
				items = append(items, keybindItem{bind: b})
			}

			l := list.New(items, keybindDelegate{}, 60, 20)
			l.Title = "sumi :: keybind explorer"
			l.SetShowStatusBar(true)
			l.SetFilteringEnabled(true)
			l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(theme.Text)

			m := keysModel{list: l}
			p := tea.NewProgram(m, tea.WithAltScreen())
			_, err = p.Run()
			return silenceQuit(err)
		},
	}
}

type keybindItem struct {
	bind keybinds.Keybind
}

func (i keybindItem) Title() string {
	mod := i.bind.Mod
	if mod != "" {
		mod += "+"
	}
	return mod + i.bind.Key
}

func (i keybindItem) Description() string {
	desc := i.bind.Description
	if desc == "" {
		desc = i.bind.Dispatcher
		if i.bind.Args != "" {
			desc += " " + i.bind.Args
		}
	}
	return desc
}

func (i keybindItem) FilterValue() string {
	return i.Title() + " " + i.Description() + " " + i.bind.Section
}

type keybindDelegate struct{}

func (d keybindDelegate) Height() int                             { return 1 }
func (d keybindDelegate) Spacing() int                            { return 0 }
func (d keybindDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d keybindDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(keybindItem)
	if !ok {
		return
	}

	keyStyle := lipgloss.NewStyle().Foreground(theme.Blue).Width(20)
	descStyle := lipgloss.NewStyle().Foreground(theme.Subtext)
	selectedKeyStyle := lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Width(20)
	selectedDescStyle := lipgloss.NewStyle().Foreground(theme.Text)

	isSelected := index == m.Index()
	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	key := i.Title()
	desc := i.Description()

	if isSelected {
		fmt.Fprint(w, prefix+selectedKeyStyle.Render(key)+selectedDescStyle.Render(desc))
	} else {
		fmt.Fprint(w, prefix+keyStyle.Render(key)+descStyle.Render(desc))
	}
}

type keysModel struct {
	list list.Model
}

func (m keysModel) Init() tea.Cmd { return nil }

func (m keysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" && !m.list.SettingFilter() {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m keysModel) View() string {
	return m.list.View()
}
