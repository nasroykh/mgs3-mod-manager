//go:build windows

package winfs

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestCheckCreateAccessDryRun(t *testing.T) {
	root := fixtureDir(t)
	before := directoryEntries(t, root)
	if err := CheckCreateAccess(root); err != nil {
		t.Fatal(err)
	}
	after := directoryEntries(t, root)
	if !equalEntries(before, after) {
		t.Fatalf("create probe changed directory: before=%v after=%v", before, after)
	}
}

func TestCheckReplaceAccessWritableDryRun(t *testing.T) {
	root := fixtureDir(t)
	name := "target.ctxr"
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	before := directoryEntries(t, root)
	if err := CheckReplaceAccess(root, name); err != nil {
		t.Fatal(err)
	}
	after := directoryEntries(t, root)
	if !equalEntries(before, after) {
		t.Fatalf("replace probe changed directory: before=%v after=%v", before, after)
	}
}

func TestCheckReplaceAccessReadOnlyTargetFails(t *testing.T) {
	root := fixtureDir(t)
	name := "readonly.ctxr"
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := windows.SetFileAttributes(windows.StringToUTF16Ptr(path), windows.FILE_ATTRIBUTE_READONLY); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = windows.SetFileAttributes(windows.StringToUTF16Ptr(path), windows.FILE_ATTRIBUTE_NORMAL) })
	if err := CheckReplaceAccess(root, name); err == nil {
		t.Fatal("read-only target unexpectedly passed replace probe")
	}
}

func TestCheckReplaceAccessExclusiveSharingFails(t *testing.T) {
	root := fixtureDir(t)
	name := "shared.ctxr"
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(windows.StringToUTF16Ptr(path), windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	if err := CheckReplaceAccess(root, name); err == nil {
		t.Fatal("exclusive sharing unexpectedly passed replace probe")
	}
}

func TestCheckReplaceAccessDACLDeniesDelete(t *testing.T) {
	root := fixtureDir(t)
	name := "acl-deny.ctxr"
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Skipf("DACL query unavailable: %v", err)
	}
	oldACL, _, err := sd.DACL()
	if err != nil {
		t.Skipf("DACL read unavailable: %v", err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Skipf("token user unavailable: %v", err)
	}
	deny := windows.EXPLICIT_ACCESS{
		AccessPermissions: windows.DELETE,
		AccessMode:        windows.DENY_ACCESS,
		Inheritance:       windows.NO_INHERITANCE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}
	newACL, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{deny}, oldACL)
	if err != nil {
		t.Skipf("DACL construction unavailable: %v", err)
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, newACL, nil); err != nil {
		t.Skipf("DACL mutation unavailable: %v", err)
	}
	t.Cleanup(func() {
		if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, oldACL, nil); err != nil {
			t.Errorf("restore DACL: %v", err)
		}
	})
	if err := CheckReplaceAccess(root, name); err == nil {
		t.Fatal("DELETE-denied target unexpectedly passed replace probe")
	}
}

func directoryEntries(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		result[entry.Name()] = struct{}{}
	}
	return result
}

func equalEntries(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for name := range a {
		if _, ok := b[name]; !ok {
			return false
		}
	}
	return true
}
