#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Theme Toggle — switch between dark/light monochrome base   ║
# ║  Accent color still comes from wallust / wallpaper          ║
# ╚══════════════════════════════════════════════════════════════╝

STATE_FILE="$HOME/.cache/sumi/theme-mode"
COLORS_CONF="$HOME/.config/hypr/conf.d/colors.conf"

FZF_THEME="--color=bg+:#1a1a1a,fg:#c8c8c8,fg+:#ffffff,hl:#7aa2f7,hl+:#7aa2f7,pointer:#7aa2f7,prompt:#7aa2f7,header:#6a6a6a,border:#3a3a3a,info:#6a6a6a"

# Detect current mode
CURRENT="dark"
[[ -f "$STATE_FILE" ]] && CURRENT=$(cat "$STATE_FILE")

set_dark() {
    cat > "$COLORS_CONF" << 'DARKEOF'
# sumi colors — DARK mode (base stays fixed, accent from wallust)
$bg       = rgb(0a0a0a)
$fg       = rgb(c8c8c8)
$surface0 = rgb(1a1a1a)
$surface1 = rgb(2a2a2a)
$surface2 = rgb(3a3a3a)
$dim      = rgb(6a6a6a)
$bright   = rgb(e8e8e8)
# Accent — overwritten by wallust
$accent     = rgb(7aa2f7)
$accent_dim = rgb(3d5178)
$warn       = rgb(e0af68)
$ok         = rgb(9ece6a)
$urgent     = rgb(f7768e)
DARKEOF
    echo "dark" > "$STATE_FILE"
}

set_light() {
    cat > "$COLORS_CONF" << 'LIGHTEOF'
# sumi colors — LIGHT mode (base stays fixed, accent from wallust)
$bg       = rgb(f0f0f0)
$fg       = rgb(1a1a1a)
$surface0 = rgb(e0e0e0)
$surface1 = rgb(d0d0d0)
$surface2 = rgb(c0c0c0)
$dim      = rgb(8a8a8a)
$bright   = rgb(0a0a0a)
# Accent — overwritten by wallust
$accent     = rgb(2e5cb8)
$accent_dim = rgb(6a8fd8)
$warn       = rgb(b08020)
$ok         = rgb(4a8a2a)
$urgent     = rgb(c83040)
LIGHTEOF
    echo "light" > "$STATE_FILE"
}

if [[ "${1:-}" == "toggle" ]]; then
    if [[ "$CURRENT" == "dark" ]]; then
        set_light
    else
        set_dark
    fi
    hyprctl reload
    notify-send -a "sumi" "Theme" "Switched to $(cat "$STATE_FILE") mode"
    exit 0
fi

# Interactive picker
MODE=$(printf "dark\t Dark mode (black bg)\nlight\t Light mode (white bg)\ntoggle\t Toggle current ($CURRENT)" \
    | fzf --delimiter='\t' --with-nth=2 \
          --header="╔══ theme ══╗  current: $CURRENT" \
          --prompt="│ mode > " \
          --height=100% --reverse --no-info \
          --border=rounded --margin=1,2 \
          $FZF_THEME \
    | cut -f1)

[[ -z "$MODE" ]] && exit 0

case "$MODE" in
    dark)   set_dark ;;
    light)  set_light ;;
    toggle) exec "$0" toggle ;;
esac

hyprctl reload
notify-send -a "sumi" "Theme" "Switched to $(cat "$STATE_FILE") mode"
