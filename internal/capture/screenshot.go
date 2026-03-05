package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CaptureOpts configures screenshot behavior.
type CaptureOpts struct {
	Clipboard bool   // copy to clipboard instead of file
	Notify    bool   // send notification
	Home      string // override $HOME
}

func screenshotDir(home string) string {
	return filepath.Join(home, "Pictures/Screenshots")
}

func timestamp() string {
	return time.Now().Format("20060102_150405")
}

func requireCaptureTool(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s not found — install with: pacman -S %s", name, name)
	}
	return nil
}

// CaptureScreen takes a fullscreen screenshot.
func CaptureScreen(opts CaptureOpts) (string, error) {
	if err := requireCaptureTool("grim"); err != nil {
		return "", err
	}
	home := homeDir(opts.Home)
	if opts.Clipboard {
		tmp := "/tmp/sumi-screenshot-last.png"
		if err := exec.Command("grim", tmp).Run(); err != nil {
			return "", fmt.Errorf("grim: %w", err)
		}
		copyToClipboard(tmp)
		notifyScreenshot(tmp, "fullscreen → clipboard", opts.Notify)
		return tmp, nil
	}

	dir := screenshotDir(home)
	os.MkdirAll(dir, 0o755) //nolint:errcheck
	file := filepath.Join(dir, "screenshot_"+timestamp()+".png")
	if err := exec.Command("grim", file).Run(); err != nil {
		return "", fmt.Errorf("grim: %w", err)
	}
	copyToClipboard(file)
	notifyScreenshot(file, "fullscreen saved + copied", opts.Notify)
	return file, nil
}

// CaptureArea takes an area selection screenshot.
func CaptureArea(opts CaptureOpts) (string, error) {
	if err := requireCaptureTool("grim"); err != nil {
		return "", err
	}
	if err := requireCaptureTool("slurp"); err != nil {
		return "", err
	}
	home := homeDir(opts.Home)
	slurpArgs := []string{"-d", "-b", "0a0a0a88", "-c", "7aa2f7", "-s", "1a1a1a44", "-w", "2"}
	geom, err := exec.Command("slurp", slurpArgs...).Output()
	if err != nil {
		return "", nil // user cancelled
	}
	g := strings.TrimSpace(string(geom))

	if opts.Clipboard {
		tmp := "/tmp/sumi-screenshot-last.png"
		if err := exec.Command("grim", "-g", g, tmp).Run(); err != nil {
			return "", fmt.Errorf("grim: %w", err)
		}
		copyToClipboard(tmp)
		notifyScreenshot(tmp, "area → clipboard", opts.Notify)
		return tmp, nil
	}

	dir := screenshotDir(home)
	os.MkdirAll(dir, 0o755) //nolint:errcheck
	file := filepath.Join(dir, "screenshot_"+timestamp()+".png")
	if err := exec.Command("grim", "-g", g, file).Run(); err != nil {
		return "", fmt.Errorf("grim: %w", err)
	}
	copyToClipboard(file)
	notifyScreenshot(file, "saved + copied", opts.Notify)
	return file, nil
}

// CaptureWindow takes a screenshot of the active window.
func CaptureWindow(opts CaptureOpts) (string, error) {
	if err := requireCaptureTool("grim"); err != nil {
		return "", err
	}
	home := homeDir(opts.Home)
	out, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil || strings.TrimSpace(string(out)) == "null" {
		return "", fmt.Errorf("no active window")
	}

	var win struct {
		At   [2]int `json:"at"`
		Size [2]int `json:"size"`
	}
	if err := json.Unmarshal(out, &win); err != nil {
		return "", fmt.Errorf("parse window: %w", err)
	}

	geom := fmt.Sprintf("%d,%d %dx%d", win.At[0], win.At[1], win.Size[0], win.Size[1])

	if opts.Clipboard {
		tmp := "/tmp/sumi-screenshot-last.png"
		if err := exec.Command("grim", "-g", geom, tmp).Run(); err != nil {
			return "", fmt.Errorf("grim: %w", err)
		}
		copyToClipboard(tmp)
		notifyScreenshot(tmp, "window → clipboard", opts.Notify)
		return tmp, nil
	}

	dir := screenshotDir(home)
	os.MkdirAll(dir, 0o755) //nolint:errcheck
	file := filepath.Join(dir, "screenshot_window_"+timestamp()+".png")
	if err := exec.Command("grim", "-g", geom, file).Run(); err != nil {
		return "", fmt.Errorf("grim: %w", err)
	}
	copyToClipboard(file)
	notifyScreenshot(file, "window saved + copied", opts.Notify)
	return file, nil
}

func copyToClipboard(file string) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	cmd := exec.Command("wl-copy")
	cmd.Stdin = f
	cmd.Run() //nolint:errcheck
}

func notifyScreenshot(file, desc string, enabled bool) {
	if !enabled {
		enabled = true // always notify for screenshots
	}
	basename := filepath.Base(file)
	exec.Command("notify-send", "-a", "sumi", "-i", file, "-t", "3000",
		"[ screenshot ]", desc+"\n"+basename).Run() //nolint:errcheck
}

func homeDir(home string) string {
	if home != "" {
		return home
	}
	return os.Getenv("HOME")
}
