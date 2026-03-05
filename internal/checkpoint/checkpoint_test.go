package checkpoint

import "testing"

func TestIsCompleted(t *testing.T) {
	s := &State{CompletedSteps: []string{"preflight/internet", "packages/aur"}}
	if !IsCompleted(s, "preflight/internet") {
		t.Error("expected preflight/internet to be completed")
	}
	if IsCompleted(s, "services/enable") {
		t.Error("expected services/enable to NOT be completed")
	}
	if IsCompleted(nil, "anything") {
		t.Error("expected nil state to return false")
	}
}

func TestResumeIndex(t *testing.T) {
	stepIDs := []string{"a", "b", "c", "d"}

	// No checkpoint — start from 0
	if idx := ResumeIndex(nil, stepIDs); idx != 0 {
		t.Errorf("ResumeIndex(nil) = %d, want 0", idx)
	}

	// a,b completed — resume from c (index 2)
	s := &State{CompletedSteps: []string{"a", "b"}}
	if idx := ResumeIndex(s, stepIDs); idx != 2 {
		t.Errorf("ResumeIndex(a,b completed) = %d, want 2", idx)
	}

	// All completed — return len
	s = &State{CompletedSteps: []string{"a", "b", "c", "d"}}
	if idx := ResumeIndex(s, stepIDs); idx != 4 {
		t.Errorf("ResumeIndex(all completed) = %d, want 4", idx)
	}
}
