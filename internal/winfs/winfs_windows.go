//go:build windows

// Package winfs contains Windows-only filesystem and process guards used by
// the manager. The guards are deliberately conservative: an uncertain or
// aliased path is an error.
package winfs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ErrBusy means another manager process owns the byte-range lock.
var ErrBusy = errors.New("winfs: busy")

const (
	lockLengthLow  = uint32(1)
	lockLengthHigh = uint32(0)

	// Windows has no public constant for this error in every supported SDK,
	// but x/sys/windows exposes the Win32 error value.
	maxProcessPath = 32768
)

var replaceFileProc = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReplaceFileW")

// CheckPath validates root and a root-relative path. Existing parents must be
// real directories with no reparse point. A non-empty rel must name an
// existing regular, non-hard-linked file; allowMissing permits only its final
// component to be absent.
func CheckPath(root, rel string, allowMissing bool) error {
	if err := validateAbsolute(root); err != nil {
		return fmt.Errorf("root: %w", err)
	}
	root = filepath.Clean(root)
	if rel == "" {
		return checkExisting(root, true, false)
	}
	parts, err := splitRelative(rel)
	if err != nil {
		return err
	}
	if err := checkExisting(root, true, false); err != nil {
		return fmt.Errorf("root: %w", err)
	}
	current := root
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			if os.IsNotExist(statErr) && i == len(parts)-1 && allowMissing {
				return nil
			}
			return fmt.Errorf("%s: %w", current, statErr)
		}
		last := i == len(parts)-1
		if last {
			if err := checkExistingInfo(current, info, false, true); err != nil {
				return err
			}
			return nil
		}
		if err := checkExistingInfo(current, info, true, false); err != nil {
			return err
		}
	}
	return nil
}

// CheckDir validates an existing absolute directory and every ancestor.
func CheckDir(path string) error {
	if err := validateAbsolute(path); err != nil {
		return err
	}
	path = filepath.Clean(path)
	return checkExisting(path, true, false)
}

// LockHandle owns the exclusive byte-range lock returned by Lock.
type LockHandle struct {
	mu     sync.Mutex
	handle windows.Handle
	over   windows.Overlapped
	closed bool
}

// Lock opens path and takes a fail-immediately exclusive byte-range lock. If
// create is false, path must already exist.
func Lock(path string, create bool) (*LockHandle, error) {
	if err := validateAbsolute(path); err != nil {
		return nil, err
	}
	path = filepath.Clean(path)
	parent := filepath.Dir(path)
	if err := CheckDir(parent); err != nil {
		return nil, fmt.Errorf("lock parent: %w", err)
	}
	base := filepath.Base(path)
	if err := CheckPath(parent, base, create); err != nil {
		return nil, fmt.Errorf("lock path: %w", err)
	}
	disposition := uint32(windows.OPEN_EXISTING)
	if create {
		disposition = windows.OPEN_ALWAYS
	}
	h, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		disposition,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("open lock: %w", err)
	}
	// OPEN_ALWAYS can race with a reparse or hard-link replacement. Validate
	// the opened name before taking ownership.
	if err := CheckPath(parent, base, false); err != nil {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("validate lock: %w", err)
	}
	lock := &LockHandle{handle: h}
	err = windows.LockFileEx(h, windows.LOCKFILE_FAIL_IMMEDIATELY|windows.LOCKFILE_EXCLUSIVE_LOCK,
		0, lockLengthLow, lockLengthHigh, &lock.over)
	if err != nil {
		_ = windows.CloseHandle(h)
		if isBusyError(err) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("lock file: %w", err)
	}
	return lock, nil
}

// ReplaceExisting replaces source over an existing target through ReplaceFileW.
// Windows preserves target security metadata during this operation. Callers
// must use this only when target exists; missing destinations require rename.
func ReplaceExisting(source, target string) error {
	if err := validateAbsolute(source); err != nil {
		return fmt.Errorf("replacement source: %w", err)
	}
	if err := validateAbsolute(target); err != nil {
		return fmt.Errorf("replacement target: %w", err)
	}
	source = filepath.Clean(source)
	target = filepath.Clean(target)
	if sameWindowsPath(source, target) {
		return fmt.Errorf("replacement source and target are identical")
	}
	if err := checkExisting(source, false, true); err != nil {
		return fmt.Errorf("replacement source: %w", err)
	}
	if err := checkExisting(target, false, true); err != nil {
		return fmt.Errorf("replacement target: %w", err)
	}
	sourcePtr, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return fmt.Errorf("replacement source: %w", err)
	}
	targetPtr, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return fmt.Errorf("replacement target: %w", err)
	}
	result, _, callErr := replaceFileProc.Call(
		uintptr(unsafe.Pointer(targetPtr)),
		uintptr(unsafe.Pointer(sourcePtr)),
		0,
		0,
		0,
		0,
	)
	if result == 0 {
		if callErr == nil || callErr == syscall.Errno(0) {
			callErr = windows.ERROR_FUNCTION_FAILED
		}
		return fmt.Errorf("replace existing: %w", callErr)
	}
	return nil
}

// Close releases lock and closes its underlying handle. It is idempotent.
func (l *LockHandle) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	unlockErr := windows.UnlockFileEx(l.handle, 0, lockLengthLow, lockLengthHigh, &l.over)
	closeErr := windows.CloseHandle(l.handle)
	if unlockErr != nil {
		return fmt.Errorf("unlock file: %w", unlockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close lock: %w", closeErr)
	}
	return nil
}

// CheckProcesses blocks when the installation's game or launcher is running.
// Toolhelp gives candidate basenames first; full paths are queried only for
// those candidates. A matching candidate whose path cannot be queried blocks.
func CheckProcesses(root string) error {
	if err := CheckDir(root); err != nil {
		return fmt.Errorf("process root: %w", err)
	}
	root = filepath.Clean(root)
	candidates := map[string]bool{
		strings.ToLower("METAL GEAR SOLID3.exe"): true,
		strings.ToLower("launcher.exe"):          true,
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("process snapshot: %w", err)
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
			return nil
		}
		return fmt.Errorf("process snapshot first: %w", err)
	}
	for {
		name := windows.UTF16ToString(entry.ExeFile[:])
		if candidates[strings.ToLower(name)] {
			pid := entry.ProcessID
			process, openErr := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
			if openErr != nil {
				if errors.Is(openErr, windows.ERROR_INVALID_PARAMETER) {
					// Process exited between snapshot and OpenProcess.
				} else {
					return fmt.Errorf("matching process %q (pid %d) is inaccessible: %w", name, pid, openErr)
				}
			} else {
				image, pathErr := processImagePath(process)
				_ = windows.CloseHandle(process)
				if pathErr != nil {
					return fmt.Errorf("matching process %q (pid %d) path is inaccessible: %w", name, pid, pathErr)
				}
				resolved, resolveErr := longPathName(image)
				if resolveErr != nil {
					return fmt.Errorf("matching process %q (pid %d) path cannot be canonicalized: %w", name, pid, resolveErr)
				}
				if sameWindowsPath(resolved, filepath.Join(root, name)) {
					return fmt.Errorf("installation process running: %s (pid %d)", image, pid)
				}
			}
		}
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				return nil
			}
			return fmt.Errorf("process snapshot next: %w", err)
		}
	}
}

// FreeBytes returns bytes available to the current caller on root's volume.
func FreeBytes(root string) (uint64, error) {
	if err := CheckDir(root); err != nil {
		return 0, err
	}
	var available, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(root), &available, &total, &free); err != nil {
		return 0, fmt.Errorf("disk free space: %w", err)
	}
	return available, nil
}

func validateAbsolute(path string) error {
	if path == "" || !filepath.IsAbs(path) || filepath.VolumeName(path) == "" {
		return fmt.Errorf("path must be an absolute Windows path: %q", path)
	}
	if strings.IndexByte(path, 0) >= 0 {
		return fmt.Errorf("path contains NUL")
	}
	volume := filepath.VolumeName(path)
	if strings.Contains(path[len(volume):], ":") {
		return fmt.Errorf("path contains alternate data stream syntax: %q", path)
	}
	return nil
}

func splitRelative(rel string) ([]string, error) {
	if strings.ContainsAny(rel, `\\:`) {
		return nil, fmt.Errorf("relative path contains backslash or colon: %q", rel)
	}
	if strings.HasPrefix(rel, "/") || strings.HasSuffix(rel, "/") {
		return nil, fmt.Errorf("relative path is not canonical: %q", rel)
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.HasSuffix(part, " ") || strings.HasSuffix(part, ".") {
			return nil, fmt.Errorf("relative path is not canonical: %q", rel)
		}
		for i := 0; i < len(part); i++ {
			if part[i] < 0x20 || part[i] > 0x7e || strings.ContainsRune(`*?[]<>|"`, rune(part[i])) {
				return nil, fmt.Errorf("relative path contains unsafe character: %q", rel)
			}
		}
	}
	return parts, nil
}

func checkExisting(path string, wantDir, rejectDir bool) error {
	volume := filepath.VolumeName(path)
	base := volume + string(filepath.Separator)
	rest := strings.TrimPrefix(path, base)
	if rest == path {
		// UNC roots use volume plus separator in filepath.Clean; fallback keeps
		// root validation correct for drive roots and UNC shares.
		rest = strings.TrimPrefix(path, volume)
		rest = strings.TrimLeft(rest, `\\/`)
	}
	if rest == "" {
		info, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return checkExistingInfo(path, info, wantDir, rejectDir)
	}
	current := base
	for _, part := range strings.FieldsFunc(rest, func(r rune) bool { return r == '\\' || r == '/' }) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("%s: %w", current, err)
		}
		last := filepath.Clean(current) == filepath.Clean(path)
		if err := checkExistingInfo(current, info, last && wantDir, last && rejectDir); err != nil {
			return err
		}
	}
	return nil
}

func checkExistingInfo(path string, info os.FileInfo, wantDir, rejectDir bool) error {
	if info == nil {
		return fmt.Errorf("%s: path does not exist", path)
	}
	attrs, err := windows.GetFileAttributes(windows.StringToUTF16Ptr(path))
	if err != nil {
		return fmt.Errorf("%s: attributes: %w", path, err)
	}
	if attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("%s: reparse point is not allowed", path)
	}
	isDir := attrs&windows.FILE_ATTRIBUTE_DIRECTORY != 0
	if err := rejectAlias(path, rejectDir && !isDir, isDir); err != nil {
		return err
	}
	if wantDir {
		if !isDir || !info.IsDir() {
			return fmt.Errorf("%s: expected directory", path)
		}
		return nil
	}
	if rejectDir {
		if isDir || !info.Mode().IsRegular() {
			return fmt.Errorf("%s: target is not a regular file", path)
		}
		return nil
	}
	return nil
}

func rejectAlias(path string, checkHardLink, isDir bool) error {
	access := uint32(windows.GENERIC_READ)
	if isDir {
		// Directory metadata can be queried with zero desired access. This also
		// permits checking ancestors where ordinary users lack list/read rights.
		access = 0
	}
	h, err := windows.CreateFile(windows.StringToUTF16Ptr(path), access,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return fmt.Errorf("%s: open for identity: %w", path, err)
	}
	defer windows.CloseHandle(h)
	longName, err := longPathName(path)
	if err != nil {
		return fmt.Errorf("%s: resolve long name: %w", path, err)
	}
	if !sameWindowsPath(path, longName) {
		return fmt.Errorf("%s: short-name or alternate spelling is not allowed (resolved %s)", path, longName)
	}
	var data windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &data); err != nil {
		return fmt.Errorf("%s: file identity: %w", path, err)
	}
	if checkHardLink && data.NumberOfLinks > 1 {
		return fmt.Errorf("%s: hard-linked file is not allowed", path)
	}
	if !isDir {
		if err := rejectStreams(h); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

// FILE_STREAM_INFO uses byte offsets and UTF-16 byte lengths. Bound every
// offset before decoding. An oversized stream inventory is rejected too.
func rejectStreams(h windows.Handle) error {
	buffer := make([]byte, 64<<10)
	err := windows.GetFileInformationByHandleEx(h, windows.FileStreamInfo, &buffer[0], uint32(len(buffer)))
	if errors.Is(err, windows.ERROR_HANDLE_EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot enumerate streams: %w", err)
	}
	for offset := 0; ; {
		if offset > len(buffer)-24 {
			return fmt.Errorf("invalid stream metadata")
		}
		next := int(binary.LittleEndian.Uint32(buffer[offset:]))
		size := int(binary.LittleEndian.Uint32(buffer[offset+4:]))
		if size%2 != 0 || size < 2 || size > len(buffer)-offset-24 {
			return fmt.Errorf("invalid stream name length")
		}
		name := make([]uint16, size/2)
		for i := range name {
			name[i] = binary.LittleEndian.Uint16(buffer[offset+24+2*i:])
		}
		if windows.UTF16ToString(name) != "::$DATA" {
			return fmt.Errorf("alternate data stream is not allowed")
		}
		if next == 0 {
			return nil
		}
		if next < 24+size || next > len(buffer)-offset-24 {
			return fmt.Errorf("invalid stream offset")
		}
		offset += next
	}
}

func longPathName(path string) (string, error) {
	ptr := windows.StringToUTF16Ptr(path)
	for size := uint32(512); size <= 32768; size *= 2 {
		buf := make([]uint16, size)
		n, err := windows.GetLongPathName(ptr, &buf[0], size)
		if err == nil {
			return windows.UTF16ToString(buf[:n]), nil
		}
		if n > 0 && n < size {
			return windows.UTF16ToString(buf[:n]), nil
		}
		if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
			return "", err
		}
	}
	return "", fmt.Errorf("path exceeds Windows maximum")
}

func sameWindowsPath(a, b string) bool {
	strip := func(path string) string {
		path = strings.TrimPrefix(path, `\\?\`)
		path = filepath.Clean(path)
		return strings.ToLower(path)
	}
	return strip(a) == strip(b)
}

func processImagePath(process windows.Handle) (string, error) {
	buf := make([]uint16, maxProcessPath)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(process, 0, &buf[0], &size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf[:size]), nil
}

func isBusyError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.Errno(windows.ERROR_LOCK_VIOLATION) || errno == syscall.Errno(windows.ERROR_LOCK_FAILED)
	}
	return err == windows.ERROR_LOCK_VIOLATION || err == windows.ERROR_LOCK_FAILED
}
