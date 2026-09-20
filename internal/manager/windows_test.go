package manager

import (
	"fmt"
	"golang.org/x/sys/windows"
	"mgs3mod/internal/winfs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLockedTargetAfterPreflight(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	var held windows.Handle
	m.config.Fault = func(point string) error {
		if point == "ready" {
			var err error
			held, err = windows.CreateFile(windows.StringToUTF16Ptr(filepath.Join(root, filepath.FromSlash(targetA))), windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
			return err
		}
		return nil
	}
	_, err := m.Run("enable", "one", Options{})
	if held != 0 {
		windows.CloseHandle(held)
	}
	m.config.Fault = nil
	if ExitCode(err) != 5 {
		t.Fatalf("sharing failure did not require recovery: %v", err)
	}
	assertBytes(t, root, targetA, "original:"+targetA)
	invoke(t, m, "recover", "")
	invoke(t, m, "verify", "")
}

func TestProcessUncertaintyBlocksWithoutWrites(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	m.config.CheckProcesses = func(string) error { return fmt.Errorf("candidate process inaccessible") }
	before := snapshot(t, root)
	if _, err := m.Run("enable", "one", Options{}); ExitCode(err) != 3 {
		t.Fatalf("process uncertainty not blocked: %v", err)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("process guard wrote")
	}
}

func TestNoOpAndDisabledRemovalDoNotNeedClosedGame(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	m.config.CheckProcesses = func(string) error { return fmt.Errorf("running") }
	invoke(t, m, "disable", "one")
	invoke(t, m, "restore", "")
	invoke(t, m, "remove", "one")
	invoke(t, m, "add", pkg)
	m.config.CheckProcesses = func(string) error { return nil }
	invoke(t, m, "enable", "one")
	m.config.CheckProcesses = func(string) error { return fmt.Errorf("running") }
	invoke(t, m, "enable", "one")
}

func TestConditionalRestoreAlreadyBaselineIgnoresExclusiveTarget(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "one")
	put(t, root, "game.exe", []byte("updated"))
	put(t, root, targetA, []byte("original:"+targetA))
	target := filepath.Join(root, filepath.FromSlash(targetA))
	held, err := windows.CreateFile(windows.StringToUTF16Ptr(target), windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(held)
	m.config.CheckProcesses = func(string) error { return fmt.Errorf("running") }
	result := invoke(t, m, "restore", "")
	if result.State.Mods["one"].Enabled || len(result.Paths) != 0 {
		t.Fatalf("logical-only restore touched exclusive target: %+v", result)
	}
}

func TestTargetSecurityDescriptorSurvivesLifecycleAndRollback(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	target := filepath.Join(root, filepath.FromSlash(targetA))
	wantSDDL := protectAndDescribeTarget(t, target)
	assertSDDL := func(step string) {
		t.Helper()
		sd, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT,
			windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatalf("%s security descriptor: %v", step, err)
		}
		if got := sd.String(); got != wantSDDL {
			t.Fatalf("%s security descriptor = %q, want %q", step, got, wantSDDL)
		}
	}

	invoke(t, m, "enable", "one")
	assertBytes(t, root, targetA, "mod:one:"+targetA)
	assertSDDL("enable")
	invoke(t, m, "disable", "one")
	assertBytes(t, root, targetA, "original:"+targetA)
	assertSDDL("disable")
	invoke(t, m, "enable", "one")
	invoke(t, m, "restore", "")
	assertBytes(t, root, targetA, "original:"+targetA)
	assertSDDL("restore")

	killAt(t, m, "enable", "one", "replaced-0")
	assertBytes(t, root, targetA, "mod:one:"+targetA)
	assertSDDL("interrupted enable")
	invoke(t, m, "recover", "")
	assertBytes(t, root, targetA, "original:"+targetA)
	assertSDDL("rollback")
}

func protectAndDescribeTarget(t *testing.T, target string) string {
	t.Helper()
	sd, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Skipf("DACL query unavailable: %v", err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Skipf("DACL read unavailable: %v", err)
	}
	if err := windows.SetNamedSecurityInfo(target, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil); err != nil {
		t.Skipf("DACL protection unavailable: %v", err)
	}
	protected, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	sddl := protected.String()
	if sddl == "" || !strings.Contains(sddl, "D:P") {
		t.Fatalf("protected security descriptor = %q", sddl)
	}
	return sddl
}

func TestWindowsUnsafeTargets(t *testing.T) {
	t.Run("hardlink", func(t *testing.T) {
		m, root := fixture(t)
		pkg := makePackage(t, root, "one", targetA)
		invoke(t, m, "init", "")
		invoke(t, m, "add", pkg)
		if err := os.Link(filepath.Join(root, filepath.FromSlash(targetA)), filepath.Join(root, "extra-link")); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Run("enable", "one", Options{}); err == nil {
			t.Fatal("hardlink accepted")
		}
	})
	t.Run("ADS", func(t *testing.T) {
		m, root := fixture(t)
		pkg := makePackage(t, root, "one", targetA)
		invoke(t, m, "init", "")
		invoke(t, m, "add", pkg)
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(targetA))+":extra", []byte("stream"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Run("enable", "one", Options{}); err == nil {
			t.Fatal("alternate stream accepted")
		}
	})
	t.Run("junction", func(t *testing.T) {
		_, root := fixture(t)
		target := filepath.Join(root, "real")
		link := filepath.Join(root, "link")
		if err := os.Mkdir(target, 0700); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", link, target)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("create fixture junction: %v %s", err, out)
		}
		if err = winfs.CheckDir(link); err == nil {
			t.Fatal("junction root accepted")
		}
		if err = winfs.CheckPath(root, "link/new.ctxr", true); err == nil {
			t.Fatal("junction parent accepted")
		}
	})
	t.Run("shortname", func(t *testing.T) {
		_, root := fixture(t)
		name := "long_texture_filename.ctxr"
		put(t, root, name, []byte("bytes"))
		full := filepath.Join(root, name)
		buf := make([]uint16, 32768)
		n, err := windows.GetShortPathName(windows.StringToUTF16Ptr(full), &buf[0], uint32(len(buf)))
		if err != nil {
			t.Fatal(err)
		}
		short := windows.UTF16ToString(buf[:n])
		alias := filepath.Base(short)
		if strings.EqualFold(alias, name) {
			t.Skip("NTFS short filename generation is disabled")
		}
		if err = winfs.CheckPath(root, alias, false); err == nil {
			t.Fatal("8.3 alias accepted")
		}
	})
}

func TestLockChild(t *testing.T) {
	if os.Getenv("MGS3_LOCK_CHILD") != "1" {
		return
	}
	root := os.Getenv("MGS3_TEST_ROOT")
	lock, err := winfs.Lock(filepath.Join(root, stateDir, "lock"), false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(92)
	}
	defer lock.Close()
	if err = os.WriteFile(filepath.Join(root, "held"), []byte("held"), 0600); err != nil {
		os.Exit(93)
	}
	time.Sleep(30 * time.Second)
	os.Exit(94)
}
func TestConcurrentProcessAndDeadOwner(t *testing.T) {
	m, root := fixture(t)
	invoke(t, m, "init", "")
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockChild$")
	cmd.Env = append(os.Environ(), "MGS3_LOCK_CHILD=1", "MGS3_TEST_ROOT="+root)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(root, "held")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child failed to acquire lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := m.Run("status", "", Options{}); ExitCode(err) != 6 {
		t.Fatalf("second process should be busy: %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	cmd.Wait()
	cmd.Process = nil
	invoke(t, m, "verify", "")
}
