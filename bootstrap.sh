#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi :: bootstrap                                          ║
# ║                                                              ║
# ║  Run this from the Arch Linux live ISO.                      ║
# ║  Partitions, encrypts, installs, and configures Arch Linux   ║
# ║  with LUKS2, btrfs, systemd-boot, then stages the rice       ║
# ║  installer for first boot.                                   ║
# ║                                                              ║
# ║  Usage:                                                      ║
# ║    Boot Arch ISO → connect to internet (or let this help) →  ║
# ║    git clone <repo> /tmp/sumi                                ║
# ║    cd /tmp/sumi && chmod +x bootstrap.sh && ./bootstrap.sh   ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_LOG="/tmp/sumi-install.log"

# ── Catppuccin Mocha palette for gum ────────────────────────────
export GUM_INPUT_CURSOR_FOREGROUND="#f38ba8"
export GUM_INPUT_PROMPT_FOREGROUND="#cba6f7"
export GUM_INPUT_HEADER_FOREGROUND="#7f849c"
export GUM_INPUT_PLACEHOLDER_FOREGROUND="#45475a"
export GUM_CONFIRM_PROMPT_FOREGROUND="#cba6f7"
export GUM_CONFIRM_SELECTED_BACKGROUND="#a6e3a1"
export GUM_CONFIRM_SELECTED_FOREGROUND="#1e1e2e"
export GUM_CONFIRM_UNSELECTED_BACKGROUND="#313244"
export GUM_CONFIRM_UNSELECTED_FOREGROUND="#cdd6f4"
export GUM_CHOOSE_CURSOR_FOREGROUND="#f38ba8"
export GUM_CHOOSE_HEADER_FOREGROUND="#cba6f7"
export GUM_CHOOSE_SELECTED_FOREGROUND="#a6e3a1"
export GUM_CHOOSE_CURSOR_PREFIX="▶ "
export GUM_CHOOSE_SELECTED_PREFIX="● "
export GUM_CHOOSE_UNSELECTED_PREFIX="  "
export GUM_SPIN_SPINNER="dot"
export GUM_SPIN_SPINNER_FOREGROUND="#cba6f7"
export GUM_SPIN_TITLE_FOREGROUND="#cdd6f4"

# Fallback ANSI (used before gum is available)
CYN='\033[0;36m'; RST='\033[0m'

# ── Helpers ──────────────────────────────────────────────────────
s_step()    { gum style --foreground '#89b4fa' "  · $1"; }
s_ok()      { gum style --foreground '#a6e3a1' "  ✓  $1"; }
s_fail()    { gum style --foreground '#f38ba8' "  ✗  $1"; }
s_warn()    { gum style --foreground '#f9e2af' "  !  $1"; }
s_section() {
    echo ""
    gum style --foreground '#cba6f7' --bold "  ── $1"
    echo ""
}

# Run a command, log output, show ✓ / ✗ — sets INSTALL_EXIT on failure
run_step() {
    local title="$1"; shift
    s_step "$title..."
    if "$@" >> "$INSTALL_LOG" 2>&1; then
        s_ok "$title"
    else
        s_fail "$title"
        INSTALL_EXIT=1
    fi
}

# ── 0. Bootstrap gum ─────────────────────────────────────────────
echo -e "${CYN}:: ${RST}Initializing..."
if ! command -v gum &>/dev/null; then
    echo -e "${CYN}:: ${RST}Installing gum (TUI toolkit)..."
    pacman -Sy --noconfirm gum &>/dev/null
fi

# ── Welcome ──────────────────────────────────────────────────────
clear
gum style \
    --border rounded \
    --border-foreground '#313244' \
    --margin "1 2" \
    --padding "1 4" \
    "$(gum style --foreground '#f38ba8' --bold 'sumi') $(gum style --foreground '#45475a' '::') $(gum style --foreground '#cdd6f4' 'arch linux bootstrap')" \
    "" \
    "$(gum style --foreground '#6c7086' 'Framework 13 AMD  ·  LUKS2  ·  btrfs  ·  Hyprland')" \
    "$(gum style --foreground '#313244' '────────────────────────────────────────────────')" \
    "$(gum style --foreground '#585b70' 'Partitions · Encrypts · Installs · Stages the rice')"
echo ""
gum confirm "  Begin installation?" || exit 0

# ── Non-ISO warning ──────────────────────────────────────────────
if [[ ! -d /run/archiso ]]; then
    echo ""
    s_warn "This doesn't look like the Arch Linux live ISO."
    gum confirm "  Continue anyway?" --default=false || exit 0
fi

# ── 1. Network ───────────────────────────────────────────────────
s_section "Network"

# Auto-bring up ethernet interfaces and request DHCP if no IP yet
gum spin --title "  Bringing up ethernet..." -- bash -c '
    for iface in $(ip -o link show | awk -F": " "{print \$2}" | grep -E "^en"); do
        ip link set "$iface" up 2>/dev/null || true
        # Only DHCP if no address yet
        if ! ip -4 addr show "$iface" 2>/dev/null | grep -q "inet "; then
            dhcpcd "$iface" --timeout 10 2>/dev/null &
        fi
    done
    sleep 5
' 2>/dev/null || true

if ! gum spin --title "  Checking connectivity..." -- \
    ping -c 1 -W 5 archlinux.org 2>/dev/null; then
    s_warn "No internet connection detected."
    echo ""
    NET_CHOICE=$(gum choose \
        --header "  How to connect?" \
        "WiFi (iwctl)" \
        "Ethernet — retry DHCP" \
        "Skip (I'll connect manually)")
    case "$NET_CHOICE" in
        "WiFi (iwctl)")
            gum style \
                --border rounded --border-foreground '#f9e2af' --padding "0 2" \
                "  station wlan0 scan" \
                "  station wlan0 get-networks" \
                "  station wlan0 connect <SSID>" \
                "  exit"
            echo ""
            clear; iwctl || true
            ;;
        "Ethernet — retry DHCP")
            gum spin --title "  Requesting DHCP on all ethernet interfaces..." -- bash -c '
                for iface in $(ip -o link show | awk -F": " "{print \$2}" | grep -E "^en"); do
                    ip link set "$iface" up 2>/dev/null || true
                    dhcpcd "$iface" --timeout 15 2>/dev/null || true
                done
            ' || true
            ;;
    esac
    sleep 2
    gum spin --title "  Rechecking..." -- ping -c 1 -W 5 archlinux.org || {
        s_fail "No internet. Connect manually and re-run."
        exit 1
    }
fi
s_ok "Internet connected"

# ── 2. Disk selection ────────────────────────────────────────────
s_section "Disk"

gum spin --title "  Scanning disks..." -- sleep 0.5 2>/dev/null || true

NVME_DEFAULT=$(lsblk -dno NAME,TYPE | grep disk | grep nvme | head -1 \
    | awk '{print "/dev/"$1}') || true

mapfile -t DISK_LIST < <(lsblk -dno NAME,SIZE,TYPE | grep disk \
    | awk '{printf "/dev/%-14s %s\n", $1, $2}')

if [[ ${#DISK_LIST[@]} -eq 0 ]]; then
    s_fail "No disks detected."
    exit 1
fi

DISK_CHOICE=$(printf '%s\n' "${DISK_LIST[@]}" \
    | gum choose --header "  Select installation disk") || { s_step "Aborted."; exit 0; }
NVME=$(echo "$DISK_CHOICE" | awk '{print $1}')

if [[ ! -b "$NVME" ]]; then
    s_fail "$NVME is not a valid block device."
    exit 1
fi

# Partition suffix: nvme uses p1/p2, sata uses 1/2
if [[ "$NVME" =~ nvme ]]; then EFI_PART="${NVME}p1"; ROOT_PART="${NVME}p2"
else                            EFI_PART="${NVME}1";  ROOT_PART="${NVME}2"
fi

DISK_SIZE_HUMAN=$(lsblk -dno SIZE "$NVME")
s_ok "Selected: $NVME  ($DISK_SIZE_HUMAN)"

echo ""
gum style \
    --border rounded --border-foreground '#f38ba8' --padding "0 2" \
    "$(gum style --foreground '#f38ba8' "  ⚠  ALL DATA ON $NVME ($DISK_SIZE_HUMAN) WILL BE ERASED")"
echo ""
gum confirm "  Confirm wipe?" \
    --prompt.foreground '#f38ba8' \
    --selected.background '#f38ba8' --selected.foreground '#1e1e2e' \
    --default=false || { s_step "Aborted."; exit 0; }

# ── 3. User configuration ────────────────────────────────────────
s_section "User Configuration"

USERNAME=$(gum input \
    --placeholder "username" \
    --prompt "  Username  › " \
    --header "  Local user account" \
    --value "user") || { s_step "Aborted."; exit 0; }
[[ -z "$USERNAME" ]] && USERNAME="user"

echo ""
while true; do
    PASSWORD=$(gum input --password \
        --placeholder "••••••••" \
        --prompt "  Password  › " \
        --header "  One password — login · root · LUKS") || { s_step "Aborted."; exit 0; }
    PASSWORD2=$(gum input --password \
        --placeholder "••••••••" \
        --prompt "  Confirm   › ") || { s_step "Aborted."; exit 0; }
    [[ "$PASSWORD" == "$PASSWORD2" ]] && { s_ok "Password set"; break; }
    s_warn "Passwords don't match — try again."
done

echo ""
HOSTNAME=$(gum input \
    --placeholder "framework" \
    --prompt "  Hostname  › " \
    --header "  Machine hostname" \
    --value "framework") || { s_step "Aborted."; exit 0; }
HOSTNAME="${HOSTNAME:-framework}"

echo ""
gum spin --title "  Detecting timezone from IP..." -- \
    bash -c 'curl -s --max-time 5 https://ipapi.co/timezone > /tmp/.sumi-tz 2>/dev/null || echo UTC > /tmp/.sumi-tz'
DETECTED_TZ=$(cat /tmp/.sumi-tz 2>/dev/null || echo "UTC")
rm -f /tmp/.sumi-tz
if [[ ! "$DETECTED_TZ" =~ ^[A-Za-z_]+/[A-Za-z_/+\-]+$ ]] && [[ "$DETECTED_TZ" != "UTC" ]]; then
    DETECTED_TZ="UTC"
fi
TIMEZONE=$(gum input \
    --placeholder "America/New_York" \
    --prompt "  Timezone  › " \
    --header "  System timezone" \
    --value "$DETECTED_TZ") || { s_step "Aborted."; exit 0; }
TIMEZONE="${TIMEZONE:-UTC}"

# ── 4. Review ────────────────────────────────────────────────────
s_section "Review"

gum style \
    --border rounded --border-foreground '#313244' --padding "1 3" \
    "$(gum style --foreground '#cba6f7' --bold '  Installation Summary')" \
    "" \
    "  $(gum style --foreground '#585b70' 'Disk      ') $NVME  ($DISK_SIZE_HUMAN)  $(gum style --foreground '#f38ba8' '← WIPED')" \
    "  $(gum style --foreground '#585b70' 'Encrypt   ') LUKS2  (argon2id)" \
    "  $(gum style --foreground '#585b70' 'Filesystem') btrfs  (@  @home  @snapshots  @var_log)" \
    "  $(gum style --foreground '#585b70' 'Bootloader') systemd-boot" \
    "  $(gum style --foreground '#585b70' 'Hostname  ') $HOSTNAME" \
    "  $(gum style --foreground '#585b70' 'User      ') $USERNAME  (sudo, shared password)" \
    "  $(gum style --foreground '#585b70' 'Timezone  ') $TIMEZONE" \
    "  $(gum style --foreground '#585b70' 'Audio     ') PipeWire" \
    "  $(gum style --foreground '#585b70' 'Network   ') NetworkManager" \
    "  $(gum style --foreground '#585b70' 'Desktop   ') Hyprland + sumi rice  (post-boot)"

echo ""
gum confirm "  Proceed with installation?" \
    --default=false --prompt.foreground '#f38ba8' || { s_step "Aborted."; exit 0; }

# ── 5. Install ───────────────────────────────────────────────────
LOCAL_IP=$(ip -4 addr show scope global \
    | awk '/inet / {split($2,a,"/"); print a[1]; exit}' 2>/dev/null || true)
[[ -z "$LOCAL_IP" ]] && LOCAL_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || true)
[[ -z "$LOCAL_IP" ]] && LOCAL_IP="this-machine"

clear
gum style \
    --border rounded --border-foreground '#313244' --padding "1 3" \
    "$(gum style --foreground '#cba6f7' --bold '  Installing Arch Linux')" \
    "$(gum style --foreground '#6c7086' "  On failure → http://${LOCAL_IP}:7777")"
echo ""

: > "$INSTALL_LOG"
INSTALL_EXIT=0
set +e

# ── Pre-flight: clean stale state ───────────────────────────────
s_step "Pre-flight cleanup..."
{
    echo "==> pre-flight"
    # Unmount /mnt (handles subvol mounts at /mnt/@ etc.)
    while IFS= read -r mnt; do
        umount "$mnt" 2>/dev/null || umount -l "$mnt" 2>/dev/null || true
    done < <(findmnt -r -n -o TARGET | grep '^/mnt' | sort -r)
    umount -R /mnt 2>/dev/null || true

    # Drop caches so the kernel releases btrfs refs
    sync
    echo 3 > /proc/sys/vm/drop_caches 2>/dev/null || true
    sleep 1

    # Close all non-ISO device mapper devices
    for name in root $(dmsetup ls 2>/dev/null | awk '{print $1}'); do
        [[ "$name" == "ventoy" ]] && continue
        [[ "$name" =~ ^sda ]] && continue
        [[ ! -b "/dev/mapper/$name" ]] && continue
        cryptsetup close "$name" 2>/dev/null || true
        [[ -b "/dev/mapper/$name" ]] && dmsetup remove --force "$name" 2>/dev/null || true
    done
} >> "$INSTALL_LOG" 2>&1 || true
s_ok "Pre-flight done"

# ── Partition ───────────────────────────────────────────────────
if [[ $INSTALL_EXIT -eq 0 ]]; then
    s_step "Partitioning $NVME..."
    {
        echo ""; echo "==> partition"
        sgdisk --zap-all "$NVME"
        sgdisk \
            --new=1:0:+1G  --typecode=1:ef00 --change-name=1:EFI \
            --new=2:0:0    --typecode=2:8309 --change-name=2:LUKS \
            "$NVME"
        partprobe "$NVME"
        udevadm settle --timeout 10
    } >> "$INSTALL_LOG" 2>&1 && s_ok "Disk partitioned" || { s_fail "Partitioning failed"; INSTALL_EXIT=1; }
fi

# ── EFI ─────────────────────────────────────────────────────────
[[ $INSTALL_EXIT -eq 0 ]] && run_step "Formatting EFI" mkfs.fat -F32 "$EFI_PART"

# ── LUKS ────────────────────────────────────────────────────────
if [[ $INSTALL_EXIT -eq 0 ]]; then
    s_step "Setting up LUKS2 (slow — hashing key)..."
    {
        echo ""; echo "==> LUKS format"
        printf '%s' "$PASSWORD" | cryptsetup luksFormat \
            --type luks2 --pbkdf argon2id --hash sha512 \
            --key-size 512 --iter-time 10000 --batch-mode \
            "$ROOT_PART" -
        echo "==> LUKS open"
        printf '%s' "$PASSWORD" | cryptsetup open "$ROOT_PART" root -
    } >> "$INSTALL_LOG" 2>&1 && s_ok "LUKS2 configured" || { s_fail "LUKS2 failed"; INSTALL_EXIT=1; }
fi

# ── btrfs ───────────────────────────────────────────────────────
[[ $INSTALL_EXIT -eq 0 ]] && run_step "Creating btrfs filesystem" mkfs.btrfs -f /dev/mapper/root

if [[ $INSTALL_EXIT -eq 0 ]]; then
    s_step "Creating btrfs subvolumes..."
    {
        echo ""; echo "==> btrfs subvolumes"
        mount /dev/mapper/root /mnt
        btrfs subvolume create /mnt/@
        btrfs subvolume create /mnt/@home
        btrfs subvolume create /mnt/@snapshots
        btrfs subvolume create /mnt/@var_log
        umount /mnt
    } >> "$INSTALL_LOG" 2>&1 && s_ok "Subvolumes created" || { s_fail "Subvolume creation failed"; INSTALL_EXIT=1; }
fi

# ── Mount ───────────────────────────────────────────────────────
if [[ $INSTALL_EXIT -eq 0 ]]; then
    s_step "Mounting filesystems..."
    {
        echo ""; echo "==> mount"
        mount -o subvol=@,compress=zstd,noatime /dev/mapper/root /mnt
        mkdir -p /mnt/{boot,home,.snapshots,var/log}
        mount "$EFI_PART" /mnt/boot
        mount -o subvol=@home,compress=zstd,noatime      /dev/mapper/root /mnt/home
        mount -o subvol=@snapshots,compress=zstd,noatime /dev/mapper/root /mnt/.snapshots
        mount -o subvol=@var_log,compress=zstd,noatime   /dev/mapper/root /mnt/var/log
    } >> "$INSTALL_LOG" 2>&1 && s_ok "Filesystems mounted" || { s_fail "Mount failed"; INSTALL_EXIT=1; }
fi

# ── Packages ────────────────────────────────────────────────────
if [[ $INSTALL_EXIT -eq 0 ]]; then
    s_step "Installing packages (5-15 min, downloading from internet)..."
    {
        echo ""; echo "==> pacstrap"
        pacstrap -K /mnt \
            base base-devel linux linux-firmware amd-ucode \
            btrfs-progs cryptsetup dosfstools efibootmgr \
            networkmanager zram-generator linux-headers \
            hyprland hyprpaper hyprlock hypridle hyprpicker \
            xdg-desktop-portal-hyprland \
            foot waybar fuzzel dunst \
            grim slurp wl-clipboard cliphist polkit-gnome \
            cava brightnessctl playerctl pulsemixer wtype \
            bluez bluez-utils \
            yazi ffmpegthumbnailer p7zip unarchiver poppler \
            fd ripgrep fzf zoxide \
            imv mpv zathura zathura-pdf-mupdf \
            btop neovim ncdu \
            git lazygit \
            greetd greetd-tuigreet plymouth \
            jq bc imagemagick python-pillow \
            starship zsh zsh-autosuggestions zsh-syntax-highlighting \
            bat eza tokei procs duf dust \
            noto-fonts noto-fonts-cjk noto-fonts-emoji \
            ttf-jetbrains-mono-nerd ttf-font-awesome \
            qt5-wayland qt6-wayland \
            pipewire pipewire-alsa pipewire-audio pipewire-jack pipewire-pulse wireplumber \
            power-profiles-daemon fprintd fwupd libfprint iio-sensor-proxy \
            mesa vulkan-radeon libva-mesa-driver \
            acpi_call acpid wf-recorder inotify-tools xdg-utils \
            gum
    } >> "$INSTALL_LOG" 2>&1 && s_ok "Packages installed" || { s_fail "Package installation failed"; INSTALL_EXIT=1; }
fi

# ── fstab ───────────────────────────────────────────────────────
if [[ $INSTALL_EXIT -eq 0 ]]; then
    run_step "Generating fstab" bash -c 'genfstab -U /mnt >> /mnt/etc/fstab'
    LUKS_UUID=$(blkid -s UUID -o value "$ROOT_PART" 2>/dev/null)
fi

# ── Chroot configuration ─────────────────────────────────────────
if [[ $INSTALL_EXIT -eq 0 ]]; then
    s_step "Configuring system in chroot..."
    mkdir -p /mnt/tmp
    CHROOT_SCRIPT=/mnt/tmp/sumi-chroot-setup.sh

    # Write the setup script — outer ${VAR} expands here (not sensitive),
    # password is passed via SUMI_PASS env var (never written to disk).
    cat > "$CHROOT_SCRIPT" << CHROOT_EOF
#!/bin/bash
set -euo pipefail

TIMEZONE="${TIMEZONE}"
HOSTNAME="${HOSTNAME}"
USERNAME="${USERNAME}"
LUKS_UUID="${LUKS_UUID}"

echo "==> timezone"
ln -sf "/usr/share/zoneinfo/\${TIMEZONE}" /etc/localtime
hwclock --systohc

echo "==> locale"
echo "en_US.UTF-8 UTF-8" > /etc/locale.gen
locale-gen
echo "LANG=en_US.UTF-8" > /etc/locale.conf
echo "KEYMAP=us" > /etc/vconsole.conf

echo "==> hostname"
echo "\${HOSTNAME}" > /etc/hostname
printf '127.0.0.1\tlocalhost\n::1\t\tlocalhost\n127.0.1.1\t%s.localdomain %s\n' \
    "\${HOSTNAME}" "\${HOSTNAME}" > /etc/hosts

echo "==> pacman (multilib + parallel downloads)"
sed -i '/^\[multilib\]/,/^Include/ s/^#//' /etc/pacman.conf
sed -i 's/^#ParallelDownloads.*/ParallelDownloads = 5/' /etc/pacman.conf

echo "==> mkinitcpio (encrypt hook for LUKS)"
sed -i 's/^HOOKS=.*/HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block encrypt filesystems fsck)/' \
    /etc/mkinitcpio.conf
mkinitcpio -P

echo "==> systemd-boot"
bootctl install
mkdir -p /boot/loader/entries
cat > /boot/loader/loader.conf << 'LOADER'
default arch.conf
timeout 3
console-mode max
editor no
LOADER
cat > /boot/loader/entries/arch.conf << ENTRY
title   Arch Linux
linux   /vmlinuz-linux
initrd  /amd-ucode.img
initrd  /initramfs-linux.img
options cryptdevice=UUID=\${LUKS_UUID}:root root=/dev/mapper/root rootflags=subvol=@ rw quiet
ENTRY

echo "==> users"
echo "root:\${SUMI_PASS}" | chpasswd
useradd -m -G wheel,video,audio,input,storage -s /bin/zsh "\${USERNAME}"
echo "\${USERNAME}:\${SUMI_PASS}" | chpasswd
echo "%wheel ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/wheel
chmod 440 /etc/sudoers.d/wheel

echo "==> services"
systemctl enable NetworkManager
systemctl enable bluetooth
systemctl enable power-profiles-daemon
systemctl enable acpid
systemctl enable fstrim.timer
systemctl enable systemd-boot-update.service

echo "==> zram"
cat > /etc/systemd/zram-generator.conf << 'ZRAM'
[zram0]
zram-size = min(ram / 2, 8192)
compression-algorithm = zstd
ZRAM

echo "==> done"
CHROOT_EOF

    chmod +x "$CHROOT_SCRIPT"
    {
        echo ""; echo "==> chroot setup"
        SUMI_PASS="$PASSWORD" arch-chroot /mnt /tmp/sumi-chroot-setup.sh
    } >> "$INSTALL_LOG" 2>&1 && s_ok "System configured" || { s_fail "Chroot configuration failed"; INSTALL_EXIT=1; }
    rm -f "$CHROOT_SCRIPT"
fi

set -e

# ── Error page ───────────────────────────────────────────────────
if [[ $INSTALL_EXIT -ne 0 ]]; then
    PORT=7777
    WEB_DIR=$(mktemp -d)
    python3 - <<PYEOF
import html, pathlib
log_path = "$INSTALL_LOG"
exit_code = "$INSTALL_EXIT"
web_dir = "$WEB_DIR"
try:
    log = pathlib.Path(log_path).read_text(errors='replace')
except Exception as e:
    log = f"Log file not available ({e})"

lines = log.splitlines()
tail = '\n'.join(lines[-80:]) if len(lines) > 80 else log

page = f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>sumi :: install error</title>
  <style>
    *{{box-sizing:border-box}}
    body{{background:#0f0f17;color:#cdd6f4;font-family:monospace;padding:2em;max-width:1400px;margin:0 auto}}
    h1{{color:#f38ba8;margin:0 0 .4em}}
    .meta{{color:#6c7086;font-size:.85em;margin-bottom:1.5em}}
    .meta span{{color:#a6adc8}}
    .tabs{{display:flex;gap:.5em;margin-bottom:1em}}
    .tabs button{{background:#1e1e2e;border:1px solid #313244;color:#cdd6f4;
                  padding:.4em 1.2em;border-radius:6px;cursor:pointer;font:inherit;font-size:.9em}}
    .tabs button.active{{background:#313244;color:#f38ba8;border-color:#f38ba8}}
    pre{{background:#1e1e2e;padding:1.5em;border-radius:8px;overflow:auto;
         line-height:1.6;font-size:.8em;white-space:pre-wrap;word-break:break-all;
         border:1px solid #313244;margin:0}}
  </style>
</head>
<body>
  <h1>sumi :: installation failed</h1>
  <p class="meta">exit code <span>{exit_code}</span> &nbsp;·&nbsp; log <span>{log_path}</span></p>
  <div class="tabs">
    <button class="active" onclick="show('tail')" id="tab-tail">Error (last 80 lines)</button>
    <button onclick="show('full')" id="tab-full">Full Log</button>
  </div>
  <pre id="pane-tail">{html.escape(tail)}</pre>
  <pre id="pane-full" style="display:none">{html.escape(log)}</pre>
  <script>
  function show(id){{
    ['tail','full'].forEach(function(p){{
      document.getElementById('pane-'+p).style.display=p===id?'':'none';
      document.getElementById('tab-'+p).className=p===id?'active':'';
    }});
  }}
  </script>
</body>
</html>"""
pathlib.Path(web_dir + "/index.html").write_text(page)
PYEOF

    ( cd "$WEB_DIR" && python3 -m http.server $PORT ) &>/dev/null &
    SERVER_PID=$!
    trap "kill $SERVER_PID 2>/dev/null; rm -rf '$WEB_DIR'" EXIT

    echo ""
    gum style \
        --border rounded --border-foreground '#f38ba8' --padding "1 3" \
        "$(gum style --foreground '#f38ba8' --bold '  Installation failed')" \
        "" \
        "  Open in a browser on your network:" \
        "$(gum style --foreground '#89b4fa' "  http://${LOCAL_IP}:${PORT}")" \
        "" \
        "$(gum style --foreground '#6c7086' '  Press Ctrl+C to stop the server and exit.')"

    wait $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo ""
s_ok "Installation complete!"

# ── 6. Stage sumi for first boot ─────────────────────────────────
s_section "Staging First Boot"

TARGET="/mnt"

if [[ ! -d "$TARGET/home/$USERNAME" ]]; then
    s_warn "Cannot find /mnt/home/$USERNAME — stage manually after reboot."
    gum style \
        --border rounded --border-foreground '#f9e2af' --padding "0 2" \
        "  git clone <repo> ~/sumi" \
        "  cd ~/sumi && ./install.sh"
    exit 0
fi

s_step "Copying sumi repo..."
SUMI_DEST="$TARGET/home/$USERNAME/sumi"
rm -rf "$SUMI_DEST" 2>/dev/null || true
cp -r "$SCRIPT_DIR" "$SUMI_DEST"
arch-chroot "$TARGET" chown -R "$USERNAME:$USERNAME" "/home/$USERNAME/sumi" 2>/dev/null || true
s_ok "Repo copied to /home/$USERNAME/sumi"

cat > "$TARGET/home/$USERNAME/.sumi-first-boot.sh" << 'FIRSTBOOT'
#!/usr/bin/env bash
# sumi first-boot installer — runs once then self-deletes
MARKER="$HOME/.sumi-first-boot-done"
[[ -f "$MARKER" ]] && exit 0

if [[ -d "$HOME/sumi" ]]; then
    echo ""
    echo "╔══════════════════════════════════════╗"
    echo "║   sumi :: first boot setup           ║"
    echo "╚══════════════════════════════════════╝"
    echo ""
    cd "$HOME/sumi"
    chmod +x install.sh
    if bash install.sh; then
        touch "$MARKER"
        rm -f "$HOME/.sumi-first-boot.sh"
        echo ""
        echo "sumi installed! Reboot for the full experience."
    else
        echo ""
        echo "Install failed. Retry on next login, or:"
        echo "  cd ~/sumi && ./install.sh"
    fi
fi
FIRSTBOOT
chmod +x "$TARGET/home/$USERNAME/.sumi-first-boot.sh"
arch-chroot "$TARGET" chown "$USERNAME:$USERNAME" "/home/$USERNAME/.sumi-first-boot.sh" 2>/dev/null || true

BASH_PROFILE="$TARGET/home/$USERNAME/.bash_profile"
if [[ ! -f "$BASH_PROFILE" ]] || ! grep -q "sumi-first-boot" "$BASH_PROFILE" 2>/dev/null; then
    cat >> "$BASH_PROFILE" << 'EOF'

# sumi first-boot hook (runs once, installs the rice, then removes itself)
[[ -f "$HOME/.sumi-first-boot.sh" ]] && bash "$HOME/.sumi-first-boot.sh"
EOF
    arch-chroot "$TARGET" chown "$USERNAME:$USERNAME" "/home/$USERNAME/.bash_profile" 2>/dev/null || true
fi
s_ok "First-boot hook installed"

# ── 7. Done ──────────────────────────────────────────────────────
echo ""
gum style \
    --border rounded --border-foreground '#a6e3a1' \
    --margin "1 2" --padding "1 4" \
    "$(gum style --foreground '#a6e3a1' --bold '  ✓  Arch Linux installed')" \
    "" \
    "$(gum style --foreground '#cba6f7' '  First boot:')" \
    "  1. Unlock LUKS  (your encryption password)" \
    "  2. Login as $(gum style --foreground '#f38ba8' "$USERNAME") at the TTY" \
    "  3. sumi installs automatically" \
    "" \
    "$(gum style --foreground '#cba6f7' '  Second reboot:')" \
    "  · Full Hyprland desktop" \
    "  · SUPER+X  control center" \
    "  · SUPER+/  keybind cheatsheet" \
    "" \
    "$(gum style --foreground '#585b70' '  If first-boot fails:  cd ~/sumi && ./install.sh')"

echo ""
gum confirm "  Reboot now?" && {
    umount -R /mnt 2>/dev/null || true
    cryptsetup close root 2>/dev/null || true
    reboot
}
