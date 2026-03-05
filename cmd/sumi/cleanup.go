package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"sumi/internal/config"
	"sumi/internal/theme"
)

func cleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Trim clipboard history, old screenshots/recordings, and temp files",
		Long:  "Runs cleanup tasks based on config values. Safe to run manually or via systemd timer.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			home := os.Getenv("HOME")
			fmt.Println(theme.Step("sumi cleanup starting..."))

			trimClipboard(cfg.Cleanup.MaxClipboard)
			cleanOldFiles(filepath.Join(home, "Pictures/Screenshots"), "screenshot_*", cfg.Cleanup.ScreenshotMaxDays, "screenshots")
			cleanOldFiles(filepath.Join(home, "Videos/Recordings"), "rec_*", cfg.Cleanup.RecordingMaxDays, "recordings")
			cleanTempFiles(home)
			cleanStalePIDs()
			pruneSessionCache(home)
			trimZshHistory(home)

			fmt.Println(theme.Ok("cleanup done"))
			return nil
		},
	}
}

func trimClipboard(max int) {
	if _, err := exec.LookPath("cliphist"); err != nil {
		return
	}
	out, err := exec.Command("cliphist", "list").Output()
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := len(lines)
	if count == 0 || (count == 1 && lines[0] == "") {
		count = 0
	}

	if count > max {
		excess := count - max
		// Delete oldest entries (tail of the list)
		for _, line := range lines[max:] {
			cmd := exec.Command("cliphist", "delete")
			cmd.Stdin = strings.NewReader(line)
			cmd.Run() //nolint:errcheck
		}
		fmt.Println(theme.Ok(fmt.Sprintf("trimmed %d old clipboard entries (kept %d)", excess, max)))
	} else {
		fmt.Println(theme.Ok(fmt.Sprintf("clipboard OK (%d entries)", count)))
	}
}

func cleanOldFiles(dir, pattern string, maxDays int, label string) {
	if maxDays <= 0 {
		return
	}
	if _, err := os.Stat(dir); err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -maxDays)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	deleted := 0
	for _, e := range entries {
		if e.IsDir() || e.Type()&os.ModeSymlink != 0 {
			continue // skip directories and symlinks
		}
		matched, _ := filepath.Match(pattern, e.Name())
		if !matched {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, e.Name())) //nolint:errcheck
			deleted++
		}
	}
	if deleted > 0 {
		fmt.Println(theme.Ok(fmt.Sprintf("deleted %d %s older than %dd", deleted, label, maxDays)))
	} else {
		fmt.Println(theme.Ok(fmt.Sprintf("%s OK (nothing older than %dd)", label, maxDays)))
	}
}

func cleanTempFiles(home string) {
	// Temp screenshot cache
	os.Remove("/tmp/sumi-screenshot-last.png") //nolint:errcheck

	// Orphaned wallust temp files
	tmpEntries, _ := os.ReadDir("/tmp")
	for _, e := range tmpEntries {
		if strings.HasPrefix(e.Name(), "wallust") {
			info, err := e.Info()
			if err != nil {
				continue
			}
			if time.Since(info.ModTime()) > 24*time.Hour {
				os.Remove(filepath.Join("/tmp", e.Name())) //nolint:errcheck
			}
		}
	}
}

func cleanStalePIDs() {
	entries, _ := filepath.Glob("/tmp/sumi-*.pid")
	for _, pidfile := range entries {
		data, err := os.ReadFile(pidfile)
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		// Check if process is alive
		proc, err := os.FindProcess(pid)
		if err != nil {
			os.Remove(pidfile) //nolint:errcheck
			fmt.Println(theme.Ok("cleaned stale PID file: " + filepath.Base(pidfile)))
			continue
		}
		// On Unix, FindProcess always succeeds; check with signal 0
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			os.Remove(pidfile) //nolint:errcheck
			fmt.Println(theme.Ok("cleaned stale PID file: " + filepath.Base(pidfile)))
		}
	}
}

func pruneSessionCache(home string) {
	sessDir := filepath.Join(home, ".cache/sumi/sessions")
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -30)
	deleted := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(sessDir, e.Name())) //nolint:errcheck
			deleted++
		}
	}
	if deleted > 0 {
		fmt.Println(theme.Ok(fmt.Sprintf("pruned %d old session saves (>30d)", deleted)))
	}
}

func trimZshHistory(home string) {
	histfile := filepath.Join(home, ".zsh_history")
	data, err := os.ReadFile(histfile)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 50000 {
		trimmed := lines[len(lines)-30000:]
		os.WriteFile(histfile, []byte(strings.Join(trimmed, "\n")), 0o600) //nolint:errcheck
		fmt.Println(theme.Ok(fmt.Sprintf("trimmed zsh history from %d to 30000 lines", len(lines))))
	}
}
