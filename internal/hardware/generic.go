package hardware

// GenericLaptop returns a hardware profile for a generic laptop.
func GenericLaptop() *Profile {
	return &Profile{
		ID:       "generic-laptop",
		Name:     "Generic Laptop",
		Services: []string{"power-profiles-daemon.service"},
	}
}

// GenericDesktop returns a hardware profile for a generic desktop.
func GenericDesktop() *Profile {
	return &Profile{
		ID:   "generic-desktop",
		Name: "Generic Desktop",
	}
}
