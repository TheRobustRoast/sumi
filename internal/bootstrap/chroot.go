package bootstrap

import (
	"fmt"
	"regexp"
	"strings"

	"sumi/internal/runner"
)

// safeShellLiteral rejects values that contain shell metacharacters.
// Used to validate hostname/username before interpolation into bash scripts.
var safeShellRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var safeTimezoneRe = regexp.MustCompile(`^[a-zA-Z0-9._/+-]+$`)

// ChrootConfigure runs all in-chroot configuration: timezone, locale, hostname,
// users, mkinitcpio, systemd-boot, services, zram.
func ChrootConfigure(send func(string), cfg *Config) error {
	// Validate inputs that get interpolated into the bash script
	for _, v := range []struct{ name, val string }{
		{"hostname", cfg.Hostname},
		{"username", cfg.Username},
	} {
		if !safeShellRe.MatchString(v.val) {
			return fmt.Errorf("%s contains unsafe characters: %q", v.name, v.val)
		}
	}
	if !safeTimezoneRe.MatchString(cfg.Timezone) || strings.Contains(cfg.Timezone, "..") {
		return fmt.Errorf("timezone contains unsafe characters: %q", cfg.Timezone)
	}

	luksUUID, err := getLUKSUUID(cfg.RootPart)
	if err != nil {
		return fmt.Errorf("get LUKS UUID: %w", err)
	}

	ucode := CPUMicrocode()
	initrd := fmt.Sprintf("/%s.img", ucode)

	// Build the chroot script. Outer vars are interpolated; password goes via env.
	script := fmt.Sprintf(`set -euo pipefail

echo "==> timezone"
ln -sf "/usr/share/zoneinfo/%[1]s" /etc/localtime
hwclock --systohc

echo "==> locale"
echo "en_US.UTF-8 UTF-8" > /etc/locale.gen
locale-gen
echo "LANG=en_US.UTF-8" > /etc/locale.conf
echo "KEYMAP=us" > /etc/vconsole.conf

echo "==> hostname"
echo "%[2]s" > /etc/hostname
printf '127.0.0.1\tlocalhost\n::1\t\tlocalhost\n127.0.1.1\t%[2]s.localdomain %[2]s\n' > /etc/hosts

echo "==> pacman (multilib + parallel downloads)"
sed -i '/^\[multilib\]/,/^Include/ s/^#//' /etc/pacman.conf
sed -i 's/^#ParallelDownloads.*/ParallelDownloads = 5/' /etc/pacman.conf

echo "==> mkinitcpio (encrypt hook for LUKS)"
sed -i 's/^HOOKS=.*/HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block encrypt filesystems fsck)/' /etc/mkinitcpio.conf
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
initrd  %[5]s
initrd  /initramfs-linux.img
options cryptdevice=UUID=%[4]s:root root=/dev/mapper/root rootflags=subvol=@ rw quiet
ENTRY

echo "==> users"
echo "root:${SUMI_PASS}" | chpasswd
useradd -m -G wheel,video,audio,input,storage -s /bin/zsh "%[3]s"
echo "%[3]s:${SUMI_PASS}" | chpasswd
echo "%%wheel ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/wheel
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
`, cfg.Timezone, cfg.Hostname, cfg.Username, luksUUID, initrd)

	send("Running chroot configuration...")
	return runChrootScript(send, cfg.Password, script)
}

// getLUKSUUID returns the UUID of the LUKS partition.
func getLUKSUUID(rootPart string) (string, error) {
	var uuid string
	err := runner.RunCmd(func(line string) {
		uuid = strings.TrimSpace(line)
	}, "blkid", "-s", "UUID", "-o", "value", rootPart)
	if err != nil {
		return "", err
	}
	if uuid == "" {
		return "", fmt.Errorf("no UUID found for %s", rootPart)
	}
	// Validate UUID format (8-4-4-4-12 hex)
	if len(uuid) != 36 || uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		return "", fmt.Errorf("invalid UUID format %q for %s", uuid, rootPart)
	}
	return uuid, nil
}

// runChrootScript pipes a script to arch-chroot /mnt bash -s, passing the
// password via SUMI_PASS environment variable.
func runChrootScript(send func(string), password, script string) error {
	return runner.RunCmdWithStdin(send, script,
		"env", fmt.Sprintf("SUMI_PASS=%s", password),
		"arch-chroot", "/mnt", "bash", "-s",
	)
}
