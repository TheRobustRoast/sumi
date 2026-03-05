package steps

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/runner"
	"sumi/internal/model"
)

// SetShell returns the step for setting the user's default shell to zsh.
func SetShell() model.Step {
	return model.Step{
		ID:      "shell/zsh",
		Section: "Shell",
		Name:    "set shell to zsh",
		RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
			return runner.Func(func(send func(string)) error {
				zsh, err := exec.LookPath("zsh")
				if err != nil {
					return fmt.Errorf("zsh not found")
				}
				return runner.RunCmd(send, "chsh", "-s", zsh)
			})
		},
		Skip: func(_ model.InstallCtx) bool {
			zsh, err := exec.LookPath("zsh")
			if err != nil {
				return true
			}
			return os.Getenv("SHELL") == zsh
		},
	}
}
