package dev

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree.
type Worktree struct {
	Path   string
	Branch string
}

// ListWorktrees parses `git worktree list --porcelain` output.
func ListWorktrees() ([]Worktree, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		if exec.Command("git", "rev-parse", "--git-dir").Run() != nil {
			return nil, fmt.Errorf("not in a git repository")
		}
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	var worktrees []Worktree
	var current Worktree

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch "):
			branch := strings.TrimPrefix(line, "branch ")
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		case line == "":
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// CreateWorktree creates a new worktree for the given branch.
func CreateWorktree(branch string) error {
	// Validate branch name to prevent ref injection
	if err := exec.Command("git", "check-ref-format", "--branch", branch).Run(); err != nil {
		return fmt.Errorf("invalid branch name %q", branch)
	}

	root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("not in a git repository")
	}
	repoRoot := strings.TrimSpace(string(root))
	wtDir := filepath.Join(filepath.Dir(repoRoot), filepath.Base(repoRoot)+"-"+branch)

	// Check if branch exists locally
	if exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil {
		if err := exec.Command("git", "worktree", "add", wtDir, branch).Run(); err != nil {
			return fmt.Errorf("git worktree add %s: %w", branch, err)
		}
		return nil
	}
	// Check remote
	if exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch).Run() == nil {
		if err := exec.Command("git", "worktree", "add", "--track", "-b", branch, wtDir, "origin/"+branch).Run(); err != nil {
			return fmt.Errorf("git worktree add %s (tracking): %w", branch, err)
		}
		return nil
	}
	// New branch from HEAD
	if err := exec.Command("git", "worktree", "add", "-b", branch, wtDir).Run(); err != nil {
		return fmt.Errorf("git worktree add -b %s: %w", branch, err)
	}
	return nil
}

// RemoveWorktree removes a worktree by path.
func RemoveWorktree(path string) error {
	if err := exec.Command("git", "worktree", "remove", path).Run(); err != nil {
		if err2 := exec.Command("git", "worktree", "remove", "--force", path).Run(); err2 != nil {
			return fmt.Errorf("git worktree remove %s: %w", path, err2)
		}
	}
	return nil
}

// PruneWorktrees cleans up stale worktree references.
func PruneWorktrees() error {
	cmd := exec.Command("git", "worktree", "prune", "-v")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		fmt.Print(string(out))
	}
	if err != nil {
		return fmt.Errorf("git worktree prune: %w", err)
	}
	return nil
}
