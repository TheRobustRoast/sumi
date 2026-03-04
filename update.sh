#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi — Config Update Script                               ║
# ║                                                              ║
# ║  Pulls latest changes and redeploys configs.                 ║
# ║  Does NOT reinstall packages, touch greetd, or Plymouth.     ║
# ║                                                              ║
# ║  Usage: ./update.sh                                          ║
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
echo -e "${CYN}║   sumi :: config update             ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
echo ""

info()  { echo -e "${CYN}:: ${RST}$1"; }
ok()    { echo -e "${GRN}   ✓${RST} $1"; }
warn()  { echo -e "${RED}   !${RST} $1"; }

deploy() {
    local src="$1"
    local dst="$2"
    mkdir -p "$(dirname "$dst")"
    cp -r "$src" "$dst"
    ok "Updated: $dst"
}

# ── Pre-flight ──────────────────────────────────────────────
if [[ "$EUID" -eq 0 ]]; then
    warn "Do not run as root."
    exit 1
fi

if [[ ! -d "$SCRIPT_DIR/.git" ]]; then
    warn "Not a git repo — cannot pull updates."
    exit 1
fi

# ── Pull latest ─────────────────────────────────────────────
info "Pulling latest from origin..."
git -C "$SCRIPT_DIR" pull --ff-only || {
    warn "git pull failed — resolve conflicts manually, then re-run."
    exit 1
}
ok "Repository up to date"
echo ""

# ── Redeploy configs ────────────────────────────────────────
info "Redeploying configuration files..."

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

# Foot
deploy "$SCRIPT_DIR/foot/foot.ini"           "$CONFIG_DIR/foot/foot.ini"

# Fuzzel
deploy "$SCRIPT_DIR/fuzzel/fuzzel.ini"       "$CONFIG_DIR/fuzzel/fuzzel.ini"

# Dunst
deploy "$SCRIPT_DIR/dunst/dunstrc"           "$CONFIG_DIR/dunst/dunstrc"

# Yazi
deploy "$SCRIPT_DIR/yazi/yazi.toml"          "$CONFIG_DIR/yazi/yazi.toml"
deploy "$SCRIPT_DIR/yazi/theme.toml"         "$CONFIG_DIR/yazi/theme.toml"

# Cava
deploy "$SCRIPT_DIR/cava/config"             "$CONFIG_DIR/cava/config"

# Lazygit
deploy "$SCRIPT_DIR/lazygit/config.yml"      "$CONFIG_DIR/lazygit/config.yml"

# Neovim
deploy "$SCRIPT_DIR/nvim/init.lua"           "$CONFIG_DIR/nvim/init.lua"

# Tmux
deploy "$SCRIPT_DIR/tmux/tmux.conf"          "$HOME/.tmux.conf"

# Starship
deploy "$SCRIPT_DIR/starship/starship.toml"  "$CONFIG_DIR/starship.toml"

# Btop
deploy "$SCRIPT_DIR/btop/btop.conf"             "$CONFIG_DIR/btop/btop.conf"
mkdir -p "$CONFIG_DIR/btop/themes"
deploy "$SCRIPT_DIR/btop/themes/sumi.theme"  "$CONFIG_DIR/btop/themes/sumi.theme"

# Wallust
deploy "$SCRIPT_DIR/wallust/wallust.toml"    "$CONFIG_DIR/wallust/wallust.toml"
mkdir -p "$CONFIG_DIR/wallust/templates"
for f in "$SCRIPT_DIR/wallust/templates/"*; do
    deploy "$f" "$CONFIG_DIR/wallust/templates/$(basename "$f")"
done

# GTK
deploy "$SCRIPT_DIR/gtk-3.0/settings.ini"   "$CONFIG_DIR/gtk-3.0/settings.ini"
deploy "$SCRIPT_DIR/gtk-4.0/settings.ini"   "$CONFIG_DIR/gtk-4.0/settings.ini"

# XDG
deploy "$SCRIPT_DIR/xdg/mimeapps.list"      "$CONFIG_DIR/mimeapps.list"
deploy "$SCRIPT_DIR/xdg/hyprland-portals.conf" "$CONFIG_DIR/xdg-desktop-portal/portals.conf"

# Zsh
deploy "$SCRIPT_DIR/zsh/.zshrc"             "$HOME/.zshrc"

# Cursor
mkdir -p "$HOME/.icons/default"
deploy "$SCRIPT_DIR/icons/default/index.theme" "$HOME/.icons/default/index.theme"

# Systemd user services
mkdir -p "$HOME/.config/systemd/user"
for f in "$SCRIPT_DIR/systemd/user/"*; do
    deploy "$f" "$HOME/.config/systemd/user/$(basename "$f")"
done
systemctl --user daemon-reload 2>/dev/null || true

echo ""

# ── Reload Hyprland if running ──────────────────────────────
if command -v hyprctl &>/dev/null && hyprctl version &>/dev/null 2>&1; then
    info "Reloading Hyprland..."
    hyprctl reload
    ok "Hyprland reloaded"
else
    info "Hyprland not running — changes will apply on next login"
fi

echo ""
echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${GRN}║   sumi :: update complete           ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
