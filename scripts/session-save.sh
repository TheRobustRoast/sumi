#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Session Save/Restore — save window layout via hyprctl      ║
# ╚══════════════════════════════════════════════════════════════╝

SESSION_DIR="$HOME/.cache/sumi/sessions"
mkdir -p "$SESSION_DIR"

FZF_THEME="--color=bg+:#1a1a1a,fg:#c8c8c8,fg+:#ffffff,hl:#7aa2f7,hl+:#7aa2f7,pointer:#7aa2f7,prompt:#7aa2f7,header:#6a6a6a,border:#3a3a3a,info:#6a6a6a"

save_session() {
    local name="${1:-$(date +%Y%m%d_%H%M%S)}"
    local file="$SESSION_DIR/$name.json"
    hyprctl clients -j > "$file"
    notify-send -a "sumi" "Session saved" "$name ($(jq length "$file") windows)"
    echo "Saved: $file"
}

restore_session() {
    local file="$1"
    [[ ! -f "$file" ]] && echo "File not found: $file" && return 1

    local restored=0
    local total=0
    # Use process substitution instead of pipe to avoid subshell counter loss
    while IFS=' ' read -r class ws x y w h floating; do
        [[ -z "$class" || "$class" == "null" ]] && continue
        ((total++))

        # Try to find running instance and move it
        addr=$(hyprctl clients -j | jq -r ".[] | select(.class == \"$class\") | .address" | head -1)
        if [[ -n "$addr" && "$addr" != "null" ]]; then
            hyprctl dispatch movetoworkspacesilent "$ws,address:$addr" 2>/dev/null
            if [[ "$floating" == "true" ]]; then
                hyprctl dispatch focuswindow "address:$addr" 2>/dev/null
                hyprctl dispatch setfloating "address:$addr" 2>/dev/null
                hyprctl dispatch moveactive "exact $x $y" 2>/dev/null
                hyprctl dispatch resizeactive "exact $w $h" 2>/dev/null
            fi
            ((restored++))
        fi
    done < <(jq -r '.[] | "\(.class) \(.workspace.id) \(.at[0]) \(.at[1]) \(.size[0]) \(.size[1]) \(.floating)"' "$file")
    notify-send -a "sumi" "Session restored" "$(basename "$file" .json) — $restored/$total windows"
}

list_sessions() {
    ls -1 "$SESSION_DIR"/*.json 2>/dev/null | while read -r f; do
        local name=$(basename "$f" .json)
        local count=$(jq length "$f" 2>/dev/null || echo "?")
        echo -e "$f\t$name ($count windows)"
    done
}

case "${1:-pick}" in
    save)
        save_session "${2:-}"
        ;;
    restore)
        if [[ -n "${2:-}" ]]; then
            restore_session "$2"
        else
            FILE=$(list_sessions | fzf --delimiter='\t' --with-nth=2 \
                --header="╔══ restore session ══╗" \
                --prompt="│ session > " \
                --height=100% --reverse --no-info \
                --border=rounded --margin=1,2 \
                $FZF_THEME \
                | cut -f1)
            [[ -n "$FILE" ]] && restore_session "$FILE"
        fi
        ;;
    pick)
        ACTION=$(printf "save\tSave current layout\nrestore\tRestore a saved layout\nlist\tList saved sessions" \
            | fzf --delimiter='\t' --with-nth=2 \
                  --header="╔══ session ══╗" \
                  --prompt="│ action > " \
                  --height=100% --reverse --no-info \
                  --border=rounded --margin=1,2 \
                  $FZF_THEME \
            | cut -f1)
        [[ -z "$ACTION" ]] && exit 0
        exec "$0" "$ACTION"
        ;;
esac
