package manager

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"mgs3mod/internal/launch"
	"mgs3mod/internal/transaction"
	"mgs3mod/internal/winfs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	launchConfigPath      = "mgs3mod-launch.json"
	launchExecutable      = "METAL GEAR SOLID3.exe"
	launchLauncher        = "launcher.exe"
	defaultLaunchTimeout  = 5 * time.Second
	defaultLaunchPoll     = 50 * time.Millisecond
	maxLaunchConfigBytes  = 1 << 20
	launchTemporaryPrefix = ".mgs3mod-launch."
	launchTemporarySuffix = ".tmp"
)

type launchProcess interface {
	PID() int
	Wait() error
	Release() error
}

type launchRunner func(executable string, arguments []string, directory string) (launchProcess, error)
type launchObserver func(pid uint32, expectedPath string) (winfs.ProcessObservation, error)

// LaunchConfig contains the current launch preferences and says whether they
// came from the root-level mgs3mod-launch.json file. When Persisted is false,
// Config is the validated in-memory seed and no file has been written.
type LaunchConfig struct {
	Config    launch.Config `json:"config"`
	Persisted bool          `json:"persisted"`
}

// LaunchReceipt records the exact native process request. PID is present only
// after a child was started. A receipt proves preflight or process observation;
// it does not claim that any game screen was reached.
type LaunchReceipt struct {
	Executable string           `json:"executable"`
	WorkingDir string           `json:"workingDir"`
	Arguments  []string         `json:"arguments"`
	Selection  launch.Selection `json:"selection"`
	PID        int              `json:"pid,omitempty"`
	DryRun     bool             `json:"dryRun,omitempty"`
	Message    string           `json:"message"`
}

type execLaunchProcess struct{ process *os.Process }

func (p *execLaunchProcess) PID() int { return p.process.Pid }
func (p *execLaunchProcess) Wait() error {
	_, err := p.process.Wait()
	return err
}
func (p *execLaunchProcess) Release() error { return p.process.Release() }

func startLaunchProcess(executable string, arguments []string, directory string) (launchProcess, error) {
	command := exec.Command(executable, arguments...)
	command.Dir = directory
	if err := command.Start(); err != nil {
		return nil, err
	}
	return &execLaunchProcess{process: command.Process}, nil
}

func (m *Manager) launchSession() (*session, history, error) {
	s, err := m.open(false)
	if err != nil {
		var detail *Error
		if errors.As(err, &detail) && detail.Code != 5 {
			return nil, history{}, err
		}
		return nil, history{}, wrap(5, "launch state unavailable; run mgs3mod init first if this installation is not initialized", err)
	}
	if err = m.compatible(); err != nil {
		s.close()
		return nil, history{}, err
	}
	h, err := s.load()
	if err != nil {
		s.close()
		return nil, history{}, wrap(5, "launch state unavailable; run mgs3mod recover for incomplete initialization", err)
	}
	if h.tail != nil {
		s.close()
		return nil, history{}, fail(5, "unresolved transaction; run recover")
	}
	return s, h, nil
}

// LoadLaunchConfig strictly loads root-level launch preferences while holding
// the manager lock. A missing file returns launch.Seed without creating it.
func (m *Manager) LoadLaunchConfig() (LaunchConfig, error) {
	s, _, err := m.launchSession()
	if err != nil {
		return LaunchConfig{}, err
	}
	defer s.close()
	return s.loadLaunchConfig()
}

func (s *session) loadLaunchConfig() (LaunchConfig, error) {
	_, err := s.root.Lstat(launchConfigPath)
	if errors.Is(err, os.ErrNotExist) {
		if err := s.guard(launchConfigPath, true); err != nil {
			return LaunchConfig{}, wrapPath(5, "unsafe launch configuration path", launchConfigPath, err)
		}
		return LaunchConfig{Config: launch.Seed()}, nil
	}
	if err != nil {
		return LaunchConfig{}, wrapPath(5, "inspect launch configuration", launchConfigPath, err)
	}
	if err := s.guard(launchConfigPath, false); err != nil {
		return LaunchConfig{}, wrapPath(5, "unsafe launch configuration path", launchConfigPath, err)
	}
	f, err := s.root.Open(launchConfigPath)
	if err != nil {
		return LaunchConfig{}, wrapPath(5, "open launch configuration", launchConfigPath, err)
	}
	data, readErr := io.ReadAll(io.LimitReader(f, maxLaunchConfigBytes+1))
	closeErr := f.Close()
	if readErr != nil {
		return LaunchConfig{}, wrapPath(5, "read launch configuration", launchConfigPath, readErr)
	}
	if closeErr != nil {
		return LaunchConfig{}, wrapPath(5, "close launch configuration", launchConfigPath, closeErr)
	}
	if len(data) > maxLaunchConfigBytes {
		return LaunchConfig{}, fail(5, "launch configuration exceeds 1 MiB", launchConfigPath)
	}
	config, err := launch.Parse(data)
	if err != nil {
		return LaunchConfig{}, wrapPath(5, "invalid launch configuration", launchConfigPath, err)
	}
	return LaunchConfig{Config: config, Persisted: true}, nil
}

// SaveLaunchConfig validates and atomically saves a complete launch.Config.
// Existing bytes are parsed first, so corrupt or future files are preserved.
func (m *Manager) SaveLaunchConfig(config launch.Config) error {
	s, _, err := m.launchSession()
	if err != nil {
		return err
	}
	defer s.close()
	current, err := s.loadLaunchConfig()
	if err != nil {
		return err
	}
	data, err := config.Marshal()
	if err != nil {
		return wrap(2, "invalid launch configuration", err)
	}
	return s.saveLaunchConfig(data, current.Persisted)
}

// AddLaunchProfile atomically merges one new profile into the latest persisted
// configuration. It refuses replacement so concurrent selectors cannot erase
// or silently overwrite profiles loaded by another session.
func (m *Manager) AddLaunchProfile(id string, profile launch.Profile) (launch.Config, error) {
	return m.updateLaunchConfig(func(config *launch.Config) error {
		if _, exists := config.Profiles[id]; exists {
			return fail(2, "launch profile "+id+" already exists")
		}
		config.Profiles[id] = profile
		return nil
	})
}

// SetDefaultLaunchProfile atomically selects a profile from the latest
// persisted configuration.
func (m *Manager) SetDefaultLaunchProfile(id string) (launch.Config, error) {
	return m.updateLaunchConfig(func(config *launch.Config) error {
		if _, exists := config.Profiles[id]; !exists {
			return fail(2, "launch profile "+id+" does not exist")
		}
		config.DefaultProfileID = id
		return nil
	})
}

func (m *Manager) updateLaunchConfig(update func(*launch.Config) error) (launch.Config, error) {
	s, _, err := m.launchSession()
	if err != nil {
		return launch.Config{}, err
	}
	defer s.close()
	current, err := s.loadLaunchConfig()
	if err != nil {
		return launch.Config{}, err
	}
	config := current.Config
	if err := update(&config); err != nil {
		return launch.Config{}, err
	}
	data, err := config.Marshal()
	if err != nil {
		return launch.Config{}, wrap(2, "invalid launch configuration", err)
	}
	if err := s.saveLaunchConfig(data, current.Persisted); err != nil {
		return launch.Config{}, err
	}
	return config, nil
}

func (s *session) saveLaunchConfig(data []byte, existed bool) error {
	if len(data) > maxLaunchConfigBytes {
		return fail(2, "launch configuration exceeds 1 MiB", launchConfigPath)
	}
	var temporary string
	for attempts := 0; attempts < 8; attempts++ {
		name, err := newLaunchTemporaryName()
		if err != nil {
			return wrap(1, "create launch configuration temporary name", err)
		}
		if err = s.guard(name, true); err != nil {
			return wrapPath(5, "unsafe launch configuration temporary path", name, err)
		}
		writer, err := transaction.Create(s.root, name, s.m.config.CreateFile)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return wrapPath(1, "create launch configuration temporary file", name, err)
		}
		temporary = name
		n, writeErr := writer.Write(data)
		if writeErr == nil && n != len(data) {
			writeErr = io.ErrShortWrite
		}
		if writeErr == nil {
			writeErr = writer.Sync()
		}
		closeErr := writer.Close()
		if writeErr == nil {
			writeErr = closeErr
		}
		if writeErr != nil {
			s.cleanupLaunchTemporary(temporary, data[:boundedWriteCount(n, len(data))])
			return wrapPath(1, "write launch configuration temporary file", temporary, writeErr)
		}
		break
	}
	if temporary == "" {
		return fail(1, "cannot allocate unique launch configuration temporary file")
	}
	if err := s.expect(temporary, transaction.Hash(data)); err != nil {
		s.cleanupLaunchTemporary(temporary, data)
		return wrapPath(1, "verify launch configuration temporary file", temporary, err)
	}

	fullTemporary := filepath.Join(s.m.config.Root, temporary)
	fullFinal := filepath.Join(s.m.config.Root, launchConfigPath)
	if existed {
		if err := s.guard(launchConfigPath, false); err != nil {
			s.cleanupLaunchTemporary(temporary, data)
			return wrapPath(5, "unsafe launch configuration path", launchConfigPath, err)
		}
		replace := s.m.config.replaceLaunch
		if replace == nil {
			replace = winfs.ReplaceExisting
		}
		if err := replace(fullTemporary, fullFinal); err != nil {
			s.cleanupLaunchTemporary(temporary, data)
			return wrapPath(1, "replace launch configuration", launchConfigPath, err)
		}
	} else {
		if err := s.guard(launchConfigPath, true); err != nil {
			s.cleanupLaunchTemporary(temporary, data)
			return wrapPath(5, "unsafe launch configuration path", launchConfigPath, err)
		}
		move := s.m.config.moveLaunch
		if move == nil {
			move = winfs.MoveNew
		}
		if err := move(fullTemporary, fullFinal); err != nil {
			s.cleanupLaunchTemporary(temporary, data)
			return wrapPath(1, "create launch configuration", launchConfigPath, err)
		}
	}

	loaded, err := s.loadLaunchConfig()
	if err != nil {
		return wrap(5, "verify promoted launch configuration", err)
	}
	if !loaded.Persisted {
		return fail(5, "promoted launch configuration is missing", launchConfigPath)
	}
	verified, err := loaded.Config.Marshal()
	if err != nil || !bytes.Equal(verified, data) {
		return fail(5, "promoted launch configuration differs", launchConfigPath)
	}
	return nil
}

func boundedWriteCount(n, length int) int {
	if n < 0 {
		return 0
	}
	if n > length {
		return length
	}
	return n
}

func (s *session) cleanupLaunchTemporary(name string, written []byte) {
	_ = s.removeExact(inventory{Path: name, SHA256: transaction.Hash(written), Bytes: int64(len(written))})
}

func newLaunchTemporaryName() (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return launchTemporaryPrefix + hex.EncodeToString(random[:]) + launchTemporarySuffix, nil
}

// waitForObservedLaunchExit keeps the manager lock held when the process
// handle cannot confirm exit. It returns the wait failure only after the PID
// observer independently reports that the process is gone.
func waitForObservedLaunchExit(process launchProcess, observer launchObserver, pid uint32, expectedPath string, poll time.Duration) error {
	waitErr := process.Wait()
	if waitErr == nil {
		return nil
	}
	for {
		observation, _ := observer(pid, expectedPath)
		if observation.State == winfs.ProcessExited {
			return waitErr
		}
		time.Sleep(poll)
	}
}

// Launch performs installation-bound preflight and optionally starts one
// resolved selection. It never reads or writes saved launch preferences.
func (m *Manager) Launch(selection launch.Selection, dryRun bool) (LaunchReceipt, error) {
	arguments, err := launch.BuildArguments(selection)
	receipt := LaunchReceipt{
		Executable: filepath.Join(m.config.Root, launchExecutable),
		WorkingDir: m.config.Root,
		Arguments:  append([]string(nil), arguments...),
		Selection:  selection,
		DryRun:     dryRun,
		Message:    "launch preflight failed",
	}
	if err != nil {
		return receipt, wrap(2, "invalid launch selection", err)
	}
	s, _, err := m.launchSession()
	if err != nil {
		return receipt, err
	}
	defer s.close()
	if err = s.guard(launchExecutable, false); err != nil {
		return receipt, wrapPath(3, "unsafe or missing game executable", launchExecutable, err)
	}
	if err = s.guard(launchLauncher, false); err != nil {
		return receipt, wrapPath(3, "unsafe or missing launcher executable", launchLauncher, err)
	}
	// Keep this guard immediately before dry-run completion or process start.
	if err = m.processes(); err != nil {
		return receipt, err
	}
	if dryRun {
		receipt.Message = "launch preflight complete; transient manager lock released"
		return receipt, nil
	}

	runner := m.config.runLaunch
	if runner == nil {
		runner = startLaunchProcess
	}
	process, err := runner(receipt.Executable, append([]string(nil), arguments...), receipt.WorkingDir)
	if err != nil {
		receipt.Message = "game process start failed"
		return receipt, wrap(1, receipt.Message, err)
	}
	pid := process.PID()
	if pid <= 0 || uint64(pid) > math.MaxUint32 {
		receipt.Message = "game process returned invalid PID"
		if waitErr := process.Wait(); waitErr != nil {
			return receipt, wrap(1, receipt.Message+"; child exit could not be confirmed", waitErr)
		}
		receipt.Message += "; launch protection remained active until child exit"
		return receipt, fail(1, receipt.Message)
	}
	receipt.PID = pid

	observer := m.config.observeLaunch
	if observer == nil {
		observer = winfs.ObserveProcess
	}
	timeout := m.config.launchTimeout
	if timeout <= 0 {
		timeout = defaultLaunchTimeout
	}
	poll := m.config.launchPoll
	if poll <= 0 {
		poll = defaultLaunchPoll
	}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		observation, observeErr := observer(uint32(pid), receipt.Executable)
		switch observation.State {
		case winfs.ProcessAliveExpected:
			if observeErr != nil {
				lastErr = observeErr
				break
			}
			if err := process.Release(); err != nil {
				receipt.Message = "game process verified but handle release failed; launch protection remained active until child exit"
				if waitErr := waitForObservedLaunchExit(process, observer, uint32(pid), receipt.Executable, poll); waitErr != nil {
					return receipt, wrap(1, receipt.Message+"; child exit confirmed by process observer after wait failure", errors.Join(err, waitErr))
				}
				return receipt, wrap(1, receipt.Message, err)
			}
			receipt.Message = "game process started and executable verified"
			return receipt, nil
		case winfs.ProcessExited:
			_ = process.Wait()
			receipt.Message = "game process exited before launch verification"
			return receipt, fail(1, receipt.Message)
		case winfs.ProcessWrongImage:
			receipt.Message = "started PID belongs to a different executable; launch protection remained active until child exit"
			if waitErr := waitForObservedLaunchExit(process, observer, uint32(pid), receipt.Executable, poll); waitErr != nil {
				return receipt, wrap(1, receipt.Message+"; child exit confirmed by process observer after wait failure", waitErr)
			}
			return receipt, fail(1, receipt.Message, observation.ImagePath)
		case winfs.ProcessInaccessible:
			receipt.Message = "started process is inaccessible during launch verification; launch protection remained active until child exit"
			if waitErr := waitForObservedLaunchExit(process, observer, uint32(pid), receipt.Executable, poll); waitErr != nil {
				if observeErr != nil {
					waitErr = errors.Join(observeErr, waitErr)
				}
				return receipt, wrap(1, receipt.Message+"; child exit confirmed by process observer after wait failure", waitErr)
			}
			if observeErr != nil {
				return receipt, wrap(1, receipt.Message, observeErr)
			}
			return receipt, fail(1, receipt.Message)
		case winfs.ProcessNotObservable:
			lastErr = observeErr
		default:
			receipt.Message = "process observer returned invalid state; launch protection remained active until child exit"
			if waitErr := waitForObservedLaunchExit(process, observer, uint32(pid), receipt.Executable, poll); waitErr != nil {
				return receipt, wrap(1, receipt.Message+"; child exit confirmed by process observer after wait failure", waitErr)
			}
			return receipt, fail(1, receipt.Message)
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if poll > remaining {
			time.Sleep(remaining)
		} else {
			time.Sleep(poll)
		}
	}
	receipt.Message = "game process observation timed out; launch protection remained active until child exit"
	if waitErr := waitForObservedLaunchExit(process, observer, uint32(pid), receipt.Executable, poll); waitErr != nil {
		if lastErr != nil {
			waitErr = errors.Join(lastErr, waitErr)
		}
		return receipt, wrap(1, receipt.Message+"; child exit confirmed by process observer after wait failure", waitErr)
	}
	if lastErr != nil {
		return receipt, wrap(1, receipt.Message, lastErr)
	}
	return receipt, fail(1, receipt.Message)
}
