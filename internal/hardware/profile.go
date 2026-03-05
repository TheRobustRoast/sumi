package hardware

// Profile describes hardware-specific configuration for the installer.
type Profile struct {
	ID           string
	Name         string
	Packages     []string // extra pacman packages
	AURPackages  []string // extra AUR packages
	KernelParams []string // extra kernel cmdline params
	Modules      []string // kernel modules to load
	Services     []string // extra systemd services to enable
	// Tweaks runs hardware-specific configuration (sysfs writes, config file edits).
	// nil means no tweaks needed.
	Tweaks func(send func(string)) error
}
