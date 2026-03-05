package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// isUserQuit returns true if the error represents a user-initiated exit
// (Ctrl+C, Esc, or q in a TUI). These should not be printed as errors.
func isUserQuit(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, huh.ErrUserAborted) ||
		errors.Is(err, tea.ErrProgramKilled)
}

// silenceQuit returns nil if the error is a user-initiated quit, otherwise
// returns the original error. Use this to wrap RunE returns so that
// Ctrl+C / Esc exits cleanly without printing "Error: ...".
func silenceQuit(err error) error {
	if isUserQuit(err) {
		return nil
	}
	return err
}

// sumiRoot returns the directory containing the sumi repo by walking up from
// the executable's location.
func sumiRoot() (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(self)
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		if sd := os.Getenv("SUMI_DIR"); sd != "" {
			return sd, nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot find sumi repo root: set SUMI_DIR env var")
		}
		return cwd, nil
	}
	return dir, nil
}

// requireTool checks that a command-line tool exists in PATH and returns an
// actionable error with install instructions if it's missing.
func requireTool(name, installHint string) error {
	if _, err := exec.LookPath(name); err != nil {
		if installHint != "" {
			return fmt.Errorf("%s not found — install with: %s", name, installHint)
		}
		return fmt.Errorf("%s not found in PATH", name)
	}
	return nil
}

func notImplemented(_ *cobra.Command, _ []string) error {
	return fmt.Errorf("not implemented yet")
}
