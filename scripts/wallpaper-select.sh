#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Wallpaper Select — Pick wallpaper via fuzzel                ║
# ╚══════════════════════════════════════════════════════════════╝

WALLPAPER_DIR="$HOME/Pictures/Wallpapers"

# List wallpapers
SELECTION=$(find "$WALLPAPER_DIR" -type f \( -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" -o -name "*.webp" \) | sort | while read -r f; do
    basename "$f"
done | fuzzel --dmenu --prompt="wallpaper> " --width=25 --lines=15)

if [[ -n "$SELECTION" ]]; then
    FULL_PATH=$(find "$WALLPAPER_DIR" -name "$SELECTION" -type f | head -1)
    if [[ -n "$FULL_PATH" ]]; then
        "$HOME/.config/hypr/scripts/wallpaper-apply.sh" "$FULL_PATH"
    fi
fi
