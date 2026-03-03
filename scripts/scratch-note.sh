#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Scratch Note — quick capture to markdown                    ║
# ║  Launch: SUPER+SHIFT+N or via control center                ║
# ╚══════════════════════════════════════════════════════════════╝

NOTES_DIR="$HOME/Documents/notes"
mkdir -p "$NOTES_DIR"

FZF_THEME="--color=bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a"

ACTION=$(printf '%s\n' \
    "new     │ create a new note" \
    "today   │ open/create today's daily note" \
    "search  │ ripgrep across all notes" \
    "recent  │ browse recent notes" \
    "todo    │ show all TODOs across notes" \
    | fzf \
        --prompt='> ' \
        --header='╔══════════════════════════════╗
║     sumi :: notes            ║
╚══════════════════════════════╝' \
        --header-first \
        --layout=reverse \
        --height=100% \
        --border=none \
        --margin=1 \
        --padding=1 \
        $FZF_THEME \
        --no-scrollbar \
        --no-mouse)

[[ -z "$ACTION" ]] && exit 0
CMD=$(echo "$ACTION" | awk -F '│' '{print $1}' | xargs)

case "$CMD" in
    new)
        # Prompt for title
        TITLE=$(echo "" | fzf --print-query --prompt='title> ' --header='Enter note title (Esc to cancel)' \
            $FZF_THEME --layout=reverse --height=10 --border=none 2>/dev/null | head -1)
        [[ -z "$TITLE" ]] && exit 0

        # Sanitize filename
        SLUG=$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')
        FILE="$NOTES_DIR/$(date +%Y%m%d)-${SLUG}.md"

        # Create with template
        cat > "$FILE" << EOF
# $TITLE

*$(date '+%Y-%m-%d %H:%M')*

---


EOF
        nvim "+normal Go" "$FILE"
        ;;

    today)
        FILE="$NOTES_DIR/daily/$(date +%Y-%m-%d).md"
        mkdir -p "$NOTES_DIR/daily"
        if [[ ! -f "$FILE" ]]; then
            cat > "$FILE" << EOF
# $(date '+%A, %B %d %Y')

## Tasks

- [ ]

## Notes


## Log

- $(date +%H:%M) —
EOF
        fi
        nvim "$FILE"
        ;;

    search)
        RESULT=$(rg --line-number --no-heading --color=always '' "$NOTES_DIR" 2>/dev/null | \
            fzf --ansi --delimiter : \
                --preview 'bat --style=numbers --color=always {1} --highlight-line {2} 2>/dev/null' \
                --preview-window 'right:60%:+{2}/3' \
                $FZF_THEME --layout=reverse --height=100% --border=none)
        [[ -z "$RESULT" ]] && exit 0
        FILE=$(echo "$RESULT" | awk -F: '{print $1}')
        LINE=$(echo "$RESULT" | awk -F: '{print $2}')
        nvim "+$LINE" "$FILE"
        ;;

    recent)
        FILE=$(find "$NOTES_DIR" -name '*.md' -type f -printf '%T@ %p\n' 2>/dev/null | \
            sort -rn | head -30 | awk '{print $2}' | \
            fzf --preview 'bat --style=numbers --color=always {} 2>/dev/null' \
                --preview-window 'right:60%' \
                $FZF_THEME --layout=reverse --height=100% --border=none)
        [[ -z "$FILE" ]] && exit 0
        nvim "$FILE"
        ;;

    todo)
        RESULT=$(rg --line-number --no-heading --color=always '\- \[ \]' "$NOTES_DIR" 2>/dev/null | \
            fzf --ansi --delimiter : \
                --preview 'bat --style=numbers --color=always {1} --highlight-line {2} 2>/dev/null' \
                --preview-window 'right:60%:+{2}/3' \
                $FZF_THEME --layout=reverse --height=100% --border=none)
        [[ -z "$RESULT" ]] && exit 0
        FILE=$(echo "$RESULT" | awk -F: '{print $1}')
        LINE=$(echo "$RESULT" | awk -F: '{print $2}')
        nvim "+$LINE" "$FILE"
        ;;
esac
