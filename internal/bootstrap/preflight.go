package bootstrap

import (
	"os"
	"strings"

	"sumi/internal/runner"
)

// Preflight cleans stale state from previous attempts: unmounts /mnt,
// drops caches, closes device-mapper devices.
func Preflight(send func(string)) error {
	send("Unmounting stale /mnt mounts...")

	// Get all mounts under /mnt, sorted in reverse for clean unmount
	out, err := captureCmd("findmnt", "-r", "-n", "-o", "TARGET")
	if err == nil {
		mounts := strings.Split(strings.TrimSpace(out), "\n")
		// Reverse order
		for i := len(mounts) - 1; i >= 0; i-- {
			m := strings.TrimSpace(mounts[i])
			if strings.HasPrefix(m, "/mnt") {
				_ = runner.RunCmd(send, "umount", m)
			}
		}
	}
	_ = runner.RunCmd(send, "umount", "-R", "/mnt")

	// Drop caches so the kernel releases btrfs refs
	send("Dropping kernel caches...")
	_ = runner.RunCmd(send, "sync")
	_ = writeFile("/proc/sys/vm/drop_caches", "3")

	// Close all non-ISO device mapper devices
	send("Closing stale device-mapper devices...")
	out, err = captureCmd("dmsetup", "ls")
	if err == nil {
		for _, line := range strings.Split(out, "\n") {
			name := strings.Fields(line)
			if len(name) == 0 {
				continue
			}
			n := name[0]
			if n == "ventoy" || strings.HasPrefix(n, "sda") || n == "No" {
				continue
			}
			_ = runner.RunCmd(send, "cryptsetup", "close", n)
			_ = runner.RunCmd(send, "dmsetup", "remove", "--force", n)
		}
	}

	// Also try to close "root" explicitly
	_ = runner.RunCmd(send, "cryptsetup", "close", "root")

	send("Pre-flight done")
	return nil
}

// captureCmd runs a command and returns its combined output as a string.
func captureCmd(args ...string) (string, error) {
	var lines []string
	err := runner.RunCmd(func(line string) { lines = append(lines, line) }, args...)
	return strings.Join(lines, "\n"), err
}

// writeFile writes content to a file, ignoring errors (best-effort).
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
