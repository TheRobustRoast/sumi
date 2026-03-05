package bootstrap

import (
	"sumi/internal/model"
	"sumi/internal/runner"

	tea "github.com/charmbracelet/bubbletea"
)

// BuildSteps returns the ordered list of bootstrap execution steps.
func BuildSteps(cfg *Config, sumiSrc string) []model.Step {
	return []model.Step{
		{
			ID:      "bootstrap/preflight",
			Section: "Pre-flight",
			Name:    "Clean stale state",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return Preflight(send)
				})
			},
		},
		{
			ID:      "bootstrap/partition",
			Section: "Disk",
			Name:    "Partition disk",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return Partition(send, cfg)
				})
			},
		},
		{
			ID:      "bootstrap/format-efi",
			Section: "Disk",
			Name:    "Format EFI partition",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return FormatEFI(send, cfg)
				})
			},
		},
		{
			ID:      "bootstrap/luks",
			Section: "Encryption",
			Name:    "LUKS2 format + open",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return FormatLUKS(send, cfg)
				})
			},
		},
		{
			ID:      "bootstrap/btrfs",
			Section: "Filesystem",
			Name:    "Create btrfs filesystem",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return CreateBtrfs(send)
				})
			},
		},
		{
			ID:      "bootstrap/subvolumes",
			Section: "Filesystem",
			Name:    "Create btrfs subvolumes",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return CreateSubvolumes(send)
				})
			},
		},
		{
			ID:      "bootstrap/mount",
			Section: "Filesystem",
			Name:    "Mount filesystems",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return MountFilesystems(send, cfg)
				})
			},
		},
		{
			ID:      "bootstrap/pacstrap",
			Section: "Packages",
			Name:    "Install packages (pacstrap)",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return Pacstrap(send, cfg)
				})
			},
		},
		{
			ID:      "bootstrap/fstab",
			Section: "Packages",
			Name:    "Generate fstab",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return GenFstab(send)
				})
			},
		},
		{
			ID:      "bootstrap/chroot",
			Section: "Configuration",
			Name:    "Configure system (chroot)",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return ChrootConfigure(send, cfg)
				})
			},
		},
		{
			ID:      "bootstrap/stage",
			Section: "Staging",
			Name:    "Stage sumi for first boot",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					return StageSumi(send, cfg, sumiSrc)
				})
			},
		},
	}
}
