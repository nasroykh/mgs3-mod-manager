//go:build windows

package winfs

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestObserveProcessCurrentImage(t *testing.T) {
	expected, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	observation, err := ObserveProcess(windows.GetCurrentProcessId(), expected)
	if err != nil {
		t.Fatal(err)
	}
	if observation.State != ProcessAliveExpected {
		t.Fatalf("state = %s, want %s", observation.State, ProcessAliveExpected)
	}
	if !sameWindowsPath(observation.ImagePath, expected) {
		t.Fatalf("image = %q, want %q", observation.ImagePath, expected)
	}
}

func TestObserveProcessRejectsWrongImage(t *testing.T) {
	expected, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrong := filepath.Join(os.Getenv("WINDIR"), "System32", "cmd.exe")
	observation, err := ObserveProcess(windows.GetCurrentProcessId(), wrong)
	if err != nil {
		t.Fatal(err)
	}
	if observation.State != ProcessWrongImage {
		t.Fatalf("state = %s, want %s", observation.State, ProcessWrongImage)
	}
	if !sameWindowsPath(observation.ImagePath, expected) {
		t.Fatalf("image = %q, want current test process %q", observation.ImagePath, expected)
	}
}

func TestObserveProcessClassifiesExitedChild(t *testing.T) {
	child := exec.Command(os.Getenv("COMSPEC"), "/d", "/c", "exit 0")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	pid := uint32(child.Process.Pid)
	if err := child.Wait(); err != nil {
		t.Fatal(err)
	}

	expected, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	observation, err := ObserveProcess(pid, expected)
	if err != nil {
		t.Fatalf("observe exited child: %v", err)
	}
	if observation.State != ProcessExited {
		t.Fatalf("state = %s, want %s", observation.State, ProcessExited)
	}
}

func TestObserveProcessRejectsInvalidPID(t *testing.T) {
	expected, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, err = ObserveProcess(0, expected)
	if err == nil {
		t.Fatal("expected invalid PID error")
	}
	var observationErr *ProcessObservationError
	if errors.As(err, &observationErr) {
		t.Fatalf("invalid PID returned observation error: %v", err)
	}
}
