package manager

import (
	"errors"
	"fmt"
	"mgs3mod/internal/launch"
	"mgs3mod/internal/transaction"
	"mgs3mod/internal/winfs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func launchFixture(t *testing.T) (*Manager, string, launch.Selection) {
	t.Helper()
	m, root := fixture(t)
	put(t, root, launchExecutable, []byte("synthetic executable; never started"))
	put(t, root, launchLauncher, []byte("synthetic launcher; never started"))
	invoke(t, m, "init", "")
	return m, root, launch.Seed().Profiles[launch.SeedProfileID].Selection
}

func alternateLaunchConfig() launch.Config {
	c := launch.Seed()
	c.DefaultProfileID = "alternate"
	c.Profiles["alternate"] = c.Profiles[launch.SeedProfileID]
	return c
}

func TestLoadLaunchConfigAbsentReturnsSeedWithoutWrite(t *testing.T) {
	m, root, _ := launchFixture(t)
	before := snapshot(t, root)
	got, err := m.LoadLaunchConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Persisted || !reflect.DeepEqual(got.Config, launch.Seed()) {
		t.Fatalf("LoadLaunchConfig() = %#v", got)
	}
	if !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Fatal("absent load wrote files")
	}
}

func TestLoadLaunchConfigRejectsStrictParseFailures(t *testing.T) {
	valid, err := launch.Seed().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"corrupt":   `{`,
		"unknown":   strings.TrimSuffix(string(valid), "}") + `,"extra":1}`,
		"trailing":  string(valid) + ` {}`,
		"duplicate": strings.Replace(string(valid), `"schema":1`, `"schema":1,"schema":1`, 1),
		"future":    strings.Replace(string(valid), `"schema":1`, `"schema":2`, 1),
		"stale":     strings.Replace(string(valid), `"na-startup"`, `"missing"`, 1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			m, root, _ := launchFixture(t)
			put(t, root, launchConfigPath, []byte(data))
			before := snapshot(t, root)
			if _, err := m.LoadLaunchConfig(); err == nil {
				t.Fatal("invalid configuration accepted")
			}
			if err := m.SaveLaunchConfig(launch.Seed()); err == nil {
				t.Fatal("invalid existing configuration overwritten")
			}
			if !reflect.DeepEqual(before, snapshot(t, root)) {
				t.Fatal("invalid configuration changed")
			}
		})
	}
}

func TestSaveLaunchConfigAtomicAndGenerationIndependent(t *testing.T) {
	m, root, _ := launchFixture(t)
	before := invoke(t, m, "list", "").State.Generation
	want := alternateLaunchConfig()
	if err := m.SaveLaunchConfig(want); err != nil {
		t.Fatal(err)
	}
	after := invoke(t, m, "list", "").State.Generation
	if after != before {
		t.Fatalf("generation changed from %d to %d", before, after)
	}
	got, err := m.LoadLaunchConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Persisted || !reflect.DeepEqual(got.Config, want) {
		t.Fatalf("saved configuration = %#v", got)
	}
	data, err := os.ReadFile(filepath.Join(root, launchConfigPath))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := launch.Parse(data)
	if err != nil || !reflect.DeepEqual(parsed, want) {
		t.Fatalf("promoted document incomplete: %#v, %v", parsed, err)
	}
}

func TestLaunchProfileUpdatesMergeLatestConfiguration(t *testing.T) {
	m, root, selection := launchFixture(t)
	other, err := m.WithRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.LoadLaunchConfig(); err != nil {
		t.Fatal(err)
	}
	if _, err = other.LoadLaunchConfig(); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AddLaunchProfile("first", launch.Profile{Selection: selection}); err != nil {
		t.Fatal(err)
	}
	if _, err = other.AddLaunchProfile("second", launch.Profile{Selection: selection}); err != nil {
		t.Fatal(err)
	}
	if _, err = other.AddLaunchProfile("first", launch.Profile{Selection: selection}); ExitCode(err) != 2 {
		t.Fatalf("duplicate profile error = %v", err)
	}
	if _, err = m.SetDefaultLaunchProfile("second"); err != nil {
		t.Fatal(err)
	}
	loaded, err := other.LoadLaunchConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Config.DefaultProfileID != "second" {
		t.Fatalf("default profile = %q", loaded.Config.DefaultProfileID)
	}
	for _, id := range []string{launch.SeedProfileID, "first", "second"} {
		if _, err := loaded.Config.Profile(id); err != nil {
			t.Fatalf("merged profile %q missing: %v", id, err)
		}
	}
}

func TestSaveLaunchConfigRejectsOversizeBeforePromotion(t *testing.T) {
	oversize := launch.Seed()
	selection := oversize.Profiles[launch.SeedProfileID].Selection
	for i := 0; i < 20000; i++ {
		oversize.Profiles[fmt.Sprintf("profile-%05d", i)] = launch.Profile{Selection: selection}
	}
	data, err := oversize.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) <= maxLaunchConfigBytes {
		t.Fatalf("fixture is only %d bytes", len(data))
	}

	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprintf("existing=%t", existing), func(t *testing.T) {
			m, root, _ := launchFixture(t)
			if existing {
				if err := m.SaveLaunchConfig(launch.Seed()); err != nil {
					t.Fatal(err)
				}
			}
			before := snapshot(t, root)
			creates := 0
			m.config.CreateFile = func(*os.Root, string) (transaction.SyncWriter, error) {
				creates++
				return nil, errors.New("must reject before creating a temporary file")
			}
			if err := m.SaveLaunchConfig(oversize); ExitCode(err) != 2 {
				t.Fatalf("oversize error = %v", err)
			}
			if creates != 0 || !reflect.DeepEqual(before, snapshot(t, root)) {
				t.Fatalf("oversize save created=%d or changed files", creates)
			}
		})
	}
}

func TestSaveLaunchConfigFailuresPreserveDestination(t *testing.T) {
	t.Run("replacement", func(t *testing.T) {
		m, root, _ := launchFixture(t)
		if err := m.SaveLaunchConfig(launch.Seed()); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, launchConfigPath)
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m.config.replaceLaunch = func(string, string) error { return errors.New("synthetic replace failure") }
		if err := m.SaveLaunchConfig(alternateLaunchConfig()); err == nil {
			t.Fatal("replacement failure accepted")
		}
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(before) {
			t.Fatalf("existing destination changed: %q, %v", after, err)
		}
	})

	t.Run("create", func(t *testing.T) {
		m, root, _ := launchFixture(t)
		m.config.CreateFile = func(*os.Root, string) (transaction.SyncWriter, error) {
			return nil, errors.New("synthetic create failure")
		}
		if err := m.SaveLaunchConfig(launch.Seed()); err == nil {
			t.Fatal("create failure accepted")
		}
		if _, err := os.Lstat(filepath.Join(root, launchConfigPath)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("destination appeared: %v", err)
		}
	})

	t.Run("sync", func(t *testing.T) {
		m, root, _ := launchFixture(t)
		before := snapshot(t, root)
		m.config.CreateFile = func(root *os.Root, path string) (transaction.SyncWriter, error) {
			writer, err := root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				return nil, err
			}
			return &failingSyncWriter{SyncWriter: writer}, nil
		}
		if err := m.SaveLaunchConfig(launch.Seed()); err == nil {
			t.Fatal("synchronization failure accepted")
		}
		if !reflect.DeepEqual(before, snapshot(t, root)) {
			t.Fatal("synchronization failure left files behind")
		}
	})

	t.Run("foreign destination race", func(t *testing.T) {
		m, root, _ := launchFixture(t)
		foreign := []byte("foreign")
		m.config.moveLaunch = func(_, target string) error {
			if err := os.WriteFile(target, foreign, 0600); err != nil {
				t.Fatal(err)
			}
			return errors.New("synthetic destination collision")
		}
		if err := m.SaveLaunchConfig(launch.Seed()); err == nil {
			t.Fatal("destination collision accepted")
		}
		got, err := os.ReadFile(filepath.Join(root, launchConfigPath))
		if err != nil || string(got) != string(foreign) {
			t.Fatalf("foreign destination changed: %q, %v", got, err)
		}
	})

	t.Run("interruption after replacement", func(t *testing.T) {
		m, _, _ := launchFixture(t)
		if err := m.SaveLaunchConfig(launch.Seed()); err != nil {
			t.Fatal(err)
		}
		want := alternateLaunchConfig()
		m.config.replaceLaunch = func(source, target string) error {
			if err := winfs.ReplaceExisting(source, target); err != nil {
				return err
			}
			return errors.New("synthetic interruption after replacement")
		}
		if err := m.SaveLaunchConfig(want); err == nil {
			t.Fatal("post-replacement interruption accepted")
		}
		got, err := m.LoadLaunchConfig()
		if err != nil || !got.Persisted || !reflect.DeepEqual(got.Config, want) {
			t.Fatalf("promoted configuration is not complete: %#v, %v", got, err)
		}
	})
}

type failingSyncWriter struct {
	transaction.SyncWriter
}

func (*failingSyncWriter) Sync() error { return errors.New("synthetic synchronization failure") }

type fakeLaunchProcess struct {
	pid        int
	waits      int
	releases   int
	waitErr    error
	releaseErr error
}

func (p *fakeLaunchProcess) PID() int { return p.pid }
func (p *fakeLaunchProcess) Wait() error {
	p.waits++
	return p.waitErr
}
func (p *fakeLaunchProcess) Release() error {
	p.releases++
	return p.releaseErr
}

func TestLaunchDryRunExactReceiptAndNoSideEffects(t *testing.T) {
	m, root, selection := launchFixture(t)
	starts := 0
	m.config.runLaunch = func(string, []string, string) (launchProcess, error) {
		starts++
		return nil, errors.New("must not start")
	}
	before := snapshot(t, root)
	receipt, err := m.Launch(selection, true)
	if err != nil {
		t.Fatal(err)
	}
	wantArguments := []string{"-region", "us", "-lan", "en", "-selfregion", "EU", "-launcherpath", "launcher.exe", "-ctrltype", "KBD"}
	if receipt.Executable != filepath.Join(root, launchExecutable) || receipt.WorkingDir != root || !reflect.DeepEqual(receipt.Arguments, wantArguments) || receipt.Selection != selection || !receipt.DryRun || receipt.PID != 0 {
		t.Fatalf("wrong dry-run receipt: %#v", receipt)
	}
	if starts != 0 || !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Fatalf("dry-run started=%d or wrote files", starts)
	}
}

func TestLaunchPreflightGuards(t *testing.T) {
	t.Run("missing init", func(t *testing.T) {
		m, root := fixture(t)
		put(t, root, launchExecutable, []byte("synthetic"))
		if _, err := m.Launch(launch.Seed().Profiles[launch.SeedProfileID].Selection, true); err == nil || !strings.Contains(err.Error(), "mgs3mod init") {
			t.Fatalf("missing initialization error = %v", err)
		}
	})

	t.Run("incomplete init", func(t *testing.T) {
		m, root := fixture(t)
		put(t, root, launchExecutable, []byte("synthetic"))
		m.config.Fault = func(point string) error {
			if point == "init-prepared" {
				return errors.New("synthetic interruption")
			}
			return nil
		}
		if _, err := m.Run("init", "", Options{}); err == nil {
			t.Fatal("fault did not interrupt initialization")
		}
		m.config.Fault = nil
		if _, err := m.Launch(launch.Seed().Profiles[launch.SeedProfileID].Selection, true); err == nil || !strings.Contains(err.Error(), "incomplete") || !strings.Contains(err.Error(), "mgs3mod recover") {
			t.Fatalf("incomplete initialization error = %v", err)
		}
	})

	t.Run("wrong fingerprint", func(t *testing.T) {
		m, root, selection := launchFixture(t)
		put(t, root, "game.exe", []byte("changed"))
		if _, err := m.Launch(selection, true); ExitCode(err) != 3 {
			t.Fatalf("wrong fingerprint error = %v", err)
		}
	})

	t.Run("unresolved recovery", func(t *testing.T) {
		m, root, selection := launchFixture(t)
		pkg := makePackage(t, root, "pending", targetA)
		m.config.Fault = func(point string) error {
			if point == "prepared" {
				return errors.New("synthetic interruption")
			}
			return nil
		}
		if _, err := m.Run("add", pkg, Options{}); err == nil {
			t.Fatal("fault did not interrupt operation")
		}
		m.config.Fault = nil
		if _, err := m.Launch(selection, true); err == nil || !strings.Contains(err.Error(), "unresolved transaction") {
			t.Fatalf("unresolved recovery error = %v", err)
		}
	})

	t.Run("process guard", func(t *testing.T) {
		m, _, selection := launchFixture(t)
		starts := 0
		m.config.CheckProcesses = func(string) error { return errors.New("synthetic running game") }
		m.config.runLaunch = func(string, []string, string) (launchProcess, error) {
			starts++
			return nil, nil
		}
		if _, err := m.Launch(selection, false); err == nil || starts != 0 {
			t.Fatalf("process guard error=%v starts=%d", err, starts)
		}
	})

	t.Run("missing executable", func(t *testing.T) {
		m, root, selection := launchFixture(t)
		if err := os.Remove(filepath.Join(root, launchExecutable)); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Launch(selection, true); err == nil {
			t.Fatal("missing executable accepted")
		}
	})

	t.Run("missing launcher", func(t *testing.T) {
		m, root, selection := launchFixture(t)
		if err := os.Remove(filepath.Join(root, launchLauncher)); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Launch(selection, true); err == nil || !strings.Contains(err.Error(), launchLauncher) {
			t.Fatalf("missing launcher error = %v", err)
		}
	})
}

func TestLaunchProcessOutcomes(t *testing.T) {
	t.Run("start failure", func(t *testing.T) {
		m, _, selection := launchFixture(t)
		m.config.runLaunch = func(string, []string, string) (launchProcess, error) {
			return nil, errors.New("synthetic start failure")
		}
		receipt, err := m.Launch(selection, false)
		if err == nil || receipt.PID != 0 || receipt.Message != "game process start failed" {
			t.Fatalf("receipt=%#v error=%v", receipt, err)
		}
	})

	t.Run("delayed visibility", func(t *testing.T) {
		m, root, selection := launchFixture(t)
		process := &fakeLaunchProcess{pid: 4242}
		starts, observations := 0, 0
		m.config.runLaunch = func(executable string, arguments []string, directory string) (launchProcess, error) {
			starts++
			want, _ := launch.BuildArguments(selection)
			if executable != filepath.Join(root, launchExecutable) || directory != root || !reflect.DeepEqual(arguments, want) {
				t.Fatalf("runner got %q %#v %q", executable, arguments, directory)
			}
			return process, nil
		}
		m.config.observeLaunch = func(pid uint32, expected string) (winfs.ProcessObservation, error) {
			observations++
			if pid != 4242 || expected != filepath.Join(root, launchExecutable) {
				t.Fatalf("observer got %d %q", pid, expected)
			}
			if observations < 3 {
				return winfs.ProcessObservation{PID: pid, State: winfs.ProcessNotObservable}, errors.New("synthetic visibility delay")
			}
			return winfs.ProcessObservation{PID: pid, State: winfs.ProcessAliveExpected, ImagePath: expected}, nil
		}
		m.config.launchTimeout = 100 * time.Millisecond
		m.config.launchPoll = time.Millisecond
		receipt, err := m.Launch(selection, false)
		if err != nil || receipt.PID != 4242 || starts != 1 || observations != 3 || process.releases != 1 || process.waits != 0 {
			t.Fatalf("receipt=%#v error=%v starts=%d observations=%d process=%#v", receipt, err, starts, observations, process)
		}
	})

	for name, state := range map[string]winfs.ProcessObservationState{
		"immediate exit": winfs.ProcessExited,
		"wrong image":    winfs.ProcessWrongImage,
	} {
		t.Run(name, func(t *testing.T) {
			m, _, selection := launchFixture(t)
			process := &fakeLaunchProcess{pid: 4343}
			starts := 0
			m.config.runLaunch = func(string, []string, string) (launchProcess, error) {
				starts++
				return process, nil
			}
			m.config.observeLaunch = func(pid uint32, _ string) (winfs.ProcessObservation, error) {
				return winfs.ProcessObservation{PID: pid, State: state, ImagePath: `C:\\wrong.exe`}, nil
			}
			if _, err := m.Launch(selection, false); err == nil || starts != 1 || process.waits != 1 || process.releases != 0 {
				t.Fatalf("error=%v starts=%d process=%#v", err, starts, process)
			}
		})
	}

	t.Run("timeout no duplicate", func(t *testing.T) {
		m, _, selection := launchFixture(t)
		process := &fakeLaunchProcess{pid: 4444}
		starts := 0
		m.config.runLaunch = func(string, []string, string) (launchProcess, error) {
			starts++
			return process, nil
		}
		m.config.observeLaunch = func(pid uint32, _ string) (winfs.ProcessObservation, error) {
			return winfs.ProcessObservation{PID: pid, State: winfs.ProcessNotObservable}, nil
		}
		m.config.launchTimeout = 3 * time.Millisecond
		m.config.launchPoll = time.Millisecond
		if _, err := m.Launch(selection, false); err == nil || starts != 1 || process.waits != 1 || process.releases != 0 {
			t.Fatalf("error=%v starts=%d process=%#v", err, starts, process)
		}
	})
}

func TestLaunchRejectsCompetingManagerLock(t *testing.T) {
	m, _, selection := launchFixture(t)
	held, err := m.open(false)
	if err != nil {
		t.Fatal(err)
	}
	defer held.close()
	if _, err := m.Launch(selection, true); ExitCode(err) != 6 {
		t.Fatalf("competing lock error = %v", err)
	}
}

type blockingLaunchProcess struct {
	pid     int
	waiting chan struct{}
	exit    chan struct{}
}

func (p *blockingLaunchProcess) PID() int { return p.pid }
func (p *blockingLaunchProcess) Wait() error {
	close(p.waiting)
	<-p.exit
	return nil
}
func (p *blockingLaunchProcess) Release() error { return nil }

func TestLaunchObservationFailureHoldsLockUntilChildExit(t *testing.T) {
	m, _, selection := launchFixture(t)
	process := &blockingLaunchProcess{pid: 4545, waiting: make(chan struct{}), exit: make(chan struct{})}
	m.config.runLaunch = func(string, []string, string) (launchProcess, error) { return process, nil }
	m.config.observeLaunch = func(pid uint32, _ string) (winfs.ProcessObservation, error) {
		return winfs.ProcessObservation{PID: pid, State: winfs.ProcessWrongImage, ImagePath: `C:\\wrong.exe`}, nil
	}
	done := make(chan error, 1)
	go func() {
		_, err := m.Launch(selection, false)
		done <- err
	}()
	select {
	case <-process.waiting:
	case <-time.After(time.Second):
		t.Fatal("launch did not wait for child exit")
	}
	if _, err := m.LoadLaunchConfig(); ExitCode(err) != 6 {
		t.Fatalf("manager lock released before child exit: %v", err)
	}
	close(process.exit)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("observation failure accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("launch did not return after child exit")
	}
	if _, err := m.LoadLaunchConfig(); err != nil {
		t.Fatalf("manager lock retained after child exit: %v", err)
	}
}

func TestLaunchWaitFailureHoldsLockUntilObserverConfirmsExit(t *testing.T) {
	m, _, selection := launchFixture(t)
	process := &fakeLaunchProcess{pid: 4646, waitErr: errors.New("synthetic wait failure")}
	allowExit := make(chan struct{})
	guarding := make(chan struct{})
	announced := false
	observations := 0
	m.config.runLaunch = func(string, []string, string) (launchProcess, error) { return process, nil }
	m.config.observeLaunch = func(pid uint32, _ string) (winfs.ProcessObservation, error) {
		observations++
		if observations == 1 {
			return winfs.ProcessObservation{PID: pid, State: winfs.ProcessWrongImage, ImagePath: `C:\wrong.exe`}, nil
		}
		select {
		case <-allowExit:
			return winfs.ProcessObservation{PID: pid, State: winfs.ProcessExited}, nil
		default:
			if !announced {
				close(guarding)
				announced = true
			}
			return winfs.ProcessObservation{PID: pid, State: winfs.ProcessNotObservable}, nil
		}
	}
	m.config.launchPoll = time.Millisecond
	done := make(chan error, 1)
	go func() {
		_, err := m.Launch(selection, false)
		done <- err
	}()
	select {
	case <-guarding:
	case <-time.After(time.Second):
		t.Fatal("launch did not retain protection after wait failure")
	}
	if _, err := m.LoadLaunchConfig(); ExitCode(err) != 6 {
		t.Fatalf("manager lock released before observer confirmed exit: %v", err)
	}
	close(allowExit)
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "wait failure") {
			t.Fatalf("wait failure result = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("launch did not return after observer confirmed exit")
	}
	if _, err := m.LoadLaunchConfig(); err != nil {
		t.Fatalf("manager lock retained after observed exit: %v", err)
	}
}

func TestLaunchConfigGuardRejectsHardLink(t *testing.T) {
	m, root, _ := launchFixture(t)
	original := filepath.Join(root, "foreign.json")
	put(t, root, "foreign.json", []byte(`{}`))
	if err := os.Link(original, filepath.Join(root, launchConfigPath)); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}
	if _, err := m.LoadLaunchConfig(); err == nil {
		t.Fatal("hard-linked launch configuration accepted")
	}
}
