// Package bootstrap implements the Arch Linux installation from live ISO.
package bootstrap


// Config holds all user-supplied values for the bootstrap process.
type Config struct {
	Disk     string // e.g. /dev/nvme0n1
	EFIPart  string // e.g. /dev/nvme0n1p1
	RootPart string // e.g. /dev/nvme0n1p2
	Username string
	Password string
	Hostname string
	Timezone string
	HWProfile string // hardware profile ID (empty = auto-detect)
}

// SetPartitions determines the correct partition paths based on disk type.
// NVMe/eMMC (name ends in digit) use "p1"/"p2", SATA/others use "1"/"2".
func (c *Config) SetPartitions() {
	// If device name ends in a digit (nvme0n1, mmcblk0), partitions use "p" prefix
	if len(c.Disk) > 0 && c.Disk[len(c.Disk)-1] >= '0' && c.Disk[len(c.Disk)-1] <= '9' {
		c.EFIPart = c.Disk + "p1"
		c.RootPart = c.Disk + "p2"
	} else {
		c.EFIPart = c.Disk + "1"
		c.RootPart = c.Disk + "2"
	}
}
