#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Screen Record TUI — wf-recorder + slurp + fzf             ║
# ╚══════════════════════════════════════════════════════════════╝

RECORD_DIR="$HOME/Videos/Recordings"
PIDFILE="/tmp/sumi-recording.pid"
mkdir -p "$RECORD_DIR"

FZF_THEME="--color=bg+:#1a1a1a,fg:#c8c8c8,fg+:#ffffff,hl:#7aa2f7,hl+:#7aa2f7,pointer:#7aa2f7,prompt:#7aa2f7,header:#6a6a6a,border:#3a3a3a,info:#6a6a6a"

# Dependency check
for cmd in wf-recorder slurp; do
    if ! command -v "$cmd" &>/dev/null; then
        notify-send -a "sumi" -u critical "Recording error" "Missing: $cmd" 2>/dev/null
        exit 1
    fi
done

# ── Stale PID cleanup ─────────────────────────────────────────
# If pidfile exists but process is dead, clean it up
if [[ -f "$PIDFILE" ]]; then
    OLDPID=$(cat "$PIDFILE" 2>/dev/null)
    if [[ -n "$OLDPID" ]] && ! kill -0 "$OLDPID" 2>/dev/null; then
        rm -f "$PIDFILE"
    fi
fi

stop_recording() {
    if [[ -f "$PIDFILE" ]]; then
        local pid
        pid=$(cat "$PIDFILE" 2>/dev/null)
        if [[ -n "$pid" ]]; then
            kill -INT "$pid" 2>/dev/null
            # Wait for graceful shutdown (up to 3s)
            for _ in {1..30}; do
                kill -0 "$pid" 2>/dev/null || break
                sleep 0.1
            done
            # Force kill if still alive
            kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null
        fi
        rm -f "$PIDFILE"
        notify-send -a "sumi" -t 3000 "Recording stopped" "Saved to $RECORD_DIR"
    fi
}

# If already recording, stop it
if [[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE" 2>/dev/null)" 2>/dev/null; then
    stop_recording
    exit 0
fi

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
SLURP_ARGS="-d -c '#7aa2f7' -b '#0a0a0a80' -s '#7aa2f720' -w 2"

start_recording() {
    local file="$1"
    shift
    wf-recorder "$@" -f "$file" &
    local pid=$!
    echo "$pid" > "$PIDFILE"

    # Verify it actually started
    sleep 0.5
    if ! kill -0 "$pid" 2>/dev/null; then
        rm -f "$PIDFILE"
        notify-send -a "sumi" -u critical "Recording failed" "wf-recorder exited immediately"
        exit 1
    fi
}

case "${1:-pick}" in
    area)
        GEOMETRY=$(slurp $SLURP_ARGS 2>/dev/null) || exit 0
        start_recording "$RECORD_DIR/rec_area_$TIMESTAMP.mp4" -g "$GEOMETRY"
        notify-send -a "sumi" -t 3000 "● Recording area" "SUPER+ALT+R to stop"
        ;;
    full)
        start_recording "$RECORD_DIR/rec_full_$TIMESTAMP.mp4" -o eDP-1
        notify-send -a "sumi" -t 3000 "● Recording fullscreen" "SUPER+ALT+R to stop"
        ;;
    area-gif)
        GEOMETRY=$(slurp $SLURP_ARGS 2>/dev/null) || exit 0
        start_recording "$RECORD_DIR/rec_area_$TIMESTAMP.gif" -g "$GEOMETRY" -c gif
        notify-send -a "sumi" -t 3000 "● Recording GIF" "SUPER+ALT+R to stop"
        ;;
    pick)
        # Show different menu based on whether recording
        if [[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE" 2>/dev/null)" 2>/dev/null; then
            ENTRIES="stop\t■ Stop current recording"
        else
            ENTRIES="area\tRecord selected region → MP4\nfull\tRecord full screen → MP4\narea-gif\tRecord selected region → GIF"
        fi

        MODE=$(printf "$ENTRIES" \
            | fzf --delimiter='\t' --with-nth=2 \
                  --header="╔══ screen-record ══╗" \
                  --prompt="│ mode > " \
                  --height=100% --reverse --no-info \
                  --border=rounded --margin=1,2 \
                  $FZF_THEME \
            | cut -f1)
        [[ -z "$MODE" ]] && exit 0
        if [[ "$MODE" == "stop" ]]; then
            stop_recording
        else
            exec "$0" "$MODE"
        fi
        ;;
esac
