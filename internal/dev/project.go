package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Project markers: files that indicate a directory is a project.
var projectMarkers = []string{
	".git", "go.mod", "Cargo.toml", "package.json",
	"pyproject.toml", "Makefile", "CMakeLists.txt",
	"flake.nix", ".envrc",
}

// ListProjects returns project directories ranked by zoxide frecency.
func ListProjects() ([]string, error) {
	home := os.Getenv("HOME")
	seen := make(map[string]bool)
	var projects []string

	// Query zoxide for frecency-ranked directories
	out, err := exec.Command("zoxide", "query", "--list", "--score").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			// Format: "score /path/to/dir"
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			dir := parts[len(parts)-1]
			if isProject(dir) && !seen[dir] {
				projects = append(projects, dir)
				seen[dir] = true
			}
		}
	}

	// Scan common project directories as fallback
	for _, scanDir := range []string{
		filepath.Join(home, "Projects"),
		filepath.Join(home, "Dev"),
		filepath.Join(home, "src"),
		filepath.Join(home, "repos"),
		filepath.Join(home, "Code"),
	} {
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(scanDir, e.Name())
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil && !seen[dir] {
				projects = append(projects, dir)
				seen[dir] = true
			}
		}
	}

	return projects, nil
}

// sanitizeTmuxName strips characters that could be interpreted by tmux or shells.
var tmuxUnsafe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func sanitizeTmuxName(s string) string {
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = tmuxUnsafe.ReplaceAllString(s, "")
	if s == "" {
		s = "project"
	}
	return s
}

// OpenInTmux opens a project directory in a tmux session.
func OpenInTmux(path string) error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux not found — install with: pacman -S tmux")
	}
	name := sanitizeTmuxName(filepath.Base(path))

	// Check if session exists
	if exec.Command("tmux", "has-session", "-t", name).Run() == nil {
		// Attach to existing
		if os.Getenv("TMUX") != "" {
			return exec.Command("tmux", "switch-client", "-t", name).Run()
		}
		cmd := exec.Command("tmux", "attach", "-t", name)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Create new session with dev layout
	if os.Getenv("TMUX") != "" {
		exec.Command("tmux", "new-session", "-d", "-s", name, "-c", path, "-n", "edit").Run() //nolint:errcheck
		exec.Command("tmux", "send-keys", "-t", name+":edit", "nvim .", "Enter").Run()        //nolint:errcheck
		exec.Command("tmux", "new-window", "-t", name, "-n", "shell", "-c", path).Run()       //nolint:errcheck
		exec.Command("tmux", "new-window", "-t", name, "-n", "run", "-c", path).Run()         //nolint:errcheck
		exec.Command("tmux", "select-window", "-t", name+":edit").Run()                       //nolint:errcheck
		return exec.Command("tmux", "switch-client", "-t", name).Run()
	}

	cmd := exec.Command("tmux", "new-session", "-s", name, "-c", path, "-n", "edit")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isProject(dir string) bool {
	for _, marker := range projectMarkers {
		p := filepath.Join(dir, marker)
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
