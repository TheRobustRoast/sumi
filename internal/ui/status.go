package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"sumi/internal/capture"
	"sumi/internal/hardware"
	"sumi/internal/mode"
	"sumi/internal/power"
	"sumi/internal/theme"
)

// StatusInfo holds the data displayed by sumi status.
type StatusInfo struct {
	Wallpaper   string
	ThemeMode   string
	Accent      string
	HWProfile   string
	Hyprland    string
	Battery     string
	Services    string
	LastUpdate  string
	GamingMode  bool
	FocusMode   bool
	Recording   bool
	PowerProfile string
}

// GatherStatus collects system state for display.
func GatherStatus(sumiDir string) StatusInfo {
	home := os.Getenv("HOME")
	info := StatusInfo{
		Wallpaper:    readFileOr(filepath.Join(home, ".cache/sumi/current-wallpaper"), "none"),
		ThemeMode:    readFileOr(filepath.Join(home, ".cache/sumi/theme-mode"), "dark"),
		Accent:       readAccent(filepath.Join(home, ".config/hypr/conf.d/colors.conf")),
		HWProfile:    hardware.Detect().Name,
		Hyprland:     hyprlandVersion(),
		Battery:      batteryStatus(),
		Services:     serviceStatus(),
		LastUpdate:   lastGitUpdate(sumiDir),
		GamingMode:   mode.IsGaming(home),
		FocusMode:    mode.IsFocus(home),
		Recording:    capture.IsRecording(),
		PowerProfile: power.CurrentProfile(),
	}
	return info
}

// RenderStatus returns a formatted status display.
func RenderStatus(info StatusInfo) string {
	titleStyle := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Dim).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(theme.Text)

	title := lipgloss.JoinHorizontal(lipgloss.Left,
		titleStyle.Render("sumi"),
		lipgloss.NewStyle().Foreground(theme.Overlay).Render(" :: "),
		lipgloss.NewStyle().Foreground(theme.Subtext).Render("status"),
	)

	row := func(label, value string) string {
		return "  " + labelStyle.Render(label) + valueStyle.Render(value)
	}

	// Active modes line
	var modes []string
	if info.GamingMode {
		modes = append(modes, lipgloss.NewStyle().Foreground(theme.Red).Render("gaming"))
	}
	if info.FocusMode {
		modes = append(modes, lipgloss.NewStyle().Foreground(theme.Yellow).Render("focus"))
	}
	if info.Recording {
		modes = append(modes, lipgloss.NewStyle().Foreground(theme.Red).Bold(true).Render("● REC"))
	}
	modesStr := "none"
	if len(modes) > 0 {
		modesStr = strings.Join(modes, " ")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		row("Wallpaper", shortenHome(info.Wallpaper)),
		row("Theme", info.ThemeMode),
		row("Accent", info.Accent),
		row("Hardware", info.HWProfile),
		row("Power", info.PowerProfile),
		row("Hyprland", info.Hyprland),
		row("Battery", info.Battery),
		row("Services", info.Services),
		row("Modes", modesStr),
		row("Last update", info.LastUpdate),
	)

	return theme.BoxStyle.Render(body)
}

func readFileOr(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}

func readAccent(colorsConf string) string {
	data, err := os.ReadFile(colorsConf)
	if err != nil {
		return "#7aa2f7"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "$accent") && strings.Contains(line, "rgb(") && !strings.Contains(line, "accent_dim") {
			start := strings.Index(line, "rgb(")
			end := strings.Index(line[start:], ")")
			if start >= 0 && end > 0 {
				return "#" + line[start+4:start+end]
			}
		}
	}
	return "#7aa2f7"
}

func hyprlandVersion() string {
	out, err := exec.Command("hyprctl", "version", "-j").Output()
	if err != nil {
		return "not running"
	}
	// Quick parse: find "tag" field
	s := string(out)
	if idx := strings.Index(s, `"tag"`); idx >= 0 {
		rest := s[idx:]
		q1 := strings.Index(rest, `": "`)
		q2 := strings.Index(rest[q1+4:], `"`)
		if q1 >= 0 && q2 >= 0 {
			return rest[q1+4 : q1+4+q2]
		}
	}
	return "running"
}

func batteryStatus() string {
	for _, bat := range []string{"BAT0", "BAT1"} {
		base := "/sys/class/power_supply/" + bat
		capData, err := os.ReadFile(base + "/capacity")
		if err != nil {
			continue
		}
		cap := strings.TrimSpace(string(capData))
		statusData, _ := os.ReadFile(base + "/status")
		status := strings.ToLower(strings.TrimSpace(string(statusData)))

		limitData, _ := os.ReadFile(base + "/charge_control_end_threshold")
		limit := strings.TrimSpace(string(limitData))

		result := cap + "% (" + status + ")"
		if limit != "" {
			result += fmt.Sprintf(", limit: %s%%", limit)
		}
		return result
	}
	return "n/a"
}

func serviceStatus() string {
	services := []string{
		"cliphist.service",
		"wallust-watcher.service",
		"lock-before-sleep.service",
		"sumi-cleanup.timer",
	}
	running := 0
	for _, svc := range services {
		out, err := exec.Command("systemctl", "--user", "is-active", svc).Output()
		if err == nil && strings.TrimSpace(string(out)) == "active" {
			running++
		}
	}
	return fmt.Sprintf("%d/%d running", running, len(services))
}

func lastGitUpdate(sumiDir string) string {
	out, err := exec.Command("git", "-C", sumiDir, "log", "-1", "--format=%cr (%h)").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func shortenHome(path string) string {
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
