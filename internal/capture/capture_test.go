package capture

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestTimestamp(t *testing.T) {
	ts := timestamp()
	if len(ts) != 15 { // YYYYMMDD_HHMMSS
		t.Errorf("unexpected timestamp format: %q", ts)
	}
	if !strings.Contains(ts, "_") {
		t.Errorf("timestamp should contain underscore: %q", ts)
	}
}

func TestScreenshotDir(t *testing.T) {
	dir := screenshotDir("/home/user")
	want := filepath.Join("/home/user", "Pictures", "Screenshots")
	if dir != want {
		t.Errorf("unexpected dir: %s, want %s", dir, want)
	}
}

func TestRecordDir(t *testing.T) {
	dir := recordDir("/home/user")
	want := filepath.Join("/home/user", "Videos", "Recordings")
	if dir != want {
		t.Errorf("unexpected dir: %s, want %s", dir, want)
	}
}

func TestHomeDir(t *testing.T) {
	if got := homeDir("/test"); got != "/test" {
		t.Errorf("expected /test, got %s", got)
	}
}
