#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Window Switcher — hyprctl + fzf TUI alt-tab                 ║
# ╚══════════════════════════════════════════════════════════════╝

WINDOWS=$(hyprctl clients -j | jq -r '.[] | select(.mapped == true) | "\(.address) │ ws:\(.workspace.id) │ \(.class) │ \(.title)"')

if [[ -z "$WINDOWS" ]]; then
    exit 0
fi

CHOICE=$(echo "$WINDOWS" | fzf \
    --prompt='win> ' \
    --header='╔════════════════════════════╗
║     window switcher        ║
╚════════════════════════════╝' \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
    --no-scrollbar)

[[ -z "$CHOICE" ]] && exit 0

ADDRESS=$(echo "$CHOICE" | awk -F '│' '{print $1}' | xargs)
hyprctl dispatch focuswindow "address:$ADDRESS"
