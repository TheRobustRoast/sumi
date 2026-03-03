#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Keybind Cheatsheet — SUPER+? to display in fzf             ║
# ╚══════════════════════════════════════════════════════════════╝

FZF_THEME="--color=bg+:#1a1a1a,fg:#c8c8c8,fg+:#ffffff,hl:#7aa2f7,hl+:#7aa2f7,pointer:#7aa2f7,prompt:#7aa2f7,header:#6a6a6a,border:#3a3a3a,info:#6a6a6a"

BINDS="
── CORE ────────────────────────────────────────
S+Return    │ Terminal (foot)
S+Q         │ Kill active window
S+D         │ App launcher (fuzzel)
S+R         │ Run prompt (fuzzel)
S+V         │ Toggle floating
S+F         │ Fullscreen
S+SH+F      │ Fake fullscreen
S+P         │ Pseudo-tile
S+S         │ Toggle split
S+O         │ Pin floating on top
S+SH+C      │ Reload Hyprland config

── FOCUS (vim-style) ───────────────────────────
S+H/J/K/L   │ Focus left/down/up/right
S+U          │ Focus urgent or last window
S+Tab        │ Window switcher (fzf)

── MOVE WINDOWS ────────────────────────────────
S+SH+H/J/K/L│ Move window left/down/up/right
S+SH+Return  │ Swap with master

── RESIZE ──────────────────────────────────────
S+CT+H/J/K/L│ Quick resize
S+CT+R       │ Enter RESIZE mode (hjkl, Esc)

── GROUPS / TABS ───────────────────────────────
S+W          │ Enter GROUP mode:
   G         │   Toggle group
   H/J/K/L   │   Absorb window from direction
   N / P     │   Next / prev tab
   O         │   Eject from group
   Esc       │   Exit mode

── WORKSPACES ──────────────────────────────────
S+1..0       │ Switch to workspace 1-10
S+SH+1..0    │ Move window to workspace 1-10
S+AL+1..0    │ Move silently (don't follow)
S+grave      │ Toggle scratchpad
S+SH+grave   │ Send to scratchpad

── SCRATCHPADS ─────────────────────────────────
F1           │ Scratch terminal
F2           │ Music visualizer (cava)
F3           │ System monitor (btop)

── TUI APPS ────────────────────────────────────
S+E          │ Files (yazi)
S+I          │ WiFi (impala)
S+B          │ Bluetooth (bluetuith)
S+A          │ Audio (pulsemixer)
S+T          │ Monitor (btop)
S+M          │ Music viz (cava)
S+G          │ Git (lazygit)
S+SH+D       │ Disk usage (ncdu)
S+X          │ Control center (all TUIs)

── DEV TOOLS ───────────────────────────────────
S+SH+P       │ Project launcher (frecency)
S+SH+G       │ Git worktree manager
S+AL+Return  │ Tmux session (attach/new)
F4           │ Dev terminal scratchpad
S+SH+F4      │ Send to dev scratchpad
S+SH+M       │ Quick notes (markdown)
S+AL+D       │ Dotfile sync (backup/restore)

── PICKERS ─────────────────────────────────────
S+SH+V       │ Clipboard history
S+Tab        │ Window switcher
S+.          │ Emoji picker
S+SH+N       │ Notification history
S+C          │ Color picker (hyprpicker)
S+=          │ Calculator (fuzzel)

── SCREENSHOTS ─────────────────────────────────
S+SH+S       │ Area → clipboard
S+Print      │ Full screen → clipboard
S+SH+Print   │ Area → file
S+AL+S       │ Screenshot picker (fzf)
S+AL+Print   │ Active window → clipboard

── SCREEN RECORDING ────────────────────────────
S+AL+R       │ Record area (toggle)
S+AL+SH+R    │ Record picker (fzf)

── THEME / WALLPAPER ───────────────────────────
S+SH+W       │ Wallpaper select
S+AL+W       │ Random wallpaper
S+SH+T       │ Theme toggle (dark/light)

── MODES ───────────────────────────────────────
F5           │ Toggle GAMING mode (perf max)
F6           │ Toggle FOCUS/DND mode
F7           │ Monitor hotplug handler
F9           │ Toggle waybar visibility

── POWER ───────────────────────────────────────
S+SH+E       │ Power menu
S+SH+X       │ Lock screen (hyprlock)

── MEDIA (with OSD bar) ────────────────────────
Bright Up/Dn │ Screen brightness ±5%
Vol Up/Dn    │ Volume ±5% (hold to repeat)
Vol Mute     │ Toggle mute
Mic Mute     │ Toggle mic mute
Play/Pause   │ Media play/pause
Next/Prev    │ Media next/prev

── MOUSE ───────────────────────────────────────
S+LMB drag   │ Move window
S+RMB drag   │ Resize window
S+scroll     │ Cycle workspaces

── TMUX (C-a prefix) ──────────────────────────
C-a v        │ Split horizontal
C-a s        │ Split vertical
C-a g        │ Lazygit popup
C-a f        │ Yazi popup
C-a t        │ Btop popup
C-a /        │ Ripgrep → fzf → nvim
C-a p        │ Session picker
Alt+hjkl     │ Pane navigation (no prefix)
Alt+1-9      │ Window switch (no prefix)

── NVIM LSP (Space leader) ────────────────────
gd/gD/gr/gi  │ Go to def/decl/refs/impl
K            │ Hover docs
Spc+ca       │ Code action
Spc+cr       │ Rename symbol
Spc+cf       │ Format buffer
Spc+xx       │ Trouble diagnostics
Spc+xt       │ TODO list
]h / [h      │ Next/prev git hunk
]d / [d      │ Next/prev diagnostic
]t / [t      │ Next/prev TODO

── GESTURES (trackpad) ─────────────────────────
3-finger swipe│ Switch workspaces
"

echo "$BINDS" | fzf \
    --header="╔══════════════════════════════════════╗
║     sumi :: keybind cheatsheet    ║
╚══════════════════════════════════════╝
  S=SUPER  SH=SHIFT  CT=CTRL  AL=ALT" \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --margin=0 \
    --padding=1 \
    $FZF_THEME \
    --no-scrollbar \
    --no-mouse \
    --info=hidden \
    --prompt="" \
    --bind="q:abort,escape:abort"
