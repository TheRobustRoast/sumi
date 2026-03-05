package hardware

import (
	"os"
	"strings"
)

// DMIReader abstracts reading DMI sysfs fields for testability.
type DMIReader interface {
	Read(field string) string
	HasBattery() bool
}

// SysfsDMI reads DMI data from the real /sys/class/dmi/id/ filesystem.
type SysfsDMI struct{}

func (SysfsDMI) Read(field string) string {
	return readDMI(field)
}

func (SysfsDMI) HasBattery() bool {
	return hasBattery()
}

// Detect reads DMI sysfs data and returns the best matching hardware profile.
func Detect() *Profile {
	return DetectWith(SysfsDMI{})
}

// DetectWith uses the provided DMIReader to detect hardware.
func DetectWith(r DMIReader) *Profile {
	vendor := r.Read("sys_vendor")
	product := r.Read("product_name")

	if strings.Contains(vendor, "Framework") && strings.Contains(product, "AMD") {
		return FrameworkAMD()
	}
	if strings.Contains(vendor, "LENOVO") && strings.Contains(strings.ToLower(product), "thinkpad") {
		return ThinkPad()
	}
	if strings.Contains(vendor, "ASUSTeK") && strings.Contains(strings.ToUpper(product), "ROG") {
		return ASUSROG()
	}
	if r.HasBattery() {
		return GenericLaptop()
	}
	return GenericDesktop()
}

func readDMI(field string) string {
	data, err := os.ReadFile("/sys/class/dmi/id/" + field)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func hasBattery() bool {
	for _, path := range []string{
		"/sys/class/power_supply/BAT0",
		"/sys/class/power_supply/BAT1",
	} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
