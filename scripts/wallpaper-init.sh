#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Wallpaper Init — Load last wallpaper or default on login    ║
# ╚══════════════════════════════════════════════════════════════╝

WALLPAPER_DIR="$HOME/Pictures/Wallpapers"
CACHE_DIR="$HOME/.cache/sumi"
CURRENT_FILE="$CACHE_DIR/current-wallpaper"

mkdir -p "$CACHE_DIR" "$WALLPAPER_DIR"

# Use cached wallpaper if it still exists
if [[ -f "$CURRENT_FILE" ]]; then
    CACHED="$(cat "$CURRENT_FILE")"
    if [[ -f "$CACHED" ]]; then
        WALLPAPER="$CACHED"
    fi
fi

# No cached wallpaper — find first image in wallpaper dir
if [[ -z "${WALLPAPER:-}" ]]; then
    WALLPAPER=$(find "$WALLPAPER_DIR" -maxdepth 1 -type f \
        \( -iname "*.png" -o -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.webp" -o -iname "*.bmp" \) \
        | sort | head -1)
fi

# Still nothing — generate a solid dark fallback
if [[ -z "$WALLPAPER" ]]; then
    FALLBACK="$WALLPAPER_DIR/default.png"
    if [[ ! -f "$FALLBACK" ]]; then
        if command -v magick &>/dev/null; then
            magick -size 3840x2160 xc:'#0a0a0a' "$FALLBACK"
        elif command -v convert &>/dev/null; then
            convert -size 3840x2160 xc:'#0a0a0a' "$FALLBACK"
        elif command -v ffmpeg &>/dev/null; then
            ffmpeg -f lavfi -i "color=c=0x0a0a0a:s=3840x2160:d=1" \
                -frames:v 1 "$FALLBACK" -y 2>/dev/null
        else
            echo ":: Warning: No image tools found. Drop a wallpaper in ~/Pictures/Wallpapers/"
            exit 0
        fi
    fi
    WALLPAPER="$FALLBACK"
fi

# Wait for Hyprland to be ready (up to 5s)
for _ in {1..50}; do
    hyprctl monitors -j &>/dev/null && break
    sleep 0.1
done

# Apply
exec "$HOME/.config/hypr/scripts/wallpaper-apply.sh" "$WALLPAPER"
