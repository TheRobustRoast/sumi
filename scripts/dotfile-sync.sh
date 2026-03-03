#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Dotfile Sync — git-based backup/restore for sumi configs   ║
# ║  Launch: via control center or `dotsync` alias              ║
# ╚══════════════════════════════════════════════════════════════╝

# Check deps
for cmd in git rsync fzf; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Missing: $cmd"
        read -r -p "Press enter..."
        exit 1
    fi
done

DOTFILE_REPO="$HOME/.dotfiles"
CONFIG="$HOME/.config"

FZF_THEME="--color=bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a"

# Tracked config paths (relative to $HOME)
TRACKED=(
    ".config/hypr"
    ".config/waybar"
    ".config/foot"
    ".config/fuzzel"
    ".config/dunst"
    ".config/nvim/init.lua"
    ".config/yazi"
    ".config/cava"
    ".config/lazygit"
    ".config/btop"
    ".config/starship.toml"
    ".config/wallust"
    ".tmux.conf"
    ".zshrc"
)

init_repo() {
    if [[ -d "$DOTFILE_REPO/.git" ]]; then
        echo "repo already exists at $DOTFILE_REPO"
        return 0
    fi
    mkdir -p "$DOTFILE_REPO"
    git -C "$DOTFILE_REPO" init
    echo "# sumi dotfiles" > "$DOTFILE_REPO/README.md"
    git -C "$DOTFILE_REPO" add README.md
    git -C "$DOTFILE_REPO" commit -m "init: sumi dotfiles"
    echo "initialized dotfile repo at $DOTFILE_REPO"
}

snapshot() {
    init_repo
    local changed=0
    for path in "${TRACKED[@]}"; do
        local src="$HOME/$path"
        local dst="$DOTFILE_REPO/$path"
        if [[ -e "$src" ]]; then
            mkdir -p "$(dirname "$dst")"
            rsync -a --delete "$src" "$(dirname "$dst")/" 2>/dev/null
            changed=1
        fi
    done

    if [[ $changed -eq 1 ]]; then
        cd "$DOTFILE_REPO" || { echo "Cannot cd to $DOTFILE_REPO"; return 1; }
        git add -A
        local status=$(git status --porcelain)
        if [[ -n "$status" ]]; then
            local msg="snapshot: $(date '+%Y-%m-%d %H:%M') — $(echo "$status" | wc -l) changes"
            git commit -m "$msg"
            local count=$(echo "$status" | wc -l)
            echo "committed $count changes"
            notify-send -a "sumi" -t 3000 "Dotfile Sync" "Snapshot saved ($count changes)"
        else
            echo "no changes to commit"
            notify-send -a "sumi" -t 2000 "Dotfile Sync" "Already up to date"
        fi
    fi
}

restore() {
    if [[ ! -d "$DOTFILE_REPO/.git" ]]; then
        echo "no dotfile repo found at $DOTFILE_REPO"
        echo "run 'dotsync snapshot' first, or clone your repo there"
        return 1
    fi

    echo "this will overwrite current configs with the last snapshot"
    echo -n "continue? [y/N] "
    read -r ans
    [[ ! "$ans" =~ ^[Yy]$ ]] && return 0

    for path in "${TRACKED[@]}"; do
        local src="$DOTFILE_REPO/$path"
        local dst="$HOME/$path"
        if [[ -e "$src" ]]; then
            mkdir -p "$(dirname "$dst")"
            rsync -a --delete "$src" "$(dirname "$dst")/"
            echo "restored: $path"
        fi
    done
    echo "restore complete — reload hyprland with SUPER+SHIFT+C"
    notify-send -a "sumi" -t 3000 "Dotfile Sync" "Configs restored — reload Hyprland"
}

diff_check() {
    init_repo
    # Sync to temp and diff
    local tmp=$(mktemp -d)
    for path in "${TRACKED[@]}"; do
        local src="$HOME/$path"
        if [[ -e "$src" ]]; then
            mkdir -p "$(dirname "$tmp/$path")"
            cp -r "$src" "$(dirname "$tmp/$path")/" 2>/dev/null
        fi
    done
    # Show diff against repo
    diff -rq "$DOTFILE_REPO" "$tmp" 2>/dev/null | grep -v '.git' | head -40
    rm -rf "$tmp"
}

log_history() {
    if [[ ! -d "$DOTFILE_REPO/.git" ]]; then
        echo "no dotfile repo"
        return 1
    fi
    git -C "$DOTFILE_REPO" log --oneline -20
}

ACTION=$(printf '%s\n' \
    "snap    │ snapshot current configs" \
    "restore │ restore from last snapshot" \
    "diff    │ show changes since last snapshot" \
    "log     │ view snapshot history" \
    "push    │ push to remote (if configured)" \
    "pull    │ pull from remote" \
    | fzf \
        --prompt='> ' \
        --header='╔══════════════════════════════╗
║     sumi :: dotsync          ║
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
    snap)    snapshot ;;
    restore) restore ;;
    diff)    diff_check; read -rp "press enter to exit..." ;;
    log)     log_history; read -rp "press enter to exit..." ;;
    push)    git -C "$DOTFILE_REPO" push 2>&1; read -rp "press enter to exit..." ;;
    pull)    git -C "$DOTFILE_REPO" pull 2>&1; read -rp "press enter to exit..." ;;
esac
