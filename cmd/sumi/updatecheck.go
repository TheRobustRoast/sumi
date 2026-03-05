package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func updateCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "update-check",
		Short:  "Check for sumi updates and notify",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			sumiDir, err := sumiRoot()
			if err != nil {
				return err
			}

			// Fetch quietly — bail on failure (no network is fine)
			fetch := exec.Command("git", "-C", sumiDir, "fetch", "--quiet")
			if err := fetch.Run(); err != nil {
				return nil
			}

			localOut, err := exec.Command("git", "-C", sumiDir, "rev-parse", "HEAD").Output()
			if err != nil {
				return nil
			}
			remoteOut, err := exec.Command("git", "-C", sumiDir, "rev-parse", "origin/main").Output()
			if err != nil {
				return nil
			}

			local := strings.TrimSpace(string(localOut))
			remote := strings.TrimSpace(string(remoteOut))

			if local == remote {
				return nil
			}

			// Count commits behind
			countOut, _ := exec.Command("git", "-C", sumiDir, "rev-list", "--count", "HEAD..origin/main").Output()
			behind := strings.TrimSpace(string(countOut))
			if behind == "" {
				behind = "new"
			}

			// Send desktop notification
			exec.Command("notify-send", //nolint:errcheck
				"-a", "sumi",
				"-i", "software-update-available",
				"-u", "low",
				"sumi update available",
				fmt.Sprintf("%s commits behind — run 'sumi update'", behind),
			).Run() //nolint:errcheck

			return nil
		},
	}
}
