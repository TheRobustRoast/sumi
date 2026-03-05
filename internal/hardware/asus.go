package hardware

import "sumi/internal/runner"

// ASUSROG returns the hardware profile for ASUS ROG laptops.
func ASUSROG() *Profile {
	return &Profile{
		ID:          "asus-rog",
		Name:        "ASUS ROG",
		Packages:    []string{"asusctl", "supergfxctl"},
		AURPackages: []string{"rog-control-center"},
		Services:    []string{"power-profiles-daemon.service", "supergfxd.service", "asusd.service"},
		Tweaks:      asusROGTweaks,
	}
}

func asusROGTweaks(send func(string)) error {
	// Configure hybrid GPU mode via supergfxctl
	runner.WriteAsSudo(send, "/etc/supergfxd.conf", //nolint:errcheck
		`{"mode":"Hybrid","vfio_enable":false,"vfio_save":false,"always_reboot":false,"no_logind":false,"logout_timeout_s":180,"hotplug_type":"None"}`+"\n")
	send("supergfxctl: Hybrid GPU mode configured")

	return nil
}
