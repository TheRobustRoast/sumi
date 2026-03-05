package session

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Monitor represents a connected display.
type Monitor struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

// DetectMonitors returns the list of connected monitors.
func DetectMonitors() ([]Monitor, error) {
	out, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl monitors: %w", err)
	}
	var monitors []Monitor
	if err := json.Unmarshal(out, &monitors); err != nil {
		return nil, fmt.Errorf("parse monitors: %w", err)
	}
	return monitors, nil
}

// ConfigureMonitors detects external monitors and assigns workspaces.
func ConfigureMonitors(home string) error {
	cacheDir := filepath.Join(home, ".cache/sumi")
	stateFile := filepath.Join(cacheDir, "monitor-state")
	os.MkdirAll(cacheDir, 0o755) //nolint:errcheck

	monitors, err := DetectMonitors()
	if err != nil {
		return err
	}

	currentCount := len(monitors)
	previousCount := readMonitorState(stateFile)

	// Find external monitor (non-eDP)
	var external string
	for _, m := range monitors {
		if m.Name != "eDP-1" {
			external = m.Name
			break
		}
	}

	// Save current state
	os.WriteFile(stateFile, []byte(fmt.Sprintf("%d", currentCount)), 0o644) //nolint:errcheck

	if currentCount == previousCount {
		// No change — just report
		if external != "" {
			notify("Monitors", fmt.Sprintf("Internal: eDP-1 + External: %s", external))
		} else {
			notify("Monitors", "Internal: eDP-1 only")
		}
		return nil
	}

	if currentCount > 1 && external != "" {
		// External connected
		notify("Monitor Connected", external+" — configuring...")

		// Move workspaces 6-10 to external
		for ws := 6; ws <= 10; ws++ {
			exec.Command("hyprctl", "dispatch", "moveworkspacetomonitor",
				fmt.Sprintf("%d %s", ws, external)).Run() //nolint:errcheck
		}

		// Restart waybar for multi-monitor
		restartWaybarForMonitor()

		notify("Monitor Ready", fmt.Sprintf("WS 1-5 → eDP-1 | WS 6-10 → %s", external))
	} else {
		// External disconnected
		notify("Monitor Disconnected", "Moving all workspaces to internal display")

		for ws := 1; ws <= 10; ws++ {
			exec.Command("hyprctl", "dispatch", "moveworkspacetomonitor",
				fmt.Sprintf("%d eDP-1", ws)).Run() //nolint:errcheck
		}

		restartWaybarForMonitor()
		notify("Single Monitor", "All workspaces on eDP-1")
	}

	return nil
}

// MonitorStatus returns a summary of current monitor configuration.
func MonitorStatus() (string, error) {
	monitors, err := DetectMonitors()
	if err != nil {
		return "", err
	}
	var parts []string
	for _, m := range monitors {
		parts = append(parts, fmt.Sprintf("%s (%dx%d)", m.Name, m.Width, m.Height))
	}
	return strings.Join(parts, " + "), nil
}

func readMonitorState(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 1
	}
	s := strings.TrimSpace(string(data))
	if s == "2" {
		return 2
	}
	return 1
}

func restartWaybarForMonitor() {
	exec.Command("pkill", "waybar").Run() //nolint:errcheck
	time.Sleep(300 * time.Millisecond)
	cmd := exec.Command("waybar")
	cmd.Start() //nolint:errcheck
}

func notify(title, body string) {
	exec.Command("notify-send", "-a", "sumi", "-t", "3000", title, body).Run() //nolint:errcheck
}
