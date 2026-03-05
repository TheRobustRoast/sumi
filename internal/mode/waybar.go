package mode

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Shared waybar visibility tracking to prevent toggle drift.

func isWaybarVisible(home string) bool {
	_, err := os.Stat(filepath.Join(home, ".cache/sumi/waybar-hidden"))
	return err != nil // visible if the file does NOT exist
}

func hideWaybar(home string) {
	if isWaybarVisible(home) {
		exec.Command("pkill", "-USR1", "waybar").Run() //nolint:errcheck
		os.WriteFile(filepath.Join(home, ".cache/sumi/waybar-hidden"), []byte("1"), 0o644) //nolint:errcheck
	}
}

func showWaybar(home string) {
	if !isWaybarVisible(home) {
		exec.Command("pkill", "-USR1", "waybar").Run() //nolint:errcheck
		os.Remove(filepath.Join(home, ".cache/sumi/waybar-hidden")) //nolint:errcheck
	}
}
