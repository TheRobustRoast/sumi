#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Wallpaper Select — Pick wallpaper via fuzzel                ║
# ╚══════════════════════════════════════════════════════════════╝

WALLPAPER_DIR="$HOME/Pictures/Wallpapers"

# Guard: directory must exist and contain images
if [[ ! -d "$WALLPAPER_DIR" ]]; then
    notify-send -a sumi -u critical -t 5000 "Wallpaper" \
        "Directory not found: ~/Pictures/Wallpapers/\nCreate it and add some images."
    exit 1
fi

IMAGES=$(find "$WALLPAPER_DIR" -type f \( -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" -o -name "*.webp" \) | sort)

if [[ -z "$IMAGES" ]]; then
    notify-send -a sumi -u low -t 5000 "Wallpaper" \
        "No images in ~/Pictures/Wallpapers/\nAdd .png, .jpg, or .webp files."
    exit 0
fi

# List wallpapers
SELECTION=$(echo "$IMAGES" | while read -r f; do
    basename "$f"
done | fuzzel --dmenu --prompt="wallpaper> " --width=25 --lines=15)

[[ -z "$SELECTION" ]] && exit 0

FULL_PATH=$(echo "$IMAGES" | grep "/$SELECTION\$" | head -1)

if [[ -n "$FULL_PATH" ]]; then
    if "$HOME/.config/hypr/scripts/wallpaper-apply.sh" "$FULL_PATH"; then
        notify-send -a sumi -t 2000 "Wallpaper" "Applied: $SELECTION"
    else
        notify-send -a sumi -u critical -t 3000 "Wallpaper" "Failed to apply: $SELECTION"
    fi
fi
