// Package checkpoint tracks installer progress for resume-from-failure.
package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// State represents the current install progress saved to disk.
type State struct {
	CompletedSteps []string          `json:"completed_steps"`
	FailedStep     string            `json:"failed_step,omitempty"`
	StepHashes     map[string]string `json:"step_hashes,omitempty"`
	Timestamp      string            `json:"timestamp"`
}

// path returns the checkpoint file path.
func path() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".cache/sumi/install-progress.json")
}

// Load reads the checkpoint from disk. Returns nil if no checkpoint exists.
func Load() *State {
	data, err := os.ReadFile(path())
	if err != nil {
		return nil
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}

// Save writes the checkpoint state to disk.
func Save(s *State) error {
	s.Timestamp = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	p := path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// MarkCompleted adds a step ID to the completed list and saves.
func MarkCompleted(s *State, stepID string) error {
	s.CompletedSteps = append(s.CompletedSteps, stepID)
	s.FailedStep = ""
	return Save(s)
}

// MarkFailed records which step failed and saves.
func MarkFailed(s *State, stepID string) error {
	s.FailedStep = stepID
	return Save(s)
}

// Clear deletes the checkpoint file.
func Clear() {
	os.Remove(path()) //nolint:errcheck
}

// IsCompleted checks if a step ID has already been completed.
func IsCompleted(s *State, stepID string) bool {
	if s == nil {
		return false
	}
	for _, id := range s.CompletedSteps {
		if id == stepID {
			return true
		}
	}
	return false
}

// SetHash stores a hash value for a step ID so future runs can detect changes.
func SetHash(s *State, stepID, hash string) {
	if s.StepHashes == nil {
		s.StepHashes = make(map[string]string)
	}
	s.StepHashes[stepID] = hash
}

// HashChanged returns true if the stored hash for stepID differs from the given hash,
// meaning the step's inputs have changed and it should be re-run.
func HashChanged(s *State, stepID, hash string) bool {
	if s == nil || s.StepHashes == nil {
		return true
	}
	stored, ok := s.StepHashes[stepID]
	if !ok {
		return true
	}
	return stored != hash
}

// LoadStepHashes returns the persisted step hash file.
// This is separate from the install-progress checkpoint — it persists
// across successful installs for incremental update detection.
func LoadStepHashes() map[string]string {
	home := os.Getenv("HOME")
	p := filepath.Join(home, ".cache/sumi/step-hashes.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var hashes map[string]string
	if err := json.Unmarshal(data, &hashes); err != nil {
		return nil
	}
	return hashes
}

// SaveStepHashes persists step hashes to disk.
func SaveStepHashes(hashes map[string]string) error {
	home := os.Getenv("HOME")
	p := filepath.Join(home, ".cache/sumi/step-hashes.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(hashes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// ResumeIndex returns the index to resume from in the step list.
// If a step has no ID, it cannot be skipped via checkpoint.
func ResumeIndex(s *State, stepIDs []string) int {
	if s == nil {
		return 0
	}
	completed := make(map[string]bool, len(s.CompletedSteps))
	for _, id := range s.CompletedSteps {
		completed[id] = true
	}
	// Find the first non-completed step
	for i, id := range stepIDs {
		if id == "" || !completed[id] {
			return i
		}
	}
	return len(stepIDs)
}
