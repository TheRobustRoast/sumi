package steps

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/hardware"
	"sumi/internal/runner"
	"sumi/internal/model"
)

// baseSystemServices are always enabled regardless of hardware.
var baseSystemServices = []string{
	"bluetooth.service",
	"NetworkManager.service",
	"power-profiles-daemon.service",
}

// userServices are user-level systemd units.
var userServices = []string{
	"cliphist.service",
	"wallust-watcher.service",
	"lock-before-sleep.service",
	"sumi-cleanup.timer",
	"sumi-update-check.timer",
}

// Services returns the step for enabling system and user services.
// The hardware profile contributes additional services.
func Services(hw *hardware.Profile) model.Step {
	sysServices := append([]string{}, baseSystemServices...)
	sysServices = append(sysServices, hw.Services...)

	return model.Step{
		ID:      "services/enable",
		Section: "Services",
		Name:    "enable system and user services",
		RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
			return runner.Func(func(send func(string)) error {
				for _, svc := range sysServices {
					runner.RunCmd(send, "sudo", "systemctl", "enable", svc) //nolint:errcheck
				}

				// TLP conflicts with power-profiles-daemon
				runner.RunCmd(send, "sudo", "systemctl", "disable", "tlp.service") //nolint:errcheck
				runner.RunCmd(send, "sudo", "systemctl", "mask", "tlp.service")    //nolint:errcheck

				runner.RunCmd(send, "systemctl", "--user", "daemon-reload") //nolint:errcheck

				for _, svc := range userServices {
					runner.RunCmd(send, "systemctl", "--user", "enable", svc) //nolint:errcheck
				}

				send("services enabled")
				return nil
			})
		},
	}
}

// DisableConflictingDMs returns the step for disabling other display managers.
func DisableConflictingDMs() model.Step {
	return model.Step{
		ID:      "services/disable-dms",
		Section: "Cleanup",
		Name:    "disable conflicting display managers",
		RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
			return runner.Func(func(send func(string)) error {
				for _, dm := range []string{"gdm", "sddm", "lightdm", "ly"} {
					out, err := exec.Command("systemctl", "is-enabled", dm+".service").Output()
					if err == nil && strings.TrimSpace(string(out)) == "enabled" {
						runner.RunCmd(send, "sudo", "systemctl", "disable", dm+".service") //nolint:errcheck
						send("disabled: " + dm)
					}
				}
				return nil
			})
		},
	}
}
