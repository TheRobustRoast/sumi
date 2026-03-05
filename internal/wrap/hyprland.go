package wrap

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CrashEntry records a single Hyprland crash.
type CrashEntry struct {
	Timestamp time.Time `json:"timestamp"`
	ExitCode  int       `json:"exit_code"`
}

// RunWithRecovery runs Hyprland with crash recovery.
// Restarts up to maxCrashes times within the given time window.
func RunWithRecovery(maxCrashes int, window time.Duration) error {
	home := os.Getenv("HOME")
	logDir := filepath.Join(home, ".cache/sumi")
	logFile := filepath.Join(logDir, "crash.log")
	crashLog := filepath.Join(logDir, "crash-log.json")

	os.MkdirAll(logDir, 0o755) //nolint:errcheck

	restartCount := 0
	var lastStart time.Time

	for {
		now := time.Now()

		// Reset counter if stable for long enough
		if !lastStart.IsZero() && now.Sub(lastStart) > window {
			restartCount = 0
		}

		if restartCount >= maxCrashes {
			appendLog(logFile, fmt.Sprintf("[%s] Hyprland crashed %d times in %s — giving up",
				now.Format(time.RFC3339), maxCrashes, window))
			fmt.Fprintf(os.Stderr, "Hyprland crashed repeatedly. Check %s\n", logFile)
			fmt.Fprintln(os.Stderr, "Press Enter to try again, or Ctrl+C to drop to TTY.")
			fmt.Scanln()
			restartCount = 0
		}

		lastStart = time.Now()
		restartCount++

		appendLog(logFile, fmt.Sprintf("[%s] Starting Hyprland (attempt %d/%d)",
			lastStart.Format(time.RFC3339), restartCount, maxCrashes))

		// Launch Hyprland
		cmd := exec.Command("Hyprland")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err := cmd.Run()

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		// Clean exit = user logged out
		if exitCode == 0 {
			appendLog(logFile, fmt.Sprintf("[%s] Hyprland exited cleanly (code 0)",
				time.Now().Format(time.RFC3339)))
			break
		}

		// Log crash
		appendLog(logFile, fmt.Sprintf("[%s] Hyprland crashed with exit code %d",
			time.Now().Format(time.RFC3339), exitCode))

		// Record crash entry
		recordCrash(crashLog, CrashEntry{
			Timestamp: time.Now(),
			ExitCode:  exitCode,
		})

		// Capture journal entries for debug
		out, _ := exec.Command("journalctl", "--user", "-b", "-n", "30", "--no-pager").Output()
		if len(out) > 0 {
			appendLog(logFile, string(out))
			appendLog(logFile, "---")
		}

		// Trim log if too large
		trimLog(logFile, 500)

		// Send notification on restart
		exec.Command("notify-send", "-a", "sumi", "-u", "critical", "-t", "5000",
			"Hyprland crashed", fmt.Sprintf("Restarting... (attempt %d/%d)", restartCount+1, maxCrashes)).Run() //nolint:errcheck

		// Brief delay before restart
		time.Sleep(1 * time.Second)
	}

	return nil
}

func appendLog(path, msg string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, msg)
}

func recordCrash(path string, entry CrashEntry) {
	var entries []CrashEntry
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &entries) //nolint:errcheck
	}
	entries = append(entries, entry)
	// Keep last 50 entries
	if len(entries) > 50 {
		entries = entries[len(entries)-50:]
	}
	out, _ := json.MarshalIndent(entries, "", "  ")
	os.WriteFile(path, out, 0o644) //nolint:errcheck
}

func trimLog(path string, maxLines int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := splitLines(string(data))
	if len(lines) > maxLines*2 {
		trimmed := lines[len(lines)-maxLines:]
		os.WriteFile(path, []byte(joinLines(trimmed)), 0o644) //nolint:errcheck
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	return result
}
