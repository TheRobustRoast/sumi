package mode

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToggleFocus toggles focus/DND mode. Returns the new state (true = on).
func ToggleFocus(home string) (bool, error) {
	cacheDir := filepath.Join(home, ".cache/sumi")
	os.MkdirAll(cacheDir, 0o755) //nolint:errcheck

	if IsFocus(home) {
		return false, disableFocus(home)
	}
	return true, enableFocus(home)
}

// IsFocus returns true if focus mode is active.
func IsFocus(home string) bool {
	data, err := os.ReadFile(filepath.Join(home, ".cache/sumi/focus-mode"))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "on"
}

func enableFocus(home string) error {
	cacheDir := filepath.Join(home, ".cache/sumi")
	os.WriteFile(filepath.Join(cacheDir, "focus-mode"), []byte("on\n"), 0o644) //nolint:errcheck

	// Notify before pausing
	exec.Command("notify-send", "-a", "sumi", "-u", "low", "-t", "1500",
		"[ FOCUS MODE ON ]", "notifications paused · waybar hidden").Run() //nolint:errcheck

	// Wait for dunst to render
	time.Sleep(1 * time.Second)

	// Pause notifications (DND)
	exec.Command("dunstctl", "set-paused", "true").Run() //nolint:errcheck

	// Hide waybar
	hideWaybar(home)

	return nil
}

func disableFocus(home string) error {
	os.Remove(filepath.Join(home, ".cache/sumi/focus-mode")) //nolint:errcheck

	// Unpause notifications
	exec.Command("dunstctl", "set-paused", "false").Run() //nolint:errcheck

	// Show waybar
	showWaybar(home)

	exec.Command("notify-send", "-a", "sumi", "-u", "low", "-t", "2000",
		"[ FOCUS MODE OFF ]", "notifications on · waybar visible").Run() //nolint:errcheck

	return nil
}
