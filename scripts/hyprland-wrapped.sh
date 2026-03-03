#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Hyprland Crash Recovery Wrapper                            ║
# ║  Auto-restarts Hyprland if it crashes (up to 3 times)       ║
# ║  Usage: set this as your greetd command instead of Hyprland ║
# ╚══════════════════════════════════════════════════════════════╝

MAX_RESTARTS=3
CRASH_WINDOW=60  # seconds — reset counter if stable for this long
LOG_DIR="$HOME/.cache/sumi"
LOG_FILE="$LOG_DIR/crash.log"

mkdir -p "$LOG_DIR"

restart_count=0
last_start=0

while true; do
    now=$(date +%s)

    # Reset counter if we've been stable for $CRASH_WINDOW seconds
    if (( now - last_start > CRASH_WINDOW )); then
        restart_count=0
    fi

    if (( restart_count >= MAX_RESTARTS )); then
        echo "[$(date)] Hyprland crashed $MAX_RESTARTS times in ${CRASH_WINDOW}s — giving up" >> "$LOG_FILE"
        echo "Hyprland crashed repeatedly. Check $LOG_FILE"
        echo "Press Enter to try again, or Ctrl+C to drop to TTY."
        read -r
        restart_count=0
    fi

    last_start=$(date +%s)
    restart_count=$((restart_count + 1))

    echo "[$(date)] Starting Hyprland (attempt $restart_count/$MAX_RESTARTS)" >> "$LOG_FILE"

    # Actually launch Hyprland
    Hyprland 2>> "$LOG_FILE"
    exit_code=$?

    # Exit code 0 = clean exit (user logged out) — don't restart
    if [[ $exit_code -eq 0 ]]; then
        echo "[$(date)] Hyprland exited cleanly (code 0)" >> "$LOG_FILE"
        break
    fi

    echo "[$(date)] Hyprland crashed with exit code $exit_code" >> "$LOG_FILE"

    # Brief delay before restart to avoid CPU spin
    sleep 1
done
