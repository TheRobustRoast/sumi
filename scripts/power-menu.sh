#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Power Menu TUI — fzf-based shutdown/reboot/suspend/lock   ║
# ╚══════════════════════════════════════════════════════════════╝

FZF_THEME="--color=bg+:#1a1a1a,fg:#c8c8c8,fg+:#ffffff,hl:#7aa2f7,hl+:#7aa2f7,pointer:#7aa2f7,prompt:#7aa2f7,header:#6a6a6a,border:#3a3a3a,info:#6a6a6a"

ACTION=$(printf "lock\t Lock screen\nsuspend\t Suspend (sleep)\nreboot\t Reboot\nshutdown\t Shutdown\nlogout\t Logout (exit Hyprland)" \
    | fzf --delimiter='\t' --with-nth=2 \
          --header="╔══ power ══╗" \
          --prompt="│ action > " \
          --height=100% --reverse --no-info \
          --border=rounded --margin=1,2 \
          $FZF_THEME \
    | cut -f1)

[[ -z "$ACTION" ]] && exit 0

confirm() {
    local ans
    ans=$(printf "yes — $1\nno — cancel" | fzf --prompt="confirm > " --height=5 --reverse --no-info $FZF_THEME)
    [[ "$ans" == yes* ]]
}

# Countdown with cancel option for destructive actions
countdown() {
    local action="$1" secs="${2:-3}"
    local nid
    nid=$(notify-send -a sumi -t $((secs * 1000)) -p "$action in ${secs}s..." "Press SUPER+ESC to cancel")
    sleep "$secs"
}

case "$ACTION" in
    lock)     hyprlock ;;
    suspend)  systemctl suspend ;;
    reboot)   confirm "Reboot" && countdown "Reboot" 3 && systemctl reboot ;;
    shutdown) confirm "Shutdown" && countdown "Shutdown" 3 && systemctl poweroff ;;
    logout)   confirm "Exit Hyprland" && hyprctl dispatch exit ;;
esac
