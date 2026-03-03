# sumi

A TUI-first Hyprland rice for Arch Linux, optimized for the Framework 13 AMD (Ryzen 7840U). Every GUI app is replaced with a terminal equivalent launched inside foot. Colors are extracted from your wallpaper and applied system-wide. Everything has a boxy, monochrome aesthetic with a single accent color — zero border-radius anywhere.

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

Boot the Arch live ISO on your Framework 13, connect to WiFi, then:

```bash
pacman -Sy git
git clone https://github.com/youruser/sumi /tmp/sumi
cd /tmp/sumi
./bootstrap.sh
```

The bootstrap script walks you through everything interactively: connects to WiFi (launches iwctl if needed), detects your NVMe, collects your username, password, LUKS encryption passphrase, hostname, and timezone, patches the archinstall configs with your settings, runs archinstall, and stages the sumi post-install to run automatically on first login.

After bootstrap finishes, it reboots into the new system. On first TTY login, `install.sh` runs automatically — installs yay, grabs AUR packages, deploys all config files, sets up greetd + Plymouth + systemd services, configures Framework 13 hardware, and changes your shell to zsh. Reboot once more for the full experience: Plymouth TUI LUKS unlock, greetd login, and the themed Hyprland desktop.

### Manual install (existing Arch system)

If you already have Arch + Hyprland installed:

```bash
git clone https://github.com/youruser/sumi ~/sumi
cd ~/sumi
./install.sh
```

The installer will: install yay if missing, install all packages via pacman and yay, back up existing configs before overwriting, deploy all config files, set up greetd + Plymouth + systemd user services, configure Framework 13 hardware (fingerprint, charge limit, AMD tuning), and change your default shell to zsh.

Reboot when prompted. Drop wallpapers in `~/Pictures/Wallpapers/` — if the directory is empty on first boot, three starter gradient wallpapers are auto-generated.

To remove everything cleanly:

```bash
chmod +x uninstall.sh
./uninstall.sh
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
├── archinstall/          # archinstall profiles (LUKS + Hyprland)
├── btop/                 # btop config + sumi theme
├── cava/                 # audio visualizer config
├── dunst/                # notification daemon
├── foot/                 # terminal emulator
├── greetd/               # login manager (tuigreet)
├── gtk-3.0/              # GTK3 dark mode + cursor
├── gtk-4.0/              # GTK4 dark mode + cursor
├── hypr/
│   ├── hyprland.conf     # main config (sources conf.d/)
│   ├── hyprlock.conf     # lockscreen (fingerprint + battery)
│   ├── hypridle.conf     # idle management
│   ├── hyprpaper.conf    # wallpaper
│   └── conf.d/
│       ├── autostart.conf
│       ├── colors.conf   # wallust-generated colors
│       ├── env.conf      # environment variables
│       ├── keybinds.conf # all keybinds
│       └── rules.conf    # window rules
├── icons/                # cursor theme index
├── lazygit/              # lazygit config
├── nvim/                 # neovim config (lazy.nvim + LSP)
├── plymouth/             # TUI boot theme for LUKS
├── scripts/              # 27 shell scripts
│   ├── control-center.sh
│   ├── project-launcher.sh
│   ├── git-worktree.sh
│   ├── wallpaper-*.sh
│   ├── screenshot.sh
│   ├── screen-record.sh
│   └── ...
├── starship/             # prompt config
├── systemd/user/         # 6 user services + timers
├── tmux/                 # tmux config
├── wallust/              # wallust config + 8 color templates
├── waybar/               # bar config + style
├── fuzzel/               # launcher config
├── xdg/                  # mime associations + portal config
├── yazi/                 # file manager config + theme
├── zsh/                  # .zshrc with TUI aliases + dev functions
├── install.sh            # full bootstrap installer
└── uninstall.sh          # clean removal
```

## Framework 13 AMD specifics

The rice is tuned for the Framework Laptop 13 with Ryzen 7840U:

- **Power:** power-profiles-daemon (not TLP — better for AMD 7040 series), amd_pstate=active
- **Display:** Scaled for 2880x1920 (2.8K), XWayland apps get 2x scaling
- **GPU:** RDNA3 iGPU optimizations — direct scanout, explicit sync, hardware cursors, reduced blur on fullscreen, tearing rules for gaming
- **Battery:** Charge capped at 80% via framework-laptop-kmod for longevity
- **Fingerprint:** fprintd enrolled and used on hyprlock
- **Sleep:** rtc_cmos.use_acpi_alarm=1 for reliable s2idle
- **Firmware:** fwupd enabled for LVFS firmware updates

## Dependencies

Installed automatically by `install.sh`:

**Core:** hyprland, foot, waybar, fuzzel, dunst, hyprlock, hypridle, hyprpaper, hyprpicker, greetd, tuigreet, plymouth

**TUI apps:** yazi, lazygit, btop, cava, impala, bluetuith, pulsemixer, ncdu, bandwhich, procs

**CLI tools:** fzf, jq, eza, bat, dust, duf, fd, ripgrep, doggo, zoxide, starship, direnv, wl-clipboard, cliphist, slurp, grim, wtype, playerctl, brightnessctl

**Dev:** neovim, tmux, zsh, zsh-autosuggestions, zsh-syntax-highlighting, nodejs, npm, python, shellcheck, shfmt, docker, lsof

**Framework:** framework-laptop-kmod-dkms-git, fprintd, fwupd, power-profiles-daemon

**Theming:** wallust, bibata-cursor-theme, JetBrainsMono Nerd Font

## License

MIT
