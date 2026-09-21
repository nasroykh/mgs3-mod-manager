package manager

import (
	"errors"
	"fmt"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"time"
)

const stateDir = ".mgs3mod"

type Error struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Paths   []string `json:"paths,omitempty"`
}

func (e *Error) Error() string                         { return e.Message }
func fail(code int, msg string, paths ...string) error { return &Error{code, msg, paths} }
func ExitCode(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 1
}

type Baseline struct {
	Target   string `json:"target"`
	SHA256   string `json:"sha256"`
	Bytes    int64  `json:"bytes"`
	Captured string `json:"captured"`
	Profile  string `json:"profile"`
	Absent   bool   `json:"absent,omitempty"`
}
type Mod struct {
	Manifest packagefmt.Manifest `json:"manifest"`
	Digest   string              `json:"digest"`
	Enabled  bool                `json:"enabled"`
}
type State struct {
	Generation uint64              `json:"generation"`
	Mods       map[string]Mod      `json:"mods"`
	Baselines  map[string]Baseline `json:"baselines"`
}

func emptyState() State { return State{Mods: map[string]Mod{}, Baselines: map[string]Baseline{}} }

type Result struct {
	Command                     string         `json:"command"`
	DryRun                      bool           `json:"dryRun,omitempty"`
	Message                     string         `json:"message"`
	State                       *State         `json:"state,omitempty"`
	Paths                       []string       `json:"paths,omitempty"`
	RequiredBytes               uint64         `json:"requiredBytes,omitempty"`
	MissingFileRecoveryExplicit bool           `json:"missingFileRecoveryExplicit,omitempty"`
	Compatible                  bool           `json:"compatible"`
	Initialized                 bool           `json:"initialized"`
	Issues                      []string       `json:"issues,omitempty"`
	Committed                   bool           `json:"committed,omitempty"`
	RecoveryRequired            bool           `json:"recoveryRequired,omitempty"`
	Targets                     []TargetStatus `json:"targets,omitempty"`
}

type TargetStatus struct {
	Path           string `json:"path"`
	Owner          string `json:"owner,omitempty"`
	ExpectedSHA256 string `json:"expectedSha256,omitempty"`
	ExpectedAbsent bool   `json:"expectedAbsent,omitempty"`
	ActualSHA256   string `json:"actualSha256,omitempty"`
	ActualAbsent   bool   `json:"actualAbsent,omitempty"`
	Issue          string `json:"issue,omitempty"`
}
type Options struct {
	DryRun         bool
	RestoreMissing []string
}

type Config struct {
	Root           string
	Core           []profile.Fingerprint
	CheckProcesses func(string) error
	CheckASILoader func(string) error
	// Fault is injected only by internal synthetic tests. The CLI never exposes it.
	Fault func(string) error
	// CreateFile is an internal test seam. Production always uses exclusive os.Root creation.
	CreateFile transaction.FileFactory
	// Launch seams are private so production callers cannot replace process or
	// atomic-move behavior. Manager tests use them without starting the game.
	runLaunch     launchRunner
	observeLaunch launchObserver
	replaceLaunch func(string, string) error
	moveLaunch    func(string, string) error
	launchTimeout time.Duration
	launchPoll    time.Duration
}
type Manager struct{ config Config }

func New(c Config) *Manager { return &Manager{c} }
func Production() *Manager  { return New(Config{Root: profile.Root, Core: profile.Core}) }
func (m *Manager) point(name string) error {
	if m.config.Fault != nil {
		return m.config.Fault(name)
	}
	return nil
}
func wrap(code int, label string, err error) error {
	if err == nil {
		return nil
	}
	var nested *Error
	if errors.As(err, &nested) {
		return fail(code, fmt.Sprintf("%s: %v", label, err), nested.Paths...)
	}
	return fail(code, fmt.Sprintf("%s: %v", label, err))
}

func wrapPath(code int, label, path string, err error) error {
	if err == nil {
		return nil
	}
	return fail(code, fmt.Sprintf("%s: %v", label, err), path)
}

type change struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  string `json:"after"`
	Game   bool   `json:"game"`
}
type inventory struct {
	Path                  string `json:"path"`
	SHA256                string `json:"sha256"`
	Bytes                 int64  `json:"bytes"`
	Target                string `json:"target,omitempty"`
	Temporary             string `json:"temporary,omitempty"`
	VerifiedObjectExisted *bool  `json:"verifiedObjectExisted,omitempty"`
}
type plan struct {
	Schema             int         `json:"schema"`
	Sequence           uint64      `json:"sequence"`
	Operation          string      `json:"operation"`
	Before             State       `json:"before"`
	After              State       `json:"after"`
	Changes            []change    `json:"changes"`
	Cleanup            []inventory `json:"cleanup"`
	Baselines          []inventory `json:"baselines"`
	ConditionalRestore bool        `json:"conditionalRestore"`
}
type marker struct {
	Plan  string `json:"plan"`
	Kind  string `json:"kind"`
	Index int    `json:"index"`
}
type initRecord struct {
	Schema  int      `json:"schema"`
	Profile string   `json:"profile"`
	Root    string   `json:"root"`
	Paths   []string `json:"paths"`
}
