package hardware

import (
	"sumi/internal/runner"
)

// ThinkPad returns the hardware profile for Lenovo ThinkPad laptops.
func ThinkPad() *Profile {
	return &Profile{
		ID:           "thinkpad",
		Name:         "ThinkPad",
		Packages:     []string{"throttled"},
		KernelParams: []string{"mem_sleep_default=deep"},
		Modules:      []string{"thinkpad_acpi"},
		Services:     []string{"power-profiles-daemon.service", "fwupd.service"},
		Tweaks:       thinkpadTweaks,
	}
}

func thinkpadTweaks(send func(string)) error {
	// TrackPoint sensitivity
	const tpSpeed = "/sys/devices/platform/i8042/serio1/serio2/sensitivity"
	if fileExists(tpSpeed) {
		runner.WriteAsSudo(send, tpSpeed, "200\n") //nolint:errcheck
		send("TrackPoint sensitivity: 200")
	}

	// Enable S3 deep sleep via tmpfiles.d
	runner.WriteAsSudo(send, "/etc/tmpfiles.d/thinkpad-sleep.conf", //nolint:errcheck
		"w /sys/power/mem_sleep - - - - deep\n")
	send("S3 deep sleep configured")

	return nil
}
