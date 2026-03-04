#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi — Update Script                                       ║
# ║                                                              ║
# ║  Configs are symlinked into the repo, so git pull is all    ║
# ║  that's needed to pick up every change instantly.           ║
# ║                                                              ║
# ║  Usage: ./update.sh                                          ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CYN='\033[0;36m'
GRN='\033[0;32m'
RED='\033[0;31m'
DIM='\033[0;90m'
RST='\033[0m'

info() { echo -e "${CYN}:: ${RST}$1"; }
ok()   { echo -e "${GRN}   ✓${RST} $1"; }
warn() { echo -e "${RED}   !${RST} $1"; }

echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${CYN}║   sumi :: update                    ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
echo ""

[[ "$EUID" -eq 0 ]] && { warn "Do not run as root."; exit 1; }
[[ ! -d "$SCRIPT_DIR/.git" ]] && { warn "Not a git repo."; exit 1; }

# ── Pull ─────────────────────────────────────────────────────
info "Pulling latest from origin..."
git -C "$SCRIPT_DIR" pull --ff-only || {
    warn "git pull failed — resolve conflicts manually, then re-run."
    exit 1
}
ok "Up to date"
echo ""

# ── Re-link configs (fixes broken symlinks after first install) ──
info "Refreshing symlinks..."
ln -sf "$SCRIPT_DIR/foot/foot.ini" "$HOME/.config/foot/foot.ini"
[[ ! -f "$HOME/.config/foot/colors.ini" ]] && \
    cp "$SCRIPT_DIR/foot/colors.ini" "$HOME/.config/foot/colors.ini"
ok "Symlinks up to date"
echo ""

# ── Reload services ──────────────────────────────────────────
systemctl --user daemon-reload 2>/dev/null || true

# ── Reload Hyprland ──────────────────────────────────────────
if command -v hyprctl &>/dev/null && hyprctl version &>/dev/null 2>&1; then
    info "Reloading Hyprland..."
    hyprctl reload
    ok "Hyprland reloaded"
else
    info "Hyprland not running — changes apply on next login"
fi

echo ""
echo -e "${DIM}╔══════════════════════════════════════╗${RST}"
echo -e "${GRN}║   sumi :: done                      ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════╝${RST}"
