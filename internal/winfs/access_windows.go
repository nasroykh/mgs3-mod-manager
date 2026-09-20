//go:build windows

package winfs

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

const (
	fileAddFile         = 0x0002
	fileAddSubdirectory = 0x0004
	fileTraverse        = 0x0020
)

// CheckReplaceAccess probes rights needed to replace an existing target.
// Probe opens are shared and are closed before returning; no filesystem state
// is created or changed. A successful probe is advisory because ACLs and
// sharing can change before the actual replacement.
func CheckReplaceAccess(root, rel string) error {
	if err := CheckPath(root, rel, false); err != nil {
		return fmt.Errorf("replace path: %w", err)
	}
	target := filepath.Join(filepath.Clean(root), filepath.FromSlash(rel))
	attrs, err := windows.GetFileAttributes(windows.StringToUTF16Ptr(target))
	if err != nil {
		return fmt.Errorf("replace attributes: %w", err)
	}
	if attrs&windows.FILE_ATTRIBUTE_READONLY != 0 {
		return fmt.Errorf("replace target is read-only: %s", target)
	}
	if err := openAccessProbe(target, windows.DELETE, 0); err != nil {
		return fmt.Errorf("replace target delete access: %w", err)
	}
	parent := filepath.Dir(target)
	if err := CheckDir(parent); err != nil {
		return fmt.Errorf("replace parent: %w", err)
	}
	if err := openAccessProbe(parent, fileAddFile|fileTraverse, windows.FILE_FLAG_BACKUP_SEMANTICS); err != nil {
		return fmt.Errorf("replace parent add-file access: %w", err)
	}
	return nil
}

// CheckCreateAccess probes rights needed to create a child in dir. It never
// creates a probe file, and closes its directory handle before returning.
func CheckCreateAccess(dir string) error {
	if err := CheckDir(dir); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	if err := openAccessProbe(dir, fileAddFile|fileAddSubdirectory|fileTraverse, windows.FILE_FLAG_BACKUP_SEMANTICS); err != nil {
		return fmt.Errorf("create access: %w", err)
	}
	return nil
}

func openAccessProbe(path string, access, flags uint32) error {
	share := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)
	h, err := windows.CreateFile(windows.StringToUTF16Ptr(path), access, share, nil, windows.OPEN_EXISTING, flags, 0)
	if err != nil {
		return err
	}
	if err := windows.CloseHandle(h); err != nil {
		return fmt.Errorf("close probe handle: %w", err)
	}
	return nil
}
