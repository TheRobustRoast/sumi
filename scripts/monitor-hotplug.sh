#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Monitor Hotplug Handler                                    ║
# ║  Detects external monitor connect/disconnect and adapts     ║
# ║  Triggered by: hyprctl dispatch dpms / Hyprland events      ║
# ║  Can be run manually: monitor-hotplug.sh                    ║
# ╚══════════════════════════════════════════════════════════════╝

CACHE_DIR="$HOME/.cache/sumi"
STATE_FILE="$CACHE_DIR/monitor-state"
SCRIPTS="$HOME/.config/hypr/scripts"

mkdir -p "$CACHE_DIR"

notify() {
    notify-send -a sumi -t 3000 "$1" "$2"
}

get_monitor_count() {
    hyprctl monitors -j 2>/dev/null | jq 'length' 2>/dev/null || echo 0
}

get_external_monitor() {
    # Returns the name of the first non-eDP monitor (external)
    hyprctl monitors -j 2>/dev/null | jq -r '.[] | select(.name != "eDP-1") | .name' 2>/dev/null | head -1
}

# ── Save/restore previous state ──────────────────────────────
save_state() {
    echo "$1" > "$STATE_FILE"
}

get_saved_state() {
    cat "$STATE_FILE" 2>/dev/null || echo "1"
}

# ── Main logic ───────────────────────────────────────────────
current_count=$(get_monitor_count)
previous_count=$(get_saved_state)
external=$(get_external_monitor)

if [[ "$current_count" -eq "$previous_count" ]]; then
    # No change — might be a manual invocation, just report
    if [[ -n "$external" ]]; then
        notify "Monitors" "Internal: eDP-1 + External: $external"
    else
        notify "Monitors" "Internal: eDP-1 only"
    fi
    exit 0
fi

save_state "$current_count"

if [[ "$current_count" -gt 1 ]] && [[ -n "$external" ]]; then
    # ── External monitor connected ───────────────────────────
    notify "Monitor Connected" "$external — configuring..."

    # Move workspaces 6-10 to external display
    for ws in 6 7 8 9 10; do
        hyprctl dispatch moveworkspacetomonitor "$ws $external" 2>/dev/null
    done

    # Restart waybar to pick up multi-monitor
    pkill waybar 2>/dev/null
    sleep 0.3
    waybar &

    notify "Monitor Ready" "WS 1-5 → eDP-1 | WS 6-10 → $external"

elif [[ "$current_count" -eq 1 ]]; then
    # ── External monitor disconnected ────────────────────────
    notify "Monitor Disconnected" "Moving all workspaces to internal display"

    # All workspaces back to internal
    for ws in 1 2 3 4 5 6 7 8 9 10; do
        hyprctl dispatch moveworkspacetomonitor "$ws eDP-1" 2>/dev/null
    done

    # Restart waybar for single-monitor mode
    pkill waybar 2>/dev/null
    sleep 0.3
    waybar &

    notify "Single Monitor" "All workspaces on eDP-1"
fi
