package wallpaper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLockfileCreation(t *testing.T) {
	tmp := t.TempDir()
	lockfile := filepath.Join(tmp, ".cache/sumi/.wallpaper-applying")

	// Lockfile should not exist yet
	if _, err := os.Stat(lockfile); err == nil {
		t.Fatal("lockfile should not exist before apply")
	}
}

func TestCurrentWallpaper(t *testing.T) {
	tmp := t.TempDir()
	cacheDir := filepath.Join(tmp, ".cache/sumi")
	os.MkdirAll(cacheDir, 0o755)

	// No file → empty
	if got := CurrentWallpaper(tmp); got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	// Write file
	os.WriteFile(filepath.Join(cacheDir, "current-wallpaper"), []byte("/path/to/wallpaper.png\n"), 0o644)
	if got := CurrentWallpaper(tmp); got != "/path/to/wallpaper.png" {
		t.Errorf("expected /path/to/wallpaper.png, got %q", got)
	}
}

func TestFindFirstImage(t *testing.T) {
	tmp := t.TempDir()

	// Empty dir → empty
	if got := findFirstImage(tmp); got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	// Create some files
	os.WriteFile(filepath.Join(tmp, "readme.txt"), []byte("hi"), 0o644)
	os.WriteFile(filepath.Join(tmp, "alpha.png"), []byte("img"), 0o644)
	os.WriteFile(filepath.Join(tmp, "beta.jpg"), []byte("img"), 0o644)

	got := findFirstImage(tmp)
	if got != filepath.Join(tmp, "alpha.png") {
		t.Errorf("expected alpha.png, got %q", got)
	}
}

func TestListImages(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.png"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(tmp, "b.jpg"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(tmp, "c.txt"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(tmp, "d.webp"), []byte(""), 0o644)

	images := listImages(tmp)
	if len(images) != 3 {
		t.Errorf("expected 3 images, got %d: %v", len(images), images)
	}
}

func TestGenerateStarters(t *testing.T) {
	tmp := t.TempDir()
	generateStarters(tmp)

	expected := []string{"sumi-deep-blue.png", "sumi-warm-dark.png", "sumi-monochrome.png"}
	for _, name := range expected {
		path := filepath.Join(tmp, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("starter %s not created: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("starter %s is empty", name)
		}
	}
}

func TestValidateTemplatesEmpty(t *testing.T) {
	tmp := t.TempDir()
	tplDir := filepath.Join(tmp, "templates")
	os.MkdirAll(tplDir, 0o755)

	issues := ValidateTemplates(tmp)
	if len(issues) != 0 {
		t.Errorf("expected no issues for empty templates dir, got %v", issues)
	}
}
