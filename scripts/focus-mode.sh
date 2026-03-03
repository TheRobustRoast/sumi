#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Focus / DND Mode Toggle — minimal distractions               ║
# ║  Pauses notifications, hides waybar, disables idle suspend    ║
# ╚══════════════════════════════════════════════════════════════╝

STATE_FILE="$HOME/.cache/sumi/focus-mode"

is_focus() {
    [[ -f "$STATE_FILE" ]] && [[ "$(cat "$STATE_FILE")" == "on" ]]
}

waybar_visible() {
    # Check if waybar is visible by looking at layer-shell
    # If waybar is hidden (SIGUSR1'd), pgrep still finds it
    # We track state ourselves to avoid toggle drift
    [[ ! -f "$HOME/.cache/sumi/waybar-hidden" ]]
}

hide_waybar() {
    if waybar_visible; then
        pkill -SIGUSR1 waybar 2>/dev/null
        touch "$HOME/.cache/sumi/waybar-hidden"
    fi
}

show_waybar() {
    if ! waybar_visible; then
        pkill -SIGUSR1 waybar 2>/dev/null
        rm -f "$HOME/.cache/sumi/waybar-hidden"
    fi
}

enable_focus() {
    mkdir -p "$(dirname "$STATE_FILE")"
    echo "on" > "$STATE_FILE"

    # Brief notification BEFORE DND kicks in
    notify-send -a "sumi" -u low -t 2000 "[ FOCUS MODE ON ]" \
        "notifications paused · waybar hidden"

    # Small delay so the notification actually shows
    sleep 2.5

    # Pause notifications (DND)
    dunstctl set-paused true 2>/dev/null

    # Hide waybar (tracked to prevent toggle drift)
    hide_waybar
}

disable_focus() {
    rm -f "$STATE_FILE"

    # Unpause notifications
    dunstctl set-paused false 2>/dev/null

    # Show waybar (tracked)
    show_waybar

    notify-send -a "sumi" -u low -t 2000 "[ FOCUS MODE OFF ]" \
        "notifications on · waybar visible"
}

if is_focus; then
    disable_focus
else
    enable_focus
fi
