#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi — Uninstall Script                                 ║
# ║                                                              ║
# ║  Removes all sumi configs and restores defaults.          ║
# ║  Does NOT uninstall packages (you may still want Hyprland). ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

RED='\033[0;31m'
GRN='\033[0;32m'
CYN='\033[0;36m'
DIM='\033[0;90m'
YLW='\033[0;33m'
RST='\033[0m'

CONFIG_DIR="$HOME/.config"

echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${RED}║   sumi :: uninstall                ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
echo ""

info()  { echo -e "${CYN}:: ${RST}$1"; }
ok()    { echo -e "${GRN}   ✓${RST} $1"; }
warn()  { echo -e "${YLW}   !${RST} $1"; }
removed() { echo -e "${RED}   ✗${RST} Removed: $1"; }

# ── Safety check ────────────────────────────────────────────
echo -e "${RED}This will remove ALL sumi configuration files.${RST}"
echo -e "${DIM}Your wallpapers, screenshots, and recordings will be kept.${RST}"
echo ""
echo -e "${CYN}:: Are you sure? Type 'yes' to continue:${RST}"
read -r confirm
if [[ "$confirm" != "yes" ]]; then
    info "Aborted."
    exit 0
fi

echo ""

# ── Stop and disable user services ──────────────────────────
info "Stopping user services..."
systemctl --user stop cliphist.service 2>/dev/null || true
systemctl --user stop wallust-watcher.service 2>/dev/null || true
systemctl --user stop sumi-cleanup.timer 2>/dev/null || true
systemctl --user disable cliphist.service 2>/dev/null || true
systemctl --user disable wallust-watcher.service 2>/dev/null || true
systemctl --user disable lock-before-sleep.service 2>/dev/null || true
systemctl --user disable sumi-cleanup.timer 2>/dev/null || true
ok "User services stopped and disabled"

# ── Remove systemd user units ───────────────────────────────
info "Removing systemd user units..."
UNITS=(
    cliphist.service
    wallust-watcher.service
    lock-before-sleep.service
    sumi-cleanup.service
    sumi-cleanup.timer
    hyprland-session.target
)
for unit in "${UNITS[@]}"; do
    if [[ -f "$CONFIG_DIR/systemd/user/$unit" ]]; then
        rm "$CONFIG_DIR/systemd/user/$unit"
        removed "$unit"
    fi
done
systemctl --user daemon-reload 2>/dev/null || true

# ── Remove sumi config directories ───────────────────────
info "Removing configuration files..."

CONFIGS=(
    "$CONFIG_DIR/hypr"
    "$CONFIG_DIR/waybar"
    "$CONFIG_DIR/foot"
    "$CONFIG_DIR/fuzzel"
    "$CONFIG_DIR/dunst"
    "$CONFIG_DIR/wallust"
    "$CONFIG_DIR/yazi"
    "$CONFIG_DIR/cava"
    "$CONFIG_DIR/lazygit"
    "$CONFIG_DIR/btop"
    "$CONFIG_DIR/starship.toml"
)

for cfg in "${CONFIGS[@]}"; do
    if [[ -e "$cfg" ]]; then
        # Check for .bak files first
        if [[ -d "$cfg" ]]; then
            # Restore most recent backup if it exists
            latest_bak=$(find "$cfg" -name "*.bak.*" -maxdepth 1 2>/dev/null | sort -r | head -1)
            if [[ -n "$latest_bak" ]]; then
                warn "Backup found at $latest_bak — keeping backups"
            fi
            rm -rf "$cfg"
        else
            rm -f "$cfg"
        fi
        removed "$cfg"
    fi
done

# ── Remove GTK/cursor theme overrides ───────────────────────
info "Removing theme overrides..."
rm -f "$CONFIG_DIR/gtk-3.0/settings.ini" 2>/dev/null && removed "gtk-3.0/settings.ini"
rm -f "$CONFIG_DIR/gtk-4.0/settings.ini" 2>/dev/null && removed "gtk-4.0/settings.ini"
rm -f "$CONFIG_DIR/mimeapps.list" 2>/dev/null && removed "mimeapps.list"
rm -rf "$CONFIG_DIR/xdg-desktop-portal" 2>/dev/null && removed "xdg-desktop-portal"
rm -rf "$HOME/.icons/default" 2>/dev/null && removed ".icons/default"

# ── Remove nvim config (careful — user may have their own) ──
if [[ -f "$CONFIG_DIR/nvim/init.lua" ]]; then
    # Only remove if it's our file (check for sumi marker)
    if grep -q "sumi" "$CONFIG_DIR/nvim/init.lua" 2>/dev/null; then
        rm -f "$CONFIG_DIR/nvim/init.lua"
        removed "nvim/init.lua (sumi config)"
    else
        warn "Skipping nvim/init.lua — appears to be user's own config"
    fi
fi

# ── Remove .zshrc (careful) ────────────────────────────────
if [[ -f "$HOME/.zshrc" ]]; then
    if grep -q "sumi" "$HOME/.zshrc" 2>/dev/null; then
        rm -f "$HOME/.zshrc"
        removed ".zshrc (sumi config)"
    else
        warn "Skipping .zshrc — appears to be user's own config"
    fi
fi

# ── Remove .tmux.conf (careful) ───────────────────────────
if [[ -f "$HOME/.tmux.conf" ]]; then
    if grep -q "sumi" "$HOME/.tmux.conf" 2>/dev/null; then
        rm -f "$HOME/.tmux.conf"
        removed ".tmux.conf (sumi config)"
    else
        warn "Skipping .tmux.conf — appears to be user's own config"
    fi
fi

# ── Remove .editorconfig ────────────────────────────────────
if [[ -f "$HOME/.editorconfig" ]]; then
    rm -f "$HOME/.editorconfig"
    removed ".editorconfig"
fi

# ── Remove battery charge limit tmpfile ─────────────────────
if [[ -f /etc/tmpfiles.d/battery-charge-limit.conf ]]; then
    sudo rm -f /etc/tmpfiles.d/battery-charge-limit.conf 2>/dev/null && removed "battery-charge-limit.conf"
fi

# ── Clean cache ─────────────────────────────────────────────
info "Cleaning sumi cache..."
rm -rf "$HOME/.cache/sumi" 2>/dev/null && removed ".cache/sumi"

# ── Plymouth (optional — requires sudo) ────────────────────
echo ""
echo -e "${CYN}:: Remove Plymouth theme and greetd config? (requires sudo) [y/N]${RST}"
read -r sys_answer
if [[ "$sys_answer" =~ ^[Yy]$ ]]; then
    sudo rm -rf /usr/share/plymouth/themes/hypr-tui 2>/dev/null && removed "Plymouth hypr-tui theme"
    if [[ -f /etc/greetd/config.toml.bak ]]; then
        sudo mv /etc/greetd/config.toml.bak /etc/greetd/config.toml
        ok "Restored greetd config from backup"
    fi
fi

# ── Done ─────────────────────────────────────────────────────
echo ""
echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${GRN}║   sumi :: uninstalled              ║${RST}"
echo -e "${DIM}╠══════════════════════════════════════╣${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Configs removed. Packages remain.   ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Wallpapers/screenshots untouched.  ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  To remove packages:                 ${DIM}║${RST}"
echo -e "${DIM}║${RST}  pacman -Rns hyprland waybar ...     ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                      ${DIM}║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
