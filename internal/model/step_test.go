package model

import "testing"

func TestStepStatusOrdering(t *testing.T) {
	// Verify the iota ordering is stable: Pending < Running < Done < Skipped < Failed
	if StepPending >= StepRunning {
		t.Error("StepPending should be less than StepRunning")
	}
	if StepRunning >= StepDone {
		t.Error("StepRunning should be less than StepDone")
	}
	if StepDone >= StepSkipped {
		t.Error("StepDone should be less than StepSkipped")
	}
	if StepSkipped >= StepFailed {
		t.Error("StepSkipped should be less than StepFailed")
	}
}

func TestInstallCtxFields(t *testing.T) {
	ctx := InstallCtx{
		SumiDir: "/home/user/sumi",
		Home:    "/home/user",
		User:    "testuser",
	}
	if ctx.SumiDir != "/home/user/sumi" {
		t.Errorf("SumiDir = %q, want %q", ctx.SumiDir, "/home/user/sumi")
	}
	if ctx.Home != "/home/user" {
		t.Errorf("Home = %q, want %q", ctx.Home, "/home/user")
	}
	if ctx.User != "testuser" {
		t.Errorf("User = %q, want %q", ctx.User, "testuser")
	}
}

func TestStepZeroValue(t *testing.T) {
	var s Step
	if s.Status != StepPending {
		t.Errorf("zero-value Step.Status = %d, want StepPending (%d)", s.Status, StepPending)
	}
	if s.Name != "" {
		t.Errorf("zero-value Step.Name = %q, want empty", s.Name)
	}
	if s.RunGo != nil {
		t.Error("zero-value Step.RunGo should be nil")
	}
	if s.RunStream != nil {
		t.Error("zero-value Step.RunStream should be nil")
	}
	if s.Skip != nil {
		t.Error("zero-value Step.Skip should be nil")
	}
}
