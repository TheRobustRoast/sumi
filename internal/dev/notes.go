package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NotesDir returns the notes directory path.
func NotesDir() string {
	return filepath.Join(os.Getenv("HOME"), "Documents/notes")
}

// OpenToday opens or creates today's daily note.
func OpenToday() error {
	dir := filepath.Join(NotesDir(), "daily")
	os.MkdirAll(dir, 0o755) //nolint:errcheck

	file := filepath.Join(dir, time.Now().Format("2006-01-02")+".md")
	if _, err := os.Stat(file); err != nil {
		// Create with template
		now := time.Now()
		content := fmt.Sprintf("# %s\n\n## Tasks\n\n- [ ]\n\n## Notes\n\n\n## Log\n\n- %s —\n",
			now.Format("Monday, January 02 2006"),
			now.Format("15:04"))
		os.WriteFile(file, []byte(content), 0o644) //nolint:errcheck
	}

	return OpenInEditor(file)
}

// NewNote creates a new note with the given title.
func NewNote(title string) error {
	dir := NotesDir()
	os.MkdirAll(dir, 0o755) //nolint:errcheck

	slug := slugify(title)
	file := filepath.Join(dir, time.Now().Format("20060102")+"-"+slug+".md")

	content := fmt.Sprintf("# %s\n\n*%s*\n\n---\n\n",
		title, time.Now().Format("2006-01-02 15:04"))
	os.WriteFile(file, []byte(content), 0o644) //nolint:errcheck

	return OpenInEditor(file)
}

// SearchNotes runs ripgrep across the notes directory.
func SearchNotes(query string) error {
	if _, err := exec.LookPath("rg"); err != nil {
		return fmt.Errorf("ripgrep (rg) not found — install with: pacman -S ripgrep")
	}
	cmd := exec.Command("rg", "--line-number", "--no-heading", "--color=always", query, NotesDir())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			return fmt.Errorf("no matches found for %q", query)
		}
		return fmt.Errorf("rg: %w", err)
	}
	return nil
}

// ListRecent returns the most recent note files.
func ListRecent(n int) ([]string, error) {
	dir := NotesDir()
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort by modification time (newest first)
	type fileWithTime struct {
		path    string
		modTime time.Time
	}
	var withTimes []fileWithTime
	for _, f := range files {
		if fi, err := os.Stat(f); err == nil {
			withTimes = append(withTimes, fileWithTime{f, fi.ModTime()})
		}
	}
	sort.Slice(withTimes, func(i, j int) bool {
		return withTimes[i].modTime.After(withTimes[j].modTime)
	})
	files = files[:0]
	for _, ft := range withTimes {
		files = append(files, ft.path)
	}

	if len(files) > n {
		files = files[:n]
	}
	return files, nil
}

// OpenInEditor opens a file in the user's preferred editor.
func OpenInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	if _, err := exec.LookPath(editor); err != nil {
		return fmt.Errorf("%s not found — set $EDITOR or install nvim: pacman -S neovim", editor)
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var buf strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			buf.WriteRune(c)
		} else {
			buf.WriteRune('-')
		}
	}
	result := buf.String()
	// Collapse multiple dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}
