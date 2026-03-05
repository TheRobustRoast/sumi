package bootstrap

import (
	"fmt"
	"os"
	"strings"

	"sumi/internal/runner"
)

// Partition wipes the disk and creates EFI + LUKS partitions using sgdisk.
func Partition(send func(string), cfg *Config) error {
	// Safety: verify disk is a block device and path is sane
	if !strings.HasPrefix(cfg.Disk, "/dev/") {
		return fmt.Errorf("disk path must start with /dev/: %q", cfg.Disk)
	}
	fi, err := os.Stat(cfg.Disk)
	if err != nil {
		return fmt.Errorf("disk not found: %w", err)
	}
	if fi.Mode()&os.ModeDevice == 0 {
		return fmt.Errorf("%s is not a device", cfg.Disk)
	}

	send(fmt.Sprintf("Wiping %s...", cfg.Disk))
	if err := runner.RunCmd(send, "sgdisk", "--zap-all", cfg.Disk); err != nil {
		return fmt.Errorf("sgdisk zap: %w", err)
	}

	send("Creating EFI (1G) + LUKS partitions...")
	if err := runner.RunCmd(send,
		"sgdisk",
		"--new=1:0:+1G", "--typecode=1:ef00", "--change-name=1:EFI",
		"--new=2:0:0", "--typecode=2:8309", "--change-name=2:LUKS",
		cfg.Disk,
	); err != nil {
		return fmt.Errorf("sgdisk partition: %w", err)
	}

	send("Probing partitions...")
	if err := runner.RunCmd(send, "partprobe", cfg.Disk); err != nil {
		return fmt.Errorf("partprobe: %w", err)
	}
	if err := runner.RunCmd(send, "udevadm", "settle", "--timeout", "10"); err != nil {
		return fmt.Errorf("udevadm settle: %w", err)
	}

	return nil
}

// FormatEFI formats the EFI partition as FAT32.
func FormatEFI(send func(string), cfg *Config) error {
	send(fmt.Sprintf("Formatting %s as FAT32...", cfg.EFIPart))
	return runner.RunCmd(send, "mkfs.fat", "-F32", cfg.EFIPart)
}
