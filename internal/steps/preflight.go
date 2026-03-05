package steps

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"sumi/internal/model"
)

// Preflight returns the pre-flight check steps (internet, arch-release, commands).
func Preflight() []model.Step {
	return []model.Step{
		{
			ID:      "preflight/internet",
			Section: "Pre-flight",
			Name:    "internet",
			RunGo: func(_ model.InstallCtx) ([]string, error) {
				conn, err := net.DialTimeout("tcp", "archlinux.org:443", 5*time.Second)
				if err != nil {
					return nil, fmt.Errorf("no internet: %w", err)
				}
				conn.Close()
				return []string{"archlinux.org reachable"}, nil
			},
		},
		{
			ID:   "preflight/arch",
			Name: "Arch Linux",
			RunGo: func(_ model.InstallCtx) ([]string, error) {
				if _, err := os.Stat("/etc/arch-release"); err != nil {
					return nil, fmt.Errorf("not running on Arch Linux")
				}
				return []string{"/etc/arch-release found"}, nil
			},
		},
		{
			ID:   "preflight/commands",
			Name: "required commands",
			RunGo: func(_ model.InstallCtx) ([]string, error) {
				var found []string
				for _, c := range []string{"git", "sudo", "pacman"} {
					if _, err := exec.LookPath(c); err != nil {
						return nil, fmt.Errorf("%s not found in PATH", c)
					}
					found = append(found, c)
				}
				return found, nil
			},
		},
	}
}
