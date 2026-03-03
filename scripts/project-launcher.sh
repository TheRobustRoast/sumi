#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Project Launcher — frecency-based project directory picker  ║
# ║  Uses zoxide's database for smart ranking                    ║
# ║  Opens project in tmux session (creates if needed)           ║
# ║  SUPER+SHIFT+P to launch                                    ║
# ╚══════════════════════════════════════════════════════════════╝

SCRIPTS="$HOME/.config/hypr/scripts"

# Check deps
for cmd in zoxide fzf tmux; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Missing: $cmd"
        read -r -p "Press enter..."
        exit 1
    fi
done

# Get directories ranked by frecency from zoxide
# Filter to likely project dirs (contain .git, Cargo.toml, package.json, etc.)
PROJECTS=$(zoxide query --list --score 2>/dev/null | sort -rn | awk '{print $2}' | while read -r dir; do
    if [[ -d "$dir/.git" ]] || \
       [[ -f "$dir/Cargo.toml" ]] || \
       [[ -f "$dir/package.json" ]] || \
       [[ -f "$dir/pyproject.toml" ]] || \
       [[ -f "$dir/go.mod" ]] || \
       [[ -f "$dir/Makefile" ]] || \
       [[ -f "$dir/CMakeLists.txt" ]] || \
       [[ -f "$dir/flake.nix" ]] || \
       [[ -f "$dir/.envrc" ]]; then
        echo "$dir"
    fi
done)

# Also scan common project directories as fallback
for scandir in "$HOME/Projects" "$HOME/Dev" "$HOME/src" "$HOME/repos" "$HOME/Code"; do
    if [[ -d "$scandir" ]]; then
        for d in "$scandir"/*/; do
            [[ -d "$d/.git" ]] && PROJECTS="$PROJECTS"$'\n'"${d%/}"
        done
    fi
done

# Deduplicate and filter empty
PROJECTS=$(echo "$PROJECTS" | sort -u | grep -v '^$')

if [[ -z "$PROJECTS" ]]; then
    echo "No projects found."
    echo "Projects are detected by: .git, Cargo.toml, package.json,"
    echo "pyproject.toml, go.mod, Makefile, CMakeLists.txt, flake.nix"
    echo ""
    echo "Tip: cd into project directories to build zoxide's database."
    read -r -p "Press enter..."
    exit 0
fi

# Pick with fzf
CHOICE=$(echo "$PROJECTS" | fzf \
    --prompt='proj> ' \
    --header='╔══════════════════════════════╗
║     project launcher        ║
╚══════════════════════════════╝
  Enter: open in tmux session' \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --preview='eza --tree --level=2 --icons --git-ignore --color=always {} 2>/dev/null || ls -la {}' \
    --preview-window=right:40%:wrap \
    --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
    --no-scrollbar)

[[ -z "$CHOICE" ]] && exit 0

# Create a tmux session name from the directory
SESSION_NAME=$(basename "$CHOICE" | tr '.' '-' | tr ' ' '-')

# Check if tmux session already exists
if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    # Attach to existing session
    if [[ -n "$TMUX" ]]; then
        tmux switch-client -t "$SESSION_NAME"
    else
        tmux attach -t "$SESSION_NAME"
    fi
else
    # Create new session with dev layout:
    # Window 1: editor (nvim)
    # Window 2: shell
    # Window 3: server/runner
    if [[ -n "$TMUX" ]]; then
        tmux new-session -d -s "$SESSION_NAME" -c "$CHOICE" -n "edit"
        tmux send-keys -t "$SESSION_NAME:edit" "nvim ." Enter
        tmux new-window -t "$SESSION_NAME" -n "shell" -c "$CHOICE"
        tmux new-window -t "$SESSION_NAME" -n "run" -c "$CHOICE"
        tmux select-window -t "$SESSION_NAME:edit"
        tmux switch-client -t "$SESSION_NAME"
    else
        tmux new-session -s "$SESSION_NAME" -c "$CHOICE" -n "edit" \; \
            send-keys "nvim ." Enter \; \
            new-window -n "shell" -c "$CHOICE" \; \
            new-window -n "run" -c "$CHOICE" \; \
            select-window -t :edit
    fi
fi
