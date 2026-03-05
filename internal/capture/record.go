package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const pidFile = "/tmp/sumi-recording.pid"

// RecordOpts configures recording behavior.
type RecordOpts struct {
	Area bool
	GIF  bool
	Home string
}

func recordDir(home string) string {
	return filepath.Join(home, "Videos/Recordings")
}

// StartRecording starts wf-recorder. If already recording, stops it.
func StartRecording(opts RecordOpts) error {
	if _, err := exec.LookPath("wf-recorder"); err != nil {
		return fmt.Errorf("wf-recorder not found — install with: pacman -S wf-recorder")
	}
	home := homeDir(opts.Home)
	dir := recordDir(home)
	os.MkdirAll(dir, 0o755) //nolint:errcheck

	// Clean stale PID
	cleanStalePID()

	// If already recording, stop
	if IsRecording() {
		return StopRecording()
	}

	ts := time.Now().Format("20060102_150405")
	var args []string
	var ext string

	if opts.GIF {
		ext = ".gif"
		args = append(args, "-c", "gif")
	} else {
		ext = ".mp4"
	}

	var file string
	if opts.Area {
		slurpArgs := []string{"-d", "-c", "#7aa2f7", "-b", "#0a0a0a80", "-s", "#7aa2f720", "-w", "2"}
		geom, err := exec.Command("slurp", slurpArgs...).Output()
		if err != nil {
			return nil // user cancelled
		}
		g := strings.TrimSpace(string(geom))
		args = append(args, "-g", g)
		file = filepath.Join(dir, fmt.Sprintf("rec_area_%s%s", ts, ext))
	} else {
		// Detect primary monitor dynamically
		if mon := primaryMonitor(); mon != "" {
			args = append(args, "-o", mon)
		}
		file = filepath.Join(dir, fmt.Sprintf("rec_full_%s%s", ts, ext))
	}

	args = append(args, "-f", file)

	cmd := exec.Command("wf-recorder", args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("wf-recorder: %w", err)
	}

	// Atomic PID file: O_CREATE|O_EXCL ensures no race condition
	pf, err := os.OpenFile(pidFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		cmd.Process.Kill() //nolint:errcheck
		return fmt.Errorf("create pid file: %w", err)
	}
	if _, err := fmt.Fprint(pf, cmd.Process.Pid); err != nil {
		pf.Close()
		cmd.Process.Kill() //nolint:errcheck
		os.Remove(pidFile)  //nolint:errcheck
		return fmt.Errorf("write pid file: %w", err)
	}
	if err := pf.Close(); err != nil {
		cmd.Process.Kill() //nolint:errcheck
		os.Remove(pidFile)  //nolint:errcheck
		return fmt.Errorf("close pid file: %w", err)
	}

	// Verify it started
	time.Sleep(500 * time.Millisecond)
	if !IsRecording() {
		os.Remove(pidFile) //nolint:errcheck
		return fmt.Errorf("wf-recorder exited immediately")
	}

	label := "area"
	if !opts.Area {
		label = "fullscreen"
	}
	exec.Command("notify-send", "-a", "sumi", "-t", "3000",
		"● Recording "+label, "SUPER+ALT+R to stop").Run() //nolint:errcheck

	return nil
}

// StopRecording stops an active recording.
func StopRecording() error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return nil
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(pidFile) //nolint:errcheck
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(pidFile) //nolint:errcheck
		return nil
	}

	// Send SIGINT for graceful shutdown
	proc.Signal(syscall.SIGINT) //nolint:errcheck

	// Wait for exit (up to 3s)
	for i := 0; i < 30; i++ {
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	os.Remove(pidFile) //nolint:errcheck

	exec.Command("notify-send", "-a", "sumi", "-t", "3000",
		"Recording stopped", "Saved to ~/Videos/Recordings/").Run() //nolint:errcheck

	return nil
}

// IsRecording returns true if a recording is in progress.
func IsRecording() bool {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// RecordingStatus returns waybar JSON for the recording indicator.
func RecordingStatus() (string, string) {
	if IsRecording() {
		return "● REC", "recording"
	}
	return "", ""
}

// primaryMonitor returns the focused monitor name from Hyprland, or empty string.
func primaryMonitor() string {
	out, err := exec.Command("hyprctl", "activeworkspace", "-j").Output()
	if err != nil {
		return ""
	}
	// Quick parse: find "monitor" field
	s := string(out)
	key := `"monitor"`
	idx := strings.Index(s, key)
	if idx < 0 {
		return ""
	}
	rest := s[idx+len(key):]
	q1 := strings.Index(rest, `"`)
	if q1 < 0 {
		return ""
	}
	q2 := strings.Index(rest[q1+1:], `"`)
	if q2 < 0 {
		return ""
	}
	return rest[q1+1 : q1+1+q2]
}

func cleanStalePID() {
	if _, err := os.Stat(pidFile); err != nil {
		return
	}
	if !IsRecording() {
		os.Remove(pidFile) //nolint:errcheck
	}
}
