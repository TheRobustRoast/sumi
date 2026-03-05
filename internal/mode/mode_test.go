package mode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGaming(t *testing.T) {
	tmp := t.TempDir()
	cacheDir := filepath.Join(tmp, ".cache/sumi")
	os.MkdirAll(cacheDir, 0o755)

	if IsGaming(tmp) {
		t.Error("expected gaming off by default")
	}

	os.WriteFile(filepath.Join(cacheDir, "gaming-mode"), []byte("on\n"), 0o644)
	if !IsGaming(tmp) {
		t.Error("expected gaming on after write")
	}

	os.WriteFile(filepath.Join(cacheDir, "gaming-mode"), []byte("off\n"), 0o644)
	if IsGaming(tmp) {
		t.Error("expected gaming off after write")
	}
}

func TestIsFocus(t *testing.T) {
	tmp := t.TempDir()
	cacheDir := filepath.Join(tmp, ".cache/sumi")
	os.MkdirAll(cacheDir, 0o755)

	if IsFocus(tmp) {
		t.Error("expected focus off by default")
	}

	os.WriteFile(filepath.Join(cacheDir, "focus-mode"), []byte("on\n"), 0o644)
	if !IsFocus(tmp) {
		t.Error("expected focus on after write")
	}
}

func TestWaybarVisibility(t *testing.T) {
	tmp := t.TempDir()
	cacheDir := filepath.Join(tmp, ".cache/sumi")
	os.MkdirAll(cacheDir, 0o755)

	if !isWaybarVisible(tmp) {
		t.Error("expected waybar visible by default")
	}

	os.WriteFile(filepath.Join(cacheDir, "waybar-hidden"), []byte("1"), 0o644)
	if isWaybarVisible(tmp) {
		t.Error("expected waybar hidden after state file")
	}

	os.Remove(filepath.Join(cacheDir, "waybar-hidden"))
	if !isWaybarVisible(tmp) {
		t.Error("expected waybar visible after remove")
	}
}
