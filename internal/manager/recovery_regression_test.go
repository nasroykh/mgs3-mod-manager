package manager

import (
	"errors"
	"fmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func rewriteManagerRecord[T any](t *testing.T, root, name string, mutate func(*T)) {
	t.Helper()
	fs, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()
	var record T
	if err := transaction.Read(fs, name, &record); err != nil {
		t.Fatal(err)
	}
	mutate(&record)
	if err := fs.Remove(name); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Write(fs, name, record); err != nil {
		t.Fatal(err)
	}
}

func requireManagerExit(t *testing.T, m *Manager, command, arg string, want int, opt Options) {
	t.Helper()
	_, err := m.Run(command, arg, opt)
	if got := ExitCode(err); got != want {
		t.Fatalf("%s %s exit code = %d, want %d (err %v)", command, arg, got, want, err)
	}
}

func TestRecoveryJournalTornAndRecomputedGeneration(t *testing.T) {
	cases := []struct {
		name  string
		point string
		file  string
		edit  func(*testing.T, string)
	}{
		{
			name:  "prepared-torn",
			point: "prepared",
			file:  txPath(1) + "/PREPARED",
			edit: func(t *testing.T, root string) {
				put(t, root, txPath(1)+"/PREPARED", []byte("torn"))
			},
		},
		{
			name:  "ready-torn",
			point: "ready",
			file:  txPath(1) + "/READY",
			edit: func(t *testing.T, root string) {
				put(t, root, txPath(1)+"/READY", []byte("torn"))
			},
		},
		{
			name:  "committed-torn",
			point: "committed",
			file:  txPath(1) + "/COMMITTED",
			edit: func(t *testing.T, root string) {
				put(t, root, txPath(1)+"/COMMITTED", []byte("torn"))
			},
		},
		{
			name:  "prepared-recomputed-invalid-generation",
			point: "prepared",
			file:  txPath(1) + "/PREPARED",
			edit: func(t *testing.T, root string) {
				rewriteManagerRecord[plan](t, root, txPath(1)+"/PREPARED", func(p *plan) {
					p.After.Generation++
				})
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			killAt(t, m, "enable", "one", tc.point)
			tc.edit(t, root)
			before := snapshot(t, root)
			requireManagerExit(t, m, "recover", "", 5, Options{})
			if !equal(before, snapshot(t, root)) {
				t.Fatalf("invalid %s evidence changed fixture", tc.file)
			}
		})
	}
}

func TestInterruptedInitializationRejectsMalformedOwnership(t *testing.T) {
	cases := []struct {
		name string
		edit func(*testing.T, string)
	}{
		{
			name: "torn-prepared-record",
			edit: func(t *testing.T, root string) {
				put(t, root, stateDir+"/INIT_PREPARED", []byte("torn"))
			},
		},
		{
			name: "recomputed-identity-mismatch",
			edit: func(t *testing.T, root string) {
				rewriteManagerRecord[initRecord](t, root, stateDir+"/INIT_PREPARED", func(r *initRecord) {
					r.Profile = "wrong-profile"
				})
			},
		},
		{
			name: "unexpected-owned-entry",
			edit: func(t *testing.T, root string) {
				put(t, root, stateDir+"/unexpected", []byte("foreign"))
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, root := fixture(t)
			killAt(t, m, "init", "", "init-prepared")
			tc.edit(t, root)
			before := snapshot(t, root)
			requireManagerExit(t, m, "recover", "", 5, Options{})
			if !equal(before, snapshot(t, root)) {
				t.Fatal("malformed incomplete initialization changed fixture")
			}
		})
	}
}

func TestMissingBeforeApplyBlocksWithoutWriting(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(targetA))); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, root)
	requireManagerExit(t, m, "enable", "one", 4, Options{})
	if !equal(before, snapshot(t, root)) {
		t.Fatal("missing target before apply created recovery state or wrote files")
	}
}

func TestMultipleMissingTargetsRequireAllFlagsBeforeWrites(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA, targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "intent-1")
	for _, target := range []string{targetA, targetB} {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(target))); err != nil {
			t.Fatal(err)
		}
	}
	before := snapshot(t, root)
	requireManagerExit(t, m, "recover", "", 5, Options{RestoreMissing: []string{targetB}})
	if !equal(before, snapshot(t, root)) {
		t.Fatal("partial missing-file authorization wrote before validating all targets")
	}
	r, err := m.Run("recover", "", Options{RestoreMissing: []string{targetA, targetB}})
	if err != nil {
		t.Fatal(err)
	}
	if !r.MissingFileRecoveryExplicit || len(r.Paths) != 2 {
		t.Fatalf("missing recovery result = %+v", r)
	}
	assertBytes(t, root, targetA, "original:"+targetA)
	assertBytes(t, root, targetB, "original:"+targetB)
	invoke(t, m, "verify", "")
}

func TestMissingRecoveryFlagErrorCategories(t *testing.T) {
	cases := []struct {
		name      string
		flag      string
		remove    bool
		coreDrift bool
		wantCode  int
	}{
		{name: "unlisted-target", flag: "textures/flatlist/_win/unlisted.ctxr", wantCode: 5},
		{name: "existing-target", flag: targetB, wantCode: 4},
		{name: "core-mismatch", flag: targetB, remove: true, coreDrift: true, wantCode: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA, targetB)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			killAt(t, m, "enable", "one", "intent-0")
			if tc.remove {
				if err := os.Remove(filepath.Join(root, filepath.FromSlash(targetB))); err != nil {
					t.Fatal(err)
				}
			}
			if tc.coreDrift {
				put(t, root, "game.exe", []byte("updated"))
			}
			before := snapshot(t, root)
			requireManagerExit(t, m, "recover", "", tc.wantCode, Options{RestoreMissing: []string{tc.flag}})
			if !equal(before, snapshot(t, root)) {
				t.Fatal("invalid missing-file authorization changed fixture")
			}
		})
	}
}

func TestUnexpectedCleanupBytesRemainPending(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "one")
	killAt(t, m, "remove", "one", "committed")
	cleanupPath := stateDir + "/library/one/payload/0.ctxr"
	put(t, root, cleanupPath, []byte("unexpected"))
	before := snapshot(t, root)
	requireManagerExit(t, m, "recover", "", 4, Options{})
	if !equal(before, snapshot(t, root)) {
		t.Fatal("unexpected cleanup bytes were deleted or state was refreshed")
	}
	assertBytes(t, root, cleanupPath, "unexpected")
	put(t, root, cleanupPath, []byte("mod:one:"+targetA))
	invoke(t, m, "recover", "")
	invoke(t, m, "verify", "")
}

func TestTornMissingAuthorizationAfterExplicitRollback(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA, targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "intent-0")
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(targetB))); err != nil {
		t.Fatal(err)
	}
	_, err := m.Run("recover", "", Options{RestoreMissing: []string{targetB}})
	if err != nil {
		t.Fatal(err)
	}
	authPath := txPath(1) + "/MISSING_AUTHORIZED"
	put(t, root, authPath, []byte("torn"))
	before := snapshot(t, root)
	requireManagerExit(t, m, "recover", "", 5, Options{})
	if !equal(before, snapshot(t, root)) {
		t.Fatal("torn authorization record changed terminal recovery evidence")
	}
}

func TestInjectedStorageFaultsRequireRecovery(t *testing.T) {
	cases := []struct {
		name      string
		point     string
		errorText string
		wantAfter bool
	}{
		{name: "injected-disk-full-before-apply", point: "ready", errorText: "injected disk-full", wantAfter: false},
		{name: "injected-denied-after-commit", point: "committed", errorText: "injected access-denied", wantAfter: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			m.config.Fault = func(point string) error {
				if point == tc.point {
					return errors.New(tc.errorText)
				}
				return nil
			}
			_, err := m.Run("enable", "one", Options{})
			m.config.Fault = nil
			if ExitCode(err) != 5 || err == nil || !containsError(err, tc.errorText) {
				t.Fatalf("injected %s returned code %d, err %v", tc.errorText, ExitCode(err), err)
			}
			if tc.wantAfter {
				assertBytes(t, root, targetA, "mod:one:"+targetA)
			} else {
				assertBytes(t, root, targetA, "original:"+targetA)
			}
			invoke(t, m, "recover", "")
			invoke(t, m, "verify", "")
		})
	}
}

func containsError(err error, want string) bool {
	return err != nil && (err.Error() == want || len(err.Error()) > len(want) && stringContains(err.Error(), want))
}

func stringContains(s, want string) bool {
	for i := 0; i+len(want) <= len(s); i++ {
		if s[i:i+len(want)] == want {
			return true
		}
	}
	return false
}

func TestConcurrentInitChild(t *testing.T) {
	if os.Getenv("MGS3_INIT_CHILD") != "1" {
		return
	}
	root := os.Getenv("MGS3_TEST_ROOT")
	ready := filepath.Join(root, "init-child-ready")
	m := New(Config{
		Root:           root,
		Core:           []profile.Fingerprint{{Path: "game.exe", SHA256: transaction.Hash([]byte("core"))}},
		CheckProcesses: func(string) error { return nil },
		Fault: func(point string) error {
			if point == "init-prepared" {
				if err := os.WriteFile(ready, []byte("ready"), 0600); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(92)
				}
				time.Sleep(300 * time.Millisecond)
			}
			return nil
		},
	})
	if _, err := m.Run("init", "", Options{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(92)
	}
	os.Exit(0)
}

func TestConcurrentInitDoesNotAdoptIncompleteState(t *testing.T) {
	m, root := fixture(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestConcurrentInitChild$")
	cmd.Env = append(os.Environ(), "MGS3_INIT_CHILD=1", "MGS3_TEST_ROOT="+root)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	ready := filepath.Join(root, "init-child-ready")
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("initializer child did not reach lock-held boundary")
		}
		time.Sleep(10 * time.Millisecond)
	}
	requireManagerExit(t, m, "init", "", 6, Options{})
	if err := cmd.Wait(); err != nil {
		t.Fatalf("initializer child failed: %v", err)
	}
	cmd.Process = nil
	invoke(t, m, "init", "")
	if _, err := os.Stat(filepath.Join(root, stateDir, "INIT_COMMITTED")); err != nil {
		t.Fatal(err)
	}
}
