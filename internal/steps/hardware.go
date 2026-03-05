package steps

import (
	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/hardware"
	"sumi/internal/runner"
	"sumi/internal/model"
)

// HardwareTweaks returns the step for applying hardware-specific configuration.
// If the profile has no tweaks, the step is skipped.
func HardwareTweaks(hw *hardware.Profile) model.Step {
	return model.Step{
		ID:      "hardware/tweaks",
		Section: "Hardware",
		Name:    hw.Name + " tweaks",
		RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
			return runner.Func(func(send func(string)) error {
				// Apply kernel params from the hardware profile
				for _, param := range hw.KernelParams {
					if modified, entry := AddKernelParam(send, param); modified {
						send(param + " added to " + entry)
					}
				}

				// Apply hardware-specific tweaks
				if hw.Tweaks != nil {
					return hw.Tweaks(send)
				}
				return nil
			})
		},
		Skip: func(_ model.InstallCtx) bool {
			return hw.Tweaks == nil && len(hw.KernelParams) == 0
		},
	}
}
