#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Gaming Mode Toggle — max perf, kill visual overhead         ║
# ║  Disables: blur, animations, shadows, waybar, notifications  ║
# ║  Enables: tearing, VRR, performance power profile             ║
# ╚══════════════════════════════════════════════════════════════╝

STATE_FILE="$HOME/.cache/sumi/gaming-mode"

mkdir -p "$(dirname "$STATE_FILE")"

is_gaming() {
    [[ -f "$STATE_FILE" ]] && [[ "$(cat "$STATE_FILE")" == "on" ]]
}

# ── Waybar state tracking (prevents toggle drift) ───────────
waybar_visible() {
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

enable_gaming() {
    echo "on" > "$STATE_FILE"

    # Kill visual overhead
    hyprctl keyword decoration:blur:enabled false
    hyprctl keyword animations:enabled false
    hyprctl keyword decoration:shadow:enabled false
    hyprctl keyword general:gaps_in 0
    hyprctl keyword general:gaps_out 0
    hyprctl keyword general:border_size 1
    hyprctl keyword misc:vfr false         # fixed frame rate = less input latency

    # Hide waybar (tracked)
    hide_waybar

    # Suppress notifications
    dunstctl set-paused true 2>/dev/null

    # Performance power profile
    powerprofilesctl set performance 2>/dev/null

    notify-send -a "sumi" -u low -t 2000 "[ GAMING MODE ON ]" \
        "blur off · anim off · perf mode · notifs paused"
}

disable_gaming() {
    rm -f "$STATE_FILE"

    # Restore visual settings
    hyprctl keyword decoration:blur:enabled true
    hyprctl keyword animations:enabled true
    hyprctl keyword general:gaps_in 3
    hyprctl keyword general:gaps_out 6
    hyprctl keyword general:border_size 2
    hyprctl keyword misc:vfr true

    # Show waybar (tracked)
    show_waybar

    # Unpause notifications
    dunstctl set-paused false 2>/dev/null

    # Balanced power profile
    powerprofilesctl set balanced 2>/dev/null

    notify-send -a "sumi" -u low -t 2000 "[ GAMING MODE OFF ]" \
        "blur on · anim on · balanced mode · notifs on"
}

if is_gaming; then
    disable_gaming
else
    enable_gaming
fi
