#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Framework Battery Info — Charge limit + health              ║
# ║  Reads from framework-laptop-kmod sysfs                      ║
# ╚══════════════════════════════════════════════════════════════╝

CHARGE_LIMIT_PATH="/sys/class/power_supply/BAT1/charge_control_end_threshold"
CAPACITY=$(cat /sys/class/power_supply/BAT1/capacity 2>/dev/null || echo "?")
STATUS=$(cat /sys/class/power_supply/BAT1/status 2>/dev/null || echo "Unknown")
POWER_NOW=$(cat /sys/class/power_supply/BAT1/power_now 2>/dev/null || echo "0")
ENERGY_FULL=$(cat /sys/class/power_supply/BAT1/energy_full 2>/dev/null || echo "0")
ENERGY_DESIGN=$(cat /sys/class/power_supply/BAT1/energy_full_design 2>/dev/null || echo "0")

# Calculate watts (power_now is in microwatts)
WATTS=$(echo "scale=1; $POWER_NOW / 1000000" | bc 2>/dev/null || echo "?")

# Battery health percentage
if [[ "$ENERGY_DESIGN" =~ ^[0-9]+$ ]] && [[ "$ENERGY_DESIGN" -gt 0 ]]; then
    HEALTH=$(echo "scale=0; $ENERGY_FULL * 100 / $ENERGY_DESIGN" | bc 2>/dev/null || echo "?")
else
    HEALTH="?"
fi

# Charge limit
if [[ -f "$CHARGE_LIMIT_PATH" ]]; then
    LIMIT=$(cat "$CHARGE_LIMIT_PATH")
else
    LIMIT="100"
fi

case "$1" in
    waybar)
        case "$STATUS" in
            Charging)    suffix="+" ;;
            Discharging) suffix="-" ;;
            Full)        suffix="=" ;;
            *)           suffix="" ;;
        esac
        TOOLTIP="Status: ${STATUS}\nPower: ${WATTS}W\nHealth: ${HEALTH}%\nCharge limit: ${LIMIT}%"
        CLASS=""
        if [[ "$CAPACITY" =~ ^[0-9]+$ ]]; then
            if [[ "$CAPACITY" -le 15 ]]; then CLASS="critical"
            elif [[ "$CAPACITY" -le 30 ]]; then CLASS="warning"
            elif [[ "$STATUS" == "Charging" ]]; then CLASS="charging"
            fi
        fi
        # percentage must be numeric for waybar, fallback to 0
        local pct="${CAPACITY}"
        [[ ! "$pct" =~ ^[0-9]+$ ]] && pct=0
        echo "{\"text\": \"bat:${CAPACITY}%${suffix} ${WATTS}W\", \"tooltip\": \"${TOOLTIP}\", \"class\": \"${CLASS}\", \"percentage\": ${pct}}"
        ;;
    set-limit)
        # Usage: framework-battery.sh set-limit 80
        if [[ -n "$2" ]] && [[ "$2" -ge 60 ]] && [[ "$2" -le 100 ]]; then
            echo "$2" | sudo tee "$CHARGE_LIMIT_PATH" > /dev/null
            notify-send -t 2000 "[ battery ]" "charge limit: ${2}%"
        else
            echo "Usage: $0 set-limit <60-100>"
        fi
        ;;
esac
