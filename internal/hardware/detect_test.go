package hardware

import "testing"

// mockDMI implements DMIReader for testing.
type mockDMI struct {
	vendor  string
	product string
	battery bool
}

func (m mockDMI) Read(field string) string {
	switch field {
	case "sys_vendor":
		return m.vendor
	case "product_name":
		return m.product
	}
	return ""
}

func (m mockDMI) HasBattery() bool {
	return m.battery
}

func TestDetectFramework(t *testing.T) {
	r := mockDMI{vendor: "Framework", product: "Laptop 13 (AMD Ryzen 7040)"}
	p := DetectWith(r)
	if p.ID != "framework-13-amd" {
		t.Errorf("got ID %q, want %q", p.ID, "framework-13-amd")
	}
}

func TestDetectThinkPad(t *testing.T) {
	r := mockDMI{vendor: "LENOVO", product: "ThinkPad X1 Carbon Gen 12", battery: true}
	p := DetectWith(r)
	if p.ID != "thinkpad" {
		t.Errorf("got ID %q, want %q", p.ID, "thinkpad")
	}
}

func TestDetectASUS(t *testing.T) {
	r := mockDMI{vendor: "ASUSTeK", product: "ROG Strix G16", battery: true}
	p := DetectWith(r)
	if p.ID != "asus-rog" {
		t.Errorf("got ID %q, want %q", p.ID, "asus-rog")
	}
}

func TestDetectGenericLaptop(t *testing.T) {
	r := mockDMI{vendor: "Dell Inc.", product: "XPS 15", battery: true}
	p := DetectWith(r)
	if p.ID != "generic-laptop" {
		t.Errorf("got ID %q, want %q", p.ID, "generic-laptop")
	}
}

func TestDetectGenericDesktop(t *testing.T) {
	r := mockDMI{vendor: "Gigabyte Technology", product: "B550 AORUS PRO", battery: false}
	p := DetectWith(r)
	if p.ID != "generic-desktop" {
		t.Errorf("got ID %q, want %q", p.ID, "generic-desktop")
	}
}

func TestDetectGenericDesktopFallback(t *testing.T) {
	r := mockDMI{vendor: "", product: "", battery: false}
	p := DetectWith(r)
	if p.ID != "generic-desktop" {
		t.Errorf("got ID %q, want %q", p.ID, "generic-desktop")
	}
}
