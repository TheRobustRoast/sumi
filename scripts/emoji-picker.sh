#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Emoji / Special Char Picker — fzf TUI                       ║
# ╚══════════════════════════════════════════════════════════════╝

EMOJI_FILE="$HOME/.cache/sumi/emoji-list.txt"

# Generate emoji list if not cached
if [[ ! -f "$EMOJI_FILE" ]]; then
    mkdir -p "$(dirname "$EMOJI_FILE")"
    # Common emoji subset — fast to load
    cat > "$EMOJI_FILE" << 'EMOJI'
😀 grinning face
😂 face with tears of joy
🥲 smiling face with tear
😊 smiling face with smiling eyes
😎 smiling face with sunglasses
🤔 thinking face
😴 sleeping face
🤯 exploding head
🥳 partying face
😤 face with steam from nose
👍 thumbs up
👎 thumbs down
👋 waving hand
🤝 handshake
✌️ victory hand
🖖 vulcan salute
🫡 saluting face
💪 flexed biceps
❤️ red heart
🔥 fire
⭐ star
✨ sparkles
💡 light bulb
⚡ high voltage
🎵 musical note
🎮 video game
💻 laptop
🖥️ desktop computer
⌨️ keyboard
🔧 wrench
🔨 hammer
⚙️ gear
📁 file folder
📂 open file folder
📝 memo
📌 pushpin
✅ check mark button
❌ cross mark
⚠️ warning
🚀 rocket
🏠 house
📡 satellite antenna
🔒 locked
🔓 unlocked
🌙 crescent moon
☀️ sun
🌊 water wave
🌲 evergreen tree
→ right arrow
← left arrow
↑ up arrow
↓ down arrow
↔ left-right arrow
⇒ rightwards double arrow
⇐ leftwards double arrow
• bullet
… ellipsis
— em dash
– en dash
© copyright
® registered
™ trademark
° degree
± plus-minus
× multiplication
÷ division
≠ not equal
≈ approximately equal
≤ less than or equal
≥ greater than or equal
∞ infinity
∑ summation
∏ product
√ square root
∫ integral
λ lambda
π pi
θ theta
α alpha
β beta
γ gamma
δ delta
ε epsilon
EMOJI
fi

CHOICE=$(cat "$EMOJI_FILE" | fzf \
    --prompt='char> ' \
    --header='╔════════════════════════════╗
║     emoji / special char   ║
╚════════════════════════════╝
  type to search, enter to copy' \
    --header-first \
    --layout=reverse \
    --height=100% \
    --border=none \
    --color='bg:#0a0a0a,bg+:#1a1a1a,fg:#6a6a6a,fg+:#c8c8c8,hl:#7aa2f7,hl+:#7aa2f7,info:#3a3a3a,prompt:#7aa2f7,pointer:#7aa2f7,marker:#9ece6a,spinner:#7aa2f7,header:#3a3a3a,border:#3a3a3a' \
    --no-scrollbar)

[[ -z "$CHOICE" ]] && exit 0

# Extract just the emoji/char (first field)
CHAR=$(echo "$CHOICE" | awk '{print $1}')
echo -n "$CHAR" | wl-copy
# Also type it into the focused window
wtype "$CHAR" 2>/dev/null || true
notify-send -t 1500 "[ char ]" "copied: $CHAR"
