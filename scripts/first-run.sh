#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  First-Run Setup — downloads a starter wallpaper + applies   ║
# ║  Called once by autostart if no wallpapers exist yet         ║
# ╚══════════════════════════════════════════════════════════════╝

WALL_DIR="$HOME/Pictures/Wallpapers"
CACHE="$HOME/.cache/sumi"
FIRST_RUN_FLAG="$CACHE/.first-run-done"
SCRIPTS="$HOME/.config/hypr/scripts"

# Skip if already done
[[ -f "$FIRST_RUN_FLAG" ]] && exit 0

mkdir -p "$WALL_DIR" "$CACHE"

# Check if user already has wallpapers
if compgen -G "$WALL_DIR/*.{jpg,jpeg,png,webp}" > /dev/null 2>&1; then
    touch "$FIRST_RUN_FLAG"
    exit 0
fi

notify-send -a sumi -t 5000 "[ sumi :: first run ]" \
    "No wallpapers found. Generating a starter wallpaper..."

# Generate a simple monochrome gradient wallpaper using imagemagick
# This ensures wallust has SOMETHING to work with on first boot
if command -v magick &>/dev/null; then
    CONVERT="magick"
elif command -v convert &>/dev/null; then
    CONVERT="convert"
else
    # No imagemagick — create a solid dark wallpaper with ffmpeg
    if command -v ffmpeg &>/dev/null; then
        ffmpeg -f lavfi -i "color=c=0x0a0a0a:s=2880x1920:d=1" \
            -frames:v 1 "$WALL_DIR/default.png" -y 2>/dev/null
    else
        # Last resort — notify user
        notify-send -a sumi -u critical -t 0 "[ sumi ]" \
            "No imagemagick or ffmpeg found.\nDrop wallpapers in ~/Pictures/Wallpapers/"
        touch "$FIRST_RUN_FLAG"
        exit 0
    fi
    # Apply the solid wallpaper
    sumi wallpaper apply "$WALL_DIR/default.png" 2>/dev/null
    touch "$FIRST_RUN_FLAG"
    exit 0
fi

# Generate 3 starter wallpapers with different moods
# 1. Dark blue gradient (default accent-friendly)
$CONVERT -size 2880x1920 \
    gradient:'#0a0a1a'-'#1a1a3a' \
    -blur 0x20 \
    "$WALL_DIR/sumi-dark-blue.png" 2>/dev/null

# 2. Warm dark gradient
$CONVERT -size 2880x1920 \
    gradient:'#1a0f0a'-'#2a1a10' \
    -blur 0x20 \
    "$WALL_DIR/sumi-warm-dark.png" 2>/dev/null

# 3. Cool grey gradient
$CONVERT -size 2880x1920 \
    gradient:'#0a0a0a'-'#2a2a2a' \
    -blur 0x20 \
    "$WALL_DIR/sumi-monochrome.png" 2>/dev/null

# Validate and apply the first wallpaper
if [[ -f "$WALL_DIR/sumi-dark-blue.png" ]]; then
    sumi wallpaper apply "$WALL_DIR/sumi-dark-blue.png" 2>/dev/null
    touch "$FIRST_RUN_FLAG"
elif [[ -f "$WALL_DIR/sumi-monochrome.png" ]]; then
    sumi wallpaper apply "$WALL_DIR/sumi-monochrome.png" 2>/dev/null
    touch "$FIRST_RUN_FLAG"
else
    notify-send -a sumi -u critical -t 0 "[ sumi ]" \
        "Wallpaper generation failed.\nDrop images in ~/Pictures/Wallpapers/ and run:\n  sumi wallpaper apply <path>"
    touch "$FIRST_RUN_FLAG"
fi

# ── Welcome tour (split into two readable notifications) ──────
sleep 1

notify-send -a sumi -t 10000 "[ sumi :: welcome ]" \
    "Your rice is ready. Essentials:

SUPER+D       → app launcher
SUPER+Return  → terminal
SUPER+X       → control center
SUPER+/       → keybind cheatsheet"

sleep 3

notify-send -a sumi -t 10000 "[ sumi :: tips ]" \
    "SUPER+SHIFT+W → pick wallpaper
SUPER+E       → file manager (yazi)

Drop wallpapers in ~/Pictures/Wallpapers/
Colors auto-adapt to your wallpaper."
