#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Control Center — fzf-based TUI quick launcher               ║
# ║  Launch: SUPER+X                                             ║
# ╚══════════════════════════════════════════════════════════════╝

ENTRIES=(
    "net     │ impala         │ wifi & network manager"
    "bt      │ bluetuith      │ bluetooth manager"
    "vol     │ pulsemixer     │ audio mixer"
    "mon     │ btop           │ system monitor"
    "files   │ yazi           │ file manager"
    "disk    │ ncdu           │ disk usage analyzer"
    "git     │ lazygit        │ git TUI"
    "viz     │ cava           │ audio visualizer"
    "clip    │ clipboard      │ clipboard history"
    "ss      │ screenshot     │ screenshot picker"
    "rec     │ screen-record  │ screen recording"
    "emoji   │ emoji-picker   │ emoji & symbols"
    "notif   │ notifications  │ notification history"
    "wins    │ window-switch  │ window switcher"
    "bw      │ bandwhich      │ bandwidth monitor"
    "proc    │ procs          │ process viewer"
    "wall    │ wallpaper      │ pick wallpaper"
    "pwr     │ power-profile  │ toggle power profile"
    "theme   │ theme-toggle   │ dark/light mode switch"
    "sess    │ session-save   │ save/restore layout"
    "keys    │ keybinds       │ keybind cheatsheet"
    "game    │ gaming-mode    │ toggle gaming mode"
    "focus   │ focus-mode     │ toggle DND / focus"
    "disp    │ monitor-plug   │ detect/configure monitors"
    "clean   │ cleanup        │ trim clipboard & old files"
    "proj    │ project-launch │ open project in tmux"
    "wt      │ git-worktree   │ manage git worktrees"
    "tmux    │ tmux-session   │ attach/create tmux session"
    "note    │ scratch-note   │ quick notes (markdown)"
    "dots    │ dotsync        │ backup/restore configs"
    "lock    │ hyprlock       │ lock screen"
    "power   │ power-menu     │ shutdown/reboot/suspend"
)

CHOICE=$(printf '%s\n' "${ENTRIES[@]}" | fzf \
    --prompt='> ' \
    --header='╔══════════════════════════════╗
║     sumi :: control       ║
╚══════════════════════════════╝' \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --margin=1 \
    --padding=1 \
    --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
    --no-scrollbar \
    --no-mouse)

[[ -z "$CHOICE" ]] && exit 0

CMD=$(echo "$CHOICE" | awk -F '│' '{print $1}' | xargs)

case "$CMD" in
    net)    impala ;;
    bt)     bluetuith ;;
    vol)    pulsemixer ;;
    mon)    btop ;;
    files)  yazi ;;
    disk)   ncdu ~ ;;
    git)    lazygit ;;
    viz)    cava ;;
    clip)   ~/.config/hypr/scripts/clipboard.sh pick ;;
    ss)     ~/.config/hypr/scripts/screenshot.sh pick ;;
    rec)    ~/.config/hypr/scripts/screen-record.sh pick ;;
    emoji)  ~/.config/hypr/scripts/emoji-picker.sh ;;
    notif)  ~/.config/hypr/scripts/notification-center.sh ;;
    wins)   ~/.config/hypr/scripts/window-switcher.sh ;;
    bw)     sudo bandwhich ;;
    proc)   procs ;;
    wall)   ~/.config/hypr/scripts/wallpaper-select.sh ;;
    pwr)    ~/.config/hypr/scripts/power-profile.sh toggle ;;
    theme)  ~/.config/hypr/scripts/theme-toggle.sh ;;
    sess)   ~/.config/hypr/scripts/session-save.sh ;;
    keys)   ~/.config/hypr/scripts/keybinds-cheatsheet.sh ;;
    game)   ~/.config/hypr/scripts/gaming-mode.sh ;;
    focus)  ~/.config/hypr/scripts/focus-mode.sh ;;
    disp)   ~/.config/hypr/scripts/monitor-hotplug.sh ;;
    clean)  ~/.config/hypr/scripts/cleanup.sh ;;
    proj)   ~/.config/hypr/scripts/project-launcher.sh ;;
    wt)     ~/.config/hypr/scripts/git-worktree.sh ;;
    tmux)   tmux new-session -A -s main ;;
    note)   ~/.config/hypr/scripts/scratch-note.sh ;;
    dots)   ~/.config/hypr/scripts/dotfile-sync.sh ;;
    lock)   hyprlock ;;
    power)  ~/.config/hypr/scripts/power-menu.sh ;;
esac
