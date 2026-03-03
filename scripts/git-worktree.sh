#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Git Worktree Manager — fzf-based worktree picker            ║
# ║  List, create, switch, and remove worktrees                  ║
# ║  SUPER+SHIFT+G to launch                                    ║
# ╚══════════════════════════════════════════════════════════════╝

# Must be in a git repo
if ! git rev-parse --is-inside-work-tree &>/dev/null; then
    echo "Not in a git repository."
    read -r -p "Press enter..."
    exit 1
fi

REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
if [[ -z "$REPO_ROOT" ]]; then
    echo "Could not determine repo root."
    read -r -p "Press enter..."
    exit 1
fi

ACTIONS=(
    "list    │ switch to existing worktree"
    "create  │ create new worktree from branch"
    "remove  │ remove a worktree"
    "prune   │ clean up stale worktrees"
)

ACTION=$(printf '%s\n' "${ACTIONS[@]}" | fzf \
    --prompt='wt> ' \
    --header='╔══════════════════════════════╗
║     git worktree manager    ║
╚══════════════════════════════╝' \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
    --no-scrollbar)

[[ -z "$ACTION" ]] && exit 0

CMD=$(echo "$ACTION" | awk -F '│' '{print $1}' | xargs)

case "$CMD" in
    list)
        # List worktrees and pick one to cd into
        WORKTREES=$(git worktree list --porcelain 2>/dev/null | \
            awk '/^worktree /{path=$2} /^branch /{branch=$2; gsub("refs/heads/","",branch); print path " │ " branch}')

        if [[ -z "$WORKTREES" ]]; then
            echo "No worktrees found."
            read -r -p "Press enter..."
            exit 0
        fi

        CHOICE=$(echo "$WORKTREES" | fzf \
            --prompt='switch> ' \
            --header='Select worktree to switch to' \
            --layout=reverse \
            --height=100% \
            --border=none \
            --preview='eza --tree --level=1 --icons --color=always $(echo {} | awk -F "│" "{print \$1}" | xargs) 2>/dev/null' \
            --preview-window=right:40% \
            --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
            --no-scrollbar)

        [[ -z "$CHOICE" ]] && exit 0

        TARGET=$(echo "$CHOICE" | awk -F '│' '{print $1}' | xargs)
        SESSION_NAME="wt-$(basename "$TARGET")"

        # Open in tmux if available
        if command -v tmux &>/dev/null && [[ -n "$TMUX" ]]; then
            if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
                tmux switch-client -t "$SESSION_NAME"
            else
                tmux new-session -d -s "$SESSION_NAME" -c "$TARGET"
                tmux switch-client -t "$SESSION_NAME"
            fi
        else
            echo "cd $TARGET"
            builtin cd "$TARGET" 2>/dev/null || echo "Could not cd to $TARGET"
        fi
        ;;

    create)
        # Show branches and create worktree from selected
        echo -n "Branch name (or new): "
        read -r BRANCH

        [[ -z "$BRANCH" ]] && exit 0

        WORKTREE_DIR="$REPO_ROOT/../$(basename "$REPO_ROOT")-$BRANCH"

        # Check if branch exists remotely or locally
        if git show-ref --verify --quiet "refs/heads/$BRANCH" 2>/dev/null; then
            git worktree add "$WORKTREE_DIR" "$BRANCH"
        elif git show-ref --verify --quiet "refs/remotes/origin/$BRANCH" 2>/dev/null; then
            git worktree add --track -b "$BRANCH" "$WORKTREE_DIR" "origin/$BRANCH"
        else
            # New branch from current HEAD
            git worktree add -b "$BRANCH" "$WORKTREE_DIR"
        fi

        if [[ $? -eq 0 ]]; then
            echo ""
            echo "Worktree created at: $WORKTREE_DIR"
            echo "Branch: $BRANCH"
        fi

        read -r -p "Press enter..."
        ;;

    remove)
        WORKTREES=$(git worktree list --porcelain 2>/dev/null | \
            awk '/^worktree /{path=$2} /^branch /{branch=$2; gsub("refs/heads/","",branch); print path " │ " branch}' | \
            grep -v "$(git rev-parse --show-toplevel) │")  # exclude main worktree

        if [[ -z "$WORKTREES" ]]; then
            echo "No secondary worktrees to remove."
            read -r -p "Press enter..."
            exit 0
        fi

        CHOICE=$(echo "$WORKTREES" | fzf \
            --prompt='remove> ' \
            --header='Select worktree to remove' \
            --layout=reverse \
            --height=100% \
            --border=none \
            --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
            --no-scrollbar)

        [[ -z "$CHOICE" ]] && exit 0

        TARGET=$(echo "$CHOICE" | awk -F '│' '{print $1}' | xargs)
        echo "Removing worktree: $TARGET"
        git worktree remove "$TARGET" 2>/dev/null || git worktree remove --force "$TARGET"
        echo "Done."
        read -r -p "Press enter..."
        ;;

    prune)
        git worktree prune -v
        echo ""
        echo "Stale worktrees cleaned."
        read -r -p "Press enter..."
        ;;
esac
