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

// WindowState stores a single window's layout.
type WindowState struct {
	Class     string `json:"class"`
	Title     string `json:"title"`
	Workspace int    `json:"workspace"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	W         int    `json:"w"`
	H         int    `json:"h"`
	Floating  bool   `json:"floating"`
}

func sessDir(home string) string {
	return filepath.Join(home, ".cache/sumi/sessions")
}

// safeName validates a user-supplied name for use in file paths.
// Rejects path traversal, absolute paths, and shell metacharacters.
func safeName(name string) error {
	if name == "" {
		return nil
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") ||
		strings.Contains(name, "\\") || strings.Contains(name, "\x00") ||
		strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid name: %q", name)
	}
	return nil
}

// Save captures the current window layout.
func Save(home, name string) error {
	if name == "" {
		name = time.Now().Format("20060102_150405")
	}
	if err := safeName(name); err != nil {
		return err
	}
	dir := sessDir(home)
	os.MkdirAll(dir, 0o755) //nolint:errcheck

	out, err := exec.Command("hyprctl", "clients", "-j").Output()
	if err != nil {
		return fmt.Errorf("hyprctl clients: %w", err)
	}

	file := filepath.Join(dir, name+".json")
	if err := os.WriteFile(file, out, 0o600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}

	// Count windows
	var clients []json.RawMessage
	json.Unmarshal(out, &clients) //nolint:errcheck
	exec.Command("notify-send", "-a", "sumi", "Session saved",
		fmt.Sprintf("%s (%d windows)", name, len(clients))).Run() //nolint:errcheck
	return nil
}

// Restore recreates a saved window layout.
func Restore(home, name string) error {
	if err := safeName(name); err != nil {
		return err
	}
	file := filepath.Join(sessDir(home), name+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read session: %w", err)
	}

	var clients []struct {
		Class     string `json:"class"`
		Workspace struct {
			ID int `json:"id"`
		} `json:"workspace"`
		At       [2]int `json:"at"`
		Size     [2]int `json:"size"`
		Floating bool   `json:"floating"`
		Address  string `json:"address"`
	}
	if err := json.Unmarshal(data, &clients); err != nil {
		return fmt.Errorf("parse session: %w", err)
	}

	restored := 0
	total := 0
	for _, c := range clients {
		if c.Class == "" {
			continue
		}
		total++

		// Find running instance by class
		out, _ := exec.Command("hyprctl", "clients", "-j").Output()
		var running []struct {
			Class   string `json:"class"`
			Address string `json:"address"`
		}
		json.Unmarshal(out, &running) //nolint:errcheck

		var addr string
		for _, r := range running {
			if r.Class == c.Class {
				addr = r.Address
				break
			}
		}
		if addr == "" {
			continue
		}

		// Move to workspace
		exec.Command("hyprctl", "dispatch", "movetoworkspacesilent",
			fmt.Sprintf("%d,address:%s", c.Workspace.ID, addr)).Run() //nolint:errcheck

		if c.Floating {
			exec.Command("hyprctl", "dispatch", "focuswindow", "address:"+addr).Run() //nolint:errcheck
			exec.Command("hyprctl", "dispatch", "setfloating", "address:"+addr).Run() //nolint:errcheck
			exec.Command("hyprctl", "dispatch", "moveactive",
				fmt.Sprintf("exact %d %d", c.At[0], c.At[1])).Run() //nolint:errcheck
			exec.Command("hyprctl", "dispatch", "resizeactive",
				fmt.Sprintf("exact %d %d", c.Size[0], c.Size[1])).Run() //nolint:errcheck
		}
		restored++
	}

	exec.Command("notify-send", "-a", "sumi", "Session restored",
		fmt.Sprintf("%s — %d/%d windows", name, restored, total)).Run() //nolint:errcheck
	return nil
}

// List returns available session names.
func List(home string) ([]string, error) {
	dir := sessDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

// Delete removes a saved session.
func Delete(home, name string) error {
	if err := safeName(name); err != nil {
		return err
	}
	file := filepath.Join(sessDir(home), name+".json")
	if err := os.Remove(file); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session %q not found", name)
		}
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
