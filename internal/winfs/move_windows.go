//go:build windows

package winfs

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// MoveNew promotes a staged file without replacing an existing destination.
// MOVEFILE_REPLACE_EXISTING is deliberately absent: an external file appearing
// after the caller's absence check must survive. Source and target stay on the
// same volume; cross-volume copying is not enabled.
func MoveNew(source, target string) error {
	if err := CheckPath(filepath.Dir(source), filepath.Base(source), false); err != nil {
		return err
	}
	if err := CheckPath(filepath.Dir(target), filepath.Base(target), true); err != nil {
		return err
	}
	src, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	dst, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(src, dst, windows.MOVEFILE_WRITE_THROUGH)
}
