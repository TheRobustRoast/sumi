#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Screenshot — area/full/window with notifications             ║
# ╚══════════════════════════════════════════════════════════════╝

SCREENSHOT_DIR="$HOME/Pictures/Screenshots"
mkdir -p "$SCREENSHOT_DIR"

TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
SLURP_ARGS="-d -b 0a0a0a88 -c 7aa2f7 -s 1a1a1a44 -w 2"

# Dependency check
for cmd in grim slurp wl-copy hyprctl jq; do
    if ! command -v "$cmd" &>/dev/null; then
        notify-send -a "sumi" -u critical "Screenshot error" "Missing: $cmd" 2>/dev/null
        exit 1
    fi
done

notify_screenshot() {
    local file="$1"
    local desc="$2"
    if [[ -n "$file" ]] && [[ -f "$file" ]]; then
        notify-send -a "sumi" -i "$file" -t 3000 \
            "[ screenshot ]" "$desc\n$(basename "$file")"
    else
        notify-send -a "sumi" -t 2000 "[ screenshot ]" "$desc"
    fi
}

case "${1:-pick}" in
    area)
        GEOM=$(slurp $SLURP_ARGS 2>/dev/null) || exit 0
        TMPFILE="/tmp/sumi-screenshot-last.png"
        grim -g "$GEOM" "$TMPFILE"
        wl-copy < "$TMPFILE"
        notify_screenshot "$TMPFILE" "area → clipboard"
        ;;
    area-save)
        GEOM=$(slurp $SLURP_ARGS 2>/dev/null) || exit 0
        FILE="$SCREENSHOT_DIR/screenshot_${TIMESTAMP}.png"
        grim -g "$GEOM" "$FILE"
        wl-copy < "$FILE"
        notify_screenshot "$FILE" "saved + copied"
        ;;
    full)
        TMPFILE="/tmp/sumi-screenshot-last.png"
        grim "$TMPFILE"
        wl-copy < "$TMPFILE"
        notify_screenshot "$TMPFILE" "fullscreen → clipboard"
        ;;
    full-save)
        FILE="$SCREENSHOT_DIR/screenshot_${TIMESTAMP}.png"
        grim "$FILE"
        wl-copy < "$FILE"
        notify_screenshot "$FILE" "fullscreen saved + copied"
        ;;
    window)
        WINDOW_JSON=$(hyprctl activewindow -j 2>/dev/null)
        if [[ -z "$WINDOW_JSON" ]] || [[ "$WINDOW_JSON" == "null" ]]; then
            notify-send -a "sumi" -u low "Screenshot" "No active window"
            exit 1
        fi
        X=$(echo "$WINDOW_JSON" | jq -r '.at[0]')
        Y=$(echo "$WINDOW_JSON" | jq -r '.at[1]')
        W=$(echo "$WINDOW_JSON" | jq -r '.size[0]')
        H=$(echo "$WINDOW_JSON" | jq -r '.size[1]')
        TMPFILE="/tmp/sumi-screenshot-last.png"
        grim -g "${X},${Y} ${W}x${H}" "$TMPFILE"
        wl-copy < "$TMPFILE"
        notify_screenshot "$TMPFILE" "window → clipboard"
        ;;
    window-save)
        FILE="$SCREENSHOT_DIR/screenshot_window_${TIMESTAMP}.png"
        WINDOW_JSON=$(hyprctl activewindow -j 2>/dev/null)
        if [[ -z "$WINDOW_JSON" ]] || [[ "$WINDOW_JSON" == "null" ]]; then
            notify-send -a "sumi" -u low "Screenshot" "No active window"
            exit 1
        fi
        X=$(echo "$WINDOW_JSON" | jq -r '.at[0]')
        Y=$(echo "$WINDOW_JSON" | jq -r '.at[1]')
        W=$(echo "$WINDOW_JSON" | jq -r '.size[0]')
        H=$(echo "$WINDOW_JSON" | jq -r '.size[1]')
        grim -g "${X},${Y} ${W}x${H}" "$FILE"
        wl-copy < "$FILE"
        notify_screenshot "$FILE" "window saved + copied"
        ;;
    pick)
        FZF_THEME="--color=bg+:#1a1a1a,fg:#c8c8c8,fg+:#ffffff,hl:#7aa2f7,hl+:#7aa2f7,pointer:#7aa2f7,prompt:#7aa2f7,header:#6a6a6a,border:#3a3a3a,info:#6a6a6a"
        ENTRIES=(
            "area       │ select region → clipboard"
            "area-save  │ select region → file"
            "full       │ full screen → clipboard"
            "full-save  │ full screen → file"
            "window     │ active window → clipboard"
            "window-save│ active window → file"
        )
        CHOICE=$(printf '%s\n' "${ENTRIES[@]}" | fzf \
            --prompt='> ' \
            --header='╔════════════════════════════╗
║     screenshot mode        ║
╚════════════════════════════╝' \
            --header-first \
            --layout=reverse \
            --height=100% \
            --border=none \
            $FZF_THEME \
            --no-scrollbar)

        [[ -z "$CHOICE" ]] && exit 0
        CMD=$(echo "$CHOICE" | awk -F '│' '{print $1}' | xargs)
        exec "$0" "$CMD"
        ;;
    *)
        echo "Usage: screenshot.sh {area|area-save|full|full-save|window|window-save|pick}"
        ;;
esac
