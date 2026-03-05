package bootstrap

import (
	"strings"

	"sumi/internal/runner"
)

// base packages for every sumi install
var basePackages = []string{
	// Core
	"base", "base-devel", "linux", "linux-firmware", "linux-headers",
	// Filesystem & boot
	"btrfs-progs", "cryptsetup", "dosfstools", "efibootmgr",
	// Network
	"networkmanager",
	// Memory
	"zram-generator",
	// Hyprland
	"hyprland", "hyprpaper", "hyprlock", "hypridle", "hyprpicker",
	"xdg-desktop-portal-hyprland",
	// Terminal & bar
	"foot", "waybar", "fuzzel", "dunst",
	// Screen & clipboard
	"grim", "slurp", "wl-clipboard", "cliphist", "polkit-gnome",
	// Media utilities
	"cava", "brightnessctl", "playerctl", "pulsemixer", "wtype",
	// Bluetooth
	"bluez", "bluez-utils",
	// File manager
	"yazi", "ffmpegthumbnailer", "p7zip", "unarchiver", "poppler",
	// CLI tools
	"fd", "ripgrep", "fzf", "zoxide",
	// Viewers
	"imv", "mpv", "zathura", "zathura-pdf-mupdf",
	// System tools
	"btop", "neovim", "ncdu",
	// Git
	"git", "lazygit",
	// Login & boot splash
	"greetd", "greetd-tuigreet", "plymouth",
	// Scripting
	"jq", "bc", "imagemagick", "python-pillow",
	// Shell
	"starship", "zsh", "zsh-autosuggestions", "zsh-syntax-highlighting",
	// Modern CLI replacements
	"bat", "eza", "tokei", "procs", "duf", "dust",
	// Fonts
	"noto-fonts", "noto-fonts-cjk", "noto-fonts-emoji",
	"ttf-jetbrains-mono-nerd", "ttf-font-awesome",
	// Wayland compatibility
	"qt5-wayland", "qt6-wayland",
	// Audio
	"pipewire", "pipewire-alsa", "pipewire-audio",
	"pipewire-jack", "pipewire-pulse", "wireplumber",
	// System services
	"power-profiles-daemon", "acpid",
	// Screen recording & utils
	"wf-recorder", "inotify-tools", "xdg-utils",
	// Build tools (for sumi itself)
	"go",
}

// CPUMicrocode returns the correct microcode package for the CPU vendor.
func CPUMicrocode() string {
	cpu, _ := captureCmd("grep", "-m1", "vendor_id", "/proc/cpuinfo")
	if strings.Contains(cpu, "GenuineIntel") {
		return "intel-ucode"
	}
	return "amd-ucode" // default to AMD
}

// Pacstrap installs all packages into /mnt.
func Pacstrap(send func(string), cfg *Config) error {
	send("Installing packages (this takes 5-15 minutes)...")

	pkgs := make([]string, 0, len(basePackages)+4)
	pkgs = append(pkgs, basePackages...)

	// Add CPU microcode
	ucode := CPUMicrocode()
	if ucode != "" {
		pkgs = append(pkgs, ucode)
	}

	// Hardware-specific packages would be added here via profile

	args := append([]string{"pacstrap", "-K", "/mnt"}, pkgs...)
	return runner.RunCmd(send, args...)
}

// GenFstab generates the fstab file from current mounts.
func GenFstab(send func(string)) error {
	send("Generating fstab...")
	return runner.RunCmd(send, "bash", "-c", "genfstab -U /mnt >> /mnt/etc/fstab")
}

