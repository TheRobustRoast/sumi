#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  OSD — On-Screen Display via dunst for vol/bright/mic       ║
# ║  Usage: osd.sh volume|brightness|mic                        ║
# ╚══════════════════════════════════════════════════════════════╝

# Dunst stacking tag ensures only one OSD shows at a time
# (new notification replaces the old one instead of stacking)

make_bar() {
    local pct="$1"
    local max=20
    local filled=$(( pct * max / 100 ))
    local empty=$(( max - filled ))
    local bar=""
    for ((i=0; i<filled; i++)); do bar+="█"; done
    for ((i=0; i<empty; i++)); do bar+="░"; done
    echo "[$bar] ${pct}%"
}

case "$1" in
    volume-up)
        wpctl set-volume @DEFAULT_AUDIO_SINK@ 5%+
        VOL=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ | awk '{printf "%.0f", $2*100}')
        MUTE=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ | grep -c MUTED)
        if [[ "$MUTE" -gt 0 ]]; then
            notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
                -h int:value:"$VOL" -t 1500 \
                "vol: muted" "$(make_bar "$VOL")"
        else
            notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
                -h int:value:"$VOL" -t 1500 \
                "vol: ${VOL}%" "$(make_bar "$VOL")"
        fi
        ;;
    volume-down)
        wpctl set-volume @DEFAULT_AUDIO_SINK@ 5%-
        VOL=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ | awk '{printf "%.0f", $2*100}')
        notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
            -h int:value:"$VOL" -t 1500 \
            "vol: ${VOL}%" "$(make_bar "$VOL")"
        ;;
    volume-mute)
        wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle
        MUTE=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ | grep -c MUTED)
        VOL=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ | awk '{printf "%.0f", $2*100}')
        if [[ "$MUTE" -gt 0 ]]; then
            notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
                -h int:value:0 -t 1500 \
                "vol: MUTED" "[░░░░░░░░░░░░░░░░░░░░] mute"
        else
            notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
                -h int:value:"$VOL" -t 1500 \
                "vol: ${VOL}%" "$(make_bar "$VOL")"
        fi
        ;;
    mic-mute)
        wpctl set-mute @DEFAULT_AUDIO_SOURCE@ toggle
        MUTE=$(wpctl get-volume @DEFAULT_AUDIO_SOURCE@ | grep -c MUTED)
        if [[ "$MUTE" -gt 0 ]]; then
            notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
                -t 1500 "mic: MUTED" "[░░░░░░░░░░░░░░░░░░░░] off"
        else
            VOL=$(wpctl get-volume @DEFAULT_AUDIO_SOURCE@ | awk '{printf "%.0f", $2*100}')
            notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
                -h int:value:"$VOL" -t 1500 \
                "mic: ${VOL}%" "$(make_bar "$VOL")"
        fi
        ;;
    brightness-up)
        brightnessctl set 5%+
        BRI=$(brightnessctl -m | awk -F, '{print $4}' | tr -d '%')
        notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
            -h int:value:"$BRI" -t 1500 \
            "lcd: ${BRI}%" "$(make_bar "$BRI")"
        ;;
    brightness-down)
        brightnessctl set 5%-
        BRI=$(brightnessctl -m | awk -F, '{print $4}' | tr -d '%')
        notify-send -a "sumi-osd" -h string:x-dunst-stack-tag:osd \
            -h int:value:"$BRI" -t 1500 \
            "lcd: ${BRI}%" "$(make_bar "$BRI")"
        ;;
    *)
        echo "Usage: osd.sh {volume-up|volume-down|volume-mute|mic-mute|brightness-up|brightness-down}"
        exit 1
        ;;
esac
