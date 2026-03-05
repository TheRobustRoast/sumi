package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"sumi/internal/theme"
	"sumi/internal/wallpaper"
)

// RunThemePicker shows the theme management TUI.
func RunThemePicker(sumiDir string) error {
	home := os.Getenv("HOME")

	var action string
	err := huh.NewSelect[string]().
		Title("sumi :: theme").
		Options(
			huh.NewOption("Set wallpaper", "wallpaper"),
			huh.NewOption("Toggle dark/light", "toggle"),
			huh.NewOption("Show current", "status"),
		).
		Value(&action).
		Run()
	if err != nil {
		return nil // user cancelled
	}

	switch action {
	case "wallpaper":
		return pickWallpaper(home, sumiDir)
	case "toggle":
		return toggleTheme(home, sumiDir)
	case "status":
		showThemeStatus(home)
		return nil
	}
	return nil
}

func pickWallpaper(home, sumiDir string) error {
	wpDir := filepath.Join(home, "Pictures/Wallpapers")
	entries, err := os.ReadDir(wpDir)
	if err != nil {
		return fmt.Errorf("cannot read ~/Pictures/Wallpapers: %w", err)
	}

	var images []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") ||
			strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".webp") {
			images = append(images, e.Name())
		}
	}

	if len(images) == 0 {
		fmt.Println("No images found in ~/Pictures/Wallpapers/")
		fmt.Println("Add .png, .jpg, or .webp files and try again.")
		return nil
	}

	var opts []huh.Option[string]
	for _, img := range images {
		opts = append(opts, huh.NewOption(img, filepath.Join(wpDir, img)))
	}

	var selected string
	err = huh.NewSelect[string]().
		Title("Select wallpaper").
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return nil // user cancelled
	}

	// Apply wallpaper via Go pipeline
	if err := wallpaper.Apply(selected, wallpaper.ApplyOpts{Home: home}); err != nil {
		return fmt.Errorf("wallpaper apply: %w", err)
	}

	fmt.Println(theme.Ok("Wallpaper set: " + filepath.Base(selected)))
	return nil
}

func toggleTheme(home, sumiDir string) error {
	newMode, err := theme.ToggleTheme(home)
	if err != nil {
		return err
	}
	theme.ReloadHyprland()
	fmt.Println(theme.Ok("Theme mode: " + newMode))
	_ = sumiDir
	return nil
}

func showThemeStatus(home string) {
	labelStyle := lipgloss.NewStyle().Foreground(theme.Dim).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(theme.Text)

	wp := readFileOr(filepath.Join(home, ".cache/sumi/current-wallpaper"), "none")
	mode := readFileOr(filepath.Join(home, ".cache/sumi/theme-mode"), "dark")
	accent := readAccent(filepath.Join(home, ".config/hypr/conf.d/colors.conf"))

	row := func(label, value string) string {
		return "  " + labelStyle.Render(label) + valueStyle.Render(value)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		row("Wallpaper", shortenHome(wp)),
		row("Mode", mode),
		row("Accent", accent),
	)
	fmt.Println(theme.BoxStyle.Render(body))
}
