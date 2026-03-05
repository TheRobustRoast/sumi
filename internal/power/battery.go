package power

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const chargeLimitPath = "/sys/class/power_supply/BAT1/charge_control_end_threshold"

// BatteryInfo holds battery state for display.
type BatteryInfo struct {
	Capacity    int
	Status      string
	ChargeLimit int
	PowerWatts  float64
	HealthPct   int
}

// GetBatteryInfo reads battery state from sysfs.
func GetBatteryInfo() BatteryInfo {
	info := BatteryInfo{
		Status:      readSysfs("/sys/class/power_supply/BAT1/status", "Unknown"),
		ChargeLimit: 100,
	}

	info.Capacity, _ = strconv.Atoi(readSysfs("/sys/class/power_supply/BAT1/capacity", "0"))

	powerNow, _ := strconv.ParseFloat(readSysfs("/sys/class/power_supply/BAT1/power_now", "0"), 64)
	info.PowerWatts = powerNow / 1_000_000

	energyFull, _ := strconv.ParseFloat(readSysfs("/sys/class/power_supply/BAT1/energy_full", "0"), 64)
	energyDesign, _ := strconv.ParseFloat(readSysfs("/sys/class/power_supply/BAT1/energy_full_design", "0"), 64)
	if energyDesign > 0 {
		info.HealthPct = int(energyFull * 100 / energyDesign)
	}

	if _, err := os.Stat(chargeLimitPath); err == nil {
		limit, _ := strconv.Atoi(readSysfs(chargeLimitPath, "100"))
		info.ChargeLimit = limit
	}

	return info
}

// SetChargeLimit writes a charge limit to sysfs (requires sudo).
func SetChargeLimit(limit int) error {
	if limit < 60 || limit > 100 {
		return fmt.Errorf("charge limit must be between 60 and 100, got %d", limit)
	}
	return os.WriteFile(chargeLimitPath, []byte(strconv.Itoa(limit)), 0o644)
}

// StatusSuffix returns a short suffix for the battery status.
func (b BatteryInfo) StatusSuffix() string {
	switch b.Status {
	case "Charging":
		return "+"
	case "Discharging":
		return "-"
	case "Full":
		return "="
	default:
		return ""
	}
}

// WaybarClass returns a CSS class based on battery state.
func (b BatteryInfo) WaybarClass() string {
	if b.Capacity <= 15 {
		return "critical"
	}
	if b.Capacity <= 30 {
		return "warning"
	}
	if b.Status == "Charging" {
		return "charging"
	}
	return ""
}

func readSysfs(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}
