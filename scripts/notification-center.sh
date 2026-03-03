#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Notification Center — dunstctl history via fzf               ║
# ╚══════════════════════════════════════════════════════════════╝

HISTORY=$(dunstctl history | jq -r '.data[0][]? | "\(.appname.data) │ \(.summary.data) │ \(.body.data)"' 2>/dev/null)

if [[ -z "$HISTORY" ]]; then
    echo "No notifications in history."
    read -r -p "Press enter to close..."
    exit 0
fi

CHOICE=$(echo "$HISTORY" | fzf \
    --prompt='notif> ' \
    --header='╔════════════════════════════╗
║   notification history     ║
╚════════════════════════════╝' \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --preview='echo {}' \
    --preview-window=down:3:wrap \
    --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
    --no-scrollbar \
    --bind='ctrl-d:execute-silent(dunstctl close-all)+abort' \
    --bind='ctrl-x:execute-silent(dunstctl history-clear)+abort')

# ctrl-d = dismiss all, ctrl-x = clear history
