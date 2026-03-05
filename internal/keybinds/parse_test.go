package keybinds

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseKeybindsConf(t *testing.T) {
	tmp := t.TempDir()
	conf := filepath.Join(tmp, "keybinds.conf")

	content := `$mod = SUPER

# ── Core ─────────────────────────────────────────────────────
bind = $mod, Return, exec, foot
bind = $mod, Q, killactive
# app launcher
bind = $mod, D, exec, fuzzel

# ── Focus (vim) ──────────────────────────────────────────────
bind = $mod, H, movefocus, l
binde = $mod CTRL, H, resizeactive, -40 0
bindm = $mod, mouse:272, movewindow
`
	os.WriteFile(conf, []byte(content), 0o644)

	binds, err := ParseKeybindsConf(conf)
	if err != nil {
		t.Fatal(err)
	}

	if len(binds) < 5 {
		t.Fatalf("expected at least 5 binds, got %d", len(binds))
	}

	// Check first bind
	if binds[0].Key != "Return" {
		t.Errorf("expected Return, got %s", binds[0].Key)
	}
	if binds[0].Section != "Core" {
		t.Errorf("expected Core section, got %s", binds[0].Section)
	}

	// Check section change
	found := false
	for _, b := range binds {
		if b.Section == "Focus (vim)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Focus (vim) section")
	}

	// Check comment-based description
	if binds[2].Description != "app launcher" {
		t.Errorf("expected 'app launcher', got %q", binds[2].Description)
	}
}

func TestFormatMod(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"$mod", "SUPER"},
		{"$mod SHIFT", "SUPER SHIFT"},
		{"$mod ALT", "SUPER ALT"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatMod(tt.input)
		if got != tt.want {
			t.Errorf("formatMod(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
