# ╔══════════════════════════════════════════════════════════════╗
# ║  Zsh Config — sumi TUI-first shell                        ║
# ╚══════════════════════════════════════════════════════════════╝

# ── History ──────────────────────────────────────────────────
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
setopt HIST_IGNORE_DUPS
setopt HIST_IGNORE_ALL_DUPS      # remove older duplicates from list
setopt HIST_FIND_NO_DUPS         # don't show dups when searching
setopt HIST_REDUCE_BLANKS        # trim whitespace from history
setopt HIST_IGNORE_SPACE
setopt SHARE_HISTORY
setopt APPEND_HISTORY
setopt INC_APPEND_HISTORY
setopt EXTENDED_HISTORY          # save timestamps in history

# ── Options ──────────────────────────────────────────────────
setopt AUTO_CD
setopt AUTO_PUSHD
setopt PUSHD_IGNORE_DUPS
setopt PUSHD_SILENT
setopt CORRECT
setopt NO_BEEP
setopt INTERACTIVE_COMMENTS
setopt EXTENDED_GLOB             # advanced globbing (#, ~, ^)
setopt GLOB_DOTS                 # include dotfiles in globs
setopt NO_CASE_GLOB              # case-insensitive globbing
setopt COMPLETE_IN_WORD          # complete from cursor position
setopt ALWAYS_TO_END             # move cursor to end on completion

# ── Vi mode ──────────────────────────────────────────────────
bindkey -v
export KEYTIMEOUT=1

# Fix backspace in vi mode
bindkey "^?" backward-delete-char
bindkey "^H" backward-delete-char

# Restore some emacs binds in vi insert mode (muscle memory)
bindkey '^A' beginning-of-line
bindkey '^E' end-of-line
bindkey '^K' kill-line
bindkey '^W' backward-kill-word
bindkey '^R' history-incremental-search-backward

# Vi mode cursor shape (block in normal, beam in insert)
function zle-keymap-select {
    if [[ $KEYMAP == vicmd ]] || [[ $1 == 'block' ]]; then
        echo -ne '\e[1 q'
    elif [[ $KEYMAP == main ]] || [[ $KEYMAP == viins ]] || [[ $1 == 'beam' ]]; then
        echo -ne '\e[5 q'
    fi
}
zle -N zle-keymap-select

function zle-line-init {
    echo -ne '\e[5 q'  # beam cursor on new prompt
}
zle -N zle-line-init

# ── Completion ───────────────────────────────────────────────
autoload -Uz compinit
# Only regenerate compinit cache once per day (faster startup)
if [[ -n ${ZDOTDIR}/.zcompdump(#qN.mh+24) ]]; then
    compinit
else
    compinit -C
fi

zstyle ':completion:*' menu select
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'
zstyle ':completion:*' list-colors "${(s.:.)LS_COLORS}"
zstyle ':completion:*' group-name ''
zstyle ':completion:*:descriptions' format '%B── %d ──%b'
zstyle ':completion:*:warnings' format '%B── no matches ──%b'
zstyle ':completion:*' squeeze-slashes true
zstyle ':completion:*' use-cache on
zstyle ':completion:*' cache-path "$HOME/.cache/zsh/compcache"

# Completion for kill/killall
zstyle ':completion:*:*:kill:*:processes' list-colors '=(#b) #([0-9]#)*=0=01;31'
zstyle ':completion:*:kill:*' command 'ps -u $USER -o pid,%cpu,tty,cputime,cmd'

# ── Plugins ──────────────────────────────────────────────────
[[ -f /usr/share/zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh ]] && \
    source /usr/share/zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh

[[ -f /usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh ]] && \
    source /usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh

# ── TUI aliases (replace coreutils with modern Rust TUIs) ────
alias ls='eza --icons --group-directories-first'
alias ll='eza -la --icons --group-directories-first --git'
alias lt='eza --tree --level=2 --icons'
alias la='eza -a --icons --group-directories-first'

alias cat='bat --style=plain --paging=never'
alias catp='bat --style=full'

alias du='dust'
alias df='duf'
alias ps='procs'
alias dig='doggo'

alias top='btop'
alias vim='nvim'
alias vi='nvim'

alias grep='rg'
alias find='fd'

# ── Git aliases ──────────────────────────────────────────────
alias g='lazygit'
alias gs='git status --short'
alias gl='git log --oneline -20'
alias gd='git diff'
alias ga='git add'
alias gc='git commit'
alias gp='git push'

# ── Sumi aliases ──────────────────────────────────────────
alias wifi='impala'
alias audio='pulsemixer'
alias blue='bluetuith'
alias files='yazi'
alias music='cava'
alias disk='ncdu'
alias sysmon='btop'
alias power='sumi power profile'
alias wall='sumi wallpaper pick'
alias wallr='sumi wallpaper random'
alias cc='sumi control'

# ── Quick edits ──────────────────────────────────────────────
alias hc='nvim ~/.config/hypr/hyprland.conf'
alias hk='nvim ~/.config/hypr/conf.d/keybinds.conf'
alias wc='nvim ~/.config/waybar/config.jsonc'
alias ws='nvim ~/.config/waybar/style.css'
alias zc='nvim ~/.zshrc'

# ── Environment ──────────────────────────────────────────────
export EDITOR="nvim"
export VISUAL="nvim"
export BROWSER="firefox"
export TERMINAL="foot"
export PAGER="bat --style=plain"
export MANPAGER="nvim +Man!"

# fzf monochrome theme
export FZF_DEFAULT_OPTS="
    --color=bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8
    --color=hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7
    --color=pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7
    --color=header:#3a3a3a,border:#3a3a3a
    --layout=reverse --height=40% --border=single --no-scrollbar
"

export FZF_DEFAULT_COMMAND='fd --type f --hidden --exclude .git'
export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
export FZF_ALT_C_COMMAND='fd --type d --hidden --exclude .git'

# ── Integrations ─────────────────────────────────────────────
# Zoxide (smart cd)
eval "$(zoxide init zsh --cmd cd)"

# Starship prompt
eval "$(starship init zsh)"

# fzf keybindings
[[ -f /usr/share/fzf/key-bindings.zsh ]] && source /usr/share/fzf/key-bindings.zsh
[[ -f /usr/share/fzf/completion.zsh ]] && source /usr/share/fzf/completion.zsh

# Direnv (per-project environment variables)
command -v direnv &>/dev/null && eval "$(direnv hook zsh)"

# Yazi shell wrapper (cd on exit)
function y() {
    local tmp="$(mktemp -t "yazi-cwd.XXXXXX")"
    yazi "$@" --cwd-file="$tmp"
    if cwd="$(cat -- "$tmp")" && [ -n "$cwd" ] && [ "$cwd" != "$PWD" ]; then
        builtin cd -- "$cwd"
    fi
    rm -f -- "$tmp"
}

# ── Dirstack ──────────────────────────────────────────────────
# Show recent dirs with `d`, jump with cd -<number>
DIRSTACKSIZE=20
setopt AUTO_PUSHD
alias d='dirs -v | head -20'

# ── Auto-create parent directories ───────────────────────────
alias mkdir='mkdir -pv'

# ── Safety nets ──────────────────────────────────────────────
alias rm='rm -I'
alias cp='cp -iv'
alias mv='mv -iv'

# ── Quick sumi aliases ────────────────────────────────────
alias dotsync='sumi sync'
alias note='sumi note'
alias hotplug='sumi monitor detect'
alias gamemode='sumi mode gaming'
alias focusmode='sumi mode focus'
alias cleanup='sumi cleanup'

# ── Dev functions ────────────────────────────────────────────

# Port killer — find and kill process on a port
port() {
    if [[ -z "$1" ]]; then
        echo "usage: port <number>  — show/kill process on port"
        return 1
    fi
    local pids=$(lsof -ti :"$1" 2>/dev/null)
    if [[ -z "$pids" ]]; then
        echo "no process on port $1"
        return 0
    fi
    echo "$pids" | while read pid; do
        local name=$(ps -p "$pid" -o comm= 2>/dev/null)
        echo "port $1 → pid $pid ($name)"
    done
    echo -n "kill? [y/N] "
    read -r ans
    [[ "$ans" =~ ^[Yy]$ ]] && echo "$pids" | xargs kill -9 && echo "killed"
}

# Pretty JSON (pipe or file)
jqp() {
    if [[ -t 0 && -n "$1" ]]; then
        jq '.' "$1" | bat --language json --style plain
    else
        jq '.' | bat --language json --style plain
    fi
}

# Quick HTTP server in current directory
serve() {
    local port="${1:-8080}"
    echo "serving $(pwd) on http://localhost:$port"
    python -m http.server "$port"
}

# Docker container shell
dsh() {
    local container=$(docker ps --format '{{.Names}}\t{{.Image}}\t{{.Status}}' 2>/dev/null | \
        fzf --prompt='container> ' --header='select container' | awk '{print $1}')
    [[ -z "$container" ]] && return 0
    docker exec -it "$container" /bin/sh -c "command -v bash >/dev/null && exec bash || exec sh"
}

# Docker logs follow with fzf picker
dlogs() {
    local container=$(docker ps -a --format '{{.Names}}\t{{.Image}}\t{{.Status}}' 2>/dev/null | \
        fzf --prompt='logs> ' --header='select container' | awk '{print $1}')
    [[ -z "$container" ]] && return 0
    docker logs -f --tail 100 "$container"
}

# Smart log tailer — picks from common log locations
logtail() {
    local logs=()
    # Journalctl units
    while IFS= read -r unit; do
        logs+=("journal:$unit")
    done < <(systemctl list-units --type=service --state=running --plain --no-legend 2>/dev/null | awk '{print $1}' | head -20)
    # Add common log files
    for f in /var/log/syslog /var/log/messages /var/log/pacman.log "$HOME/.cache/sumi/crash.log"; do
        [[ -f "$f" ]] && logs+=("file:$f")
    done
    local choice=$(printf '%s\n' "${logs[@]}" | fzf --prompt='log> ' --header='select log source')
    [[ -z "$choice" ]] && return 0
    case "$choice" in
        journal:*) journalctl -u "${choice#journal:}" -f --no-hostname -n 50 ;;
        file:*)    tail -f -n 50 "${choice#file:}" ;;
    esac
}

# Git branch cleanup — delete merged branches
gbclean() {
    local branches=$(git branch --merged main 2>/dev/null || git branch --merged master 2>/dev/null)
    branches=$(echo "$branches" | grep -vE '^\*|main|master|develop|dev' | sed 's/^[[:space:]]*//')
    if [[ -z "$branches" ]]; then
        echo "no merged branches to clean"
        return 0
    fi
    echo "$branches"
    echo -n "delete these merged branches? [y/N] "
    read -r ans
    [[ "$ans" =~ ^[Yy]$ ]] && echo "$branches" | xargs git branch -d
}

# Quick env var lookup with fzf
envs() {
    env | sort | fzf --prompt='env> ' --preview='echo {}' --preview-window=down:3:wrap
}

# Tmux session launcher for projects
proj() {
    sumi project
}

# Git worktree manager
wt() {
    sumi worktree
}

# Man page with fzf search
fman() {
    local page=$(man -k . 2>/dev/null | fzf --prompt='man> ' | awk '{print $1}' | sed 's/(.*//')
    [[ -n "$page" ]] && man "$page"
}

# Systemd unit browser
sysf() {
    local unit=$(systemctl list-units --all --plain --no-legend 2>/dev/null | \
        fzf --prompt='unit> ' --preview='systemctl status {1} 2>/dev/null' --preview-window=right:50%:wrap | \
        awk '{print $1}')
    [[ -n "$unit" ]] && systemctl status "$unit"
}

# ── Tmux auto-attach ────────────────────────────────────────
# Attach to existing session or create new one in interactive shells
if command -v tmux &>/dev/null && [[ -z "$TMUX" && -n "$PS1" && "$TERM_PROGRAM" != "vscode" ]]; then
    # Only auto-attach in foot terminal, not in nested shells
    if [[ "$TERM" == "foot" || "$TERM" == "foot-extra" ]]; then
        tmux attach -t main 2>/dev/null || tmux new-session -s main
    fi
fi

# ── sumi CLI is a Go binary at ~/.local/bin/sumi ──────────
# Shell aliases above provide shortcuts; run `sumi --help` for all commands.

# ── Create cache dirs ────────────────────────────────────────
[[ -d "$HOME/.cache/zsh" ]] || mkdir -p "$HOME/.cache/zsh"
