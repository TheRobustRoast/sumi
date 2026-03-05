package mode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ToggleGaming toggles gaming mode. Returns the new state (true = on).
func ToggleGaming(home string) (bool, error) {
	if _, err := exec.LookPath("hyprctl"); err != nil {
		return false, fmt.Errorf("hyprctl not found — gaming mode requires Hyprland")
	}
	cacheDir := filepath.Join(home, ".cache/sumi")
	os.MkdirAll(cacheDir, 0o755) //nolint:errcheck

	if IsGaming(home) {
		return false, disableGaming(home)
	}
	return true, enableGaming(home)
}

// IsGaming returns true if gaming mode is active.
func IsGaming(home string) bool {
	data, err := os.ReadFile(filepath.Join(home, ".cache/sumi/gaming-mode"))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "on"
}

func enableGaming(home string) error {
	cacheDir := filepath.Join(home, ".cache/sumi")
	os.WriteFile(filepath.Join(cacheDir, "gaming-mode"), []byte("on\n"), 0o644) //nolint:errcheck

	// Kill visual overhead
	hyprSet("decoration:blur:enabled", "false")
	hyprSet("animations:enabled", "false")
	hyprSet("decoration:shadow:enabled", "false")
	hyprSet("general:gaps_in", "0")
	hyprSet("general:gaps_out", "0")
	hyprSet("general:border_size", "1")
	hyprSet("misc:vfr", "false")

	// Hide waybar
	hideWaybar(home)

	// Suppress notifications
	exec.Command("dunstctl", "set-paused", "true").Run() //nolint:errcheck

	// Performance power profile
	exec.Command("powerprofilesctl", "set", "performance").Run() //nolint:errcheck

	exec.Command("notify-send", "-a", "sumi", "-u", "low", "-t", "2000",
		"[ GAMING MODE ON ]", "blur off · anim off · perf mode · notifs paused").Run() //nolint:errcheck

	return nil
}

func disableGaming(home string) error {
	os.Remove(filepath.Join(home, ".cache/sumi/gaming-mode")) //nolint:errcheck

	// Restore visual settings
	hyprSet("decoration:blur:enabled", "true")
	hyprSet("animations:enabled", "true")
	hyprSet("decoration:shadow:enabled", "true")
	hyprSet("general:gaps_in", "3")
	hyprSet("general:gaps_out", "6")
	hyprSet("general:border_size", "2")
	hyprSet("misc:vfr", "true")

	// Show waybar
	showWaybar(home)

	// Unpause notifications
	exec.Command("dunstctl", "set-paused", "false").Run() //nolint:errcheck

	// Balanced power profile
	exec.Command("powerprofilesctl", "set", "balanced").Run() //nolint:errcheck

	exec.Command("notify-send", "-a", "sumi", "-u", "low", "-t", "2000",
		"[ GAMING MODE OFF ]", "blur on · anim on · balanced mode · notifs on").Run() //nolint:errcheck

	return nil
}

func hyprSet(key, value string) {
	exec.Command("hyprctl", "keyword", key, value).Run() //nolint:errcheck
}
