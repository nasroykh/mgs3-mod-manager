package manager

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mgs3mod/internal/transaction"
	"mgs3mod/internal/winfs"
	"os"
	"path"
	"path/filepath"
)

type session struct {
	m    *Manager
	root *os.Root
	lock *winfs.LockHandle
}

func (s *session) close() { s.root.Close(); s.lock.Close() }
func (m *Manager) open(create bool) (*session, error) {
	if err := winfs.CheckDir(m.config.Root); err != nil {
		return nil, wrap(3, "unsafe installation root", err)
	}
	statePath := filepath.Join(m.config.Root, stateDir)
	if create {
		if err := os.Mkdir(statePath, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, err
		}
	}
	if err := winfs.CheckDir(statePath); err != nil {
		return nil, wrap(5, "manager not initialized or unsafe state directory", err)
	}
	lock, err := winfs.Lock(filepath.Join(statePath, "lock"), create)
	if errors.Is(err, winfs.ErrBusy) {
		return nil, fail(6, "another manager holds the lock")
	}
	if err != nil {
		return nil, wrap(5, "state lock", err)
	}
	root, err := os.OpenRoot(m.config.Root)
	if err != nil {
		lock.Close()
		return nil, err
	}
	if err = m.point("lock-acquired"); err != nil {
		root.Close()
		lock.Close()
		return nil, err
	}
	return &session{m, root, lock}, nil
}
func (s *session) guard(p string, missing bool) error {
	return winfs.CheckPath(s.m.config.Root, p, missing)
}
func (s *session) mkdir(p string) error {
	full := filepath.Join(s.m.config.Root, filepath.FromSlash(p))
	nearest := full
	for {
		_, err := os.Lstat(nearest)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		parent := filepath.Dir(nearest)
		if parent == nearest {
			return fmt.Errorf("no safe parent")
		}
		nearest = parent
	}
	if err := winfs.CheckDir(nearest); err != nil {
		return err
	}
	if err := s.root.MkdirAll(p, 0700); err != nil {
		return err
	}
	return winfs.CheckDir(full)
}
func (s *session) exists(p string) bool {
	_, err := s.root.Lstat(p)
	return !errors.Is(err, os.ErrNotExist)
}
func (s *session) hash(p string) (string, int64, error) {
	if err := s.guard(p, false); err != nil {
		return "", 0, err
	}
	f, err := s.root.Open(p)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}
func (s *session) expect(p, h string) error {
	actual, _, err := s.hash(p)
	if err != nil {
		return wrapPath(4, "cannot verify "+p, p, err)
	}
	if actual != h {
		return fail(4, "unexpected file contents", p)
	}
	return nil
}

// observe returns the content hash of a regular file, or absent=true when the
// path is safely missing. An empty expected hash is reserved for that missing
// state and is never a file-content hash.
func (s *session) observe(p string) (hash string, absent bool, err error) {
	if !s.exists(p) {
		if err := s.guard(p, true); err != nil {
			return "", false, err
		}
		return "", true, nil
	}
	hash, _, err = s.hash(p)
	return hash, false, err
}

func (s *session) expectState(p, h string) error {
	actual, absent, err := s.observe(p)
	if err != nil {
		return wrapPath(4, "cannot verify "+p, p, err)
	}
	if h == "" {
		if !absent {
			return fail(4, "expected file to be absent", p)
		}
		return nil
	}
	if absent || actual != h {
		return fail(4, "unexpected file contents", p)
	}
	return nil
}
func (s *session) write(p string, v any) error {
	if err := s.guard(p, true); err != nil {
		return err
	}
	return transaction.WriteUsing(s.root, p, v, s.m.config.CreateFile)
}
func read[T any](s *session, p string, v *T) error {
	if err := s.guard(p, false); err != nil {
		return err
	}
	return transaction.Read(s.root, p, v)
}
func (s *session) bytes(p string, b []byte) error {
	if err := s.guard(p, true); err != nil {
		return err
	}
	return transaction.WriteBytesUsing(s.root, p, b, s.m.config.CreateFile)
}
func (s *session) copy(src, dst, want string) error {
	if err := s.guard(src, false); err != nil {
		return err
	}
	if err := s.guard(dst, true); err != nil {
		return err
	}
	in, err := s.root.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := transaction.Create(s.root, dst, s.m.config.CreateFile)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(out, h), in)
	if err == nil {
		err = out.Sync()
	}
	closeErr := out.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if hex.EncodeToString(h.Sum(nil)) != want {
		return fail(4, "source changed while copying", src)
	}
	return s.expect(dst, want)
}
func (s *session) replace(src, dst, want string) error {
	if err := s.guard(src, false); err != nil {
		return err
	}
	if err := s.guard(dst, true); err != nil {
		return err
	}
	if s.exists(dst) {
		source := filepath.Join(s.m.config.Root, filepath.FromSlash(src))
		target := filepath.Join(s.m.config.Root, filepath.FromSlash(dst))
		if err := winfs.ReplaceExisting(source, target); err != nil {
			return err
		}
	} else {
		if err := s.root.Rename(src, dst); err != nil {
			return err
		}
	}
	return s.expect(dst, want)
}

// promoteNew atomically gives a transaction-owned stage its final name only
// when that name is still absent. Unlike replace, it must never overwrite a
// file that appeared after preflight.
func (s *session) promoteNew(src, dst, want string) error {
	if err := s.guard(src, false); err != nil {
		return err
	}
	if err := s.guard(dst, true); err != nil {
		return err
	}
	source := filepath.Join(s.m.config.Root, filepath.FromSlash(src))
	target := filepath.Join(s.m.config.Root, filepath.FromSlash(dst))
	if err := winfs.MoveNew(source, target); err != nil {
		return err
	}
	return s.expect(dst, want)
}
func (s *session) removeExact(item inventory) error {
	if !s.exists(item.Path) {
		return nil
	}
	h, n, err := s.hash(item.Path)
	if err != nil {
		return err
	}
	if h != item.SHA256 || n != item.Bytes {
		return fail(5, "cleanup preserves unexpected contents", item.Path)
	}
	return s.root.Remove(item.Path)
}
func txPath(seq uint64) string  { return fmt.Sprintf("%s/transactions/%020d", stateDir, seq) }
func blob(dir, h string) string { return path.Join(dir, "blobs", h) }
func baseline(h string) string  { return stateDir + "/baseline/" + h }
