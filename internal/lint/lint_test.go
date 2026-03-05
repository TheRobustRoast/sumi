package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckHyprlandDeprecated(t *testing.T) {
	tmp := t.TempDir()
	hyprDir := filepath.Join(tmp, "hypr")
	confD := filepath.Join(hyprDir, "conf.d")
	os.MkdirAll(confD, 0o755)

	// Main config
	os.WriteFile(filepath.Join(hyprDir, "hyprland.conf"), []byte("# empty"), 0o644)

	// Config with deprecated syntax
	os.WriteFile(filepath.Join(confD, "test.conf"), []byte(`
windowrulev2 = float, class:^(firefox)$
workspace_swipe = true
`), 0o644)

	issues := CheckHyprland(hyprDir)
	if len(issues) < 2 {
		t.Errorf("expected at least 2 issues for deprecated syntax, got %d", len(issues))
	}
}

func TestCheckFootDpiAware(t *testing.T) {
	tmp := t.TempDir()
	footIni := filepath.Join(tmp, "foot.ini")

	os.WriteFile(footIni, []byte(`[main]
dpi-aware = auto
font = monospace:size=11
`), 0o644)

	issues := CheckFoot(footIni)
	found := false
	for _, issue := range issues {
		if issue.Message == "invalid: dpi-aware only accepts yes/no (not auto)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dpi-aware=auto issue")
	}
}

func TestCheckFootTweak(t *testing.T) {
	tmp := t.TempDir()
	footIni := filepath.Join(tmp, "foot.ini")

	os.WriteFile(footIni, []byte(`[main]
font = monospace:size=11

[tweak]
overflowing-glyphs = true
`), 0o644)

	issues := CheckFoot(footIni)
	found := false
	for _, issue := range issues {
		if issue.Severity == SevWarning && issue.Message == "removed section: [tweak] is no longer valid in foot" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected [tweak] section warning")
	}
}

func TestRunAllEmpty(t *testing.T) {
	tmp := t.TempDir()
	issues := RunAll(tmp, tmp)
	// Should not panic on empty dirs
	_ = issues
}
