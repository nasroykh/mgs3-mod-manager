package manager

import (
	"golang.org/x/sys/windows"
	"io"
	"mgs3mod/internal/transaction"
	"os"
	"strings"
	"testing"
)

type failingFile struct {
	*os.File
	mode string
}

func (f *failingFile) Write(b []byte) (int, error) {
	switch f.mode {
	case "disk-full":
		n, err := f.File.Write(b[:min(7, len(b))])
		if err != nil {
			return n, err
		}
		return n, windows.ERROR_DISK_FULL
	case "short":
		return f.File.Write(b[:min(7, len(b))])
	}
	return f.File.Write(b)
}
func (f *failingFile) Sync() error {
	if f.mode == "sync" {
		return windows.ERROR_WRITE_FAULT
	}
	return f.File.Sync()
}
func (f *failingFile) Close() error {
	err := f.File.Close()
	if f.mode == "close" {
		return windows.ERROR_WRITE_FAULT
	}
	return err
}

func TestActualWriteBoundaryFailures(t *testing.T) {
	cases := []struct {
		name, match, mode  string
		corrupt, committed bool
	}{
		{"blob-disk-full", "/blobs/", "disk-full", false, false},
		{"baseline-disk-full", "/stage/baseline-", "disk-full", false, false},
		{"blob-access-denied", "/blobs/", "deny", false, false},
		{"prepared-short-write", "/PREPARED", "short", true, false},
		{"ready-short-write", "/READY", "short", true, false},
		{"ready-sync-failure", "/READY", "sync", false, false},
		{"commit-short-write", "/COMMITTED", "short", false, false},
		{"commit-sync-failure", "/COMMITTED", "sync", false, false},
		{"commit-close-failure", "/COMMITTED", "close", false, false},
		{"cache-disk-full", "/state.json", "disk-full", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			injected := false
			m.config.CreateFile = func(root *os.Root, path string) (transaction.SyncWriter, error) {
				active := !injected && strings.Contains(path, tc.match)
				if active && tc.mode == "deny" {
					injected = true
					return nil, &os.PathError{Op: "create", Path: path, Err: windows.ERROR_ACCESS_DENIED}
				}
				f, err := root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
				if err != nil {
					return nil, err
				}
				if active {
					injected = true
					return &failingFile{f, tc.mode}, nil
				}
				return f, nil
			}
			result, err := m.Run("enable", "one", Options{})
			if !injected || ExitCode(err) != 5 || !result.RecoveryRequired {
				t.Fatalf("injection not reported: injected=%v result=%+v err=%v", injected, result, err)
			}
			m.config.CreateFile = nil
			if tc.corrupt {
				before := snapshot(t, root)
				if _, err = m.Run("recover", "", Options{}); ExitCode(err) != 5 {
					t.Fatalf("torn record recovery should block: %v", err)
				}
				if !equal(before, snapshot(t, root)) {
					t.Fatal("corrupt journal recovery changed evidence")
				}
				return
			}
			if result.Committed != tc.committed {
				t.Fatalf("commit reporting=%v want %v", result.Committed, tc.committed)
			}
			invoke(t, m, "recover", "")
			want := "original:" + targetA
			if tc.committed {
				want = "mod:one:" + targetA
			}
			assertBytes(t, root, targetA, want)
			invoke(t, m, "verify", "")
		})
	}
}

func TestRollbackStageWriteFailuresAreResumable(t *testing.T) {
	for _, mode := range []string{"disk-full", "short", "sync"} {
		t.Run(mode, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			killAt(t, m, "enable", "one", "replaced-0")
			injected := false
			m.config.CreateFile = func(r *os.Root, name string) (transaction.SyncWriter, error) {
				f, err := r.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
				if err != nil {
					return nil, err
				}
				if !injected && strings.Contains(name, "/stage/rollback-") {
					injected = true
					return &failingFile{f, mode}, nil
				}
				return f, nil
			}
			result, err := m.Run("recover", "", Options{})
			if !injected || ExitCode(err) != 5 || !result.RecoveryRequired || result.Message != "recovery failed" || len(result.Paths) != 0 {
				t.Fatalf("rollback fault result: injected=%v result=%+v err=%v", injected, result, err)
			}
			m.config.CreateFile = nil
			invoke(t, m, "recover", "")
			assertBytes(t, root, targetA, "original:"+targetA)
			invoke(t, m, "verify", "")
		})
	}
}

func TestRollbackStageProcessInterruptionResumes(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "replaced-0")
	killAt(t, m, "recover", "", "rollback-staged-0")
	invoke(t, m, "recover", "")
	assertBytes(t, root, targetA, "original:"+targetA)
	invoke(t, m, "verify", "")
}

func TestCommitPromotionCrashBoundary(t *testing.T) {
	for _, tc := range []struct {
		point, want string
	}{
		{point: "commit-staged", want: "original:" + targetA},
		{point: "commit-promoted", want: "mod:one:" + targetA},
	} {
		t.Run(tc.point, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			killAt(t, m, "enable", "one", tc.point)
			invoke(t, m, "recover", "")
			assertBytes(t, root, targetA, tc.want)
			invoke(t, m, "verify", "")
		})
	}
}

func TestRecoveryCacheFailureAfterDurableRollbackReportsResolvedJournal(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "replaced-0")

	injected := false
	m.config.CreateFile = func(r *os.Root, name string) (transaction.SyncWriter, error) {
		f, err := r.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return nil, err
		}
		if !injected && strings.HasSuffix(name, "/state.json") {
			injected = true
			return &failingFile{File: f, mode: "sync"}, nil
		}
		return f, nil
	}
	result, err := m.Run("recover", "", Options{})
	if !injected || ExitCode(err) != 1 || result.RecoveryRequired || result.Committed || result.State == nil {
		t.Fatalf("durable rollback result: injected=%v result=%+v err=%v", injected, result, err)
	}
	if result.Message != "recovery failed without an unresolved journal" || result.State.Generation != 1 || result.State.Mods["one"].Enabled {
		t.Fatalf("durable rollback state/message: %+v", result)
	}
	assertBytes(t, root, targetA, "original:"+targetA)

	m.config.CreateFile = nil
	invoke(t, m, "verify", "")
}

func TestDiskFullCopyActuallyLeavesPartialOwnedFile(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	var partial string
	m.config.CreateFile = func(r *os.Root, path string) (transaction.SyncWriter, error) {
		f, err := r.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return nil, err
		}
		if partial == "" && strings.Contains(path, "/blobs/") {
			partial = path
			return &failingFile{f, "disk-full"}, nil
		}
		return f, nil
	}
	if _, err := m.Run("enable", "one", Options{}); err == nil {
		t.Fatal("disk full accepted")
	}
	m.config.CreateFile = nil
	data, err := os.ReadFile(pathJoin(root, partial))
	if err != nil || len(data) != 7 {
		t.Fatalf("actual partial output: %d bytes %v", len(data), err)
	}
	invoke(t, m, "recover", "")
	if _, err = os.Stat(pathJoin(root, partial)); !os.IsNotExist(err) {
		t.Fatalf("partial owned staging remains: %v", err)
	}
	assertBytes(t, root, targetA, "original:"+targetA)
}

var _ io.Writer = (*failingFile)(nil)
