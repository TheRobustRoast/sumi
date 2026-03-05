package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/runner"
	"sumi/internal/model"
)

// Greetd returns the step for configuring the greetd login manager.
func Greetd() model.Step {
	return model.Step{
		ID:      "display/greetd",
		Section: "Display Manager",
		Name:    "configure greetd",
		RunStream: func(ctx model.InstallCtx) (*runner.Stream, tea.Cmd) {
			return runner.Func(func(send func(string)) error {
				return configureGreetd(send, ctx)
			})
		},
	}
}

func configureGreetd(send func(string), ctx model.InstallCtx) error {
	if err := runner.RunCmd(send, "sudo", "mkdir", "-p", "/etc/greetd"); err != nil {
		return err
	}
	runner.RunCmd(send, "sudo", "cp", "/etc/greetd/config.toml", "/etc/greetd/config.toml.bak") //nolint:errcheck

	sumiWrap := ctx.Home + "/.local/bin/sumi wrap hyprland"
	cfg := fmt.Sprintf(
		"[terminal]\nvt = 1\n\n"+
			"# Cold-boot autologin: LUKS unlock is authentication — no second prompt.\n"+
			"[initial_session]\ncommand = %q\nuser = %q\n\n"+
			"# After logout: interactive tuigreet.\n"+
			"[default_session]\ncommand = \"tuigreet --time --remember --remember-session --asterisks --cmd %s\"\nuser = \"greeter\"\n",
		sumiWrap, ctx.User, sumiWrap,
	)
	if err := runner.WriteAsSudo(send, "/etc/greetd/config.toml", cfg); err != nil {
		return err
	}
	send("greetd.toml written")
	return runner.RunCmd(send, "sudo", "systemctl", "enable", "greetd.service")
}

// Plymouth returns the step for configuring the Plymouth boot theme.
func Plymouth() model.Step {
	return model.Step{
		ID:      "display/plymouth",
		Section: "Plymouth",
		Name:    "configure plymouth",
		RunStream: func(ctx model.InstallCtx) (*runner.Stream, tea.Cmd) {
			return runner.Func(func(send func(string)) error {
				return configurePlymouth(send, ctx)
			})
		},
	}
}

func configurePlymouth(send func(string), ctx model.InstallCtx) error {
	themeSrc := filepath.Join(ctx.SumiDir, "plymouth/themes/hypr-tui")
	themeDst := "/usr/share/plymouth/themes/hypr-tui"

	if err := runner.RunCmd(send, "sudo", "mkdir", "-p", themeDst); err != nil {
		return err
	}

	entries, err := os.ReadDir(themeSrc)
	if err != nil {
		send("Warning: plymouth theme source not found, skipping copy")
	} else {
		for _, e := range entries {
			runner.RunCmd(send, "sudo", "cp", filepath.Join(themeSrc, e.Name()), filepath.Join(themeDst, e.Name())) //nolint:errcheck
		}
		runner.RunCmd(send, "sudo", "plymouth-set-default-theme", "hypr-tui") //nolint:errcheck
		send("Plymouth theme set")
	}

	if err := addPlymouthHook(send); err != nil {
		send("Warning: could not modify mkinitcpio.conf: " + err.Error())
	}

	if modified, entry := AddKernelParam(send, "splash"); modified {
		send("splash added to " + entry)
	}

	return runner.RunCmd(send, "sudo", "mkinitcpio", "-P")
}

func addPlymouthHook(send func(string)) error {
	content, err := runner.ReadFile("/etc/mkinitcpio.conf")
	if err != nil {
		return err
	}
	if strings.Contains(content, "plymouth") {
		send("plymouth already in mkinitcpio HOOKS")
		return nil
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "HOOKS=") {
			lines[i] = strings.Replace(line, "udev", "udev plymouth", 1)
		}
	}
	if err := runner.WriteAsSudo(send, "/etc/mkinitcpio.conf", strings.Join(lines, "\n")); err != nil {
		return err
	}
	send("plymouth hook added to mkinitcpio")
	return nil
}

// AddKernelParam adds a parameter to the first boot loader entry if not present.
// Returns (true, entryPath) if the file was modified.
func AddKernelParam(send func(string), param string) (bool, string) {
	entries, err := filepath.Glob("/boot/loader/entries/*.conf")
	if err != nil || len(entries) == 0 {
		return false, ""
	}
	entry := entries[0]
	content, err := runner.ReadFile(entry)
	if err != nil {
		send("Warning: cannot read " + entry)
		return false, ""
	}
	if strings.Contains(content, param) {
		return false, ""
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "options ") {
			lines[i] = line + " " + param
		}
	}
	if err := runner.WriteAsSudo(send, entry, strings.Join(lines, "\n")); err != nil {
		send("Warning: cannot write " + entry + ": " + err.Error())
		return false, ""
	}
	return true, entry
}
