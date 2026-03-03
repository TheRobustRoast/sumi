#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Cleanup — auto-trim clipboard, old screenshots/recordings  ║
# ║  Run via systemd timer or manually                           ║
# ╚══════════════════════════════════════════════════════════════╝

MAX_CLIPBOARD=500
MAX_SCREENSHOT_DAYS=30
MAX_RECORDING_DAYS=14

echo ":: sumi cleanup starting..."

# ── Trim clipboard history ─────────────────────────────────────
if command -v cliphist &>/dev/null; then
    COUNT=$(cliphist list 2>/dev/null | wc -l)
    if [[ "$COUNT" -gt "$MAX_CLIPBOARD" ]]; then
        EXCESS=$(( COUNT - MAX_CLIPBOARD ))
        cliphist list | tail -n "$EXCESS" | while read -r line; do
            echo "$line" | cliphist delete 2>/dev/null
        done
        echo ":: Trimmed $EXCESS old clipboard entries (kept $MAX_CLIPBOARD)"
    else
        echo ":: Clipboard OK ($COUNT entries)"
    fi
fi

# ── Clean old screenshots ──────────────────────────────────────
SCREENSHOT_DIR="$HOME/Pictures/Screenshots"
if [[ -d "$SCREENSHOT_DIR" ]]; then
    OLD_SS=$(find "$SCREENSHOT_DIR" -type f -name "screenshot_*" -mtime +"$MAX_SCREENSHOT_DAYS" 2>/dev/null)
    if [[ -n "$OLD_SS" ]]; then
        COUNT=$(echo "$OLD_SS" | wc -l)
        echo "$OLD_SS" | xargs rm -f
        echo ":: Deleted $COUNT screenshots older than ${MAX_SCREENSHOT_DAYS}d"
    else
        echo ":: Screenshots OK (nothing older than ${MAX_SCREENSHOT_DAYS}d)"
    fi
fi

# ── Clean old recordings ──────────────────────────────────────
RECORDING_DIR="$HOME/Videos/Recordings"
if [[ -d "$RECORDING_DIR" ]]; then
    OLD_REC=$(find "$RECORDING_DIR" -type f -name "rec_*" -mtime +"$MAX_RECORDING_DAYS" 2>/dev/null)
    if [[ -n "$OLD_REC" ]]; then
        COUNT=$(echo "$OLD_REC" | wc -l)
        echo "$OLD_REC" | xargs rm -f
        echo ":: Deleted $COUNT recordings older than ${MAX_RECORDING_DAYS}d"
    else
        echo ":: Recordings OK (nothing older than ${MAX_RECORDING_DAYS}d)"
    fi
fi

# ── Clean temp screenshot cache ────────────────────────────────
rm -f /tmp/sumi-screenshot-last.png 2>/dev/null

# ── Clean stale PID files ──────────────────────────────────────
for pidfile in /tmp/sumi-*.pid; do
    [[ ! -f "$pidfile" ]] && continue
    PID=$(cat "$pidfile" 2>/dev/null)
    if [[ -n "$PID" ]] && ! kill -0 "$PID" 2>/dev/null; then
        rm -f "$pidfile"
        echo ":: Cleaned stale PID file: $pidfile"
    fi
done

echo ":: Cleanup done."
