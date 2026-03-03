#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Power Profile Toggle — Framework 13 AMD                     ║
# ║  Cycles: power-saver → balanced → performance                ║
# ║  Uses power-profiles-daemon (recommended for AMD 7040)       ║
# ╚══════════════════════════════════════════════════════════════╝

current=$(powerprofilesctl get)

case "$1" in
    toggle)
        case "$current" in
            power-saver)  powerprofilesctl set balanced    ;;
            balanced)     powerprofilesctl set performance ;;
            performance)  powerprofilesctl set power-saver ;;
        esac
        new=$(powerprofilesctl get)
        notify-send -t 2000 "[ power ]" "profile: $new"
        ;;
    status)
        echo "$current"
        ;;
    waybar)
        # JSON output for waybar custom module
        case "$current" in
            power-saver)  icon="eco" ;;
            balanced)     icon="bal" ;;
            performance)  icon="prf" ;;
        esac
        echo "{\"text\": \"pwr:${icon}\", \"tooltip\": \"Power profile: ${current}\", \"class\": \"${current}\"}"
        ;;
esac
