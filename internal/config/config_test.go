package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Battery.ChargeLimit != 80 {
		t.Errorf("default ChargeLimit = %d, want 80", cfg.Battery.ChargeLimit)
	}
	if cfg.Theme.Mode != "dark" {
		t.Errorf("default Theme.Mode = %q, want %q", cfg.Theme.Mode, "dark")
	}
	if cfg.Cleanup.MaxClipboard != 500 {
		t.Errorf("default MaxClipboard = %d, want 500", cfg.Cleanup.MaxClipboard)
	}
	if cfg.Cleanup.ScreenshotMaxDays != 30 {
		t.Errorf("default ScreenshotMaxDays = %d, want 30", cfg.Cleanup.ScreenshotMaxDays)
	}
	if cfg.Cleanup.RecordingMaxDays != 14 {
		t.Errorf("default RecordingMaxDays = %d, want 14", cfg.Cleanup.RecordingMaxDays)
	}
	if cfg.Hardware.Profile != "" {
		t.Errorf("default Hardware.Profile = %q, want empty", cfg.Hardware.Profile)
	}
}

func TestLoadFromMissing(t *testing.T) {
	cfg := LoadFrom(filepath.Join(t.TempDir(), "nonexistent.toml"))
	want := Default()
	if cfg != want {
		t.Errorf("LoadFrom(missing) = %+v, want defaults %+v", cfg, want)
	}
}

func TestLoadFromPartial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`
[theme]
mode = "light"
`), 0o644)

	cfg := LoadFrom(path)
	if cfg.Theme.Mode != "light" {
		t.Errorf("Theme.Mode = %q, want %q", cfg.Theme.Mode, "light")
	}
	// Other fields should be defaults
	if cfg.Battery.ChargeLimit != 80 {
		t.Errorf("ChargeLimit = %d, want 80 (default)", cfg.Battery.ChargeLimit)
	}
	if cfg.Cleanup.MaxClipboard != 500 {
		t.Errorf("MaxClipboard = %d, want 500 (default)", cfg.Cleanup.MaxClipboard)
	}
}

func TestLoadFromCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`this is not valid toml {{{{`), 0o644)

	cfg := LoadFrom(path)
	want := Default()
	if cfg != want {
		t.Errorf("LoadFrom(corrupt) = %+v, want defaults %+v", cfg, want)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	original := Config{
		Battery: BatteryConfig{ChargeLimit: 60},
		Theme:   ThemeConfig{Mode: "light"},
		Cleanup: CleanupConfig{
			MaxClipboard:      200,
			ScreenshotMaxDays: 7,
			RecordingMaxDays:  3,
		},
		Hardware: HardwareConfig{Profile: "framework-13-amd"},
	}

	if err := SaveTo(original, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	loaded := LoadFrom(path)
	if loaded.Battery.ChargeLimit != 60 {
		t.Errorf("ChargeLimit = %d, want 60", loaded.Battery.ChargeLimit)
	}
	if loaded.Theme.Mode != "light" {
		t.Errorf("Theme.Mode = %q, want %q", loaded.Theme.Mode, "light")
	}
	if loaded.Cleanup.MaxClipboard != 200 {
		t.Errorf("MaxClipboard = %d, want 200", loaded.Cleanup.MaxClipboard)
	}
	if loaded.Cleanup.ScreenshotMaxDays != 7 {
		t.Errorf("ScreenshotMaxDays = %d, want 7", loaded.Cleanup.ScreenshotMaxDays)
	}
	if loaded.Cleanup.RecordingMaxDays != 3 {
		t.Errorf("RecordingMaxDays = %d, want 3", loaded.Cleanup.RecordingMaxDays)
	}
	if loaded.Hardware.Profile != "framework-13-amd" {
		t.Errorf("Hardware.Profile = %q, want %q", loaded.Hardware.Profile, "framework-13-amd")
	}
}

func TestLoadFromDefaultsForZeroValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	// Write config with explicit zero values
	os.WriteFile(path, []byte(`
[battery]
charge_limit = 0

[cleanup]
max_clipboard = 0
`), 0o644)

	cfg := LoadFrom(path)
	// Zero values should be replaced with defaults
	if cfg.Battery.ChargeLimit != 80 {
		t.Errorf("ChargeLimit = %d, want 80 (default for zero)", cfg.Battery.ChargeLimit)
	}
	if cfg.Cleanup.MaxClipboard != 500 {
		t.Errorf("MaxClipboard = %d, want 500 (default for zero)", cfg.Cleanup.MaxClipboard)
	}
	// Mode="" should become "dark"
	if cfg.Theme.Mode != "dark" {
		t.Errorf("Theme.Mode = %q, want %q (default for empty)", cfg.Theme.Mode, "dark")
	}
}
