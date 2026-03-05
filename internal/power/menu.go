package power

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/huh"
)

// RunPowerMenu shows an interactive power menu.
func RunPowerMenu() error {
	var action string
	err := huh.NewSelect[string]().
		Title("sumi :: power").
		Options(
			huh.NewOption("Lock screen", "lock"),
			huh.NewOption("Suspend (sleep)", "suspend"),
			huh.NewOption("Reboot", "reboot"),
			huh.NewOption("Shutdown", "shutdown"),
			huh.NewOption("Logout (exit Hyprland)", "logout"),
		).
		Value(&action).
		Run()
	if err != nil {
		return nil // user cancelled
	}

	switch action {
	case "lock":
		if err := exec.Command("hyprlock").Run(); err != nil {
			return fmt.Errorf("hyprlock: %w — install with: pacman -S hyprlock", err)
		}
	case "suspend":
		if err := exec.Command("systemctl", "suspend").Run(); err != nil {
			return fmt.Errorf("suspend: %w", err)
		}
	case "reboot":
		if !confirm("Reboot the system?") {
			return nil
		}
		if err := exec.Command("systemctl", "reboot").Run(); err != nil {
			return fmt.Errorf("reboot: %w", err)
		}
	case "shutdown":
		if !confirm("Shut down the system?") {
			return nil
		}
		if err := exec.Command("systemctl", "poweroff").Run(); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	case "logout":
		if !confirm("Exit Hyprland?") {
			return nil
		}
		if err := exec.Command("hyprctl", "dispatch", "exit").Run(); err != nil {
			return fmt.Errorf("hyprctl exit: %w", err)
		}
	}
	return nil
}

func confirm(msg string) bool {
	var yes bool
	err := huh.NewConfirm().
		Title(msg).
		Affirmative("Yes").
		Negative("No").
		Value(&yes).
		Run()
	if err != nil {
		return false
	}
	fmt.Println()
	return yes
}
