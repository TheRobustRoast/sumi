package bootstrap

import (
	"fmt"
	"os"

	"sumi/internal/runner"
)

// CreateBtrfs formats /dev/mapper/root as btrfs.
func CreateBtrfs(send func(string)) error {
	send("Creating btrfs filesystem...")
	return runner.RunCmd(send, "mkfs.btrfs", "-f", "/dev/mapper/root")
}

// CreateSubvolumes creates the btrfs subvolume layout: @, @home, @snapshots, @var_log.
func CreateSubvolumes(send func(string)) error {
	send("Mounting root for subvolume creation...")
	if err := runner.RunCmd(send, "mount", "/dev/mapper/root", "/mnt"); err != nil {
		return fmt.Errorf("mount root: %w", err)
	}

	for _, sv := range []string{"@", "@home", "@snapshots", "@var_log"} {
		send(fmt.Sprintf("Creating subvolume %s...", sv))
		if err := runner.RunCmd(send, "btrfs", "subvolume", "create", "/mnt/"+sv); err != nil {
			return fmt.Errorf("btrfs subvolume create %s: %w", sv, err)
		}
	}

	send("Unmounting...")
	return runner.RunCmd(send, "umount", "/mnt")
}

// MountFilesystems mounts all btrfs subvolumes and the EFI partition.
// On failure, it unmounts everything that was already mounted.
func MountFilesystems(send func(string), cfg *Config) error {
	// Check prerequisite: /dev/mapper/root must exist (created by FormatLUKS)
	if _, err := os.Stat("/dev/mapper/root"); err != nil {
		return fmt.Errorf("/dev/mapper/root not found — LUKS must be opened first: %w", err)
	}

	cleanup := func() {
		_ = runner.RunCmd(send, "umount", "-R", "/mnt")
	}

	send("Mounting @ subvolume...")
	if err := runner.RunCmd(send,
		"mount", "-o", "subvol=@,compress=zstd,noatime",
		"/dev/mapper/root", "/mnt",
	); err != nil {
		return fmt.Errorf("mount @: %w", err)
	}

	// Create mount points
	for _, dir := range []string{"/mnt/boot", "/mnt/home", "/mnt/.snapshots", "/mnt/var/log"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			cleanup()
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	send("Mounting EFI partition...")
	if err := runner.RunCmd(send, "mount", cfg.EFIPart, "/mnt/boot"); err != nil {
		cleanup()
		return fmt.Errorf("mount EFI: %w", err)
	}

	mounts := []struct{ subvol, target string }{
		{"@home", "/mnt/home"},
		{"@snapshots", "/mnt/.snapshots"},
		{"@var_log", "/mnt/var/log"},
	}
	for _, m := range mounts {
		send(fmt.Sprintf("Mounting %s...", m.subvol))
		if err := runner.RunCmd(send,
			"mount", "-o", fmt.Sprintf("subvol=%s,compress=zstd,noatime", m.subvol),
			"/dev/mapper/root", m.target,
		); err != nil {
			cleanup()
			return fmt.Errorf("mount %s: %w", m.subvol, err)
		}
	}

	// Pre-seed /etc so pacstrap's mkinitcpio hook doesn't warn
	if err := os.MkdirAll("/mnt/etc", 0o755); err != nil {
		cleanup()
		return err
	}
	return os.WriteFile("/mnt/etc/vconsole.conf", []byte("KEYMAP=us\n"), 0o644)
}
