package manager

import (
	"errors"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/winfs"
	"os"
)

func (m *Manager) compatible() error {
	return wrapPath(3, "installation fingerprints", m.config.Root, profile.Check(m.config.Root, m.config.Core))
}
func (m *Manager) processes() error {
	check := m.config.CheckProcesses
	if check == nil {
		check = winfs.CheckProcesses
	}
	return wrap(3, "close the game and launcher before continuing", check(m.config.Root))
}
func (s *session) initSpec() initRecord {
	return initRecord{1, profile.ID, s.m.config.Root, []string{"baseline", "library", "transactions", "installation.json", "state.json", "lock", "INIT_PREPARED", "INIT_COMMITTED"}}
}
func (s *session) checkInit() error {
	want := s.initSpec()
	for _, p := range []string{"INIT_PREPARED", "installation.json", "INIT_COMMITTED"} {
		var got initRecord
		if err := read(s, stateDir+"/"+p, &got); err != nil {
			return wrap(5, "incomplete or invalid initialization", err)
		}
		if !equal(got, want) {
			return fail(5, "installation record mismatch")
		}
	}
	return nil
}
func (s *session) finishInit() error {
	spec := s.initSpec()
	var prepared initRecord
	if err := read(s, stateDir+"/INIT_PREPARED", &prepared); err != nil {
		return wrap(5, "cannot adopt incomplete state directory", err)
	}
	if !equal(prepared, spec) {
		return fail(5, "initialization identity mismatch")
	}
	f, err := s.root.Open(stateDir)
	if err != nil {
		return err
	}
	entries, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		return err
	}
	allowed := map[string]bool{}
	for _, p := range spec.Paths {
		allowed[p] = true
	}
	for _, e := range entries {
		if e.Name() == "state.json" {
			return fail(5, "unexpected state cache before initialization committed")
		}
		if !allowed[e.Name()] {
			return fail(5, "unexpected initialization contents", e.Name())
		}
		if e.IsDir() {
			if err := winfs.CheckDir(pathJoin(s.m.config.Root, stateDir+"/"+e.Name())); err != nil {
				return err
			}
			d, err := s.root.Open(stateDir + "/" + e.Name())
			if err != nil {
				return err
			}
			children, err := d.ReadDir(-1)
			d.Close()
			if err != nil {
				return err
			}
			if len(children) > 0 {
				return fail(5, "nonempty incomplete initialization directory", e.Name())
			}
		}
	}
	for _, dir := range []string{"baseline", "library", "transactions"} {
		if err = s.mkdir(stateDir + "/" + dir); err != nil {
			return err
		}
		if err = s.m.point("init-directory-" + dir); err != nil {
			return err
		}
	}
	installation := stateDir + "/installation.json"
	if s.exists(installation) {
		var got initRecord
		if err = read(s, installation, &got); err != nil {
			return wrap(5, "invalid installation record", err)
		}
		if !equal(got, spec) {
			return fail(5, "installation record differs")
		}
	} else if err = s.write(installation, spec); err != nil {
		return err
	}
	if err = s.m.point("init-installation"); err != nil {
		return err
	}
	if err = s.write(stateDir+"/INIT_COMMITTED", spec); err != nil {
		return err
	}
	return s.m.point("init-committed")
}
func (m *Manager) init(dry bool) (Result, error) {
	result := Result{Command: "init", DryRun: dry, Message: "initialize local manager"}
	if err := m.compatible(); err != nil {
		return result, err
	}
	result.Compatible = true
	statePath := pathJoin(m.config.Root, stateDir)
	_, statErr := os.Lstat(statePath)
	if statErr == nil {
		s, err := m.open(false)
		if err != nil {
			return result, err
		}
		defer s.close()
		if err = m.compatible(); err != nil {
			return result, err
		}
		if err = s.checkInit(); err != nil {
			return result, err
		}
		if _, err = s.load(); err != nil {
			return result, err
		}
		result.Initialized = true
		result.Message = "already initialized"
		return result, nil
	}
	if !errors.Is(statErr, os.ErrNotExist) {
		return result, statErr
	}
	if err := winfs.CheckCreateAccess(m.config.Root); err != nil {
		return result, wrapPath(1, "cannot initialize manager", m.config.Root, err)
	}
	if dry {
		return result, nil
	}
	s, err := m.open(true)
	if err != nil {
		return result, err
	}
	defer s.close()
	if err = m.compatible(); err != nil {
		return result, err
	}
	// A concurrent init may have finished while this caller waited for the lock.
	if s.exists(stateDir + "/INIT_COMMITTED") {
		err = s.checkInit()
		result.Initialized = err == nil
		return result, err
	}
	f, err := s.root.Open(stateDir)
	if err != nil {
		return result, err
	}
	entries, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if entry.Name() != "lock" {
			return result, fail(5, "incomplete initialization; use recover")
		}
	}
	if err = s.write(stateDir+"/INIT_PREPARED", s.initSpec()); err != nil {
		return result, err
	}
	if err = m.point("init-prepared"); err != nil {
		return result, err
	}
	if err = s.finishInit(); err != nil {
		return result, err
	}
	result.Initialized = true
	return result, nil
}
