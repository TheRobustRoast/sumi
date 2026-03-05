package hardware

import (
	"os/exec"

	"sumi/internal/runner"
)

// FrameworkAMD returns the hardware profile for the Framework Laptop 13 (AMD Ryzen 7040).
func FrameworkAMD() *Profile {
	return &Profile{
		ID:           "framework-13-amd",
		Name:         "Framework 13 AMD",
		AURPackages:  []string{"framework-laptop-kmod-dkms-git"},
		KernelParams: []string{"amd_pstate=active", "rtc_cmos.use_acpi_alarm=1"},
		Modules:      []string{"cros_ec", "cros_ec_lpcs"},
		Services:     []string{"fprintd.service", "fwupd.service"},
		Tweaks:       frameworkTweaks,
	}
}

func frameworkTweaks(send func(string)) error {
	// framework-laptop-kmod modules
	if out, err := exec.Command("pacman", "-Qi", "framework-laptop-kmod-dkms-git").Output(); err == nil && len(out) > 0 {
		if err := runner.WriteAsSudo(send, "/etc/modules-load.d/framework.conf", "cros_ec\ncros_ec_lpcs\n"); err == nil {
			send("framework-laptop-kmod configured")
		}
	}

	// Battery charge limit
	const chargeThreshold = "/sys/class/power_supply/BAT1/charge_control_end_threshold"
	if fileExists(chargeThreshold) {
		runner.WriteAsSudo(send, chargeThreshold, "80\n")                                     //nolint:errcheck
		runner.WriteAsSudo(send, "/etc/tmpfiles.d/battery-charge-limit.conf",                  //nolint:errcheck
			"w /sys/class/power_supply/BAT1/charge_control_end_threshold - - - - 80\n")
		send("battery charge limit: 80%")
	}

	return nil
}
