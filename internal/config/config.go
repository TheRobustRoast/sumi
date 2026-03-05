// Package config manages the sumi configuration file at ~/.config/sumi/config.toml.
// All values have sensible defaults — the file is optional.
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the sumi configuration schema.
type Config struct {
	Battery  BatteryConfig  `toml:"battery"`
	Theme    ThemeConfig    `toml:"theme"`
	Cleanup  CleanupConfig  `toml:"cleanup"`
	Hardware HardwareConfig `toml:"hardware"`
}

// BatteryConfig controls battery charge behavior.
type BatteryConfig struct {
	// ChargeLimit is the maximum battery percentage (0 = disabled).
	ChargeLimit int `toml:"charge_limit"`
}

// ThemeConfig controls the visual theme.
type ThemeConfig struct {
	// Mode is "dark" or "light".
	Mode string `toml:"mode"`
}

// CleanupConfig controls the sumi-cleanup timer.
type CleanupConfig struct {
	// MaxClipboard is the max clipboard history entries to keep.
	MaxClipboard int `toml:"max_clipboard"`
	// ScreenshotMaxDays is the max age of screenshots in days (0 = keep forever).
	ScreenshotMaxDays int `toml:"screenshot_max_days"`
	// RecordingMaxDays is the max age of recordings in days (0 = keep forever).
	RecordingMaxDays int `toml:"recording_max_days"`
}

// HardwareConfig controls hardware profile selection.
type HardwareConfig struct {
	// Profile overrides auto-detection. Empty = auto-detect.
	Profile string `toml:"profile"`
}

// Default returns the default configuration.
func Default() Config {
	return Config{
		Battery: BatteryConfig{
			ChargeLimit: 80,
		},
		Theme: ThemeConfig{
			Mode: "dark",
		},
		Cleanup: CleanupConfig{
			MaxClipboard:      500,
			ScreenshotMaxDays: 30,
			RecordingMaxDays:  14,
		},
		Hardware: HardwareConfig{
			Profile: "",
		},
	}
}

// Path returns the config file path.
func Path() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config/sumi/config.toml")
}

// Load reads the config from disk, falling back to defaults for missing values.
func Load() Config {
	return LoadFrom(Path())
}

// LoadFrom reads the config from the given path, falling back to defaults for missing values.
func LoadFrom(path string) Config {
	cfg := Default()
	if _, err := os.Stat(path); err != nil {
		return cfg
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Default()
	}
	// Apply defaults for zero values
	if cfg.Theme.Mode == "" {
		cfg.Theme.Mode = "dark"
	}
	if cfg.Battery.ChargeLimit == 0 {
		cfg.Battery.ChargeLimit = 80
	}
	if cfg.Cleanup.MaxClipboard == 0 {
		cfg.Cleanup.MaxClipboard = 500
	}
	return cfg
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	return SaveTo(cfg, Path())
}

// SaveTo writes the config to the given path, creating parent directories as needed.
func SaveTo(cfg Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}
