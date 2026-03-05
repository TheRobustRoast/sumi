package power

import (
	"fmt"
	"os/exec"
	"strings"
)

var profiles = []string{"power-saver", "balanced", "performance"}

// CurrentProfile returns the active power profile.
func CurrentProfile() string {
	out, err := exec.Command("powerprofilesctl", "get").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// SetProfile sets the power profile by name.
func SetProfile(name string) error {
	if _, err := exec.LookPath("powerprofilesctl"); err != nil {
		return fmt.Errorf("powerprofilesctl not found — install with: pacman -S power-profiles-daemon")
	}
	valid := false
	for _, p := range profiles {
		if name == p {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid profile %q — must be one of: %s", name, strings.Join(profiles, ", "))
	}
	if err := exec.Command("powerprofilesctl", "set", name).Run(); err != nil {
		return fmt.Errorf("powerprofilesctl set %s: %w", name, err)
	}
	return nil
}

// CycleProfile cycles through power-saver → balanced → performance.
// Returns the new profile name.
func CycleProfile() (string, error) {
	current := CurrentProfile()
	var next string
	switch current {
	case "power-saver":
		next = "balanced"
	case "balanced":
		next = "performance"
	default:
		next = "power-saver"
	}
	if err := SetProfile(next); err != nil {
		return "", err
	}
	return next, nil
}

// AvailableProfiles returns the list of supported profiles.
func AvailableProfiles() []string {
	return profiles
}

// ProfileIcon returns a short label for waybar display.
func ProfileIcon(profile string) string {
	switch profile {
	case "power-saver":
		return "eco"
	case "balanced":
		return "bal"
	case "performance":
		return "prf"
	default:
		return "?"
	}
}
