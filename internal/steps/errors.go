package steps

import "strings"

// errorPattern maps a substring in an error message to a suggested fix.
type errorPattern struct {
	Pattern    string
	Suggestion string
}

var knownErrors = []errorPattern{
	{
		Pattern:    "unable to lock database",
		Suggestion: "Another pacman instance is running. Try: sudo rm /var/lib/pacman/db.lck",
	},
	{
		Pattern:    "target not found",
		Suggestion: "Package not in repos. Run: sudo pacman -Sy to refresh, or check the package name.",
	},
	{
		Pattern:    "failed to commit transaction",
		Suggestion: "Pacman transaction failed. Try: sudo pacman -Syu to do a full system upgrade first.",
	},
	{
		Pattern:    "could not resolve host",
		Suggestion: "DNS resolution failed. Check your network connection and /etc/resolv.conf.",
	},
	{
		Pattern:    "no internet",
		Suggestion: "Cannot reach archlinux.org. Check network: ip link, ping 1.1.1.1",
	},
	{
		Pattern:    "not running on Arch Linux",
		Suggestion: "sumi requires Arch Linux. /etc/arch-release not found.",
	},
	{
		Pattern:    "permission denied",
		Suggestion: "Insufficient permissions. Make sure your user is in the wheel group and sudo works.",
	},
	{
		Pattern:    "command not found",
		Suggestion: "A required binary is missing. Run: sumi doctor to check your setup.",
	},
	{
		Pattern:    "GPGME error",
		Suggestion: "Keyring issue. Try: sudo pacman-key --init && sudo pacman-key --populate archlinux",
	},
	{
		Pattern:    "failed to init transaction",
		Suggestion: "Pacman database issue. Try: sudo pacman -Sy or check /var/lib/pacman/",
	},
	{
		Pattern:    "error: failed to synchronize",
		Suggestion: "Mirror sync failed. Try updating mirrorlist: sudo reflector --latest 10 --sort rate --save /etc/pacman.d/mirrorlist",
	},
	{
		Pattern:    "makepkg",
		Suggestion: "AUR build failed. Check build dependencies: pacman -S --needed base-devel",
	},
	{
		Pattern:    "disk space",
		Suggestion: "Not enough disk space. Free up space or run: sumi cleanup",
	},
	{
		Pattern:    "symlink",
		Suggestion: "Symlink operation failed. Check if the target already exists as a regular file.",
	},
	{
		Pattern:    "chsh",
		Suggestion: "Shell change failed. Ensure zsh is in /etc/shells: grep zsh /etc/shells",
	},
}

// DiagnoseError maps a step name and error to a human-readable suggestion.
// Returns empty string if no match is found.
func DiagnoseError(stepName string, err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	for _, p := range knownErrors {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(p.Pattern)) {
			return p.Suggestion
		}
	}
	return ""
}

// DoctorFix returns a fix suggestion for a doctor check failure.
func DoctorFix(checkName string) string {
	fixes := map[string]string{
		"config symlinks":          "Run: sumi install (re-links dotfiles)",
		"core packages installed":  "Run: sumi install (re-installs missing packages)",
		"user services":            "Run: systemctl --user daemon-reload && sumi install",
		"system services":          "Run: sudo systemctl enable <service>",
		"wallpaper pipeline":       "Install wallust (yay -S wallust) and set a wallpaper with: sumi theme",
		"free space":               "Free disk space or run: sumi cleanup",
	}
	if fix, ok := fixes[checkName]; ok {
		return fix
	}
	return ""
}
