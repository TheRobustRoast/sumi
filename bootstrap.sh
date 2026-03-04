#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  sumi :: bootstrap                                          ║
# ║                                                              ║
# ║  Run this from the Arch Linux live ISO.                      ║
# ║  It connects to WiFi, collects your settings, patches the   ║
# ║  archinstall configs, runs archinstall, then stages the      ║
# ║  post-install script to run on first boot.                   ║
# ║                                                              ║
# ║  Usage:                                                      ║
# ║    Boot Arch ISO → connect to internet (or let this help) →  ║
# ║    git clone <repo> /tmp/sumi                                ║
# ║    cd /tmp/sumi && chmod +x bootstrap.sh && ./bootstrap.sh   ║
# ╚══════════════════════════════════════════════════════════════╝

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONF="$SCRIPT_DIR/archinstall/user_configuration.json"
CREDS="$SCRIPT_DIR/archinstall/user_credentials.json"

# ── Colors ──────────────────────────────────────────────────
RED='\033[0;31m'
GRN='\033[0;32m'
CYN='\033[0;36m'
DIM='\033[0;90m'
YLW='\033[0;33m'
BLD='\033[1m'
RST='\033[0m'

clear
echo -e "${DIM}╔══════════════════════════════════════════╗${RST}"
echo -e "${CYN}║   sumi :: arch linux bootstrap           ║${RST}"
echo -e "${DIM}║   Framework 13 AMD · Hyprland · TUI      ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════════╝${RST}"
echo ""

info()    { echo -e "${CYN}:: ${RST}$1"; }
ok()      { echo -e "${GRN}   ✓${RST} $1"; }
warn()    { echo -e "${YLW}   !${RST} $1"; }
err()     { echo -e "${RED}   ✗${RST} $1"; }
prompt()  { echo -ne "${CYN}:: ${RST}$1"; }

# ── 0. Sanity checks ───────────────────────────────────────
if [[ ! -f "$CONF" ]]; then
    err "Cannot find: $CONF"
    err "Make sure you're running this from the sumi repo root."
    err "Expected: cd /path/to/sumi && ./bootstrap.sh"
    exit 1
fi

if [[ ! -f "$CREDS" ]]; then
    warn "user_credentials.json not found — generating template..."
    mkdir -p "$(dirname "$CREDS")"
    cat > "$CREDS" << 'CREDSTPL'
{
    "!root-password": "!CHANGE_ME!",
    "!users": [
        {
            "!password": "!CHANGE_ME!",
            "sudo": true,
            "username": "user"
        }
    ],
    "encryption_password": "!CHANGE_ME!"
}
CREDSTPL
    ok "Generated $CREDS"
fi

if [[ ! -d /run/archiso ]]; then
    warn "This doesn't look like the Arch live ISO."
    warn "This script is meant to be run from the installer."
    prompt "Continue anyway? [y/N] "
    read -r cont
    [[ ! "$cont" =~ ^[Yy]$ ]] && exit 0
fi

# ── 1. Network ─────────────────────────────────────────────
echo ""
info "Checking network connectivity..."

if ping -c 1 -W 3 archlinux.org &>/dev/null; then
    ok "Internet connection detected"
else
    warn "No internet connection."
    echo ""
    info "Connect via WiFi using iwctl:"
    echo -e "${DIM}   The WiFi setup wizard will now launch.${RST}"
    echo -e "${DIM}   Commands: station wlan0 scan${RST}"
    echo -e "${DIM}             station wlan0 get-networks${RST}"
    echo -e "${DIM}             station wlan0 connect <SSID>${RST}"
    echo -e "${DIM}             exit${RST}"
    echo ""
    prompt "Press Enter to open iwctl (or Ctrl+C to abort)... "
    read -r
    iwctl || true

    # Re-check
    sleep 2
    if ping -c 1 -W 3 archlinux.org &>/dev/null; then
        ok "Connected!"
    else
        err "Still no internet. Connect manually and re-run this script."
        exit 1
    fi
fi

# ── 2. Detect disk ─────────────────────────────────────────
echo ""
info "Detecting NVMe drive..."

DISKS=$(lsblk -dno NAME,SIZE,TYPE | grep disk | awk '{print "/dev/"$1" ("$2")"}') || true
NVME=$(lsblk -dno NAME,SIZE,TYPE | grep disk | grep nvme | head -1 | awk '{print "/dev/"$1}') || true

echo -e "${DIM}   Available disks:${RST}"
echo "$DISKS" | while read -r line; do
    echo -e "${DIM}     $line${RST}"
done
echo ""

if [[ -n "$NVME" ]]; then
    info "Detected NVMe: $NVME"
    prompt "Use $NVME? [Y/n] "
    read -r use_nvme
    if [[ "$use_nvme" =~ ^[Nn]$ ]]; then
        prompt "Enter device path (e.g., /dev/sda): "
        read -r NVME
    fi
else
    warn "No NVMe detected."
    prompt "Enter device path (e.g., /dev/nvme0n1 or /dev/sda): "
    read -r NVME
fi

if [[ ! -b "$NVME" ]]; then
    err "$NVME is not a valid block device."
    exit 1
fi

# Calculate root partition size from actual disk size.
# EFI ends at 1025 MiB; leave 1 MiB at the end for the backup GPT header.
DISK_SIZE_MIB=$(lsblk -bdno SIZE "$NVME" | awk '{print int($1/1024/1024)}')
ROOT_SIZE_MIB=$((DISK_SIZE_MIB - 1025 - 1))
ok "Disk size: ${DISK_SIZE_MIB} MiB → root partition: ${ROOT_SIZE_MIB} MiB"

# Detect GPU for the correct archinstall gfx_driver value
if lspci 2>/dev/null | grep -qi "nvidia"; then
    GFX_DRIVER="Nvidia (open-source)"
elif lspci 2>/dev/null | grep -qiE "intel.*(graphics|vga|display)"; then
    GFX_DRIVER="Intel (open-source)"
else
    GFX_DRIVER="AMD / ATI (open-source)"   # covers AMD, ATI, and unknown
fi
ok "GPU driver: $GFX_DRIVER"

echo -e "${RED}${BLD}"
echo "   ⚠  WARNING: This will WIPE $NVME completely."
echo -e "${RST}"
prompt "Type 'yes' to confirm: "
read -r wipe_confirm
if [[ "$wipe_confirm" != "yes" ]]; then
    info "Aborted."
    exit 0
fi

# ── 3. Collect user settings ──────────────────────────────
echo ""
echo -e "${DIM}╔══════════════════════════════════════════╗${RST}"
echo -e "${CYN}║   user configuration                     ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════════╝${RST}"
echo ""

# Username
prompt "Username: "
read -r USERNAME
if [[ -z "$USERNAME" ]]; then
    USERNAME="user"
    warn "Defaulting to 'user'"
fi

# Single password — used for user, root, and LUKS disk encryption
info "One password will be used for: user login, root, and LUKS disk encryption."
info "On cold boot, unlocking LUKS will automatically log you in."
while true; do
    prompt "Password: "
    read -rs PASSWORD
    echo ""
    prompt "Confirm password: "
    read -rs PASSWORD2
    echo ""
    if [[ "$PASSWORD" == "$PASSWORD2" ]]; then
        break
    else
        warn "Passwords don't match. Try again."
    fi
done
ROOT_PASSWORD="$PASSWORD"
LUKS_PASS="$PASSWORD"
ok "User, root, and LUKS passwords set to the same value"

# Hostname
echo ""
prompt "Hostname [framework]: "
read -r HOSTNAME
HOSTNAME="${HOSTNAME:-framework}"

# Timezone
echo ""
info "Timezone detection..."
DETECTED_TZ=$(curl -s --max-time 5 https://ipapi.co/timezone 2>/dev/null || echo "")
# Validate: a real timezone contains "/" (e.g. America/New_York) or is "UTC".
# Discard rate-limit messages, HTML, or other garbage the API might return.
if [[ "$DETECTED_TZ" =~ ^[A-Za-z_]+/[A-Za-z_/+\-]+$ ]] || [[ "$DETECTED_TZ" == "UTC" ]]; then
    prompt "Timezone [$DETECTED_TZ]: "
    read -r TIMEZONE
    TIMEZONE="${TIMEZONE:-$DETECTED_TZ}"
else
    [[ -n "$DETECTED_TZ" ]] && warn "Timezone detection returned invalid value: '$DETECTED_TZ'"
    prompt "Timezone (e.g., America/New_York) [UTC]: "
    read -r TIMEZONE
    TIMEZONE="${TIMEZONE:-UTC}"
fi

# ── 4. Patch the JSON configs ─────────────────────────────
echo ""
info "Generating archinstall configuration..."

# Ensure we have jq (should be on the live ISO)
if ! command -v jq &>/dev/null; then
    pacman -Sy --noconfirm jq
fi

# Patch user_configuration.json
PATCHED_CONF=$(mktemp)
PATCHED_CREDS=$(mktemp)

# Trap ensures password temp files are always cleaned up, even on crash/set -e exit
cleanup_secrets() {
    rm -f "$PATCHED_CONF" "$PATCHED_CREDS"
}
trap cleanup_secrets EXIT

jq \
    --arg disk "$NVME" \
    --arg host "$HOSTNAME" \
    --arg tz "$TIMEZONE" \
    --argjson root_mib "$ROOT_SIZE_MIB" \
    --arg gfx "$GFX_DRIVER" \
    '
    .disk_config.device_modifications[0].device = $disk |
    .disk_config.device_modifications[0].partitions[1].size.value = $root_mib |
    .disk_config.device_modifications[0].partitions[1].size.unit = "MiB" |
    .hostname = $host |
    .timezone = $tz |
    .profile_config.gfx_driver = $gfx
    ' "$CONF" > "$PATCHED_CONF"

ok "Patched user_configuration.json"
echo -e "${DIM}     disk:     $NVME${RST}"
echo -e "${DIM}     hostname: $HOSTNAME${RST}"
echo -e "${DIM}     timezone: $TIMEZONE${RST}"
echo -e "${DIM}     encrypt:  LUKS on root${RST}"
echo -e "${DIM}     gpu:      $GFX_DRIVER${RST}"

# Patch user_credentials.json
jq \
    --arg user "$USERNAME" \
    --arg pass "$PASSWORD" \
    --arg root "$ROOT_PASSWORD" \
    --arg luks "$LUKS_PASS" \
    '
    ."!root-password" = $root |
    ."!users"[0].username = $user |
    ."!users"[0]."!password" = $pass |
    .encryption_password = $luks
    ' "$CREDS" > "$PATCHED_CREDS"

ok "Patched user_credentials.json"
echo -e "${DIM}     user:     $USERNAME (sudo)${RST}"

# ── 5. Review before install ──────────────────────────────
echo ""
echo -e "${DIM}╔══════════════════════════════════════════╗${RST}"
echo -e "${YLW}║   review before install                  ║${RST}"
echo -e "${DIM}╚══════════════════════════════════════════╝${RST}"
echo ""
echo -e "  Disk:        ${BLD}$NVME${RST} (will be ${RED}WIPED${RST})"
echo -e "  Filesystem:  btrfs (zstd, subvols: @, @home, @snapshots, @var_log)"
echo -e "  Encryption:  LUKS on root partition"
echo -e "  Bootloader:  systemd-boot"
echo -e "  Hostname:    $HOSTNAME"
echo -e "  User:        $USERNAME (sudo, root & LUKS share same password)"
echo -e "  Timezone:    $TIMEZONE"
echo -e "  Audio:       PipeWire"
echo -e "  Network:     NetworkManager"
echo -e "  Desktop:     Hyprland + sumi rice"
echo -e "  Packages:    ~85 from repos + 5 from AUR (post-boot)"
echo ""
echo -e "${RED}${BLD}  This is your last chance to cancel.${RST}"
echo ""
prompt "Proceed with installation? [y/N] "
read -r go
if [[ ! "$go" =~ ^[Yy]$ ]]; then
    info "Aborted."
    exit 0
fi

# ── 6. Run archinstall ────────────────────────────────────
echo ""
info "Starting archinstall..."
echo -e "${DIM}   This will take 5-15 minutes depending on your connection.${RST}"
echo ""

# Make sure archinstall is up to date
pacman -Sy --noconfirm archinstall 2>/dev/null || true

# ── Pre-flight: clear any leftover mounts from a previous attempt ─────────
# archinstall 3.x bug: umount_all_existing() passes btrfs subvolume
# mountpoints (e.g. /mnt/var/log) to lsblk, which only accepts block
# devices. Unmounting /mnt beforehand prevents that cleanup path entirely.
if mountpoint -q /mnt 2>/dev/null; then
    info "Unmounting leftover /mnt mounts from previous attempt..."
    umount -R /mnt 2>/dev/null || true
fi

# Capture exit code without letting set -e kill the script
# (trap cleanup_secrets EXIT handles temp file deletion)
INSTALL_EXIT=0
archinstall \
    --config "$PATCHED_CONF" \
    --creds "$PATCHED_CREDS" \
    --silent || INSTALL_EXIT=$?

if [[ $INSTALL_EXIT -ne 0 ]]; then
    err "archinstall failed with exit code $INSTALL_EXIT"
    echo ""
    ARCHLOG="/var/log/archinstall/install.log"
    if [[ -f "$ARCHLOG" ]]; then
        info "Uploading install log for debugging..."
        LOG_URL=$(curl -s -F "file=@${ARCHLOG}" https://0x0.st 2>/dev/null || echo "")
        if [[ -n "$LOG_URL" ]]; then
            echo -e "${CYN}   Log URL: ${BLD}${LOG_URL}${RST}"
        else
            warn "Upload failed. Log is at: $ARCHLOG"
        fi
    else
        warn "No install log found at $ARCHLOG"
    fi
    exit 1
fi

ok "archinstall completed successfully!"

# ── 7. Create btrfs subvolumes ────────────────────────────
# archinstall 3.x has a bug with btrfs subvolume configs so we create
# them here manually after the install, then update fstab accordingly.
echo ""
info "Creating btrfs subvolumes..."

# Find the root block device archinstall used (the LUKS-mapped or raw partition)
ROOT_DEV=$(findmnt -n -o SOURCE /mnt 2>/dev/null || true)

if [[ -z "$ROOT_DEV" ]]; then
    warn "Could not detect mounted root device — skipping btrfs subvolumes."
    warn "Create them manually after reboot if needed."
else
    # archinstall already installed into the root of the btrfs fs.
    # Strategy: snapshot current root as @, create siblings, update fstab.
    BTRFS_MNT=$(mktemp -d)
    mount -o subvolid=5 "$ROOT_DEV" "$BTRFS_MNT"

    # Create a snapshot of the existing root content as the @ subvolume
    btrfs subvolume snapshot "$BTRFS_MNT" "$BTRFS_MNT/@" 2>/dev/null \
        || btrfs subvolume create "$BTRFS_MNT/@"

    # Create empty sibling subvolumes
    btrfs subvolume create "$BTRFS_MNT/@home"      2>/dev/null || true
    btrfs subvolume create "$BTRFS_MNT/@snapshots" 2>/dev/null || true
    btrfs subvolume create "$BTRFS_MNT/@var_log"   2>/dev/null || true

    # Populate @home and @var_log from the installed system if present
    [[ -d "$BTRFS_MNT/@/home" ]]    && cp -a "$BTRFS_MNT/@/home/."    "$BTRFS_MNT/@home/"
    [[ -d "$BTRFS_MNT/@/var/log" ]] && cp -a "$BTRFS_MNT/@/var/log/." "$BTRFS_MNT/@var_log/"

    umount "$BTRFS_MNT"
    rmdir  "$BTRFS_MNT"

    # Re-mount /mnt with the @ subvolume as root
    umount -R /mnt 2>/dev/null || true
    mount -o subvol=@,compress=zstd,noatime "$ROOT_DEV" /mnt
    mkdir -p /mnt/{boot,home,.snapshots,var/log}
    mount "${NVME}p1" /mnt/boot  2>/dev/null \
        || mount "${NVME}1" /mnt/boot 2>/dev/null || true
    mount -o subvol=@home,compress=zstd,noatime      "$ROOT_DEV" /mnt/home
    mount -o subvol=@snapshots,compress=zstd,noatime "$ROOT_DEV" /mnt/.snapshots
    mount -o subvol=@var_log,compress=zstd,noatime   "$ROOT_DEV" /mnt/var/log

    # Rewrite fstab with the correct subvolume mount options
    ROOT_UUID=$(blkid -s UUID -o value "$ROOT_DEV")
    EFI_UUID=$(blkid -s UUID -o value "${NVME}p1" 2>/dev/null \
               || blkid -s UUID -o value "${NVME}1" 2>/dev/null || echo "")
    {
        echo "# Generated by sumi bootstrap"
        [[ -n "$EFI_UUID" ]] && \
            echo "UUID=$EFI_UUID  /boot       vfat  defaults             0 2"
        echo "UUID=$ROOT_UUID  /            btrfs subvol=@,compress=zstd,noatime 0 0"
        echo "UUID=$ROOT_UUID  /home        btrfs subvol=@home,compress=zstd,noatime 0 0"
        echo "UUID=$ROOT_UUID  /.snapshots  btrfs subvol=@snapshots,compress=zstd,noatime 0 0"
        echo "UUID=$ROOT_UUID  /var/log     btrfs subvol=@var_log,compress=zstd,noatime 0 0"
    } > /mnt/etc/fstab

    ok "btrfs subvolumes created and fstab updated"
    echo -e "${DIM}     @            → /${RST}"
    echo -e "${DIM}     @home        → /home${RST}"
    echo -e "${DIM}     @snapshots   → /.snapshots${RST}"
    echo -e "${DIM}     @var_log     → /var/log${RST}"
fi

# ── 8. Stage sumi for first boot ──────────────────────────
echo ""
info "Staging sumi for first boot..."

# archinstall mounts the new system at /mnt
TARGET="/mnt"
if [[ ! -d "$TARGET/home/$USERNAME" ]]; then
    warn "Cannot find installed system at /mnt/home/$USERNAME"
    warn "archinstall may have used a different mount point or the user wasn't created."
    warn "You'll need to run install.sh manually after reboot."
    warn "Steps:"
    warn "  1. Reboot into the new system"
    warn "  2. git clone <repo> ~/sumi"
    warn "  3. cd ~/sumi && chmod +x install.sh && ./install.sh"
    exit 0
fi

# Copy sumi repo to the new user's home (use cp -rT to replace, not nest, on re-run)
SUMI_DEST="$TARGET/home/$USERNAME/sumi"
rm -rf "$SUMI_DEST" 2>/dev/null || true
cp -r "$SCRIPT_DIR" "$SUMI_DEST"
# Fix ownership (archinstall creates the user, we need their UID)
arch-chroot "$TARGET" chown -R "$USERNAME:$USERNAME" "/home/$USERNAME/sumi" 2>/dev/null || true
ok "Copied sumi to /home/$USERNAME/sumi"

# Create a first-boot script that runs install.sh automatically
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
    echo "Running sumi post-install..."
    echo ""
    cd "$HOME/sumi"
    chmod +x install.sh
    if bash install.sh; then
        touch "$MARKER"
        rm -f "$HOME/.sumi-first-boot.sh"
        echo ""
        echo "sumi installed successfully! Reboot for the full experience."
    else
        echo ""
        echo "Install failed. It will retry on next login, or run manually:"
        echo "  cd ~/sumi && ./install.sh"
    fi
fi
FIRSTBOOT
chmod +x "$TARGET/home/$USERNAME/.sumi-first-boot.sh"
arch-chroot "$TARGET" chown "$USERNAME:$USERNAME" "/home/$USERNAME/.sumi-first-boot.sh" 2>/dev/null || true

# Add to .bash_profile so it runs on first TTY login (before zsh is default shell)
BASH_PROFILE="$TARGET/home/$USERNAME/.bash_profile"
if [[ ! -f "$BASH_PROFILE" ]] || ! grep -q "sumi-first-boot" "$BASH_PROFILE" 2>/dev/null; then
    cat >> "$BASH_PROFILE" << 'EOF'

# sumi first-boot hook (runs once, installs the rice, then removes itself)
[[ -f "$HOME/.sumi-first-boot.sh" ]] && bash "$HOME/.sumi-first-boot.sh"
EOF
    arch-chroot "$TARGET" chown "$USERNAME:$USERNAME" "/home/$USERNAME/.bash_profile" 2>/dev/null || true
fi
ok "First-boot hook installed"

# ── 8. Done ───────────────────────────────────────────────
echo ""
echo -e "${DIM}╔══════════════════════════════════════════╗${RST}"
echo -e "${GRN}║   sumi :: bootstrap complete             ║${RST}"
echo -e "${DIM}╠══════════════════════════════════════════╣${RST}"
echo -e "${DIM}║${RST}                                          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  Arch Linux is installed. On first boot:  ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  1. Unlock LUKS (your encryption pass)   ${DIM}║${RST}"
echo -e "${DIM}║${RST}  2. Login at the TTY as ${BLD}$USERNAME${RST}          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  3. sumi installs automatically          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  4. Reboot again for greetd + plymouth   ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  After the second reboot, you'll get:    ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • Plymouth TUI unlock → auto login      ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • Full Hyprland desktop with theming    ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • SUPER+X for control center            ${DIM}║${RST}"
echo -e "${DIM}║${RST}  • SUPER+/ for keybind cheatsheet        ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                          ${DIM}║${RST}"
echo -e "${DIM}║${RST}  If first-boot auto-install fails, run:  ${DIM}║${RST}"
echo -e "${DIM}║${RST}  cd ~/sumi && ./install.sh               ${DIM}║${RST}"
echo -e "${DIM}║${RST}                                          ${DIM}║${RST}"
echo -e "${DIM}╚══════════════════════════════════════════╝${RST}"
echo ""
prompt "Reboot now? [Y/n] "
read -r reboot_answer
if [[ ! "$reboot_answer" =~ ^[Nn]$ ]]; then
    umount -R /mnt 2>/dev/null || true
    reboot
fi
