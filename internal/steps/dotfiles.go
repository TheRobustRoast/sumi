package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sumi/internal/model"
)

// Dotfiles returns the step that symlinks all config files.
func Dotfiles() model.Step {
	return model.Step{
		ID:      "configs/dotfiles",
		Section: "Configs",
		Name:    "link dotfiles",
		RunGo:   linkDotfiles,
	}
}

func linkDotfiles(ctx model.InstallCtx) ([]string, error) {
	var lines []string

	link := func(src, dst string) {
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			lines = append(lines, "✗ mkdir "+filepath.Dir(dst)+": "+err.Error())
			return
		}
		if fi, err := os.Lstat(dst); err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				os.Remove(dst) //nolint:errcheck
			} else {
				bak := fmt.Sprintf("%s.bak.%d", dst, time.Now().Unix())
				os.Rename(dst, bak) //nolint:errcheck
				lines = append(lines, "! backed up "+filepath.Base(dst))
			}
		}
		if err := os.Symlink(src, dst); err != nil {
			lines = append(lines, "✗ "+filepath.Base(dst)+": "+err.Error())
		} else {
			lines = append(lines, "→ "+filepath.Base(dst))
		}
	}

	sd, h := ctx.SumiDir, ctx.Home

	os.MkdirAll(filepath.Join(h, ".config/hypr/conf.d"), 0o755) //nolint:errcheck
	for _, name := range []string{"hyprland.conf", "hyprlock.conf", "hypridle.conf", "hyprpaper.conf"} {
		link(filepath.Join(sd, "hypr", name), filepath.Join(h, ".config/hypr", name))
	}
	if entries, err := os.ReadDir(filepath.Join(sd, "hypr/conf.d")); err == nil {
		for _, e := range entries {
			link(
				filepath.Join(sd, "hypr/conf.d", e.Name()),
				filepath.Join(h, ".config/hypr/conf.d", e.Name()),
			)
		}
	}
	link(filepath.Join(sd, "scripts"), filepath.Join(h, ".config/hypr/scripts"))

	for _, d := range []string{"waybar", "yazi", "btop", "wallust", "nvim"} {
		link(filepath.Join(sd, d), filepath.Join(h, ".config", d))
	}

	for _, p := range [][2]string{
		{filepath.Join(sd, "foot/foot.ini"), filepath.Join(h, ".config/foot/foot.ini")},
		{filepath.Join(sd, "fuzzel/fuzzel.ini"), filepath.Join(h, ".config/fuzzel/fuzzel.ini")},
		{filepath.Join(sd, "dunst/dunstrc"), filepath.Join(h, ".config/dunst/dunstrc")},
		{filepath.Join(sd, "cava/config"), filepath.Join(h, ".config/cava/config")},
		{filepath.Join(sd, "lazygit/config.yml"), filepath.Join(h, ".config/lazygit/config.yml")},
		{filepath.Join(sd, "gtk-3.0/settings.ini"), filepath.Join(h, ".config/gtk-3.0/settings.ini")},
		{filepath.Join(sd, "gtk-4.0/settings.ini"), filepath.Join(h, ".config/gtk-4.0/settings.ini")},
		{filepath.Join(sd, "xdg/mimeapps.list"), filepath.Join(h, ".config/mimeapps.list")},
		{filepath.Join(sd, "xdg/hyprland-portals.conf"), filepath.Join(h, ".config/xdg-desktop-portal/portals.conf")},
		{filepath.Join(sd, "icons/default/index.theme"), filepath.Join(h, ".icons/default/index.theme")},
		{filepath.Join(sd, "starship/starship.toml"), filepath.Join(h, ".config/starship.toml")},
		{filepath.Join(sd, "tmux/tmux.conf"), filepath.Join(h, ".tmux.conf")},
		{filepath.Join(sd, "zsh/.zshrc"), filepath.Join(h, ".zshrc")},
	} {
		link(p[0], p[1])
	}

	if src := filepath.Join(sd, ".editorconfig"); fileExists(src) {
		link(src, filepath.Join(h, ".editorconfig"))
	}

	// foot/colors.ini — copy once; wallust will overwrite on first wallpaper set
	if dst := filepath.Join(h, ".config/foot/colors.ini"); !fileExists(dst) {
		if data, err := os.ReadFile(filepath.Join(sd, "foot/colors.ini")); err == nil {
			os.WriteFile(dst, data, 0o644) //nolint:errcheck
			lines = append(lines, "→ foot/colors.ini (seed)")
		}
	}

	if svcDir := filepath.Join(sd, "systemd/user"); fileExists(svcDir) {
		if entries, err := os.ReadDir(svcDir); err == nil {
			for _, e := range entries {
				link(filepath.Join(svcDir, e.Name()), filepath.Join(h, ".config/systemd/user", e.Name()))
			}
		}
	}

	if binDir := filepath.Join(sd, "bin"); fileExists(binDir) {
		os.MkdirAll(filepath.Join(h, ".local/bin"), 0o755) //nolint:errcheck
		if entries, err := os.ReadDir(binDir); err == nil {
			for _, e := range entries {
				src := filepath.Join(binDir, e.Name())
				os.Chmod(src, 0o755) //nolint:errcheck
				link(src, filepath.Join(h, ".local/bin", e.Name()))
			}
		}
	}

	// wallust color seeds — copy once if missing (wallust overwrites later)
	seedColorFile(sd, h, "wallust/templates/colors-waybar.css", ".config/waybar/colors.css", &lines)
	seedColorFile(sd, h, "wallust/templates/colors-fuzzel.ini", ".config/fuzzel/colors.ini", &lines)

	return lines, nil
}

func seedColorFile(sd, h, src, dst string, lines *[]string) {
	dstPath := filepath.Join(h, dst)
	if fileExists(dstPath) {
		return
	}
	if data, err := os.ReadFile(filepath.Join(sd, src)); err == nil {
		os.WriteFile(dstPath, data, 0o644) //nolint:errcheck
		*lines = append(*lines, "→ "+filepath.Base(dst)+" (seed)")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Dirs returns the step that creates user directories.
func Dirs() model.Step {
	return model.Step{
		ID:      "configs/dirs",
		Section: "Directories",
		Name:    "create user directories",
		RunGo:   createDirs,
	}
}

func createDirs(ctx model.InstallCtx) ([]string, error) {
	dirs := []string{
		filepath.Join(ctx.Home, "Pictures/Wallpapers"),
		filepath.Join(ctx.Home, "Pictures/Screenshots"),
		filepath.Join(ctx.Home, "Videos/Recordings"),
		filepath.Join(ctx.Home, ".cache/sumi"),
		filepath.Join(ctx.Home, ".local/share/sumi"),
		filepath.Join(ctx.Home, ".local/bin"),
	}
	var lines []string
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return lines, fmt.Errorf("mkdir %s: %w", d, err)
		}
		lines = append(lines, "→ ~/"+strings.TrimPrefix(d, ctx.Home+"/"))
	}
	return lines, nil
}
