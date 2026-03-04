#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi — rice installer                                       ║
# ║                                                              ║
# ║  Run as your normal user after the first boot.               ║
# ║  Usage:  ~/sumi/install.sh                                   ║
# ║  Update: sumi-update  (git pull + re-run)                    ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Catppuccin Mocha gum env ─────────────────────────────────────
export GUM_INPUT_CURSOR_FOREGROUND="#f38ba8"
export GUM_INPUT_PROMPT_FOREGROUND="#cba6f7"
export GUM_CONFIRM_SELECTED_BACKGROUND="#a6e3a1"
export GUM_CONFIRM_SELECTED_FOREGROUND="#1e1e2e"
export GUM_CONFIRM_UNSELECTED_BACKGROUND="#313244"
export GUM_CONFIRM_UNSELECTED_FOREGROUND="#cdd6f4"
export GUM_SPIN_SPINNER="dot"
export GUM_SPIN_SPINNER_FOREGROUND="#cba6f7"
export GUM_SPIN_TITLE_FOREGROUND="#cdd6f4"

# ── Helpers ──────────────────────────────────────────────────────
s_step()    { gum style --foreground '#89b4fa' "  · $1"; }
s_ok()      { gum style --foreground '#a6e3a1' "  ✓  $1"; }
s_fail()    { gum style --foreground '#f38ba8' "  ✗  $1"; }
s_warn()    { gum style --foreground '#f9e2af' "  !  $1"; }
s_section() { echo ""; gum style --foreground '#cba6f7' --bold "  ── $1"; echo ""; }

# Symlink src → dst.  Re-running is always safe — just refreshes the link.
# Backs up real files/dirs on first encounter; replaces existing symlinks silently.
link() {
    local src="$1" dst="$2"
    mkdir -p "$(dirname "$dst")"
    if [[ -L "$dst" ]]; then
        rm -f "$dst"
    elif [[ -e "$dst" ]]; then
        local bak="${dst}.bak.$(date +%s)"
        mv "$dst" "$bak"
        s_warn "Backed up: $(basename "$dst") → $(basename "$bak")"
    fi
    ln -sf "$src" "$dst"
    s_ok "→ $(basename "$dst")"
}

# ── Welcome ──────────────────────────────────────────────────────
clear
gum style \
    --border rounded --border-foreground '#313244' \
    --margin "1 2" --padding "1 4" \
    "$(gum style --foreground '#f38ba8' --bold 'sumi') $(gum style --foreground '#45475a' '::') $(gum style --foreground '#cdd6f4' 'rice installer')" \
    "" \
    "$(gum style --foreground '#6c7086' 'Framework 13 AMD  ·  Hyprland  ·  Catppuccin Mocha')"
echo ""

# ── 0. Pre-flight ────────────────────────────────────────────────
s_section "Pre-flight"

[[ "$EUID" -eq 0 ]] && { s_fail "Run as your normal user, not root."; exit 1; }
[[ ! -f /etc/arch-release ]] && { s_fail "Arch Linux only."; exit 1; }

gum spin --title "  Checking internet..." -- ping -c 1 -W 5 archlinux.org || {
    s_fail "No internet connection."
    exit 1
}

for cmd in git sudo pacman; do
    command -v "$cmd" &>/dev/null && s_ok "$cmd found" || { s_fail "$cmd not found"; exit 1; }
done

s_step "Refreshing keyring..."
gum spin --title "  Updating archlinux-keyring..." -- \
    sudo pacman -Sy --noconfirm archlinux-keyring
gum spin --title "  Populating keys..." -- \
    sudo pacman-key --populate archlinux
s_ok "Keyring up to date"

s_step "Installing keyring hook..."
sudo mkdir -p /etc/pacman.d/hooks
sudo install -m 644 "$SCRIPT_DIR/pacman/hooks/keyring.hook" /etc/pacman.d/hooks/keyring.hook
s_ok "Keyring hook installed"

# ── 1. AUR helper (yay) ──────────────────────────────────────────
s_section "AUR Helper"

if ! command -v yay &>/dev/null; then
    s_step "Installing yay..."
    gum spin --title "  Cloning yay-bin..." -- \
        git clone https://aur.archlinux.org/yay-bin.git /tmp/yay-bin
    (cd /tmp/yay-bin && makepkg -si --noconfirm)
    rm -rf /tmp/yay-bin
    s_ok "yay installed"
else
    s_ok "yay already installed"
fi

# ── 2. AUR packages ──────────────────────────────────────────────
s_section "AUR Packages"

s_step "Installing AUR packages..."
gum spin --title "  yay (this may take a while)..." -- \
    yay -S --needed --noconfirm \
        wallust \
        bluetuith \
        framework-laptop-kmod-dkms-git \
        bibata-cursor-theme \
    || s_warn "Some AUR packages failed — check output above"
s_ok "AUR packages done"

# ── 3. Extra pacman packages ─────────────────────────────────────
s_section "Extra Packages"

gum spin --title "  pacman --needed..." -- \
    sudo pacman -S --needed --noconfirm \
        shellcheck shfmt \
        docker docker-compose \
        lsof direnv rsync \
    || s_warn "Some packages failed — check output above"
s_ok "Packages up to date"

# ── 4. Link configs ──────────────────────────────────────────────
s_section "Configs"
s_step "Linking dotfiles (symlinks — git pull = instant update)..."

# Hyprland configs (individual files so ~/.config/hypr/ stays a real dir,
# allowing other apps / scripts to create files alongside these).
mkdir -p "$HOME/.config/hypr/conf.d"
link "$SCRIPT_DIR/hypr/hyprland.conf"  "$HOME/.config/hypr/hyprland.conf"
link "$SCRIPT_DIR/hypr/hyprlock.conf"  "$HOME/.config/hypr/hyprlock.conf"
link "$SCRIPT_DIR/hypr/hypridle.conf"  "$HOME/.config/hypr/hypridle.conf"
link "$SCRIPT_DIR/hypr/hyprpaper.conf" "$HOME/.config/hypr/hyprpaper.conf"
for f in "$SCRIPT_DIR/hypr/conf.d/"*; do
    link "$f" "$HOME/.config/hypr/conf.d/$(basename "$f")"
done

# Scripts: symlinked as a dir so the whole folder updates on git pull.
# Keybinds reference $HOME/.config/hypr/scripts/...
link "$SCRIPT_DIR/scripts" "$HOME/.config/hypr/scripts"

# Self-contained config dirs — symlink the whole dir.
link "$SCRIPT_DIR/waybar"   "$HOME/.config/waybar"
link "$SCRIPT_DIR/yazi"     "$HOME/.config/yazi"
link "$SCRIPT_DIR/btop"     "$HOME/.config/btop"
link "$SCRIPT_DIR/wallust"  "$HOME/.config/wallust"
link "$SCRIPT_DIR/nvim"     "$HOME/.config/nvim"

# Single config files.
link "$SCRIPT_DIR/foot/foot.ini"               "$HOME/.config/foot/foot.ini"
# Copy (not symlink) so wallust can overwrite it on first wallpaper set
[[ ! -f "$HOME/.config/foot/colors.ini" ]] && \
    cp "$SCRIPT_DIR/foot/colors.ini" "$HOME/.config/foot/colors.ini" && s_ok "→ colors.ini (default)"
link "$SCRIPT_DIR/fuzzel/fuzzel.ini"           "$HOME/.config/fuzzel/fuzzel.ini"
link "$SCRIPT_DIR/dunst/dunstrc"               "$HOME/.config/dunst/dunstrc"
link "$SCRIPT_DIR/cava/config"                 "$HOME/.config/cava/config"
link "$SCRIPT_DIR/lazygit/config.yml"          "$HOME/.config/lazygit/config.yml"
link "$SCRIPT_DIR/gtk-3.0/settings.ini"        "$HOME/.config/gtk-3.0/settings.ini"
link "$SCRIPT_DIR/gtk-4.0/settings.ini"        "$HOME/.config/gtk-4.0/settings.ini"
link "$SCRIPT_DIR/xdg/mimeapps.list"           "$HOME/.config/mimeapps.list"
link "$SCRIPT_DIR/xdg/hyprland-portals.conf"   "$HOME/.config/xdg-desktop-portal/portals.conf"
link "$SCRIPT_DIR/icons/default/index.theme"   "$HOME/.icons/default/index.theme"
link "$SCRIPT_DIR/starship/starship.toml"      "$HOME/.config/starship.toml"
link "$SCRIPT_DIR/tmux/tmux.conf"              "$HOME/.tmux.conf"
link "$SCRIPT_DIR/zsh/.zshrc"                  "$HOME/.zshrc"
[[ -f "$SCRIPT_DIR/.editorconfig" ]] && \
    link "$SCRIPT_DIR/.editorconfig"           "$HOME/.editorconfig"

# Systemd user services.
for f in "$SCRIPT_DIR/systemd/user/"*; do
    link "$f" "$HOME/.config/systemd/user/$(basename "$f")"
done

# bin/ commands → ~/.local/bin/ (ensures sumi-update etc. are in PATH).
if [[ -d "$SCRIPT_DIR/bin" ]]; then
    mkdir -p "$HOME/.local/bin"
    for f in "$SCRIPT_DIR/bin/"*; do
        chmod +x "$f"
        link "$f" "$HOME/.local/bin/$(basename "$f")"
    done
fi

# Wallust color seed files — copy once if missing (wallust overwrites later).
[[ ! -f "$HOME/.config/waybar/colors.css" ]] && \
    cp "$SCRIPT_DIR/wallust/templates/colors-waybar.css" \
       "$HOME/.config/waybar/colors.css" 2>/dev/null || true
[[ ! -f "$HOME/.config/fuzzel/colors.ini" ]] && \
    cp "$SCRIPT_DIR/wallust/templates/colors-fuzzel.ini" \
       "$HOME/.config/fuzzel/colors.ini" 2>/dev/null || true

s_ok "All dotfiles linked"

# ── 5. greetd ────────────────────────────────────────────────────
s_section "Display Manager"

[[ -f /etc/greetd/config.toml ]] && \
    sudo cp /etc/greetd/config.toml /etc/greetd/config.toml.bak 2>/dev/null || true

sudo tee /etc/greetd/config.toml > /dev/null << GREETDCFG
[terminal]
vt = 1

# Cold-boot autologin: LUKS passphrase is authentication — no second prompt.
[initial_session]
command = "$HOME/.config/hypr/scripts/hyprland-wrapped.sh"
user = "$USER"

# After logout: interactive tuigreet login.
[default_session]
command = "tuigreet --time --remember --remember-session --asterisks --cmd '$HOME/.config/hypr/scripts/hyprland-wrapped.sh'"
user = "greeter"
GREETDCFG

sudo systemctl enable greetd.service 2>/dev/null || true
s_ok "greetd configured (LUKS unlock → autologin as $USER)"

# ── 6. Plymouth ──────────────────────────────────────────────────
s_section "Plymouth"

sudo mkdir -p /usr/share/plymouth/themes/hypr-tui
sudo cp "$SCRIPT_DIR/plymouth/themes/hypr-tui/"* \
    /usr/share/plymouth/themes/hypr-tui/
sudo plymouth-set-default-theme hypr-tui 2>/dev/null || \
    s_warn "Plymouth theme set failed — run manually"

if grep -q "^HOOKS=" /etc/mkinitcpio.conf && \
   ! grep -q "plymouth" /etc/mkinitcpio.conf; then
    sudo sed -i 's/^HOOKS=(\(.*\)udev\(.*\))/HOOKS=(\1udev plymouth\2)/' \
        /etc/mkinitcpio.conf
    s_ok "Plymouth hook added to mkinitcpio"
fi

LOADER_ENTRY=$(find /boot/loader/entries/ -name "*.conf" 2>/dev/null | head -1 || true)
if [[ -n "$LOADER_ENTRY" ]] && ! grep -q "splash" "$LOADER_ENTRY" 2>/dev/null; then
    sudo sed -i 's/^options.*/& splash/' "$LOADER_ENTRY"
    s_ok "splash added to kernel cmdline"
fi

gum spin --title "  Regenerating initramfs..." -- sudo mkinitcpio -P
s_ok "Initramfs regenerated"

# ── 7. Directories ───────────────────────────────────────────────
s_section "Directories"

mkdir -p \
    "$HOME/Pictures/Wallpapers" \
    "$HOME/Pictures/Screenshots" \
    "$HOME/Videos/Recordings" \
    "$HOME/.cache/sumi" \
    "$HOME/.local/share/sumi" \
    "$HOME/.local/bin"
s_ok "~/Pictures/Wallpapers  ~/Pictures/Screenshots  ~/Videos/Recordings"

# ── 8. Services ──────────────────────────────────────────────────
s_section "Services"

sudo systemctl enable \
    bluetooth.service \
    NetworkManager.service \
    power-profiles-daemon.service \
    fprintd.service \
    fwupd.service \
    2>/dev/null || true

# TLP conflicts with power-profiles-daemon on AMD 7040.
sudo systemctl disable tlp.service 2>/dev/null || true
sudo systemctl mask    tlp.service 2>/dev/null || true

systemctl --user daemon-reload 2>/dev/null || true
systemctl --user enable \
    cliphist.service \
    wallust-watcher.service \
    lock-before-sleep.service \
    sumi-cleanup.timer \
    2>/dev/null || true

s_ok "System and user services enabled"

# ── 9. Framework 13 AMD hardware ─────────────────────────────────
s_section "Framework 13 Hardware"

if pacman -Qi framework-laptop-kmod-dkms-git &>/dev/null; then
    printf 'cros_ec\ncros_ec_lpcs\n' | \
        sudo tee /etc/modules-load.d/framework.conf > /dev/null
    s_ok "framework-laptop-kmod configured (charge limit, LEDs)"
fi

if [[ -f /sys/class/power_supply/BAT1/charge_control_end_threshold ]]; then
    echo 80 | sudo tee \
        /sys/class/power_supply/BAT1/charge_control_end_threshold > /dev/null
    printf 'w /sys/class/power_supply/BAT1/charge_control_end_threshold - - - - 80\n' | \
        sudo tee /etc/tmpfiles.d/battery-charge-limit.conf > /dev/null
    s_ok "Battery charge limit → 80%"
fi

LOADER_FW=$(find /boot/loader/entries/ -name "*.conf" 2>/dev/null | head -1 || true)
if [[ -n "$LOADER_FW" ]] && ! grep -q "amd_pstate=active" "$LOADER_FW" 2>/dev/null; then
    sudo sed -i 's/^options.*/& amd_pstate=active rtc_cmos.use_acpi_alarm=1/' "$LOADER_FW"
    s_ok "AMD pstate=active + RTC alarm fix added"
fi

# ── 10. Shell ────────────────────────────────────────────────────
s_section "Shell"

if command -v zsh &>/dev/null && [[ "$SHELL" != "$(command -v zsh)" ]]; then
    chsh -s "$(command -v zsh)"
    s_ok "Default shell → zsh"
else
    s_ok "Shell already zsh"
fi

# ── 11. Disable conflicting display managers ─────────────────────
for dm in gdm sddm lightdm ly; do
    if systemctl is-enabled "$dm.service" 2>/dev/null | grep -q "enabled"; then
        sudo systemctl disable "$dm.service" 2>/dev/null
        s_warn "Disabled conflicting DM: $dm"
    fi
done

# ── 12. Migrations ───────────────────────────────────────────────
s_section "Migrations"

MIGRATIONS_DIR="$SCRIPT_DIR/migrations"
APPLIED_FILE="$HOME/.local/share/sumi/applied-migrations"
mkdir -p "$HOME/.local/share/sumi"
touch "$APPLIED_FILE"

ran=0
if [[ -d "$MIGRATIONS_DIR" ]]; then
    for migration in "$MIGRATIONS_DIR"/[0-9]*.sh; do
        [[ -f "$migration" ]] || continue
        num=$(basename "$migration" | grep -oE '^[0-9]+')
        grep -qxF "$num" "$APPLIED_FILE" && continue
        s_step "$(basename "$migration")..."
        if bash "$migration"; then
            echo "$num" >> "$APPLIED_FILE"
            s_ok "Applied: $(basename "$migration")"
            ((ran++)) || true
        else
            s_fail "Migration failed: $(basename "$migration")"
            exit 1
        fi
    done
fi
[[ $ran -eq 0 ]] && s_ok "No pending migrations" || s_ok "$ran migration(s) applied"

# ── 13. Fingerprint ──────────────────────────────────────────────
echo ""
if gum confirm "  Enroll fingerprint now?" --default=false; then
    fprintd-enroll || s_warn "Fingerprint enrollment failed — retry after reboot"
fi

# ── Done ─────────────────────────────────────────────────────────
echo ""
gum style \
    --border rounded --border-foreground '#a6e3a1' \
    --margin "1 2" --padding "1 4" \
    "$(gum style --foreground '#a6e3a1' --bold '  ✓  sumi installed')" \
    "" \
    "  $(gum style --foreground '#cba6f7' '1.')  Drop wallpapers in $(gum style --foreground '#f38ba8' '~/Pictures/Wallpapers/')" \
    "  $(gum style --foreground '#cba6f7' '2.')  Reboot → LUKS unlock → autologin → Hyprland" \
    "  $(gum style --foreground '#cba6f7' '3.')  SUPER+X  control center    SUPER+/  keybinds" \
    "" \
    "  To update later: $(gum style --foreground '#89b4fa' 'sumi-update')"
echo ""
gum confirm "  Reboot now?" && sudo reboot || true
