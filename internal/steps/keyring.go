package steps

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/runner"
	"sumi/internal/model"
)

// Keyring returns steps for refreshing the pacman keyring and installing the hook.
func Keyring() []model.Step {
	return []model.Step{
		{
			ID:      "keyring/refresh",
			Section: "Keyring",
			Name:    "refresh keyring",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					send("pacman -Sy archlinux-keyring")
					if err := runner.RunCmd(send, "sudo", "pacman", "-Sy", "--noconfirm", "archlinux-keyring"); err != nil {
						return err
					}
					send("pacman-key --populate archlinux")
					return runner.RunCmd(send, "sudo", "pacman-key", "--populate", "archlinux")
				})
			},
		},
		{
			ID:   "keyring/hook",
			Name: "install keyring hook",
			RunStream: func(ctx model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					if err := runner.RunCmd(send, "sudo", "mkdir", "-p", "/etc/pacman.d/hooks"); err != nil {
						return err
					}
					src := filepath.Join(ctx.SumiDir, "pacman/hooks/keyring.hook")
					return runner.RunCmd(send, "sudo", "install", "-m", "644", src, "/etc/pacman.d/hooks/keyring.hook")
				})
			},
			Skip: func(_ model.InstallCtx) bool {
				_, err := os.Stat("/etc/pacman.d/hooks/keyring.hook")
				return err == nil
			},
		},
	}
}
