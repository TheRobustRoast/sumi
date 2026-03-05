package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"sumi/internal/bootstrap"
	"sumi/internal/model"
	"sumi/internal/runner"
	"sumi/internal/theme"
)

// RunBootstrap runs the two-phase bootstrap TUI:
// Phase A: Interactive configuration (huh forms)
// Phase B: Execution (step runner)
func RunBootstrap(sumiSrc string) error {
	// Check if we're on the Arch ISO
	if _, err := os.Stat("/run/archiso"); err != nil {
		fmt.Println(theme.Warn("This doesn't look like the Arch Linux live ISO."))
		var cont bool
		confirm := huh.NewConfirm().
			Title("Continue anyway?").
			Value(&cont)
		if err := confirm.Run(); err != nil || !cont {
			return nil
		}
	}

	// Phase A: Collect configuration
	cfg, err := collectConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		return nil // user aborted
	}

	// Phase B: Execute installation
	return executeBootstrap(cfg, sumiSrc)
}

func collectConfig() (*bootstrap.Config, error) {
	cfg := &bootstrap.Config{
		Username: "user",
		Hostname: "sumi",
		Timezone: "UTC",
	}

	// ── Network ──
	fmt.Println()
	fmt.Println(theme.Section("Network"))

	bootstrap.BringUpEthernet(func(s string) { fmt.Println(theme.SubtextStyle.Render("  " + s)) })

	if !bootstrap.CheckNetwork(func(s string) {}) {
		fmt.Println(theme.Warn("No internet connection detected."))

		var netChoice string
		netForm := huh.NewSelect[string]().
			Title("How to connect?").
			Options(
				huh.NewOption("WiFi (iwctl)", "wifi"),
				huh.NewOption("Ethernet — retry DHCP", "dhcp"),
				huh.NewOption("Skip (I'll connect manually)", "skip"),
			).
			Value(&netChoice)
		if err := netForm.Run(); err != nil {
			return nil, nil
		}

		switch netChoice {
		case "wifi":
			fmt.Println(theme.SubtextStyle.Render("  Launching iwctl... (type 'exit' when connected)"))
			cmd := tea.ExecProcess(newIwctlCmd(), nil)
			_ = cmd
			// Fall through to iwctl
			_ = iwctl()
		case "dhcp":
			bootstrap.RetryDHCP(func(s string) { fmt.Println(theme.SubtextStyle.Render("  " + s)) })
		}

		if !bootstrap.CheckNetwork(func(s string) {}) {
			return nil, fmt.Errorf("no internet connection — connect manually and re-run")
		}
	}
	fmt.Println(theme.Ok("Internet connected"))

	// ── Disk Selection ──
	fmt.Println()
	fmt.Println(theme.Section("Disk"))

	disks, err := bootstrap.ListDisks()
	if err != nil || len(disks) == 0 {
		return nil, fmt.Errorf("no disks detected")
	}

	diskOptions := make([]huh.Option[string], len(disks))
	for i, d := range disks {
		diskOptions[i] = huh.NewOption(d.String(), d.Path)
	}

	diskForm := huh.NewSelect[string]().
		Title("Select installation disk").
		Options(diskOptions...).
		Value(&cfg.Disk)
	if err := diskForm.Run(); err != nil {
		return nil, nil
	}
	cfg.SetPartitions()

	// Find disk size for display
	var diskSize string
	for _, d := range disks {
		if d.Path == cfg.Disk {
			diskSize = d.Size
			break
		}
	}
	fmt.Println(theme.Ok(fmt.Sprintf("Selected: %s  (%s)", cfg.Disk, diskSize)))

	// Confirm wipe
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(theme.Red).Bold(true).Render(
		fmt.Sprintf("  ALL DATA ON %s (%s) WILL BE ERASED", cfg.Disk, diskSize),
	))
	var confirmWipe bool
	wipeConfirm := huh.NewConfirm().
		Title("Confirm wipe?").
		Affirmative("Yes, wipe it").
		Negative("Cancel").
		Value(&confirmWipe)
	if err := wipeConfirm.Run(); err != nil || !confirmWipe {
		fmt.Println(theme.Step("Aborted."))
		return nil, nil
	}

	// ── User Configuration ──
	fmt.Println()
	fmt.Println(theme.Section("User Configuration"))

	var password, password2 string
	nameValidate := func(s string) error {
		if len(s) == 0 {
			return fmt.Errorf("cannot be empty")
		}
		for _, c := range s {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
				return fmt.Errorf("only lowercase letters, digits, - _ . allowed")
			}
		}
		if s[0] == '-' || s[0] == '.' {
			return fmt.Errorf("cannot start with - or .")
		}
		return nil
	}
	userForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Description("Local user account").
				Value(&cfg.Username).
				Validate(nameValidate),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Password").
				Description("One password — login, root, LUKS").
				EchoMode(huh.EchoModePassword).
				Value(&password).
				Validate(func(s string) error {
					if len(s) < 4 {
						return fmt.Errorf("password too short (min 4 chars)")
					}
					return nil
				}),
			huh.NewInput().
				Title("Confirm password").
				EchoMode(huh.EchoModePassword).
				Value(&password2).
				Validate(func(s string) error {
					if s != password {
						return fmt.Errorf("passwords don't match")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Hostname").
				Description("Machine hostname").
				Value(&cfg.Hostname).
				Validate(nameValidate),
		),
	)
	if err := userForm.Run(); err != nil {
		return nil, nil
	}
	cfg.Password = password

	// Timezone
	detected := bootstrap.DetectTimezone(func(s string) {})
	if detected != "" {
		cfg.Timezone = detected
	}

	tzForm := huh.NewInput().
		Title("Timezone").
		Description("System timezone").
		Value(&cfg.Timezone)
	if err := tzForm.Run(); err != nil {
		return nil, nil
	}

	// ── Review ──
	fmt.Println()
	fmt.Println(theme.Section("Review"))

	review := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Surface2).
		Padding(1, 3).
		Render(strings.Join([]string{
			lipgloss.NewStyle().Foreground(theme.Text).Bold(true).Render("  Installation Summary"),
			"",
			fmt.Sprintf("  %-12s %s  (%s)  %s", dim("Disk"), cfg.Disk, diskSize, lipgloss.NewStyle().Foreground(theme.Red).Render("← WIPED")),
			fmt.Sprintf("  %-12s LUKS2  (argon2id)", dim("Encrypt")),
			fmt.Sprintf("  %-12s btrfs  (@  @home  @snapshots  @var_log)", dim("Filesystem")),
			fmt.Sprintf("  %-12s systemd-boot", dim("Bootloader")),
			fmt.Sprintf("  %-12s %s", dim("Hostname"), cfg.Hostname),
			fmt.Sprintf("  %-12s %s  (sudo, shared password)", dim("User"), cfg.Username),
			fmt.Sprintf("  %-12s %s", dim("Timezone"), cfg.Timezone),
			fmt.Sprintf("  %-12s PipeWire", dim("Audio")),
			fmt.Sprintf("  %-12s NetworkManager", dim("Network")),
			fmt.Sprintf("  %-12s Hyprland + sumi rice  (post-boot)", dim("Desktop")),
		}, "\n"))
	fmt.Println(review)

	var confirmInstall bool
	installConfirm := huh.NewConfirm().
		Title("Proceed with installation?").
		Affirmative("Install").
		Negative("Cancel").
		Value(&confirmInstall)
	if err := installConfirm.Run(); err != nil || !confirmInstall {
		fmt.Println(theme.Step("Aborted."))
		return nil, nil
	}

	return cfg, nil
}

func executeBootstrap(cfg *bootstrap.Config, sumiSrc string) error {
	logPath := "/tmp/sumi-install.log"

	localIP := bootstrap.LocalIP()

	steps := bootstrap.BuildSteps(cfg, sumiSrc)

	br := NewBootstrapRunner(BootstrapRunnerConfig{
		Steps:   steps,
		LogPath: logPath,
		LocalIP: localIP,
	})

	p := tea.NewProgram(br, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		// Ctrl+C during bootstrap: clean up mounts/LUKS
		if errors.Is(err, tea.ErrProgramKilled) {
			bootstrap.Preflight(func(s string) {})
			return nil
		}
		return err
	}

	result := m.(BootstrapRunner)
	if result.failed {
		// Clean up mounts/LUKS so system isn't left in broken state
		bootstrap.Preflight(func(s string) {}) //nolint:errcheck

		// Serve error page
		fmt.Println()
		fmt.Println(theme.Fail("Installation failed"))
		fmt.Printf("  Open in a browser: http://%s:7777\n", localIP)
		fmt.Println(theme.SubtextStyle.Render("  Press Ctrl+C to stop the server and exit."))
		bootstrap.ServeErrorPage(logPath) // blocks
		return fmt.Errorf("installation failed")
	}

	// Success — offer reboot
	fmt.Println()
	fmt.Println(theme.Ok("Arch Linux installed"))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(theme.Text).Render("  First boot:"))
	fmt.Println("  1. Unlock LUKS  (your encryption password)")
	fmt.Printf("  2. Login as %s at the TTY\n", lipgloss.NewStyle().Foreground(theme.Red).Render(cfg.Username))
	fmt.Println("  3. sumi installs automatically")
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(theme.Text).Render("  Second reboot:"))
	fmt.Println("  · Full Hyprland desktop")
	fmt.Println("  · SUPER+X  control center")
	fmt.Println("  · SUPER+/  keybind cheatsheet")
	fmt.Println()

	var reboot bool
	rebootConfirm := huh.NewConfirm().
		Title("Reboot now?").
		Value(&reboot)
	if err := rebootConfirm.Run(); err == nil && reboot {
		_ = runReboot()
	}

	return nil
}

func dim(s string) string {
	return lipgloss.NewStyle().Foreground(theme.Dim).Render(s)
}

func iwctl() error {
	cmd := newIwctlCmd()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func newIwctlCmd() *exec.Cmd {
	return exec.Command("iwctl")
}

func runReboot() error {
	_ = exec.Command("umount", "-R", "/mnt").Run()
	_ = exec.Command("cryptsetup", "close", "root").Run()
	return exec.Command("reboot").Run()
}

// BootstrapRunnerConfig configures the bootstrap step runner.
type BootstrapRunnerConfig struct {
	Steps   []model.Step
	LogPath string
	LocalIP string
}

// BootstrapRunner is the Bubble Tea model for the bootstrap execution phase.
// Unlike StepRunner, it stops on first failure and offers the error page.
type BootstrapRunner struct {
	StepRunner
	logPath string
	localIP string
	failed  bool
}

// NewBootstrapRunner creates a bootstrap-specific step runner that stops on failure.
func NewBootstrapRunner(cfg BootstrapRunnerConfig) BootstrapRunner {
	return BootstrapRunner{
		StepRunner: NewStepRunner(StepRunnerConfig{
			Title:    "bootstrap",
			Subtitle: fmt.Sprintf("LUKS2 · btrfs · Hyprland  ·  On failure → http://%s:7777", cfg.LocalIP),
			Steps:    cfg.Steps,
			DoneMsg:  "Arch Linux installed",
		}),
		logPath: cfg.LogPath,
		localIP: cfg.LocalIP,
	}
}

func (m BootstrapRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case runner.DoneMsg:
		if msg.Err != nil {
			// Stop on failure — don't continue to next step
			m.failed = true
			m.StepRunner.cfg.Steps[m.StepRunner.cur].Status = model.StepFailed
			m.StepRunner.done = true
			return m, nil
		}
	}

	// Delegate to StepRunner for everything else
	updated, cmd := m.StepRunner.Update(msg)
	m.StepRunner = updated.(StepRunner)
	return m, cmd
}

func (m BootstrapRunner) View() string {
	return m.StepRunner.View()
}
