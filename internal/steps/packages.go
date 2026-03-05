package steps

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/hardware"
	"sumi/internal/runner"
	"sumi/internal/model"
)

// baseAURPackages are AUR packages installed regardless of hardware.
var baseAURPackages = []string{"wallust", "bluetuith", "bibata-cursor-theme"}

// baseExtraPackages are extra pacman packages installed regardless of hardware.
var baseExtraPackages = []string{
	"shellcheck", "shfmt", "docker", "docker-compose", "lsof", "direnv", "rsync",
}

// Packages returns steps for installing the AUR helper and all packages.
// The hardware profile contributes additional packages.
func Packages(hw *hardware.Profile) []model.Step {
	aurPkgs := append([]string{}, baseAURPackages...)
	aurPkgs = append(aurPkgs, hw.AURPackages...)

	extraPkgs := append([]string{}, baseExtraPackages...)
	extraPkgs = append(extraPkgs, hw.Packages...)

	return []model.Step{
		{
			ID:      "packages/yay",
			Section: "AUR Helper",
			Name:    "yay",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					send("cloning yay-bin...")
					if err := runner.RunCmd(send, "git", "clone", "https://aur.archlinux.org/yay-bin.git", "/tmp/yay-bin"); err != nil {
						return err
					}
					send("makepkg -si --noconfirm")
					if err := runner.RunCmdDir(send, "/tmp/yay-bin", "makepkg", "-si", "--noconfirm"); err != nil {
						return err
					}
					return runner.RunCmd(send, "rm", "-rf", "/tmp/yay-bin")
				})
			},
			Skip: func(_ model.InstallCtx) bool {
				_, err := exec.LookPath("yay")
				return err == nil
			},
		},
		{
			ID:      "packages/aur",
			Section: "AUR Packages",
			Name:    "AUR packages",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					args := append([]string{"-S", "--needed", "--noconfirm"}, aurPkgs...)
					return runner.RunCmd(send, append([]string{"yay"}, args...)...)
				})
			},
		},
		{
			ID:      "packages/extra",
			Section: "Extra Packages",
			Name:    "extra pacman packages",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					args := append([]string{"pacman", "-S", "--needed", "--noconfirm"}, extraPkgs...)
					return runner.RunCmd(send, append([]string{"sudo"}, args...)...)
				})
			},
		},
	}
}
