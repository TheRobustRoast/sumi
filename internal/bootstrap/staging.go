package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumi/internal/runner"
)

// StageSumi copies the sumi repo into the target system and installs
// a first-boot hook so the rice installer runs on first login.
func StageSumi(send func(string), cfg *Config, sumiSrc string) error {
	userHome := fmt.Sprintf("/mnt/home/%s", cfg.Username)

	if _, err := os.Stat(userHome); err != nil {
		return fmt.Errorf("user home %s not found — stage manually after reboot", userHome)
	}

	// Copy sumi repo
	sumiDest := filepath.Join(userHome, "sumi")
	send("Copying sumi repo...")
	_ = os.RemoveAll(sumiDest)
	if err := runner.RunCmd(send, "cp", "-r", sumiSrc, sumiDest); err != nil {
		return fmt.Errorf("copy sumi: %w", err)
	}
	_ = runner.RunCmd(send, "arch-chroot", "/mnt",
		"chown", "-R", fmt.Sprintf("%s:%s", cfg.Username, cfg.Username),
		fmt.Sprintf("/home/%s/sumi", cfg.Username),
	)
	send("Repo copied")

	// Create first-boot script
	firstBoot := filepath.Join(userHome, ".sumi-first-boot.sh")
	if err := os.WriteFile(firstBoot, []byte(firstBootScript), 0o755); err != nil {
		return fmt.Errorf("write first-boot script: %w", err)
	}
	_ = runner.RunCmd(send, "arch-chroot", "/mnt",
		"chown", fmt.Sprintf("%s:%s", cfg.Username, cfg.Username),
		fmt.Sprintf("/home/%s/.sumi-first-boot.sh", cfg.Username),
	)

	// Add hook to .bash_profile
	bashProfile := filepath.Join(userHome, ".bash_profile")
	hook := "\n# sumi first-boot hook (runs once, installs the rice, then removes itself)\n" +
		"[[ -f \"$HOME/.sumi-first-boot.sh\" ]] && bash \"$HOME/.sumi-first-boot.sh\"\n"

	existing, _ := os.ReadFile(bashProfile)
	if len(existing) == 0 || !strings.Contains(string(existing), "sumi-first-boot") {
		f, err := os.OpenFile(bashProfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open .bash_profile: %w", err)
		}
		defer f.Close()
		if _, err := f.WriteString(hook); err != nil {
			return fmt.Errorf("write .bash_profile: %w", err)
		}

		_ = runner.RunCmd(send, "arch-chroot", "/mnt",
			"chown", fmt.Sprintf("%s:%s", cfg.Username, cfg.Username),
			fmt.Sprintf("/home/%s/.bash_profile", cfg.Username),
		)
	}

	send("First-boot hook installed")
	return nil
}

const firstBootScript = `#!/usr/bin/env bash
# sumi first-boot installer — runs once then self-deletes
MARKER="$HOME/.sumi-first-boot-done"
[[ -f "$MARKER" ]] && exit 0

if [[ -d "$HOME/sumi" ]]; then
    echo ""
    echo "sumi :: first boot setup"
    echo ""
    cd "$HOME/sumi"
    if command -v sumi &>/dev/null; then
        sumi install
    else
        go build -o sumi ./cmd/sumi && ./sumi install
    fi
    if [[ $? -eq 0 ]]; then
        touch "$MARKER"
        rm -f "$HOME/.sumi-first-boot.sh"
        echo ""
        echo "sumi installed! Reboot for the full experience."
    else
        echo ""
        echo "Install failed. Retry on next login, or:"
        echo "  cd ~/sumi && sumi install"
    fi
fi
`
