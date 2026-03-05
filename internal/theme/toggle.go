package theme

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const darkColors = `# sumi colors — DARK mode (base stays fixed, accent from wallust)
$bg       = rgb(0a0a0a)
$fg       = rgb(c8c8c8)
$surface0 = rgb(1a1a1a)
$surface1 = rgb(2a2a2a)
$surface2 = rgb(3a3a3a)
$dim      = rgb(6a6a6a)
$bright   = rgb(e8e8e8)
# Accent — overwritten by wallust
$accent     = rgb(7aa2f7)
$accent_dim = rgb(3d5178)
$warn       = rgb(e0af68)
$ok         = rgb(9ece6a)
$urgent     = rgb(f7768e)
`

const lightColors = `# sumi colors — LIGHT mode (base stays fixed, accent from wallust)
$bg       = rgb(f0f0f0)
$fg       = rgb(1a1a1a)
$surface0 = rgb(e0e0e0)
$surface1 = rgb(d0d0d0)
$surface2 = rgb(c0c0c0)
$dim      = rgb(8a8a8a)
$bright   = rgb(0a0a0a)
# Accent — overwritten by wallust
$accent     = rgb(2e5cb8)
$accent_dim = rgb(6a8fd8)
$warn       = rgb(b08020)
$ok         = rgb(4a8a2a)
$urgent     = rgb(c83040)
`

// SetThemeMode writes the color config for the given mode ("dark" or "light")
// and persists the mode to the state file.
func SetThemeMode(home, mode string) error {
	colorsConf := filepath.Join(home, ".config/hypr/conf.d/colors.conf")
	stateFile := filepath.Join(home, ".cache/sumi/theme-mode")

	var content string
	switch mode {
	case "light":
		content = lightColors
	default:
		content = darkColors
		mode = "dark"
	}

	if err := os.MkdirAll(filepath.Dir(colorsConf), 0o755); err != nil {
		return fmt.Errorf("mkdir colors.conf dir: %w", err)
	}
	if err := os.WriteFile(colorsConf, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write colors.conf: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}
	if err := os.WriteFile(stateFile, []byte(mode+"\n"), 0o644); err != nil {
		return fmt.Errorf("write theme-mode: %w", err)
	}

	return nil
}

// ToggleTheme reads the current mode and switches to the opposite.
// Returns the new mode.
func ToggleTheme(home string) (string, error) {
	current := CurrentMode(home)
	newMode := "light"
	if current == "light" {
		newMode = "dark"
	}
	if err := SetThemeMode(home, newMode); err != nil {
		return "", err
	}
	return newMode, nil
}

// CurrentMode reads the persisted theme mode, defaulting to "dark".
func CurrentMode(home string) string {
	data, err := os.ReadFile(filepath.Join(home, ".cache/sumi/theme-mode"))
	if err != nil {
		return "dark"
	}
	mode := strings.TrimSpace(string(data))
	if mode == "light" {
		return "light"
	}
	return "dark"
}

// ReloadHyprland tells Hyprland to reload its config.
func ReloadHyprland() {
	exec.Command("hyprctl", "reload").Run() //nolint:errcheck
}
