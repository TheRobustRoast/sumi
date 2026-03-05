package steps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sumi/internal/model"
)

// Migrations returns the step that runs numbered migration scripts.
func Migrations() model.Step {
	return model.Step{
		ID:      "migrations/run",
		Section: "Migrations",
		Name:    "run migrations",
		RunGo:   runMigrations,
	}
}

// ParseMigrationName extracts the numeric prefix from a migration filename.
// Returns the number string and whether the filename is valid.
// Valid: "0001-some-name.sh" → ("0001", true)
// Invalid: "readme.txt" → ("", false)
func ParseMigrationName(filename string) (string, bool) {
	if !strings.HasSuffix(filename, ".sh") {
		return "", false
	}
	parts := strings.SplitN(filename, "-", 2)
	if len(parts) < 2 {
		return "", false
	}
	num := parts[0]
	if len(num) == 0 {
		return "", false
	}
	for _, c := range num {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	return num, true
}

func runMigrations(ctx model.InstallCtx) ([]string, error) {
	migrationsDir := filepath.Join(ctx.SumiDir, "migrations")
	appliedFile := filepath.Join(ctx.Home, ".local/share/sumi/applied-migrations")

	if err := os.MkdirAll(filepath.Dir(appliedFile), 0o755); err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	if data, err := os.ReadFile(appliedFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if s := strings.TrimSpace(line); s != "" {
				applied[s] = true
			}
		}
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"no migrations"}, nil
		}
		return nil, err
	}

	var lines []string
	ran := 0
	for _, e := range entries {
		name := e.Name()
		num, valid := ParseMigrationName(name)
		if !valid || applied[num] {
			continue
		}

		out, err := exec.Command("/bin/bash", filepath.Join(migrationsDir, name)).CombinedOutput()
		if err != nil {
			return lines, fmt.Errorf("migration %s failed: %w\n%s", name, err, out)
		}

		f, ferr := os.OpenFile(appliedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if ferr != nil {
			return lines, ferr
		}
		if _, werr := fmt.Fprintln(f, num); werr != nil {
			f.Close()
			return lines, fmt.Errorf("write applied-migrations: %w", werr)
		}
		f.Close()
		lines = append(lines, "applied: "+name)
		ran++
	}

	if ran == 0 {
		lines = append(lines, "no pending migrations")
	}
	return lines, nil
}
