package steps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sumi/internal/model"
	"sumi/internal/wallpaper"
)

// DoctorChecks returns health-check steps for sumi doctor.
func DoctorChecks() []model.Step {
	return []model.Step{
		{
			ID:      "doctor/symlinks",
			Section: "Symlinks",
			Name:    "config symlinks",
			RunGo:   checkSymlinks,
		},
		{
			ID:      "doctor/packages",
			Section: "Packages",
			Name:    "core packages installed",
			RunGo:   checkPackages,
		},
		{
			ID:      "doctor/user-services",
			Section: "Services",
			Name:    "user services",
			RunGo:   checkUserServices,
		},
		{
			ID:    "doctor/system-services",
			Name:  "system services",
			RunGo: checkSystemServices,
		},
		{
			ID:      "doctor/wallpaper",
			Section: "Theming",
			Name:    "wallpaper pipeline",
			RunGo:   checkWallpaperPipeline,
		},
		{
			ID:      "doctor/wallust-templates",
			Name:    "wallust template validation",
			RunGo:   checkWallustTemplates,
		},
		{
			ID:      "doctor/disk",
			Section: "Disk",
			Name:    "free space",
			RunGo:   checkDiskSpace,
		},
	}
}

func checkSymlinks(ctx model.InstallCtx) ([]string, error) {
	var lines []string
	var missing int

	check := func(target string) {
		fi, err := os.Lstat(target)
		if err != nil {
			lines = append(lines, "  missing: "+shortenPath(target, ctx.Home))
			missing++
			return
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			lines = append(lines, "  not a symlink: "+shortenPath(target, ctx.Home))
			missing++
			return
		}
		lines = append(lines, "  ok: "+shortenPath(target, ctx.Home))
	}

	h := ctx.Home
	for _, d := range []string{"hypr", "waybar", "yazi", "btop", "wallust", "nvim"} {
		check(filepath.Join(h, ".config", d))
	}
	for _, f := range []string{
		".config/foot/foot.ini",
		".config/fuzzel/fuzzel.ini",
		".config/dunst/dunstrc",
		".config/starship.toml",
		".tmux.conf",
		".zshrc",
	} {
		check(filepath.Join(h, f))
	}

	if missing > 0 {
		lines = append(lines, "")
		lines = append(lines, "  fix: sumi install (re-creates symlinks)")
		return lines, fmt.Errorf("%d symlink(s) missing or broken", missing)
	}
	return lines, nil
}

func checkPackages(_ model.InstallCtx) ([]string, error) {
	core := []string{
		"hyprland", "foot", "waybar", "fuzzel", "dunst",
		"hyprlock", "hypridle", "hyprpaper",
		"yazi", "lazygit", "btop", "neovim",
		"zsh", "starship", "tmux", "fzf", "ripgrep",
	}
	var lines []string
	var missing int
	for _, pkg := range core {
		if err := exec.Command("pacman", "-Qi", pkg).Run(); err != nil {
			lines = append(lines, "  missing: "+pkg)
			missing++
		} else {
			lines = append(lines, "  ok: "+pkg)
		}
	}
	if missing > 0 {
		var missingPkgs []string
		for _, pkg := range core {
			if exec.Command("pacman", "-Qi", pkg).Run() != nil {
				missingPkgs = append(missingPkgs, pkg)
			}
		}
		lines = append(lines, "")
		lines = append(lines, "  fix: sudo pacman -S "+strings.Join(missingPkgs, " "))
		return lines, fmt.Errorf("%d package(s) not installed", missing)
	}
	return lines, nil
}

func checkUserServices(_ model.InstallCtx) ([]string, error) {
	services := []string{
		"cliphist.service",
		"wallust-watcher.service",
		"lock-before-sleep.service",
		"sumi-cleanup.timer",
		"sumi-update-check.timer",
	}
	var lines []string
	var failed int
	for _, svc := range services {
		out, _ := exec.Command("systemctl", "--user", "is-active", svc).Output()
		status := strings.TrimSpace(string(out))
		if status == "active" {
			lines = append(lines, "  active: "+svc)
		} else {
			lines = append(lines, "  "+status+": "+svc)
			failed++
		}
	}
	if failed > 0 {
		lines = append(lines, "")
		lines = append(lines, "  fix: systemctl --user daemon-reload && systemctl --user enable --now <service>")
		return lines, fmt.Errorf("%d user service(s) not active", failed)
	}
	return lines, nil
}

func checkSystemServices(_ model.InstallCtx) ([]string, error) {
	services := []string{
		"bluetooth.service",
		"NetworkManager.service",
		"greetd.service",
	}
	var lines []string
	var failed int
	for _, svc := range services {
		out, _ := exec.Command("systemctl", "is-enabled", svc).Output()
		status := strings.TrimSpace(string(out))
		if status == "enabled" {
			lines = append(lines, "  enabled: "+svc)
		} else {
			lines = append(lines, "  "+status+": "+svc)
			failed++
		}
	}
	if failed > 0 {
		return lines, fmt.Errorf("%d system service(s) not enabled", failed)
	}
	return lines, nil
}

func checkWallpaperPipeline(ctx model.InstallCtx) ([]string, error) {
	var lines []string
	var issues int

	// wallust binary
	if _, err := exec.LookPath("wallust"); err != nil {
		lines = append(lines, "  missing: wallust binary")
		issues++
	} else {
		lines = append(lines, "  ok: wallust installed")
	}

	// wallust templates
	tplDir := filepath.Join(ctx.Home, ".config/wallust/templates")
	if entries, err := os.ReadDir(tplDir); err != nil {
		lines = append(lines, "  missing: wallust templates dir")
		issues++
	} else {
		lines = append(lines, fmt.Sprintf("  ok: %d wallust templates", len(entries)))
	}

	// current-wallpaper file
	cwp := filepath.Join(ctx.Home, ".cache/sumi/current-wallpaper")
	if _, err := os.Stat(cwp); err != nil {
		lines = append(lines, "  missing: current-wallpaper cache (set a wallpaper first)")
		// This is a warning, not a failure
	} else {
		lines = append(lines, "  ok: current-wallpaper cached")
	}

	// wallpaper directory
	wpDir := filepath.Join(ctx.Home, "Pictures/Wallpapers")
	if entries, err := os.ReadDir(wpDir); err != nil {
		lines = append(lines, "  missing: ~/Pictures/Wallpapers/")
		issues++
	} else {
		count := 0
		for _, e := range entries {
			if !e.IsDir() {
				count++
			}
		}
		lines = append(lines, fmt.Sprintf("  ok: %d wallpaper(s) in ~/Pictures/Wallpapers/", count))
	}

	if issues > 0 {
		return lines, fmt.Errorf("%d wallpaper pipeline issue(s)", issues)
	}
	return lines, nil
}

func checkWallustTemplates(ctx model.InstallCtx) ([]string, error) {
	wallustDir := filepath.Join(ctx.Home, ".config/wallust")
	issues := wallpaper.ValidateTemplates(wallustDir)
	var lines []string
	if len(issues) == 0 {
		lines = append(lines, "  ok: all wallust templates valid")
		return lines, nil
	}
	for _, issue := range issues {
		lines = append(lines, "  warning: "+issue)
	}
	return lines, fmt.Errorf("%d wallust template issue(s)", len(issues))
}

func checkDiskSpace(_ model.InstallCtx) ([]string, error) {
	out, err := exec.Command("df", "-h", "/", "/home").Output()
	if err != nil {
		return nil, fmt.Errorf("df: %w", err)
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		lines = append(lines, "  "+line)
	}
	return lines, nil
}

func shortenPath(path, home string) string {
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
