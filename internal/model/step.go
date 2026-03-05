// Package model defines shared types used across the ui and steps packages.
package model

import (
	tea "github.com/charmbracelet/bubbletea"

	"sumi/internal/runner"
)

// StepStatus tracks the state of a single installer step.
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepDone
	StepSkipped
	StepFailed
)

// InstallCtx is the immutable context passed to all step functions.
type InstallCtx struct {
	SumiDir string
	Home    string
	User    string
}

// Step is a single unit of work in a step-runner TUI.
type Step struct {
	ID        string // stable identifier (e.g. "packages/aur") for checkpointing
	Section   string // section header (printed when it changes)
	Name      string
	RunGo     func(ctx InstallCtx) ([]string, error)
	RunStream func(ctx InstallCtx) (*runner.Stream, tea.Cmd)
	Skip      func(ctx InstallCtx) bool
	Status    StepStatus
}

// GoResultMsg carries the result of a RunGo step.
type GoResultMsg struct {
	Lines []string
	Err   error
}

// NextStepMsg advances the step-runner state machine.
type NextStepMsg struct{}
