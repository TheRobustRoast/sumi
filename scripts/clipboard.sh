#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Clipboard Manager — cliphist + fzf TUI picker               ║
# ╚══════════════════════════════════════════════════════════════╝

# Dependency check
for cmd in cliphist fzf wl-copy; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Missing dependency: $cmd"
        echo "Install with: pacman -S ${cmd}"
        read -r -p "Press enter to close..."
        exit 1
    fi
done

case "$1" in
    pick)
        if ! cliphist list 2>/dev/null | head -1 | grep -q .; then
            echo "Clipboard history is empty."
            read -r -p "Press enter to close..."
            exit 0
        fi

        cliphist list | fzf \
            --prompt='clip> ' \
            --header='╔════════════════════════════╗
║   clipboard history        ║
╚════════════════════════════╝' \
            --header-first \
            --layout=reverse \
            --height=100% \
            --border=none \
            --preview='echo {}' \
            --preview-window=down:3:wrap \
            --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
            --no-scrollbar | cliphist decode | wl-copy
        ;;
    wipe)
        cliphist wipe
        notify-send -t 2000 "[ clipboard ]" "history cleared"
        ;;
    *)
        echo "Usage: clipboard.sh {pick|wipe}"
        ;;
esac
