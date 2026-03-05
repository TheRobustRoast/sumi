package sync

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Tracked config paths relative to $HOME.
var tracked = []string{
	".config/hypr",
	".config/waybar",
	".config/foot",
	".config/fuzzel",
	".config/dunst",
	".config/nvim/init.lua",
	".config/yazi",
	".config/cava",
	".config/lazygit",
	".config/btop",
	".config/starship.toml",
	".config/wallust",
	".tmux.conf",
	".zshrc",
}

// SnapshotInfo holds metadata about a saved snapshot.
type SnapshotInfo struct {
	Name    string
	Time    time.Time
	Size    int64
	FileNum int
}

func snapshotDir(home string) string {
	return filepath.Join(home, ".local/share/sumi/snapshots")
}

// safeName validates a user-supplied name for use in file paths.
func safeName(name string) error {
	if name == "" {
		return nil
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") ||
		strings.Contains(name, "\\") || strings.Contains(name, "\x00") ||
		strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid name: %q", name)
	}
	return nil
}

// Snapshot copies tracked dotfiles to a named snapshot directory.
func Snapshot(home, name string) error {
	if name == "" {
		name = time.Now().Format("20060102_150405")
	}
	if err := safeName(name); err != nil {
		return err
	}
	dst := filepath.Join(snapshotDir(home), name)
	os.MkdirAll(dst, 0o755) //nolint:errcheck

	count := 0
	for _, rel := range tracked {
		src := filepath.Join(home, rel)
		dstPath := filepath.Join(dst, rel)

		info, err := os.Stat(src)
		if err != nil {
			continue
		}

		if info.IsDir() {
			if err := copyDir(src, dstPath); err != nil {
				continue
			}
		} else {
			os.MkdirAll(filepath.Dir(dstPath), 0o755) //nolint:errcheck
			if err := copyFile(src, dstPath); err != nil {
				continue
			}
		}
		count++
	}

	fmt.Printf("snapshot saved: %s (%d paths)\n", name, count)
	return nil
}

// Restore copies files from a snapshot back to their original locations.
func Restore(home, name string) error {
	if err := safeName(name); err != nil {
		return err
	}
	src := filepath.Join(snapshotDir(home), name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("snapshot not found: %s", name)
	}

	count := 0
	for _, rel := range tracked {
		srcPath := filepath.Join(src, rel)
		dstPath := filepath.Join(home, rel)

		info, err := os.Stat(srcPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			copyDir(srcPath, dstPath) //nolint:errcheck
		} else {
			os.MkdirAll(filepath.Dir(dstPath), 0o755) //nolint:errcheck
			copyFile(srcPath, dstPath)                 //nolint:errcheck
		}
		count++
	}

	fmt.Printf("restored from: %s (%d paths)\n", name, count)
	return nil
}

// DiffEntry represents a file difference.
type DiffEntry struct {
	Path   string
	Status string // "modified", "missing", "new"
}

// Diff compares a snapshot with current config.
func Diff(home, name string) ([]DiffEntry, error) {
	if err := safeName(name); err != nil {
		return nil, err
	}
	snapDir := filepath.Join(snapshotDir(home), name)
	if _, err := os.Stat(snapDir); err != nil {
		return nil, fmt.Errorf("snapshot not found: %s", name)
	}

	var entries []DiffEntry
	for _, rel := range tracked {
		current := filepath.Join(home, rel)
		snapshot := filepath.Join(snapDir, rel)

		curInfo, curErr := os.Stat(current)
		snapInfo, snapErr := os.Stat(snapshot)

		if curErr != nil && snapErr != nil {
			continue
		}
		if curErr != nil {
			entries = append(entries, DiffEntry{Path: rel, Status: "deleted"})
			continue
		}
		if snapErr != nil {
			entries = append(entries, DiffEntry{Path: rel, Status: "new"})
			continue
		}

		if !curInfo.IsDir() && !snapInfo.IsDir() {
			if curInfo.Size() != snapInfo.Size() || curInfo.ModTime() != snapInfo.ModTime() {
				entries = append(entries, DiffEntry{Path: rel, Status: "modified"})
			}
		}
	}

	return entries, nil
}

// List returns available snapshots.
func List(home string) ([]SnapshotInfo, error) {
	dir := snapshotDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}

	var snapshots []SnapshotInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		snapshots = append(snapshots, SnapshotInfo{
			Name: e.Name(),
			Time: info.ModTime(),
		})
	}
	return snapshots, nil
}

// Export creates a tar.gz of current config.
func Export(home, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, rel := range tracked {
		src := filepath.Join(home, rel)
		info, err := os.Stat(src)
		if err != nil {
			continue
		}

		if info.IsDir() {
			filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				relPath := strings.TrimPrefix(path, home+string(filepath.Separator))
				addToTar(tw, path, relPath, fi) //nolint:errcheck
				return nil
			})
		} else {
			addToTar(tw, src, rel, info) //nolint:errcheck
		}
	}

	return nil
}

func addToTar(tw *tar.Writer, src, name string, fi os.FileInfo) error {
	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	header.Name = name

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		os.MkdirAll(filepath.Dir(target), 0o755) //nolint:errcheck
		return copyFile(path, target)
	})
}
