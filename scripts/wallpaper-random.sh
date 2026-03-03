#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Wallpaper Random — Pick a random wallpaper                  ║
# ╚══════════════════════════════════════════════════════════════╝

WALLPAPER_DIR="$HOME/Pictures/Wallpapers"

WALLPAPER=$(find "$WALLPAPER_DIR" -type f \( -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" -o -name "*.webp" \) | shuf -n 1)

if [[ -n "$WALLPAPER" ]]; then
    "$HOME/.config/hypr/scripts/wallpaper-apply.sh" "$WALLPAPER"
fi
