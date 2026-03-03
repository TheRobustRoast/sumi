#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi — Post-Install Bootstrap Script                     ║
# ║                                                              ║
# ║  Run this AFTER archinstall completes and you've rebooted    ║
# ║  into your new system. It deploys all config files and       ║
# ║  sets up the full rice.                                      ║
# ║                                                              ║
# ║  Usage: git clone <repo> ~/sumi && cd ~/sumi           ║
# ║         chmod +x install.sh && ./install.sh                  ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$HOME/.config"
RED='\033[0;31m'
GRN='\033[0;32m'
CYN='\033[0;36m'
DIM='\033[0;90m'
RST='\033[0m'

echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${CYN}║   sumi :: framework 13 installer   ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
echo ""

# ── Helper functions ─────────────────────────────────────────
info()  { echo -e "${CYN}:: ${RST}$1"; }
ok()    { echo -e "${GRN}   ✓${RST} $1"; }
warn()  { echo -e "${RED}   !${RST} $1"; }

backup_if_exists() {
    if [[ -e "$1" ]]; then
        local backup="$1.bak.$(date +%s)"
        mv "$1" "$backup"
        warn "Backed up existing: $1 → $backup"
    fi
}

deploy() {
    local src="$1"
    local dst="$2"
    mkdir -p "$(dirname "$dst")"
    backup_if_exists "$dst"
    cp -r "$src" "$dst"
    ok "Deployed: $dst"
}

# ── 0. Pre-flight checks ──────────────────────────────────
info "Running pre-flight checks..."

# Must not be root
if [[ "$EUID" -eq 0 ]]; then
    warn "Do not run this script as root. Run as your normal user."
    exit 1
fi

# Must be on Arch
if [[ ! -f /etc/arch-release ]]; then
    warn "This script is designed for Arch Linux."
    exit 1
fi

# Must have internet
if ! ping -c 1 -W 3 archlinux.org &>/dev/null; then
    warn "No internet connection detected."
    exit 1
fi

# Check critical base commands
for cmd in git sudo pacman hyprctl; do
    if command -v "$cmd" &>/dev/null; then
        ok "$cmd found"
    else
        warn "$cmd not found — install it first"
        exit 1
    fi
done

ok "Pre-flight checks passed"
echo ""

# ── 1. Install AUR helper (yay) ─────────────────────────────
info "Checking for yay..."
if ! command -v yay &>/dev/null; then
    info "Installing yay..."
    cd /tmp
    git clone https://aur.archlinux.org/yay-bin.git
    cd yay-bin
    makepkg -si --noconfirm
    cd "$SCRIPT_DIR"
    ok "yay installed"
else
    ok "yay already installed"
fi

# ── 2. Install additional AUR packages ──────────────────────
info "Installing AUR packages..."
yay -S --needed --noconfirm \
    wallust \
    bluetuith \
    framework-laptop-kmod-dkms-git \
    wtype \
    bibata-cursor-theme \
    2>/dev/null || warn "Some AUR packages may have failed — check manually"

# ── 2a. Ensure required pacman packages ────────────────────
info "Verifying required packages..."
sudo pacman -S --needed --noconfirm \
    inotify-tools \
    slurp grim \
    wl-clipboard cliphist \
    jq fzf \
    hyprpicker \
    zsh-autosuggestions zsh-syntax-highlighting \
    python bc \
    playerctl brightnessctl \
    tmux \
    npm nodejs \
    shellcheck shfmt \
    docker docker-compose \
    lsof \
    direnv \
    rsync \
    fuzzel \
    doggo \
    2>/dev/null || warn "Some packages may have failed — check manually"

# ── 3. Deploy configs ───────────────────────────────────────
echo ""
info "Deploying configuration files..."

# Hyprland
deploy "$SCRIPT_DIR/hypr/hyprland.conf"     "$CONFIG_DIR/hypr/hyprland.conf"
deploy "$SCRIPT_DIR/hypr/hyprlock.conf"      "$CONFIG_DIR/hypr/hyprlock.conf"
deploy "$SCRIPT_DIR/hypr/hypridle.conf"      "$CONFIG_DIR/hypr/hypridle.conf"
deploy "$SCRIPT_DIR/hypr/hyprpaper.conf"     "$CONFIG_DIR/hypr/hyprpaper.conf"
mkdir -p "$CONFIG_DIR/hypr/conf.d"
for f in "$SCRIPT_DIR/hypr/conf.d/"*; do
    deploy "$f" "$CONFIG_DIR/hypr/conf.d/$(basename "$f")"
done

# Scripts
mkdir -p "$CONFIG_DIR/hypr/scripts"
for f in "$SCRIPT_DIR/scripts/"*.sh; do
    deploy "$f" "$CONFIG_DIR/hypr/scripts/$(basename "$f")"
    chmod +x "$CONFIG_DIR/hypr/scripts/$(basename "$f")"
done

# Waybar
deploy "$SCRIPT_DIR/waybar/config.jsonc"     "$CONFIG_DIR/waybar/config.jsonc"
deploy "$SCRIPT_DIR/waybar/style.css"        "$CONFIG_DIR/waybar/style.css"
# Create default colors.css for waybar (pre-wallust)
cp "$SCRIPT_DIR/wallust/templates/colors-waybar.css" "$CONFIG_DIR/waybar/colors.css" 2>/dev/null || true

# Foot
deploy "$SCRIPT_DIR/foot/foot.ini"           "$CONFIG_DIR/foot/foot.ini"

# Fuzzel
deploy "$SCRIPT_DIR/fuzzel/fuzzel.ini"       "$CONFIG_DIR/fuzzel/fuzzel.ini"
cp "$SCRIPT_DIR/wallust/templates/colors-fuzzel.ini" "$CONFIG_DIR/fuzzel/colors.ini" 2>/dev/null || true

# Dunst
deploy "$SCRIPT_DIR/dunst/dunstrc"           "$CONFIG_DIR/dunst/dunstrc"
mkdir -p "$CONFIG_DIR/dunst/dunstrc.d"

# Yazi (TUI file manager)
deploy "$SCRIPT_DIR/yazi/yazi.toml"          "$CONFIG_DIR/yazi/yazi.toml"
deploy "$SCRIPT_DIR/yazi/theme.toml"         "$CONFIG_DIR/yazi/theme.toml"

# Cava (TUI audio visualizer)
deploy "$SCRIPT_DIR/cava/config"             "$CONFIG_DIR/cava/config"

# Lazygit
deploy "$SCRIPT_DIR/lazygit/config.yml"      "$CONFIG_DIR/lazygit/config.yml"

# Neovim
deploy "$SCRIPT_DIR/nvim/init.lua"           "$CONFIG_DIR/nvim/init.lua"

# Tmux
deploy "$SCRIPT_DIR/tmux/tmux.conf"          "$HOME/.tmux.conf"

# Starship prompt
deploy "$SCRIPT_DIR/starship/starship.toml"  "$CONFIG_DIR/starship.toml"

# Btop
deploy "$SCRIPT_DIR/btop/btop.conf"                "$CONFIG_DIR/btop/btop.conf"
mkdir -p "$CONFIG_DIR/btop/themes"
deploy "$SCRIPT_DIR/btop/themes/sumi.theme"     "$CONFIG_DIR/btop/themes/sumi.theme"

# Wallust
deploy "$SCRIPT_DIR/wallust/wallust.toml"    "$CONFIG_DIR/wallust/wallust.toml"
mkdir -p "$CONFIG_DIR/wallust/templates"
for f in "$SCRIPT_DIR/wallust/templates/"*; do
    deploy "$f" "$CONFIG_DIR/wallust/templates/$(basename "$f")"
done

# GTK theming (dark mode + cursor + font)
deploy "$SCRIPT_DIR/gtk-3.0/settings.ini"         "$CONFIG_DIR/gtk-3.0/settings.ini"
deploy "$SCRIPT_DIR/gtk-4.0/settings.ini"         "$CONFIG_DIR/gtk-4.0/settings.ini"

# XDG defaults (mime associations + portal config)
deploy "$SCRIPT_DIR/xdg/mimeapps.list"            "$CONFIG_DIR/mimeapps.list"
deploy "$SCRIPT_DIR/xdg/hyprland-portals.conf"    "$CONFIG_DIR/xdg-desktop-portal/portals.conf"

# Cursor theme
mkdir -p "$HOME/.icons/default"
deploy "$SCRIPT_DIR/icons/default/index.theme"    "$HOME/.icons/default/index.theme"

# Systemd user services
mkdir -p "$HOME/.config/systemd/user"
for f in "$SCRIPT_DIR/systemd/user/"*; do
    deploy "$f" "$HOME/.config/systemd/user/$(basename "$f")"
done

# ── 4. Setup greetd ─────────────────────────────────────────
echo ""
info "Setting up greetd login manager..."
if [[ -f /etc/greetd/config.toml ]]; then
    sudo cp /etc/greetd/config.toml /etc/greetd/config.toml.bak
fi
sudo cp "$SCRIPT_DIR/greetd/config.toml" /etc/greetd/config.toml
sudo systemctl enable greetd.service 2>/dev/null || true
ok "greetd configured and enabled"

# ── 5. Setup Plymouth ───────────────────────────────────────
info "Setting up Plymouth TUI boot theme..."
sudo mkdir -p /usr/share/plymouth/themes/hypr-tui
sudo cp "$SCRIPT_DIR/plymouth/themes/hypr-tui/"* /usr/share/plymouth/themes/hypr-tui/
sudo plymouth-set-default-theme hypr-tui 2>/dev/null || warn "Plymouth theme set failed — run manually"

# Add plymouth to mkinitcpio HOOKS
if grep -q "^HOOKS=" /etc/mkinitcpio.conf; then
    if ! grep -q "plymouth" /etc/mkinitcpio.conf; then
        info "Adding plymouth hook to mkinitcpio..."
        sudo sed -i 's/^HOOKS=(\(.*\)udev\(.*\))/HOOKS=(\1udev plymouth\2)/' /etc/mkinitcpio.conf
        ok "Plymouth hook added"
    else
        ok "Plymouth hook already present"
    fi
fi

# Add plymouth to kernel cmdline for systemd-boot
LOADER_ENTRY=$(find /boot/loader/entries/ -name "*.conf" 2>/dev/null | head -1)
if [[ -n "$LOADER_ENTRY" ]]; then
    if ! grep -q "splash" "$LOADER_ENTRY" 2>/dev/null; then
        info "Adding 'splash' to kernel cmdline..."
        sudo sed -i 's/^options.*/& splash/' "$LOADER_ENTRY"
        ok "Splash added to boot entry"
    fi
fi

# Regenerate initramfs
info "Regenerating initramfs..."
sudo mkinitcpio -P
ok "Initramfs regenerated"

# ── 6. Create wallpaper directory ────────────────────────────
echo ""
info "Setting up wallpaper directory..."
mkdir -p "$HOME/Pictures/Wallpapers"
mkdir -p "$HOME/Pictures/Screenshots"
mkdir -p "$HOME/Videos/Recordings"
mkdir -p "$HOME/.cache/sumi"
ok "Created ~/Pictures/Wallpapers — drop your wallpapers here"
ok "Created ~/Pictures/Screenshots & ~/Videos/Recordings"

# ── 7. Enable services ──────────────────────────────────────
info "Enabling services..."
sudo systemctl enable bluetooth.service 2>/dev/null || true
sudo systemctl enable NetworkManager.service 2>/dev/null || true
ok "Core services enabled"

# Enable user services (cliphist, wallust watcher)
info "Enabling user services..."
systemctl --user daemon-reload 2>/dev/null || true
systemctl --user enable cliphist.service 2>/dev/null || true
systemctl --user enable wallust-watcher.service 2>/dev/null || true
systemctl --user enable lock-before-sleep.service 2>/dev/null || true
systemctl --user enable sumi-cleanup.timer 2>/dev/null || true
systemctl --user start sumi-cleanup.timer 2>/dev/null || true
ok "User services enabled (cliphist, wallust-watcher, lock-before-sleep, cleanup-timer)"

# ── 7a. Framework 13 AMD specific services ──────────────────
echo ""
info "Setting up Framework 13 AMD hardware..."

# Power Profiles Daemon (recommended over TLP for AMD 7040)
sudo systemctl enable power-profiles-daemon.service 2>/dev/null || true
# Make sure TLP is NOT running (conflicts with PPD on AMD)
sudo systemctl disable tlp.service 2>/dev/null || true
sudo systemctl mask tlp.service 2>/dev/null || true
ok "power-profiles-daemon enabled (TLP disabled — AMD 7040 recommendation)"

# Fingerprint reader
sudo systemctl enable fprintd.service 2>/dev/null || true
ok "fprintd enabled"

# Firmware updates
sudo systemctl enable fwupd.service 2>/dev/null || true
ok "fwupd enabled (for Framework firmware updates)"

# Framework laptop kernel module
if pacman -Qi framework-laptop-kmod-dkms-git &>/dev/null; then
    # Ensure cros_ec modules load at boot
    echo -e "cros_ec\ncros_ec_lpcs" | sudo tee /etc/modules-load.d/framework.conf > /dev/null
    ok "framework-laptop-kmod configured (battery charge limit, LEDs)"
else
    warn "framework-laptop-kmod not installed — install from AUR for charge limit control"
fi

# Enroll fingerprint
echo ""
info "Fingerprint enrollment..."
echo -e "${DIM}   Would you like to enroll a fingerprint now? [y/N]${RST}"
read -r fp_answer
if [[ "$fp_answer" =~ ^[Yy]$ ]]; then
    fprintd-enroll || warn "Fingerprint enrollment failed — try again after reboot"
fi

# Set charge limit to 80% for battery longevity
if [[ -f /sys/class/power_supply/BAT1/charge_control_end_threshold ]]; then
    echo 80 | sudo tee /sys/class/power_supply/BAT1/charge_control_end_threshold > /dev/null
    ok "Battery charge limit set to 80% (recommended for longevity)"
fi

# AMD-specific kernel parameters for better s2idle
LOADER_ENTRY_FW=$(find /boot/loader/entries/ -name "*.conf" 2>/dev/null | head -1)
if [[ -n "$LOADER_ENTRY_FW" ]]; then
    if ! grep -q "amd_pstate=active" "$LOADER_ENTRY_FW" 2>/dev/null; then
        info "Adding AMD power tuning to kernel cmdline..."
        sudo sed -i 's/^options.*/& amd_pstate=active rtc_cmos.use_acpi_alarm=1/' "$LOADER_ENTRY_FW"
        ok "AMD pstate=active and RTC alarm fix added"
    fi
fi

# ── 8. Setup Zsh + Shell Environment ─────────────────────────
echo ""
info "Setting up Zsh shell environment..."

# Change default shell to zsh
if command -v zsh &>/dev/null; then
    if [[ "$SHELL" != "$(which zsh)" ]]; then
        chsh -s "$(which zsh)"
        ok "Default shell changed to zsh"
    else
        ok "Shell already set to zsh"
    fi
fi

# Deploy .zshrc
deploy "$SCRIPT_DIR/zsh/.zshrc" "$HOME/.zshrc"
ok "Zsh config deployed with TUI aliases"

# ── 9. Disable other display managers ────────────────────────
info "Checking for conflicting display managers..."
for dm in gdm sddm lightdm ly; do
    if systemctl is-enabled "$dm.service" 2>/dev/null | grep -q "enabled"; then
        sudo systemctl disable "$dm.service" 2>/dev/null
        warn "Disabled conflicting DM: $dm"
    fi
done

# ── Done ─────────────────────────────────────────────────────
echo ""
echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${GRN}║   sumi :: framework 13 ready       ║${RST}"
echo -e "${DIM}╠══════════════════════════════════════╣${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  1. Drop wallpapers in:              ${DIM}║${RST}"
echo -e "${DIM}║${RST}     ~/Pictures/Wallpapers/           ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  2. Reboot for greetd + plymouth     ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  3. SUPER+X for TUI control center   ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  TUI apps:                           ${DIM}║${RST}"
echo -e "${DIM}║${RST}  SUPER+E files  SUPER+I wifi         ${DIM}║${RST}"
echo -e "${DIM}║${RST}  SUPER+B blue   SUPER+A audio        ${DIM}║${RST}"
echo -e "${DIM}║${RST}  SUPER+G git    SUPER+T monitor      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  SUPER+M cava   SUPER+X control      ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Pickers & tools:                    ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+SH+V clip   S+Tab   windows      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+.    emoji  S+SH+N  notifs       ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+SH+S shot   S+AL+S  shot-pick    ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+AL+R rec    S+SH+E  power        ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+SH+W wall   S+SH+T  theme        ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+=    calc   S+/     cheatsheet   ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Submaps (modal):                    ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+CT+R → resize (hjkl, Esc exit)   ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+W    → group  (g/h/l/n/p/o)      ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Modes:                              ${DIM}║${RST}"
echo -e "${DIM}║${RST}  F5 gaming  F6 focus  F7 monitor     ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Scratchpads: F1 term  F2 music      ${DIM}║${RST}"
echo -e "${DIM}║${RST}               F3 monitor F4 devterm  ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Dev tools:                          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+SH+P projects S+SH+G worktree    ${DIM}║${RST}"
echo -e "${DIM}║${RST}  S+AL+Ret tmux   nvim has LSP+cmp   ${DIM}║${RST}"
echo -e "${DIM}║${RST}  tmux: C-a g=git f=files t=btop     ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Framework 13 AMD:                   ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • Fingerprint on lockscreen         ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • pwr: in bar cycles profile        ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • Charge capped at 80%              ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
echo ""
echo -e "${CYN}:: Reboot now? [y/N]${RST}"
read -r answer
if [[ "$answer" =~ ^[Yy]$ ]]; then
    sudo reboot
fi
