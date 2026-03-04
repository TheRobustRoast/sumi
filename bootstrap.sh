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

# Colors (used for non-dialog terminal output)
RED='\033[0;31m'
GRN='\033[0;32m'
CYN='\033[0;36m'
YLW='\033[0;33m'
BLD='\033[1m'
RST='\033[0m'

BT="sumi :: arch linux bootstrap"

info()  { echo -e "${CYN}:: ${RST}$1"; }
ok()    { echo -e "${GRN}   ✓${RST} $1"; }
warn()  { echo -e "${YLW}   !${RST} $1"; }
err()   { echo -e "${RED}   ✗${RST} $1"; }

# ── 0. Sanity checks ───────────────────────────────────────────
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

# ── Install dialog if missing ──────────────────────────────────
if ! command -v dialog &>/dev/null; then
    info "Installing dialog..."
    pacman -Sy --noconfirm dialog &>/dev/null
fi

# ── Welcome ────────────────────────────────────────────────────
dialog --backtitle "$BT" \
    --title " sumi " \
    --msgbox "\n  Framework 13 AMD  ·  Hyprland  ·  TUI\n\n  This script installs Arch Linux with:\n    · Full disk encryption (LUKS2)\n    · btrfs with subvolumes\n    · Hyprland + sumi rice (post-boot)\n\n  Press OK to begin." 13 52

# ── Non-ISO warning ────────────────────────────────────────────
if [[ ! -d /run/archiso ]]; then
    dialog --backtitle "$BT" \
        --title " Warning " \
        --defaultno \
        --yesno "\n  This doesn't look like the Arch Linux live ISO.\n\n  This script is designed to run from the installer.\n\n  Continue anyway?" 11 52 || exit 0
fi

# ── 1. Network ─────────────────────────────────────────────────
dialog --backtitle "$BT" \
    --title " Network " \
    --infobox "\n  Checking internet connectivity..." 5 42

if ! ping -c 1 -W 3 archlinux.org &>/dev/null; then
    dialog --backtitle "$BT" \
        --title " No Internet " \
        --msgbox "\n  No internet connection detected.\n\n  Press OK to open iwctl and connect to WiFi.\n\n  Commands:\n    station wlan0 scan\n    station wlan0 get-networks\n    station wlan0 connect <SSID>\n    exit" 15 52

    clear
    iwctl || true

    sleep 2
    if ! ping -c 1 -W 3 archlinux.org &>/dev/null; then
        dialog --backtitle "$BT" \
            --title " Error " \
            --msgbox "\n  Still no internet connection.\n\n  Please connect manually and re-run this script." 9 52
        exit 1
    fi
fi

# ── 2. Detect disk ─────────────────────────────────────────────
dialog --backtitle "$BT" \
    --title " Disk Detection " \
    --infobox "\n  Scanning for disks..." 5 38

NVME_DEFAULT=$(lsblk -dno NAME,TYPE | grep disk | grep nvme | head -1 | awk '{print "/dev/"$1}') || true

MENU_ITEMS=()
while read -r name size; do
    MENU_ITEMS+=("/dev/$name" "$size")
done < <(lsblk -dno NAME,SIZE,TYPE | grep disk | awk '{print $1, $2}')

if [[ ${#MENU_ITEMS[@]} -eq 0 ]]; then
    dialog --backtitle "$BT" --title " Error " \
        --msgbox "\n  No disks detected.\n\n  Please check your hardware and try again." 9 48
    exit 1
fi

NVME=$(dialog --backtitle "$BT" \
    --title " Select Installation Disk " \
    --stdout \
    --default-item "${NVME_DEFAULT:-${MENU_ITEMS[0]}}" \
    --menu "\n  Select the disk to install Arch Linux on.\n\n  WARNING: The selected disk will be WIPED." 14 56 6 \
    "${MENU_ITEMS[@]}") || { info "Aborted."; exit 0; }

if [[ ! -b "$NVME" ]]; then
    dialog --backtitle "$BT" --title " Error " \
        --msgbox "\n  $NVME is not a valid block device." 7 48
    exit 1
fi

# Calculate root partition size from actual disk size.
# EFI ends at 1025 MiB; leave 1 MiB at the end for the backup GPT header.
DISK_SIZE_MIB=$(lsblk -bdno SIZE "$NVME" | awk '{print int($1/1024/1024)}')
ROOT_SIZE_MIB=$((DISK_SIZE_MIB - 1025 - 1))

# Detect GPU for the correct archinstall gfx_driver value
if lspci 2>/dev/null | grep -qi "nvidia"; then
    GFX_DRIVER="Nvidia (open-source)"
elif lspci 2>/dev/null | grep -qiE "intel.*(graphics|vga|display)"; then
    GFX_DRIVER="Intel (open-source)"
else
    GFX_DRIVER="AMD / ATI (open-source)"   # covers AMD, ATI, and unknown
fi

# Wipe confirmation
DISK_SIZE_HUMAN=$(lsblk -dno SIZE "$NVME")
dialog --backtitle "$BT" \
    --title " ⚠  DESTRUCTIVE ACTION " \
    --defaultno \
    --yesno "\n  $NVME  ($DISK_SIZE_HUMAN)\n\n  ALL DATA ON THIS DISK WILL BE PERMANENTLY ERASED.\n  This cannot be undone.\n\n  Are you absolutely sure?" 12 56 || { info "Aborted."; exit 0; }

# ── 3. Collect user settings ───────────────────────────────────

# Username
USERNAME=$(dialog --backtitle "$BT" \
    --title " Username " \
    --stdout \
    --inputbox "\n  Enter your username:" 8 44 "user") || { info "Aborted."; exit 0; }
[[ -z "$USERNAME" ]] && USERNAME="user"

# Single password — used for user, root, and LUKS disk encryption
while true; do
    PASSWORD=$(dialog --backtitle "$BT" \
        --title " Password " \
        --stdout \
        --passwordbox "\n  One password for: login, root, and LUKS encryption.\n\n  Enter password:" 10 54) || { info "Aborted."; exit 0; }
    PASSWORD2=$(dialog --backtitle "$BT" \
        --title " Password " \
        --stdout \
        --passwordbox "\n  Confirm password:" 8 54) || { info "Aborted."; exit 0; }
    if [[ "$PASSWORD" == "$PASSWORD2" ]]; then
        break
    fi
    dialog --backtitle "$BT" --title " Mismatch " \
        --msgbox "\n  Passwords don't match. Please try again." 7 46
done
ROOT_PASSWORD="$PASSWORD"
LUKS_PASS="$PASSWORD"

# Hostname
HOSTNAME=$(dialog --backtitle "$BT" \
    --title " Hostname " \
    --stdout \
    --inputbox "\n  Enter hostname:" 8 44 "framework") || { info "Aborted."; exit 0; }
HOSTNAME="${HOSTNAME:-framework}"

# Timezone
dialog --backtitle "$BT" \
    --title " Timezone " \
    --infobox "\n  Detecting timezone from IP..." 5 38

DETECTED_TZ=$(curl -s --max-time 5 https://ipapi.co/timezone 2>/dev/null || echo "")
# Validate: a real timezone contains "/" (e.g. America/New_York) or is "UTC".
if [[ "$DETECTED_TZ" =~ ^[A-Za-z_]+/[A-Za-z_/+\-]+$ ]] || [[ "$DETECTED_TZ" == "UTC" ]]; then
    TZ_DEFAULT="$DETECTED_TZ"
else
    TZ_DEFAULT="UTC"
fi

TIMEZONE=$(dialog --backtitle "$BT" \
    --title " Timezone " \
    --stdout \
    --inputbox "\n  Enter timezone (e.g., America/New_York):" 8 54 "$TZ_DEFAULT") || { info "Aborted."; exit 0; }
TIMEZONE="${TIMEZONE:-UTC}"

# ── 4. Patch the JSON configs ──────────────────────────────────
dialog --backtitle "$BT" \
    --title " Configuration " \
    --infobox "\n  Generating archinstall configuration..." 5 46

# Ensure we have jq (should be on the live ISO)
if ! command -v jq &>/dev/null; then
    pacman -Sy --noconfirm jq
fi

PATCHED_CONF=$(mktemp)
PATCHED_CREDS=$(mktemp)

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

# ── 5. Review before install ───────────────────────────────────
dialog --backtitle "$BT" \
    --title " Review — Final Confirmation " \
    --defaultno \
    --yesno "\
  Disk:        $NVME  ($DISK_SIZE_HUMAN)  ← WIPED
  Filesystem:  btrfs + LUKS2 encryption
  Bootloader:  systemd-boot
  Hostname:    $HOSTNAME
  User:        $USERNAME  (sudo, shared password)
  Timezone:    $TIMEZONE
  GPU:         $GFX_DRIVER
  Audio:       PipeWire
  Network:     NetworkManager
  Desktop:     Hyprland + sumi rice

  This is your last chance to cancel.

  Proceed with installation?" 19 58 || { info "Aborted."; exit 0; }

# ── 6. Run archinstall ─────────────────────────────────────────

# Detect local IP now (used in infobox and error page)
LOCAL_IP=$(ip -4 addr show scope global \
    | awk '/inet / {split($2,a,"/"); print a[1]; exit}' 2>/dev/null || true)
[[ -z "$LOCAL_IP" ]] && LOCAL_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || true)
[[ -z "$LOCAL_IP" ]] && LOCAL_IP="this-machine"

# Make sure archinstall is up to date
info "Updating archinstall..."
pacman -Sy --noconfirm archinstall 2>/dev/null || true

# Patch archinstall luks.py: treat cryptsetup exit 5 ("device already open")
# as success instead of crashing. In archinstall 3.x, udevd auto-opens the
# newly luksFormat'd partition and the subsequent cryptsetup open call in
# format_encrypted then fails with CRYPT_BUSY (exit 5).
python3 - << 'LUKS_PATCH'
import glob, pathlib, sys
matches = glob.glob('/usr/lib/python*/site-packages/archinstall/lib/luks.py')
if not matches:
    print("WARNING: luks.py not found, skipping patch")
    sys.exit(0)
p = pathlib.Path(matches[0])
src = p.read_text()
if 'returncode == 5' in src:
    sys.exit(0)  # already patched
lines = src.splitlines()
for i, line in enumerate(lines):
    if 'result = run(cmd, input_data=passphrase)' in line:
        ind = ' ' * (len(line) - len(line.lstrip()))
        lines[i] = '\n'.join([
            f'{ind}try:',
            f'{ind}    result = run(cmd, input_data=passphrase)',
            f'{ind}except Exception as _e:',
            f'{ind}    if getattr(_e, "returncode", None) == 5:',
            f'{ind}        pass  # mapper already open (udevd race) — reuse it',
            f'{ind}    else:',
            f'{ind}        raise',
        ])
        p.write_text('\n'.join(lines))
        print(f"patched {p}")
        break
else:
    print("WARNING: unlock target line not found in luks.py (version mismatch)")
LUKS_PATCH

# ── Pre-flight: clear ALL stale disk state from previous attempts ──────────
# archinstall 3.x: /dev/mapper/root from a prior run causes cryptsetup
# to fail with exit 5 ("device already exists") when it tries to open the
# freshly-formatted LUKS partition. Even after umount, the kernel's btrfs
# subsystem holds a reference that blocks dmsetup remove --force.
# Fix: drop the page/dentry/inode cache to release kernel FS references,
# then explicitly close the root mapper before the general DM cleanup.

# 1. Unmount everything under /mnt
if mountpoint -q /mnt 2>/dev/null; then
    info "Unmounting leftover /mnt mounts..."
    umount -R /mnt 2>/dev/null || true
fi

# 2. Drop caches so the kernel releases btrfs inode/dentry refs on the mapper
sync
echo 3 > /proc/sys/vm/drop_caches 2>/dev/null || true
sleep 1

# 3. Close all non-ISO device mapper devices
# Try cryptsetup first (graceful), then dmsetup --force (nuclear)
# "root" is archinstall's LUKS mapper name — handle it explicitly first
for name in root $(dmsetup ls 2>/dev/null | awk '{print $1}'); do
    [[ "$name" == "ventoy" ]] && continue
    [[ "$name" =~ ^sda ]] && continue
    [[ ! -b "/dev/mapper/$name" ]] && continue
    cryptsetup close "$name" 2>/dev/null || true
    [[ -b "/dev/mapper/$name" ]] && dmsetup remove --force "$name" 2>/dev/null || true
done

# 4. Wipe all FS/partition-table signatures so archinstall finds a blank disk
wipefs -af "$NVME" 2>/dev/null || true
partprobe "$NVME" 2>/dev/null || true
udevadm settle --timeout 10 2>/dev/null || true

# Clear the install log so the error page only shows the current run
mkdir -p /var/log/archinstall
: > /var/log/archinstall/install.log

dialog --backtitle "$BT" \
    --title " Installing Arch Linux " \
    --infobox "\
  archinstall is now running silently.

  This will take 5-15 minutes depending
  on your internet connection speed.

  Do NOT close this terminal.

  If it fails, open in a browser:
  http://${LOCAL_IP}:7777" 14 52

# Capture exit code without letting set -e kill the script.
# Redirect stderr to a temp file — archinstall crashes (unhandled
# Python exceptions) print to stderr, not the log file.
INSTALL_EXIT=0
STDERR_LOG=$(mktemp)
archinstall \
    --config "$PATCHED_CONF" \
    --creds "$PATCHED_CREDS" \
    --silent 2>"$STDERR_LOG" || INSTALL_EXIT=$?

if [[ $INSTALL_EXIT -ne 0 ]]; then
    ARCHLOG="/var/log/archinstall/install.log"

    # Append stderr to the log so the error page shows the full picture
    {
        echo ""
        echo "=== stderr (unhandled exceptions / crash output) ==="
        cat "$STDERR_LOG" 2>/dev/null || echo "(no stderr output)"
        echo "=== archinstall exit code: $INSTALL_EXIT ==="
    } >> "$ARCHLOG"
    rm -f "$STDERR_LOG"
    PORT=7777

    # Build a self-contained HTML error page from the install log
    WEB_DIR=$(mktemp -d)
    python3 - <<PYEOF
import html, pathlib
log_path = "$ARCHLOG"
exit_code = "$INSTALL_EXIT"
web_dir = "$WEB_DIR"
try:
    log = pathlib.Path(log_path).read_text(errors='replace')
except Exception as e:
    log = f"Log file not available ({e})"

# Clean view: collapse each Python traceback to just the exception message.
# Full traceback is still available in the "Full Log" tab.
import re as _re
lines = log.splitlines()
cleaned = []
i = 0
while i < len(lines):
    line = lines[i]
    if '- ERROR -' in line:
        # Strip "Traceback (most recent call last):" from the header line
        cleaned.append(_re.sub(r'\s*Traceback \(most recent call last\):\s*$', '', line))
        # Collect the traceback block (everything until the next timestamp)
        j, block = i + 1, []
        while j < len(lines) and not lines[j].startswith('['):
            block.append(lines[j])
            j += 1
        # Find the last meaningful line = the actual exception class: message
        for bl in reversed(block):
            s = bl.strip()
            if s and not s.startswith('File ') and not all(c in ' ~^' for c in s) and s != '...':
                cleaned.append('  ' + s)
                break
        i = j
    else:
        cleaned.append(line)
        i += 1
error_section = '\n'.join(cleaned)

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
  <h1>sumi :: archinstall failed</h1>
  <p class="meta">exit code <span>{exit_code}</span> &nbsp;·&nbsp; log <span>{log_path}</span></p>
  <div class="tabs">
    <button class="active" onclick="show('error')" id="tab-error">Error</button>
    <button onclick="show('full')" id="tab-full">Full Log</button>
  </div>
  <pre id="pane-error">{html.escape(error_section)}</pre>
  <pre id="pane-full" style="display:none">{html.escape(log)}</pre>
  <script>
  function show(id){{
    ['error','full'].forEach(function(p){{
      document.getElementById('pane-'+p).style.display=p===id?'':'none';
      document.getElementById('tab-'+p).className=p===id?'active':'';
    }});
  }}
  </script>
</body>
</html>"""
pathlib.Path(web_dir + "/index.html").write_text(page)
PYEOF

    # Start HTTP server in background
    ( cd "$WEB_DIR" && python3 -m http.server $PORT ) &>/dev/null &
    SERVER_PID=$!

    # Replace EXIT trap so we clean up server + web dir on exit
    trap "kill $SERVER_PID 2>/dev/null; rm -rf '$WEB_DIR'; rm -f '$PATCHED_CONF' '$PATCHED_CREDS'" EXIT

    dialog --backtitle "$BT" \
        --title " Installation Failed " \
        --msgbox "\n  archinstall exited with code $INSTALL_EXIT.\n\n  Open this in a browser on your network:\n\n    http://${LOCAL_IP}:${PORT}\n\n  Press OK — the server will keep running.\n  Press Ctrl+C in the terminal to stop it." 14 56

    echo ""
    echo -e "${BLD}${CYN}  http://${LOCAL_IP}:${PORT}${RST}"
    echo ""
    echo -e "  Press Ctrl+C to stop the server and exit."
    echo ""

    wait $SERVER_PID 2>/dev/null || true
    exit 1
fi

rm -f "$STDERR_LOG"
ok "archinstall completed successfully!"

# ── 7. Create btrfs subvolumes ────────────────────────────────
# archinstall 3.x has a bug with btrfs subvolume configs so we create
# them here manually after the install, then update fstab accordingly.
dialog --backtitle "$BT" \
    --title " Post-Install " \
    --infobox "\n  Creating btrfs subvolumes..." 5 38

ROOT_DEV=$(findmnt -n -o SOURCE /mnt 2>/dev/null || true)

if [[ -z "$ROOT_DEV" ]]; then
    warn "Could not detect mounted root device — skipping btrfs subvolumes."
    warn "Create them manually after reboot if needed."
else
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
fi

# ── 8. Stage sumi for first boot ──────────────────────────────
dialog --backtitle "$BT" \
    --title " Post-Install " \
    --infobox "\n  Staging sumi for first boot..." 5 40

TARGET="/mnt"
if [[ ! -d "$TARGET/home/$USERNAME" ]]; then
    dialog --backtitle "$BT" \
        --title " Warning " \
        --msgbox "\n  Cannot find installed system at /mnt/home/$USERNAME\n\n  archinstall may have used a different mount point or\n  the user wasn't created.\n\n  After reboot, run manually:\n    git clone <repo> ~/sumi\n    cd ~/sumi && ./install.sh" 14 58
    exit 0
fi

# Copy sumi repo to the new user's home
SUMI_DEST="$TARGET/home/$USERNAME/sumi"
rm -rf "$SUMI_DEST" 2>/dev/null || true
cp -r "$SCRIPT_DIR" "$SUMI_DEST"
arch-chroot "$TARGET" chown -R "$USERNAME:$USERNAME" "/home/$USERNAME/sumi" 2>/dev/null || true

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

# Add to .bash_profile so it runs on first TTY login
BASH_PROFILE="$TARGET/home/$USERNAME/.bash_profile"
if [[ ! -f "$BASH_PROFILE" ]] || ! grep -q "sumi-first-boot" "$BASH_PROFILE" 2>/dev/null; then
    cat >> "$BASH_PROFILE" << 'EOF'

# sumi first-boot hook (runs once, installs the rice, then removes itself)
[[ -f "$HOME/.sumi-first-boot.sh" ]] && bash "$HOME/.sumi-first-boot.sh"
EOF
    arch-chroot "$TARGET" chown "$USERNAME:$USERNAME" "/home/$USERNAME/.bash_profile" 2>/dev/null || true
fi

# ── 9. Done ────────────────────────────────────────────────────
dialog --backtitle "$BT" \
    --title " Bootstrap Complete " \
    --msgbox "\
  Arch Linux is installed successfully!

  On first boot:
    1. Unlock LUKS  (your encryption password)
    2. Login as $USERNAME at the TTY
    3. sumi installs automatically

  On second reboot:
    · Plymouth TUI unlock → auto login
    · Full Hyprland desktop with theming
    · SUPER+X for control center
    · SUPER+/ for keybind cheatsheet

  If first-boot auto-install fails:
    cd ~/sumi && ./install.sh" 20 54

dialog --backtitle "$BT" \
    --title " Reboot " \
    --yesno "\n  Reboot into your new system now?" 7 42 && {
    umount -R /mnt 2>/dev/null || true
    reboot
}
