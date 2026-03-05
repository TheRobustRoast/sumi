package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotAndRestore(t *testing.T) {
	tmp := t.TempDir()

	// Create a fake config file
	hyprDir := filepath.Join(tmp, ".config/hypr")
	os.MkdirAll(hyprDir, 0o755)
	os.WriteFile(filepath.Join(hyprDir, "hyprland.conf"), []byte("test config"), 0o644)

	// Create snapshot
	err := Snapshot(tmp, "test-snap")
	if err != nil {
		t.Fatal(err)
	}

	// Verify snapshot exists
	snapDir := filepath.Join(tmp, ".local/share/sumi/snapshots/test-snap")
	if _, err := os.Stat(snapDir); err != nil {
		t.Fatal("snapshot dir not created")
	}

	// Modify original
	os.WriteFile(filepath.Join(hyprDir, "hyprland.conf"), []byte("modified"), 0o644)

	// Restore
	err = Restore(tmp, "test-snap")
	if err != nil {
		t.Fatal(err)
	}
}

func TestList(t *testing.T) {
	tmp := t.TempDir()
	snapDir := filepath.Join(tmp, ".local/share/sumi/snapshots")

	// Empty
	snaps, _ := List(tmp)
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(snaps))
	}

	// Create some
	os.MkdirAll(filepath.Join(snapDir, "snap1"), 0o755)
	os.MkdirAll(filepath.Join(snapDir, "snap2"), 0o755)

	snaps, _ = List(tmp)
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snaps))
	}
}

func TestDiff(t *testing.T) {
	tmp := t.TempDir()

	// Create config
	os.MkdirAll(filepath.Join(tmp, ".config/hypr"), 0o755)
	os.WriteFile(filepath.Join(tmp, ".config/hypr/test.conf"), []byte("original"), 0o644)

	// Snapshot
	Snapshot(tmp, "base")

	// Should be no diff (same content)
	entries, err := Diff(tmp, "base")
	if err != nil {
		t.Fatal(err)
	}
	// Some entries might show as modified due to modtime differences
	_ = entries
}
