# sumi

A TUI-first Hyprland rice for Arch Linux with hardware-agnostic support. Every GUI app is replaced with a terminal equivalent launched inside foot. Colors are extracted from your wallpaper and applied system-wide. Everything has a boxy, monochrome aesthetic with a single accent color — zero border-radius anywhere. Managed entirely through a Go CLI with Bubble Tea TUIs.

![Arch](https://img.shields.io/badge/Arch_Linux-1793D1?style=flat&logo=archlinux&logoColor=white)
![Hyprland](https://img.shields.io/badge/Hyprland-58E1FF?style=flat&logo=wayland&logoColor=black)
![License](https://img.shields.io/badge/License-MIT-blue)

## What's included

**80 files** across 25 directories covering the full desktop stack:

| Layer | Tool | Purpose |
|-------|------|---------|
| Compositor | Hyprland | Tiling Wayland compositor |
| Terminal | foot | Fast Wayland-native terminal |
| Bar | Waybar | Status bar with clickable modules |
| Launcher | fuzzel | App launcher and run menu |
| Notifications | dunst | Notification daemon |
| Lock | hyprlock | Lockscreen with fingerprint support |
| Idle | hypridle | Auto-lock and DPMS |
| Wallpaper | hyprpaper + wallust | Wallpaper display + color extraction |
| Login | greetd + tuigreet | TUI login manager |
| Boot | Plymouth hypr-tui | TUI-style LUKS unlock screen |
| Editor | Neovim + lazy.nvim | Full LSP IDE with 20+ plugins |
| Multiplexer | tmux | Terminal multiplexer with dev popups |
| Files | yazi | TUI file manager |
| Git | lazygit | TUI git interface |
| Shell | zsh + starship | Shell with prompt, completions, vi mode |
| Audio | pulsemixer / cava | Volume mixer / audio visualizer |
| Bluetooth | bluetuith | TUI bluetooth manager |
| Network | impala | TUI network manager |
| Monitor | btop | System monitor |
| Disk | ncdu | Disk usage analyzer |

## Design philosophy

**Monochrome + single accent.** The base palette is fixed grayscale (#0a0a0a background through #c8c8c8 foreground). Only the accent color changes — it comes from your wallpaper via wallust. This means swapping wallpapers recolors the entire desktop without disrupting readability.

**Everything is a TUI.** GUI apps are replaced with terminal alternatives launched in floating foot windows. Every picker, selector, and menu uses fzf inside foot with a consistent boxy style.

**Vim-style navigation everywhere.** HJKL for window focus, tmux panes, fzf selections, and neovim. SUPER is the only modifier you need for daily use.

**Framework 13 AMD tuned.** AMD pstate=active, RDNA3 iGPU optimizations (direct scanout, explicit sync, reduced blur), fingerprint on lockscreen, battery charge capped at 80%, power-profiles-daemon over TLP.

## Installation

### One-command bootstrap (from Arch ISO)

Boot the Arch live ISO, connect to the internet, then:

```bash
pacman -Sy git go
git clone https://github.com/TheRobustRoast/sumi /tmp/sumi
cd /tmp/sumi && CGO_ENABLED=0 go build -o sumi ./cmd/sumi
./sumi bootstrap
```

The bootstrap TUI walks you through everything interactively: connects to WiFi (launches iwctl if needed), selects your disk, collects your username/password/hostname/timezone, then partitions with LUKS2 encryption, creates btrfs subvolumes, runs pacstrap, configures the system in chroot, and stages the rice installer for first login. On failure, a debug page is available at `http://<your-ip>:7777` so you can view the full log from another device.

After bootstrap finishes, reboot into the new system. On first TTY login, `sumi install` runs automatically — installs yay, grabs AUR packages, deploys all config files, sets up greetd + Plymouth + systemd services, applies hardware-specific tweaks, and changes your shell to zsh. Reboot once more for the full experience.

### Manual install (existing Arch system)

If you already have Arch + Hyprland installed:

```bash
git clone https://github.com/TheRobustRoast/sumi ~/sumi
cd ~/sumi && go build -o sumi ./cmd/sumi
./sumi install
```

### CLI commands

| Command | Description |
|---------|-------------|
| `sumi bootstrap` | Install Arch Linux from the live ISO |
| `sumi install` | Install the sumi rice on an existing Arch system |
| `sumi update` | Pull latest changes, rebuild, and re-install |
| `sumi uninstall` | Remove the sumi rice and restore defaults |
| `sumi theme` | Manage wallpaper and color theme |
| `sumi config` | Edit settings interactively (~/.config/sumi/config.toml) |
| `sumi doctor` | Check system health and configuration |
| `sumi status` | Show current state (wallpaper, theme, hardware, services) |

### Uninstall

```bash
sumi uninstall
```

## Wallpaper theming

wallust extracts colors from your wallpaper and writes them into 9 template files that are sourced by Hyprland, foot, fuzzel, waybar, dunst, cava, btop, tmux, and hyprlock. A systemd user service watches `~/Pictures/Wallpapers/` for changes and re-applies colors automatically.

Change wallpapers with `SUPER+SHIFT+W` (picker) or `SUPER+ALT+W` (random).

## Keybinds

Press `SUPER+/` for the full interactive cheatsheet. Here are the essentials:

### Core

| Key | Action |
|-----|--------|
| `SUPER+Return` | Terminal |
| `SUPER+Q` | Kill window |
| `SUPER+D` | App launcher |
| `SUPER+HJKL` | Focus left/down/up/right |
| `SUPER+SHIFT+HJKL` | Move window |
| `SUPER+1-0` | Workspace 1-10 |
| `SUPER+F` | Fullscreen |
| `SUPER+V` | Toggle float |
| `SUPER+X` | Control center (30 TUI tools) |

### TUI apps

| Key | App |
|-----|-----|
| `SUPER+E` | Files (yazi) |
| `SUPER+G` | Git (lazygit) |
| `SUPER+T` | Monitor (btop) |
| `SUPER+A` | Audio (pulsemixer) |
| `SUPER+I` | WiFi (impala) |
| `SUPER+B` | Bluetooth (bluetuith) |
| `SUPER+M` | Music viz (cava) |

### Dev tools

| Key | Action |
|-----|--------|
| `SUPER+SHIFT+P` | Project launcher (frecency-based) |
| `SUPER+SHIFT+G` | Git worktree manager |
| `SUPER+ALT+Return` | Tmux session |
| `F4` | Dev terminal scratchpad (tmux) |

### Pickers and utilities

| Key | Action |
|-----|--------|
| `SUPER+SHIFT+V` | Clipboard history |
| `SUPER+Tab` | Window switcher |
| `SUPER+.` | Emoji picker |
| `SUPER+SHIFT+S` | Screenshot (area) |
| `SUPER+=` | Calculator |
| `SUPER+C` | Color picker |

### Modes and submaps

| Key | Action |
|-----|--------|
| `SUPER+CTRL+R` | Resize mode (HJKL, Esc to exit) |
| `SUPER+W` | Group/tab mode |
| `F5` | Gaming mode (perf max, no animations) |
| `F6` | Focus/DND mode |

### Scratchpads

| Key | Scratchpad |
|-----|------------|
| `F1` | Terminal |
| `F2` | Music (cava) |
| `F3` | System monitor (btop) |
| `F4` | Dev terminal (tmux) |

## Neovim

The nvim config auto-bootstraps lazy.nvim and installs everything on first launch. Plugins included:

**LSP & Completion** — mason.nvim auto-installs 11 language servers (Lua, Python, Rust, TypeScript, Bash, CSS, HTML, JSON, Go, C/C++, TOML). nvim-cmp provides autocompletion with Tab cycling, LuaSnip for snippets.

**Formatting & Linting** — conform.nvim formats on save across 14 filetypes (prettier, rustfmt, black/ruff, gofmt, shfmt, clang-format, stylua). nvim-lint runs shellcheck and ruff on write.

**Navigation** — Telescope for fuzzy finding (`Space+f` files, `Space+/` grep, `Space+b` buffers). Trouble.nvim for diagnostics list. Todo-comments highlights TODO/FIXME with `Space+xt` to list them.

**Git** — Gitsigns for gutters, `]h`/`[h` for hunk navigation, `Space+gp` preview, `Space+gb` blame.

**Editing** — Treesitter (13 languages), Comment.nvim (`gc`), autopairs, surround (`ys`/`ds`/`cs`), indent guides, which-key.

All themed monochrome with the single accent color. The statusline shows mode, git branch, filename, LSP client, diagnostics, filetype, and position.

## Tmux

Prefix is `C-a`. Vim-style splits and navigation:

| Key | Action |
|-----|--------|
| `C-a v` | Horizontal split |
| `C-a s` | Vertical split |
| `Alt+HJKL` | Pane navigation (no prefix) |
| `Alt+1-9` | Window switch (no prefix) |
| `C-a g` | Lazygit popup |
| `C-a f` | Yazi popup |
| `C-a t` | Btop popup |
| `C-a /` | Ripgrep → fzf → nvim |
| `C-a p` | Session picker |

## Shell functions

The zsh config includes dev-focused functions:

| Command | What it does |
|---------|-------------|
| `port 3000` | Find and optionally kill process on a port |
| `jqp file.json` | Pretty-print JSON with syntax highlighting |
| `serve` | HTTP server in current directory (default :8080) |
| `dsh` | fzf-pick a Docker container and shell into it |
| `dlogs` | fzf-pick a Docker container and tail its logs |
| `logtail` | fzf-pick from journalctl units and log files |
| `gbclean` | Delete git branches already merged into main |
| `proj` | Project launcher (frecency-based, opens tmux) |
| `wt` | Git worktree manager (list/create/remove) |
| `fman` | fzf-search man pages |
| `sysf` | Browse systemd units with live status preview |
| `envs` | fzf-search environment variables |

## Project structure

```
sumi/
├── cmd/sumi/             # CLI entry point (cobra subcommands)
├── internal/
│   ├── bootstrap/        # Arch ISO installation logic
│   ├── config/           # TOML config (~/.config/sumi/config.toml)
│   ├── hardware/         # Hardware detection + profiles
│   ├── model/            # Shared types (Step, InstallCtx)
│   ├── runner/           # Subprocess streaming (Stream, RunCmd)
│   ├── steps/            # Install/doctor/uninstall step implementations
│   ├── theme/            # Monochrome palette + lipgloss styles
│   └── ui/               # Bubble Tea TUIs (installer, steprunner, bootstrap, config, theme, status)
├── btop/                 # btop config + sumi theme
├── cava/                 # audio visualizer config
├── dunst/                # notification daemon
├── foot/                 # terminal emulator
├── greetd/               # login manager (tuigreet)
├── hypr/                 # Hyprland config (sources conf.d/)
├── nvim/                 # neovim config (lazy.nvim + LSP)
├── plymouth/             # TUI boot theme for LUKS
├── scripts/              # shell scripts called by keybinds
├── systemd/user/         # user services + timers
├── wallust/              # color extraction config + templates
├── waybar/               # bar config + style
├── zsh/                  # .zshrc with TUI aliases + dev functions
└── Makefile              # build, static, install targets
```

## Hardware profiles

sumi auto-detects your hardware via DMI sysfs and applies the correct profile. Override with `sumi config` or set `hardware.profile` in `~/.config/sumi/config.toml`.

| Profile | Detection | Extras |
|---------|-----------|--------|
| Framework 13 AMD | DMI vendor match | framework-laptop-kmod, battery charge limit, amd_pstate, RDNA3 tweaks, fingerprint |
| Generic Laptop | Battery present | power-profiles-daemon, brightnessctl |
| Generic Desktop | No battery | Minimal, no power management |

## Dependencies

Installed automatically by `sumi install` and `sumi bootstrap`:

**Core:** hyprland, foot, waybar, fuzzel, dunst, hyprlock, hypridle, hyprpaper, hyprpicker, greetd, tuigreet, plymouth

**TUI apps:** yazi, lazygit, btop, cava, pulsemixer, ncdu, procs

**CLI tools:** fzf, jq, eza, bat, dust, duf, fd, ripgrep, zoxide, starship, wl-clipboard, cliphist, slurp, grim, wtype, playerctl, brightnessctl

**Dev:** neovim, zsh, zsh-autosuggestions, zsh-syntax-highlighting, go

**Hardware-specific:** Installed per profile (e.g. framework-laptop-kmod-dkms-git, fprintd, fwupd for Framework 13)

**Theming:** wallust, bibata-cursor-theme, JetBrainsMono Nerd Font

## License

MIT
