//go:build windows

package winfs

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// ProcessObservationState describes what was learned about a PID.
type ProcessObservationState uint8

const (
	// ProcessNotObservable means Windows did not expose a stable process image
	// yet. Callers may retry this state during a bounded launch handoff.
	ProcessNotObservable ProcessObservationState = iota
	// ProcessAliveExpected means PID is alive and has expected image.
	ProcessAliveExpected
	// ProcessExited means PID no longer exists.
	ProcessExited
	// ProcessWrongImage means PID is alive but image differs from expected.
	ProcessWrongImage
	// ProcessInaccessible means PID exists but its handle or image is inaccessible.
	ProcessInaccessible
)

func (s ProcessObservationState) String() string {
	switch s {
	case ProcessNotObservable:
		return "not observable"
	case ProcessAliveExpected:
		return "alive expected"
	case ProcessExited:
		return "exited"
	case ProcessWrongImage:
		return "wrong image"
	case ProcessInaccessible:
		return "inaccessible"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// ProcessObservation is one point-in-time observation of a PID.
type ProcessObservation struct {
	PID       uint32
	State     ProcessObservationState
	ImagePath string
}

// ProcessObservationError carries an inaccessible or otherwise unobservable
// observation while retaining its typed state for bounded polling callers.
type ProcessObservationError struct {
	Observation ProcessObservation
	Err         error
}

func (e *ProcessObservationError) Error() string {
	return fmt.Sprintf("process %d %s: %v", e.Observation.PID, e.Observation.State, e.Err)
}

func (e *ProcessObservationError) Unwrap() error { return e.Err }

// ObserveProcess opens pid, reads its full executable image, and compares it
// with expectedPath. It never trusts the process basename. A PID that exits
// during observation is classified as ProcessExited; transient query failures
// are ProcessNotObservable and may be retried by the caller.
func ObserveProcess(pid uint32, expectedPath string) (ProcessObservation, error) {
	observation := ProcessObservation{PID: pid, State: ProcessNotObservable}
	if pid == 0 {
		return observation, fmt.Errorf("process pid must be nonzero")
	}
	if err := validateAbsolute(expectedPath); err != nil {
		return observation, fmt.Errorf("expected process image: %w", err)
	}
	expectedPath = filepath.Clean(expectedPath)
	if resolved, err := longPathName(expectedPath); err == nil {
		expectedPath = resolved
	}

	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			observation.State = ProcessExited
			return observation, nil
		}
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			observation.State = ProcessInaccessible
			return observation, &ProcessObservationError{Observation: observation, Err: err}
		}
		return observation, &ProcessObservationError{Observation: observation, Err: err}
	}
	defer windows.CloseHandle(process)

	imagePath, err := processImagePath(process)
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) || errors.Is(err, windows.ERROR_INVALID_HANDLE) {
			observation.State = ProcessExited
			return observation, nil
		}
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			observation.State = ProcessInaccessible
		} else {
			observation.State = ProcessNotObservable
		}
		return observation, &ProcessObservationError{Observation: observation, Err: err}
	}
	observation.ImagePath = imagePath
	resolvedImage, err := longPathName(imagePath)
	if err != nil {
		observation.State = ProcessNotObservable
		return observation, &ProcessObservationError{Observation: observation, Err: err}
	}
	observation.ImagePath = resolvedImage
	if sameWindowsPath(resolvedImage, expectedPath) {
		observation.State = ProcessAliveExpected
	} else {
		observation.State = ProcessWrongImage
	}
	return observation, nil
}
