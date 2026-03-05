package wallpaper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ApplyOpts configures wallpaper application behavior.
type ApplyOpts struct {
	NoReload  bool // skip waybar/dunst restart
	NoWallust bool // skip color extraction
	Home      string
}

// Apply sets the wallpaper at path, runs wallust, and reloads programs.
func Apply(path string, opts ApplyOpts) error {
	home := opts.Home
	if home == "" {
		home = os.Getenv("HOME")
	}
	cacheDir := filepath.Join(home, ".cache/sumi")

	// Validate input
	if path == "" {
		return fmt.Errorf("no wallpaper path provided")
	}

	// Resolve to absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	path = abs

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("wallpaper not found: %s", path)
	}

	// Lockfile — prevent inotify feedback loops from wallust-watcher
	lockfile := filepath.Join(cacheDir, ".wallpaper-applying")
	if _, err := os.Stat(lockfile); err == nil {
		return nil // another apply is in progress
	}

	os.MkdirAll(cacheDir, 0o755) //nolint:errcheck
	if err := os.WriteFile(lockfile, []byte("1"), 0o644); err != nil {
		return fmt.Errorf("create lockfile: %w", err)
	}
	defer os.Remove(lockfile)

	// Save current wallpaper path
	os.WriteFile(filepath.Join(cacheDir, "current-wallpaper"), []byte(path+"\n"), 0o644) //nolint:errcheck

	// Set wallpaper via hyprpaper IPC
	if err := setHyprpaper(path, home); err != nil {
		fmt.Fprintf(os.Stderr, "hyprpaper: %v\n", err)
	}

	// Generate colors with wallust
	if !opts.NoWallust {
		if err := runWallust(path); err != nil {
			fmt.Fprintf(os.Stderr, "wallust: %v (keeping existing colors)\n", err)
		}
	}

	// Reload programs
	if !opts.NoReload {
		restartWaybar()
		restartDunst(home)
		reloadCava()
		reloadTmux(home)
		reloadFoot()
	}

	// Notify
	basename := filepath.Base(path)
	exec.Command("notify-send", "-a", "sumi", "-t", "3000", "-i", path,
		"[ wallpaper ]", "applied: "+basename).Run() //nolint:errcheck

	return nil
}

// CurrentWallpaper reads the cached current wallpaper path.
func CurrentWallpaper(home string) string {
	data, err := os.ReadFile(filepath.Join(home, ".cache/sumi/current-wallpaper"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func setHyprpaper(path, home string) error {
	// Try IPC first if hyprpaper is running
	if isProcessRunning("hyprpaper") {
		err1 := exec.Command("hyprctl", "hyprpaper", "preload", path).Run()
		err2 := exec.Command("hyprctl", "hyprpaper", "wallpaper", ","+path).Run()
		if err1 == nil && err2 == nil {
			exec.Command("hyprctl", "hyprpaper", "unload", "unused").Run() //nolint:errcheck
			return nil
		}
	}

	// IPC failed or hyprpaper not running — restart it
	exec.Command("pkill", "hyprpaper").Run() //nolint:errcheck
	time.Sleep(300 * time.Millisecond)

	conf := filepath.Join(home, ".config/hypr/hyprpaper.conf")
	content := fmt.Sprintf("preload = %s\nwallpaper = ,%s\nsplash = false\nipc = on\n", path, path)
	os.WriteFile(conf, []byte(content), 0o644) //nolint:errcheck

	cmd := exec.Command("hyprpaper")
	cmd.Start() //nolint:errcheck

	// Wait for hyprpaper to be ready (up to 2s)
	for i := 0; i < 20; i++ {
		if isProcessRunning("hyprpaper") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func runWallust(path string) error {
	if _, err := exec.LookPath("wallust"); err != nil {
		return fmt.Errorf("wallust not found — install with: pacman -S wallust")
	}
	if err := exec.Command("wallust", "run", path).Run(); err != nil {
		return fmt.Errorf("wallust run: %w", err)
	}
	return nil
}

func restartWaybar() {
	if isProcessRunning("waybar") {
		exec.Command("pkill", "waybar").Run() //nolint:errcheck
		time.Sleep(300 * time.Millisecond)
	}
	cmd := exec.Command("waybar")
	cmd.Start() //nolint:errcheck
}

func restartDunst(home string) {
	// Check if dunst was paused (DND/focus mode)
	wasPaused := false
	out, _ := exec.Command("dunstctl", "is-paused").Output()
	if strings.TrimSpace(string(out)) == "true" {
		wasPaused = true
	}

	exec.Command("pkill", "dunst").Run() //nolint:errcheck
	time.Sleep(200 * time.Millisecond)
	cmd := exec.Command("dunst")
	cmd.Start() //nolint:errcheck

	// Restore DND state
	if wasPaused {
		time.Sleep(300 * time.Millisecond)
		exec.Command("dunstctl", "set-paused", "true").Run() //nolint:errcheck
	}
}

func reloadCava() {
	if isProcessRunning("cava") {
		exec.Command("pkill", "-USR1", "cava").Run() //nolint:errcheck
	}
}

func reloadTmux(home string) {
	if !isProcessRunning("tmux") {
		return
	}
	tmuxColors := filepath.Join(home, ".config/sumi/tmux-colors.conf")
	if _, err := os.Stat(tmuxColors); err == nil {
		exec.Command("tmux", "source-file", tmuxColors).Run() //nolint:errcheck
	}
}

func reloadFoot() {
	if isProcessRunning("foot") {
		exec.Command("pkill", "-USR1", "foot").Run() //nolint:errcheck
	}
}

func isProcessRunning(name string) bool {
	return exec.Command("pgrep", "-x", name).Run() == nil
}
