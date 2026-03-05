package main

import (
	"os"

	"github.com/spf13/cobra"

	"sumi/internal/theme"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "sumi",
		Short:   "sumi — Hyprland rice manager",
		Long: `sumi manages your Hyprland rice on Arch Linux.

Lifecycle: bootstrap, install, update, configure, and sync your dotfiles.
Desktop: wallpaper pipeline, screenshots, modes, power, sessions, keybinds.
Tools: clipboard, emoji picker, calculator, notes, project launcher.`,
		Version: version,
	}

	// Silence cobra's default error printing — we handle it below
	root.SilenceErrors = true

	// Command groups for organized help output
	root.AddGroup(
		&cobra.Group{ID: "lifecycle", Title: "Lifecycle:"},
		&cobra.Group{ID: "desktop", Title: "Desktop:"},
		&cobra.Group{ID: "tools", Title: "Tools:"},
		&cobra.Group{ID: "dev", Title: "Dev:"},
	)

	// Lifecycle commands
	for _, cmd := range []*cobra.Command{
		bootstrapCmd(), installCmd(), updateCmd(), uninstallCmd(),
		configCmd(), doctorCmd(), statusCmd(), lintCmd(),
		cleanupCmd(), welcomeCmd(), syncCmd(),
	} {
		cmd.GroupID = "lifecycle"
		root.AddCommand(cmd)
	}

	// Desktop commands
	for _, cmd := range []*cobra.Command{
		wallpaperCmd(), themeCmd(), modeCmd(), powerCmd(),
		captureCmd(), sessionCmd(), monitorCmd(),
		osdCmd(), controlCmd(), keysCmd(), notifyCmd(),
	} {
		cmd.GroupID = "desktop"
		root.AddCommand(cmd)
	}

	// Tools
	for _, cmd := range []*cobra.Command{
		clipboardCmd(), emojiCmd(), calcCmd(),
	} {
		cmd.GroupID = "tools"
		root.AddCommand(cmd)
	}

	// Dev commands
	for _, cmd := range []*cobra.Command{
		projectCmd(), worktreeCmd(), noteCmd(),
	} {
		cmd.GroupID = "dev"
		root.AddCommand(cmd)
	}

	// Internal (hidden)
	root.AddCommand(wrapCmd(), updateCheckCmd())

	if err := root.Execute(); err != nil {
		if !isUserQuit(err) {
			os.Stderr.WriteString(theme.Fail(err.Error()) + "\n")
			os.Exit(1)
		}
	}
}
