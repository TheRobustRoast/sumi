package ui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"

	"sumi/internal/config"
	"sumi/internal/theme"
)

// RunConfigEditor shows the interactive configuration TUI.
func RunConfigEditor() error {
	cfg := config.Load()

	chargeStr := strconv.Itoa(cfg.Battery.ChargeLimit)
	clipStr := strconv.Itoa(cfg.Cleanup.MaxClipboard)
	ssStr := strconv.Itoa(cfg.Cleanup.ScreenshotMaxDays)
	recStr := strconv.Itoa(cfg.Cleanup.RecordingMaxDays)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Theme mode").
				Description("Controls the base palette (accent always comes from wallpaper)").
				Options(
					huh.NewOption("Dark", "dark"),
					huh.NewOption("Light", "light"),
				).
				Value(&cfg.Theme.Mode),
		).Title("Theme"),

		huh.NewGroup(
			huh.NewInput().
				Title("Charge limit (%)").
				Description("Max battery charge percentage (0 = disabled)").
				Value(&chargeStr).
				Validate(validateInt(1, 100)),
		).Title("Battery"),

		huh.NewGroup(
			huh.NewInput().
				Title("Max clipboard entries").
				Value(&clipStr).
				Validate(validateInt(1, 10000)),
			huh.NewInput().
				Title("Screenshot max age (days, 0 = forever)").
				Value(&ssStr).
				Validate(validateInt(0, 365)),
			huh.NewInput().
				Title("Recording max age (days, 0 = forever)").
				Value(&recStr).
				Validate(validateInt(0, 365)),
		).Title("Cleanup"),

		huh.NewGroup(
			huh.NewInput().
				Title("Hardware profile override").
				Description("Leave empty for auto-detection").
				Value(&cfg.Hardware.Profile),
		).Title("Hardware"),
	)

	err := form.Run()
	if err != nil {
		return nil // user cancelled
	}

	// Parse string fields back to ints
	cfg.Battery.ChargeLimit, _ = strconv.Atoi(chargeStr)
	cfg.Cleanup.MaxClipboard, _ = strconv.Atoi(clipStr)
	cfg.Cleanup.ScreenshotMaxDays, _ = strconv.Atoi(ssStr)
	cfg.Cleanup.RecordingMaxDays, _ = strconv.Atoi(recStr)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println(theme.Ok("Config saved to " + config.Path()))
	return nil
}

func validateInt(min, max int) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if n < min || n > max {
			return fmt.Errorf("must be between %d and %d", min, max)
		}
		return nil
	}
}
