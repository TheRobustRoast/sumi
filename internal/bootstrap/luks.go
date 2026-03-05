package bootstrap

import (
	"fmt"

	"sumi/internal/runner"
)

// FormatLUKS creates a LUKS2 volume on the root partition using the given password.
func FormatLUKS(send func(string), cfg *Config) error {
	send("Formatting LUKS2 (argon2id, this is slow)...")
	if err := runner.RunCmdWithStdin(send, cfg.Password,
		"cryptsetup", "luksFormat",
		"--type", "luks2",
		"--pbkdf", "argon2id",
		"--hash", "sha512",
		"--key-size", "512",
		"--iter-time", "10000",
		"--batch-mode",
		cfg.RootPart, "-",
	); err != nil {
		return fmt.Errorf("cryptsetup luksFormat: %w", err)
	}

	send("Opening LUKS volume...")
	if err := runner.RunCmdWithStdin(send, cfg.Password,
		"cryptsetup", "open", cfg.RootPart, "root", "-",
	); err != nil {
		return fmt.Errorf("cryptsetup open: %w", err)
	}

	return nil
}
