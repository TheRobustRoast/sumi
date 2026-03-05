You are helping with the **sumi** Hyprland rice project — an Arch Linux desktop configuration for a Framework 13 AMD laptop.

## Project context

- **Configs are symlinked** from the repo into `~/.config/` via `install.sh` (uses `ln -sf`). Editing a file in the repo edits it live on the system.
- **`sumi update`** — pulls latest, rebuilds, re-installs, reloads Hyprland
- **Theming** — wallust generates color files from the wallpaper and writes them to `~/.config/*/colors.*`. Static defaults live in the repo (e.g. `foot/colors.ini`, `waybar/colors.css`) and are seeded on install/update.

## Key files

| Area | Repo path |
|------|-----------|
| Hyprland main config | `hypr/hyprland.conf` |
| Window rules | `hypr/conf.d/rules.conf` |
| Keybinds | `hypr/conf.d/keybinds.conf` |
| Waybar layout | `waybar/config.jsonc` |
| Waybar style | `waybar/style.css` |
| Waybar colors (default) | `waybar/colors.css` |
| Foot terminal | `foot/foot.ini` |
| Foot colors (default) | `foot/colors.ini` |
| Fuzzel launcher | `fuzzel/fuzzel.ini` |
| Dunst notifications | `dunst/dunstrc` |
| Wallpaper apply | `sumi wallpaper apply` |
| Wallpaper init | `sumi wallpaper init` |
| Wallpaper select | `sumi wallpaper pick` |
| Wallust config | `wallust/wallust.toml` |
| Wallust templates | `wallust/templates/` |

## Hyprland version notes (0.50–0.53 breaking changes)

- `windowrulev2` replaced by `windowrule = match:class regex, rule value`
- `layerrule` syntax: `layerrule = blur on, match:namespace waybar`
- `render:explicit_sync` / `render:explicit_sync_kms` removed (0.50, now always-on)
- `first_launch_animation` replaced by `animation = monitorAdded, 0`
- Gestures: `workspace_swipe*` removed → `gesture = 3, horizontal, workspace`
- `new_window_takes_over_fullscreen` renamed to `on_focus_under_fullscreen`
- Field renames: `noblur`→`no_blur`, `idleinhibit`→`idle_inhibit`, `ignorezero`→`ignore_alpha 0.2`, `suppressevent`→`suppress_event`

## Foot version notes

- `dpi-aware` only accepts `yes`/`no` (not `auto`)
- `bold-text-uses-bright-colors`, `url.protocols`, `bell.command=` (empty), `[tweak]` section — all removed/invalid in installed version
- `Control+Shift+u` reserved for unicode input — don't bind it

## Wallpaper / theming pipeline

1. `sumi wallpaper apply <image>` — sets wallpaper via hyprpaper IPC, runs wallust, restarts waybar/dunst
2. Lockfile at `~/.cache/sumi/.wallpaper-applying` prevents inotify feedback loop
3. `wallust-watcher.service` watches `~/.cache/sumi/current-wallpaper` and re-applies on change

## Conventions

- Always commit with a clear message and `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>`
- After changing configs, commit + push so `sumi update` picks them up
- Static color defaults live in the repo; wallust overwrites them at runtime — never symlink color files
- Show the full commit message in every response when committing

---

Now help the user with their request about this rice/frontend configuration:

$ARGUMENTS
