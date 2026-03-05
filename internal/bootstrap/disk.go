package bootstrap

import (
	"fmt"
	"strings"

	"sumi/internal/runner"
)

// DiskInfo represents a block device available for installation.
type DiskInfo struct {
	Path string // e.g. /dev/nvme0n1
	Size string // e.g. 476.9G
}

// String returns a formatted display string.
func (d DiskInfo) String() string {
	return fmt.Sprintf("%-20s %s", d.Path, d.Size)
}

// ListDisks returns all available block devices of type "disk".
func ListDisks() ([]DiskInfo, error) {
	var lines []string
	err := runner.RunCmd(func(line string) {
		lines = append(lines, line)
	}, "lsblk", "-dno", "NAME,SIZE,TYPE")
	if err != nil {
		return nil, err
	}

	var disks []DiskInfo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[2] != "disk" {
			continue
		}
		disks = append(disks, DiskInfo{
			Path: "/dev/" + fields[0],
			Size: fields[1],
		})
	}
	return disks, nil
}
