#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Wallpaper Apply — Set wallpaper + regenerate theme          ║
# ║  Usage: wallpaper-apply.sh /path/to/image.png               ║
# ╚══════════════════════════════════════════════════════════════╝

set -uo pipefail  # no -e: handle errors gracefully

WALLPAPER="${1:-}"
CACHE_DIR="$HOME/.cache/sumi"

mkdir -p "$CACHE_DIR"

# ── Validate input ─────────────────────────────────────────────
if [[ -z "$WALLPAPER" ]]; then
    echo "Error: No wallpaper path provided"
    echo "Usage: wallpaper-apply.sh /path/to/image.png"
    exit 1
fi

if [[ ! -f "$WALLPAPER" ]]; then
    notify-send -a "sumi" -u critical "Wallpaper error" "File not found: $WALLPAPER" 2>/dev/null
    exit 1
fi

# Resolve to absolute path
WALLPAPER="$(realpath "$WALLPAPER")"
echo ":: Applying wallpaper: $WALLPAPER"

# ── Save current wallpaper path ──────────────────────────────
echo "$WALLPAPER" > "$CACHE_DIR/current-wallpaper"

# ── Set wallpaper via hyprpaper ──────────────────────────────
_need_restart=false

if pgrep -x hyprpaper > /dev/null; then
    # Try IPC first (faster, no flicker)
    if hyprctl hyprpaper preload "$WALLPAPER" 2>/dev/null \
       && hyprctl hyprpaper wallpaper ",$WALLPAPER" 2>/dev/null; then
        # Unload old wallpapers to free VRAM
        hyprctl hyprpaper unload unused 2>/dev/null || true
        echo ":: Wallpaper set via IPC"
    else
        echo ":: IPC failed, restarting hyprpaper"
        _need_restart=true
    fi
else
    _need_restart=true
fi

if [[ "$_need_restart" == "true" ]]; then
    pkill hyprpaper 2>/dev/null || true
    sleep 0.3

    cat > "$HOME/.config/hypr/hyprpaper.conf" <<EOF
preload = $WALLPAPER
wallpaper = ,$WALLPAPER
splash = false
ipc = on
EOF

    hyprpaper &
    disown

    # Wait for hyprpaper to be ready (up to 2s)
    for _ in {1..20}; do
        pgrep -x hyprpaper > /dev/null && break
        sleep 0.1
    done
fi

# ── Generate color scheme with wallust ───────────────────────
echo ":: Running wallust..."
if ! wallust run "$WALLPAPER" 2>/dev/null; then
    echo ":: wallust failed — keeping existing colors"
    notify-send -a "sumi" -u low "Theme" "wallust failed — colors unchanged" 2>/dev/null
fi

# ── Reload waybar ──────────────────────────────────────────────
echo ":: Reloading waybar..."
if pgrep -x waybar > /dev/null; then
    pkill waybar 2>/dev/null || true
    sleep 0.3
fi
waybar &
disown

# ── Reload dunst ───────────────────────────────────────────────
echo ":: Reloading dunst..."
_was_paused=false
if dunstctl is-paused 2>/dev/null | grep -q true; then
    _was_paused=true
fi
pkill dunst 2>/dev/null || true
sleep 0.2
dunst &
disown
# Restore DND state if it was paused (focus/gaming mode)
if [[ "$_was_paused" == "true" ]]; then
    sleep 0.3
    dunstctl set-paused true 2>/dev/null
fi

# ── Reload cava if running ───────────────────────────────────
if pgrep -x cava > /dev/null; then
    echo ":: Reloading cava..."
    pkill -USR1 cava 2>/dev/null || true
fi

# ── Btop theme note ────────────────────────────────────────────
if [[ -f "$HOME/.config/btop/themes/wallust.theme" ]]; then
    echo ":: btop theme updated (applies on next launch)"
fi

# ── Notify ─────────────────────────────────────────────────────
sleep 0.3
BASENAME=$(basename "$WALLPAPER")
notify-send -a "sumi" -t 3000 -i "$WALLPAPER" \
    "[ wallpaper ]" "applied: $BASENAME" 2>/dev/null

echo ":: Done!"
