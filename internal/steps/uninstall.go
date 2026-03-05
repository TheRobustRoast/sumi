package steps

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/model"
	"sumi/internal/runner"
)

// UninstallSteps returns the steps for removing sumi configs.
func UninstallSteps() []model.Step {
	return []model.Step{
		{
			ID:      "uninstall/stop-services",
			Section: "Services",
			Name:    "stop and disable user services",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(stopUserServices)
			},
		},
		{
			ID:    "uninstall/systemd-units",
			Name:  "remove systemd user units",
			RunGo: removeSystemdUnits,
		},
		{
			ID:      "uninstall/config-symlinks",
			Section: "Configs",
			Name:    "remove config symlinks",
			RunGo:   removeConfigSymlinks,
		},
		{
			ID:    "uninstall/theme-overrides",
			Name:  "remove theme overrides",
			RunGo: removeThemeOverrides,
		},
		{
			ID:    "uninstall/dotfiles",
			Name:  "remove managed dotfiles",
			RunGo: removeManagedDotfiles,
		},
		{
			ID:      "uninstall/cache",
			Section: "Cache",
			Name:    "clean sumi cache",
			RunGo:   cleanCache,
		},
		{
			ID:      "uninstall/pacman-hooks",
			Section: "System",
			Name:    "remove pacman hooks",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(func(send func(string)) error {
					f := "/etc/pacman.d/hooks/keyring.hook"
					if _, err := os.Stat(f); err == nil {
						runner.RunCmd(send, "sudo", "rm", "-f", f) //nolint:errcheck
						send("removed: " + f)
					}
					return nil
				})
			},
		},
		{
			ID:   "uninstall/battery",
			Name: "remove battery charge limit config",
			RunStream: func(_ model.InstallCtx) (*runner.Stream, tea.Cmd) {
				return runner.Func(removeBatteryConfig)
			},
			Skip: func(_ model.InstallCtx) bool {
				_, err := os.Stat("/etc/tmpfiles.d/battery-charge-limit.conf")
				return err != nil
			},
		},
	}
}

func stopUserServices(send func(string)) error {
	services := []string{
		"cliphist.service",
		"wallust-watcher.service",
		"sumi-cleanup.timer",
		"sumi-update-check.timer",
	}
	for _, svc := range services {
		runner.RunCmd(send, "systemctl", "--user", "stop", svc)    //nolint:errcheck
		runner.RunCmd(send, "systemctl", "--user", "disable", svc) //nolint:errcheck
	}
	runner.RunCmd(send, "systemctl", "--user", "disable", "lock-before-sleep.service") //nolint:errcheck
	send("user services stopped and disabled")
	return nil
}

func removeSystemdUnits(ctx model.InstallCtx) ([]string, error) {
	units := []string{
		"cliphist.service",
		"wallust-watcher.service",
		"lock-before-sleep.service",
		"sumi-cleanup.service",
		"sumi-cleanup.timer",
		"sumi-update-check.service",
		"sumi-update-check.timer",
		"hyprland-session.target",
	}
	var lines []string
	unitDir := filepath.Join(ctx.Home, ".config/systemd/user")
	for _, unit := range units {
		path := filepath.Join(unitDir, unit)
		if err := os.Remove(path); err == nil {
			lines = append(lines, "removed: " + unit)
		}
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run() //nolint:errcheck
	if len(lines) == 0 {
		lines = append(lines, "no units to remove")
	}
	return lines, nil
}

func removeConfigSymlinks(ctx model.InstallCtx) ([]string, error) {
	h := ctx.Home
	configs := []string{
		filepath.Join(h, ".config/hypr"),
		filepath.Join(h, ".config/waybar"),
		filepath.Join(h, ".config/foot"),
		filepath.Join(h, ".config/fuzzel"),
		filepath.Join(h, ".config/dunst"),
		filepath.Join(h, ".config/wallust"),
		filepath.Join(h, ".config/yazi"),
		filepath.Join(h, ".config/cava"),
		filepath.Join(h, ".config/lazygit"),
		filepath.Join(h, ".config/btop"),
		filepath.Join(h, ".config/starship.toml"),
	}

	var lines []string
	for _, cfg := range configs {
		fi, err := os.Lstat(cfg)
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			os.Remove(cfg) //nolint:errcheck
			lines = append(lines, "removed: "+shortenPath(cfg, h))
		} else {
			os.RemoveAll(cfg) //nolint:errcheck
			lines = append(lines, "removed: "+shortenPath(cfg, h))
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "no configs to remove")
	}
	return lines, nil
}

func removeThemeOverrides(ctx model.InstallCtx) ([]string, error) {
	h := ctx.Home
	paths := []string{
		filepath.Join(h, ".config/gtk-3.0/settings.ini"),
		filepath.Join(h, ".config/gtk-4.0/settings.ini"),
		filepath.Join(h, ".config/mimeapps.list"),
		filepath.Join(h, ".config/xdg-desktop-portal"),
		filepath.Join(h, ".icons/default"),
	}
	var lines []string
	for _, p := range paths {
		fi, err := os.Lstat(p)
		if err != nil {
			continue
		}
		if fi.IsDir() {
			os.RemoveAll(p) //nolint:errcheck
		} else {
			os.Remove(p) //nolint:errcheck
		}
		lines = append(lines, "removed: "+shortenPath(p, h))
	}
	if len(lines) == 0 {
		lines = append(lines, "no overrides to remove")
	}
	return lines, nil
}

func removeManagedDotfiles(ctx model.InstallCtx) ([]string, error) {
	h := ctx.Home
	var lines []string

	// Only remove if it's a sumi-managed file (symlink or contains "sumi")
	for _, path := range []string{
		filepath.Join(h, ".zshrc"),
		filepath.Join(h, ".tmux.conf"),
		filepath.Join(h, ".editorconfig"),
	} {
		fi, err := os.Lstat(path)
		if err != nil {
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			os.Remove(path) //nolint:errcheck
			lines = append(lines, "removed: "+filepath.Base(path))
			continue
		}
		// Regular file — only remove if it contains "sumi"
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "sumi") {
			os.Remove(path) //nolint:errcheck
			lines = append(lines, "removed: "+filepath.Base(path)+" (sumi config)")
		} else {
			lines = append(lines, "skipped: "+filepath.Base(path)+" (user's own config)")
		}
	}

	// nvim — check for sumi marker
	nvimInit := filepath.Join(h, ".config/nvim/init.lua")
	if data, err := os.ReadFile(nvimInit); err == nil {
		if strings.Contains(string(data), "sumi") {
			os.Remove(nvimInit) //nolint:errcheck
			lines = append(lines, "removed: nvim/init.lua (sumi config)")
		} else {
			lines = append(lines, "skipped: nvim/init.lua (user's own config)")
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "no dotfiles to remove")
	}
	return lines, nil
}

func cleanCache(ctx model.InstallCtx) ([]string, error) {
	cacheDir := filepath.Join(ctx.Home, ".cache/sumi")
	if err := os.RemoveAll(cacheDir); err != nil {
		return nil, err
	}
	return []string{"removed: ~/.cache/sumi"}, nil
}

func removeBatteryConfig(send func(string)) error {
	runner.RunCmd(send, "sudo", "rm", "-f", "/etc/tmpfiles.d/battery-charge-limit.conf") //nolint:errcheck
	send("removed battery-charge-limit.conf")
	return nil
}
