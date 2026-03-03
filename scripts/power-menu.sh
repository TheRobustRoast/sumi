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

case "$ACTION" in
    lock)     hyprlock ;;
    suspend)  systemctl suspend ;;
    reboot)   systemctl reboot ;;
    shutdown) systemctl poweroff ;;
    logout)   hyprctl dispatch exit ;;
esac
