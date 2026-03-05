package wallpaper

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Init loads the cached wallpaper, or picks the first image in the wallpapers dir,
// or generates a solid dark fallback. Called on login via autostart.
func Init(home string) error {
	wpDir := filepath.Join(home, "Pictures/Wallpapers")
	cacheDir := filepath.Join(home, ".cache/sumi")

	os.MkdirAll(cacheDir, 0o755) //nolint:errcheck
	os.MkdirAll(wpDir, 0o755)    //nolint:errcheck

	var wallpaper string

	// Try cached wallpaper
	cached := CurrentWallpaper(home)
	if cached != "" {
		if _, err := os.Stat(cached); err == nil {
			wallpaper = cached
		}
	}

	// Find first image in wallpapers dir
	if wallpaper == "" {
		wallpaper = findFirstImage(wpDir)
	}

	// Generate starter wallpapers if dir is empty
	if wallpaper == "" {
		generateStarters(wpDir)
		wallpaper = findFirstImage(wpDir)
	}

	// Absolute fallback — generate solid dark
	if wallpaper == "" {
		fallback := filepath.Join(wpDir, "default.png")
		if err := generateSolid(fallback, color.RGBA{10, 10, 10, 255}); err != nil {
			return fmt.Errorf("generate fallback: %w", err)
		}
		wallpaper = fallback
	}

	// Wait for Hyprland to be ready (up to 5s)
	for i := 0; i < 50; i++ {
		if exec.Command("hyprctl", "monitors", "-j").Run() == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return Apply(wallpaper, ApplyOpts{Home: home})
}

// Random picks a random wallpaper from ~/Pictures/Wallpapers and applies it.
func Random(home string) error {
	wpDir := filepath.Join(home, "Pictures/Wallpapers")
	images := listImages(wpDir)
	if len(images) == 0 {
		return fmt.Errorf("no wallpapers found in %s", wpDir)
	}

	// Use /dev/urandom via shuf for randomness
	shufArgs := append([]string{"-n", "1", "-e"}, images...)
	out, err := exec.Command("shuf", shufArgs...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		// Fallback: use time-based selection
		idx := int(time.Now().UnixNano()) % len(images)
		return Apply(images[idx], ApplyOpts{Home: home})
	}
	return Apply(strings.TrimSpace(string(out)), ApplyOpts{Home: home})
}

// ListWallpapers returns all image paths in ~/Pictures/Wallpapers.
func ListWallpapers(home string) []string {
	return listImages(filepath.Join(home, "Pictures/Wallpapers"))
}

func findFirstImage(dir string) string {
	images := listImages(dir)
	if len(images) > 0 {
		return images[0]
	}
	return ""
}

func listImages(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var images []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg") ||
			strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".webp") ||
			strings.HasSuffix(name, ".bmp") {
			images = append(images, filepath.Join(dir, e.Name()))
		}
	}
	return images
}

// generateStarters creates 3 gradient wallpapers if the wallpaper dir is empty.
func generateStarters(dir string) {
	generateGradient(filepath.Join(dir, "sumi-deep-blue.png"),
		color.RGBA{5, 5, 20, 255}, color.RGBA{10, 25, 50, 255})
	generateGradient(filepath.Join(dir, "sumi-warm-dark.png"),
		color.RGBA{15, 8, 5, 255}, color.RGBA{30, 15, 10, 255})
	generateGradient(filepath.Join(dir, "sumi-monochrome.png"),
		color.RGBA{8, 8, 8, 255}, color.RGBA{20, 20, 20, 255})
}

func generateGradient(path string, top, bottom color.RGBA) error {
	const w, h = 3840, 2160
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		t := float64(y) / float64(h-1)
		r := lerp(float64(top.R), float64(bottom.R), t)
		g := lerp(float64(top.G), float64(bottom.G), t)
		b := lerp(float64(top.B), float64(bottom.B), t)
		c := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func generateSolid(path string, c color.RGBA) error {
	const w, h = 3840, 2160
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func lerp(a, b, t float64) float64 {
	return math.Round(a + (b-a)*t)
}
