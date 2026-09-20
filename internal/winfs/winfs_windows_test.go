//go:build windows

package winfs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestCheckPathAndDirectoryGuards(t *testing.T) {
	root := fixtureDir(t)
	if err := CheckDir(root); err != nil {
		t.Fatal(err)
	}
	if err := CheckPath(root, "", false); err != nil {
		t.Fatal(err)
	}
	if err := CheckPath(root, "missing.ctxr", true); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "target.ctxr"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := CheckPath(root, "target.ctxr", false); err != nil {
		t.Fatal(err)
	}
	if err := CheckPath(root, "target.ctxr", true); err != nil {
		t.Fatal(err)
	}
	if err := CheckPath(root, "target.ctxr/child", true); err == nil {
		t.Fatal("expected non-directory target to fail parent validation")
	}
	if err := CheckPath(root, "target.ctxr", false); err != nil {
		t.Fatal(err)
	}
	if err := CheckPath(root, "missing/target.ctxr", true); err == nil {
		t.Fatal("expected missing parent to fail")
	}
	if err := CheckPath(root, "..\\target.ctxr", true); err == nil {
		t.Fatal("expected non-canonical path to fail")
	}
}

func TestLockBusyAndRelease(t *testing.T) {
	root := fixtureDir(t)
	path := filepath.Join(root, "lock")
	first, err := Lock(path, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Lock(path, false)
	if !errors.Is(err, ErrBusy) {
		t.Fatalf("second lock error = %v, want ErrBusy", err)
	}
	if second != nil {
		_ = second.Close()
		t.Fatal("busy lock unexpectedly returned handle")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	third, err := Lock(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := third.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestHeldLockPathCannotBeRenamedOrDeleted(t *testing.T) {
	root := fixtureDir(t)
	path := filepath.Join(root, "lock")
	lock, err := Lock(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()

	if err := os.Rename(path, filepath.Join(root, "replacement")); err == nil {
		t.Fatal("held lock path was renamed")
	}
	if err := os.Remove(path); err == nil {
		t.Fatal("held lock path was deleted")
	}
}

func TestReplaceExistingPreservesProtectedDACLAndBytes(t *testing.T) {
	root := fixtureDir(t)
	target := filepath.Join(root, "target.ctxr")
	source := filepath.Join(root, "source.ctxr")
	original := []byte("original target bytes")
	replacement := []byte("replacement bytes with exact length independent of target")
	if err := os.WriteFile(target, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, replacement, 0600); err != nil {
		t.Fatal(err)
	}
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
	protected, err := protectedDACL(target)
	if err != nil {
		t.Skipf("DACL control query unavailable: %v", err)
	}
	if !protected {
		t.Fatal("fixture DACL was not protected")
	}
	beforeSD, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	beforeSDDL := beforeSD.String()
	if beforeSDDL == "" {
		t.Fatal("fixture security descriptor could not be rendered")
	}

	if err := ReplaceExisting(source, target); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, replacement) {
		t.Fatalf("replacement bytes = %q, want %q", got, replacement)
	}
	protected, err = protectedDACL(target)
	if err != nil {
		t.Fatal(err)
	}
	if !protected {
		t.Fatal("replacement removed target DACL protection")
	}
	afterSD, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if afterSDDL := afterSD.String(); afterSDDL != beforeSDDL {
		t.Fatalf("replacement security descriptor = %q, want %q", afterSDDL, beforeSDDL)
	}
}

func TestCheckProcessesAndFreeBytes(t *testing.T) {
	root := fixtureDir(t)
	if err := CheckProcesses(root); err != nil {
		t.Fatal(err)
	}
	free, err := FreeBytes(root)
	if err != nil {
		t.Fatal(err)
	}
	if free == 0 {
		t.Fatal("free bytes unexpectedly zero")
	}
}

func fixtureDir(t *testing.T) string {
	t.Helper()
	parent, err := filepath.Abs(filepath.Join("..", "..", ".cache", "winfs-tests"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp(parent, "case-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	return root
}

func protectedDACL(path string) (bool, error) {
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return false, err
	}
	proc := windows.NewLazySystemDLL("advapi32.dll").NewProc("GetSecurityDescriptorControl")
	var control uint16
	var revision uint32
	result, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(sd)),
		uintptr(unsafe.Pointer(&control)),
		uintptr(unsafe.Pointer(&revision)),
	)
	if result == 0 {
		if callErr == nil || callErr == syscall.Errno(0) {
			callErr = errors.New("GetSecurityDescriptorControl failed")
		}
		return false, callErr
	}
	return control&uint16(windows.SE_DACL_PROTECTED) != 0, nil
}
