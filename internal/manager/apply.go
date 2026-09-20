package manager

import (
	"fmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"mgs3mod/internal/winfs"
	"path"
	"sort"
)

type input struct {
	Path string
	Data []byte
}

func (s *session) execute(p plan, inputs map[string]input) error {
	dir := txPath(p.Sequence)
	if err := validatePlan(p, p.Before, p.Sequence); err != nil {
		return wrap(1, "internal plan validation", err)
	}
	if err := s.mkdir(dir); err != nil {
		return err
	}
	if err := s.write(dir+"/PREPARED", p); err != nil {
		return err
	}
	if err := s.m.point("prepared"); err != nil {
		return err
	}
	for _, d := range []string{dir + "/blobs", dir + "/stage"} {
		if err := s.mkdir(d); err != nil {
			return err
		}
	}
	hashes := []string{}
	for hash := range inputs {
		hashes = append(hashes, hash)
	}
	sort.Strings(hashes)
	for _, hash := range hashes {
		src := inputs[hash]
		dst := blob(dir, hash)
		var err error
		if src.Path != "" {
			err = s.copy(src.Path, dst, hash)
		} else {
			if transaction.Hash(src.Data) != hash {
				return fail(4, "payload changed during preparation")
			}
			err = s.bytes(dst, src.Data)
		}
		if err != nil {
			return err
		}
		if err = s.expect(dst, hash); err != nil {
			return err
		}
		if err = s.m.point("blob-" + hash); err != nil {
			return err
		}
	}
	promotedBaselines := map[string]bool{}
	for _, item := range p.Baselines {
		verifiedExisted := p.Schema == 2 && item.VerifiedObjectExisted != nil && *item.VerifiedObjectExisted
		if verifiedExisted || p.Schema == 1 && s.exists(item.Path) {
			if err := s.expect(item.Path, item.SHA256); err != nil {
				return err
			}
			continue
		}
		if s.exists(item.Path) {
			if !promotedBaselines[item.Path] {
				return fail(5, "baseline appeared during preparation", item.Path)
			}
			if err := s.expect(item.Path, item.SHA256); err != nil {
				return err
			}
			continue
		}
		temp := item.Temporary
		if p.Schema == 1 {
			temp = dir + "/stage/baseline-" + item.SHA256
		}
		if err := s.copy(blob(dir, item.SHA256), temp, item.SHA256); err != nil {
			return err
		}
		if err := s.m.point("baseline-staged"); err != nil {
			return err
		}
		if s.exists(item.Path) {
			return fail(5, "baseline appeared during preparation", item.Path)
		}
		if err := s.replace(temp, item.Path, item.SHA256); err != nil {
			return err
		}
		promotedBaselines[item.Path] = true
		if err := s.m.point("baseline-promoted"); err != nil {
			return err
		}
	}
	for i, c := range p.Changes {
		if !c.Game {
			if err := s.mkdir(path.Dir(c.Path)); err != nil {
				return err
			}
		}
		if c.After != "" {
			if err := s.copy(blob(dir, c.After), fmt.Sprintf("%s/stage/apply-%04d", dir, i), c.After); err != nil {
				return err
			}
		}
	}
	if err := s.mark(dir, "READY", p, -1); err != nil {
		return err
	}
	if err := s.m.point("ready"); err != nil {
		return err
	}
	for i, c := range p.Changes {
		if err := s.mark(dir, fmt.Sprintf("APPLY_INTENT_%04d", i), p, i); err != nil {
			return err
		}
		if err := s.m.point(fmt.Sprintf("intent-%d", i)); err != nil {
			return err
		}
		if err := s.expectState(c.Path, c.Before); err != nil {
			return err
		}
		if c.Game {
			if err := s.m.processes(); err != nil {
				return err
			}
			if err := s.expectState(c.Path, c.Before); err != nil {
				return err
			}
		}
		if c.After == "" {
			if err := s.root.Remove(c.Path); err != nil {
				return err
			}
			if err := s.expectState(c.Path, ""); err != nil {
				return err
			}
		} else {
			stage := fmt.Sprintf("%s/stage/apply-%04d", dir, i)
			var err error
			if c.Before == "" {
				err = s.promoteNew(stage, c.Path, c.After)
			} else {
				err = s.replace(stage, c.Path, c.After)
			}
			if err != nil {
				return err
			}
		}
		if err := s.m.point(fmt.Sprintf("replaced-%d", i)); err != nil {
			return err
		}
		if err := s.mark(dir, fmt.Sprintf("APPLIED_%04d", i), p, i); err != nil {
			return err
		}
		if err := s.m.point(fmt.Sprintf("applied-%d", i)); err != nil {
			return err
		}
	}
	if err := s.commit(dir, p); err != nil {
		return err
	}
	if err := s.m.point("committed"); err != nil {
		return err
	}
	if err := s.cache(p.After); err != nil {
		return err
	}
	if err := s.m.point("cached"); err != nil {
		return err
	}
	return s.cleanup(p)
}

const commitStage = "COMMITTED.stage"

// Commit becomes authoritative only after complete marker bytes were synced,
// closed, and promoted from the exact transaction-owned staging name.
func (s *session) commit(dir string, p plan) error {
	stage := path.Join(dir, commitStage)
	if err := s.write(stage, marker{planHash(p), "COMMITTED", -1}); err != nil {
		return err
	}
	if err := s.m.point("commit-staged"); err != nil {
		return err
	}
	final := path.Join(dir, "COMMITTED")
	if err := s.guard(stage, false); err != nil {
		return err
	}
	if s.exists(final) {
		return fail(5, "commit marker appeared before promotion")
	}
	if err := s.guard(final, true); err != nil {
		return err
	}
	if err := s.root.Rename(stage, final); err != nil {
		return err
	}
	if ok, err := s.hasMarker(dir, "COMMITTED", p, -1); err != nil || !ok {
		if err != nil {
			return err
		}
		return fail(5, "commit promotion missing")
	}
	return s.m.point("commit-promoted")
}
func (s *session) cleanup(p plan) error {
	for i, item := range p.Cleanup {
		if err := s.removeExact(item); err != nil {
			return err
		}
		if err := s.m.point(fmt.Sprintf("cleanup-%d", i)); err != nil {
			return err
		}
	}
	if err := s.prunePackageDirs(p.Cleanup); err != nil {
		return err
	}
	if err := s.mark(txPath(p.Sequence), "CLEANED", p, -1); err != nil {
		return err
	}
	return s.m.point("cleaned")
}

// Only empty, known package directories can be removed. Never recursively delete.
func (s *session) prunePackageDirs(items []inventory) error {
	dirs := map[string]bool{}
	for _, item := range items {
		for d := path.Dir(item.Path); d != stateDir+"/library" && d != "."; d = path.Dir(d) {
			dirs[d] = true
		}
	}
	sorted := []string{}
	for d := range dirs {
		sorted = append(sorted, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(sorted)))
	for _, d := range sorted {
		if s.exists(d) {
			if err := winfs.CheckDir(pathJoin(s.m.config.Root, d)); err != nil {
				return err
			}
			if err := s.root.Remove(d); err != nil {
				return wrap(5, "package directory contains unexpected files or cannot be removed", err)
			}
		}
	}
	return nil
}

func (s *session) pruneImport(p plan) error {
	if p.Operation != "add" {
		return nil
	}
	for id, mod := range p.After.Mods {
		if _, existed := p.Before.Mods[id]; !existed {
			return s.prunePackageDirs(libraryItems(mod))
		}
	}
	return nil
}

type missingAuthorization struct {
	Plan                        string   `json:"plan"`
	Paths                       []string `json:"paths"`
	MissingFileRecoveryExplicit bool     `json:"missingFileRecoveryExplicit"`
}

func (s *session) recover(h history, opt Options, compatible bool) ([]string, error) {
	if h.tail == nil {
		if len(opt.RestoreMissing) > 0 {
			return nil, fail(5, "no unresolved operation authorizes missing-file recovery")
		}
		if !opt.DryRun {
			return nil, s.cache(h.state)
		}
		return nil, nil
	}
	tail := h.tail
	p := tail.plan
	dir := txPath(p.Sequence)
	if tail.committed {
		if len(opt.RestoreMissing) > 0 {
			return nil, fail(5, "committed operation cannot authorize missing-file recreation")
		}
		if compatible {
			if err := s.verifyState(p.After, true); err != nil {
				return nil, err
			}
		} else {
			for _, c := range p.Changes {
				if err := s.expectState(c.Path, c.After); err != nil {
					return nil, err
				}
				if c.Game {
					b := p.After.Baselines[profile.Key(c.Path)]
					if !b.Absent {
						if err := s.expect(baseline(b.SHA256), b.SHA256); err != nil {
							return nil, err
						}
					}
				}
			}
		}
		for _, item := range p.Cleanup {
			if s.exists(item.Path) {
				if err := s.expect(item.Path, item.SHA256); err != nil {
					return nil, err
				}
			}
		}
		if opt.DryRun {
			return nil, nil
		}
		if err := s.cache(p.After); err != nil {
			return nil, err
		}
		return nil, s.cleanup(p)
	}
	if !tail.ready {
		if len(opt.RestoreMissing) > 0 {
			return nil, fail(5, "operation never reached READY")
		}
		return nil, s.abandonPreparation(p, opt.DryRun)
	}
	auth := map[string]bool{}
	for _, v := range opt.RestoreMissing {
		if profile.Target(v) != nil {
			return nil, fail(2, "invalid missing-file target", v)
		}
		key := profile.Key(v)
		if auth[key] {
			return nil, fail(2, "duplicate missing-file target", v)
		}
		auth[key] = true
	}
	if len(auth) > 0 && !compatible {
		return nil, fail(3, "missing-file recreation requires matching core fingerprints")
	}
	allowedExisting := map[string]bool{}
	// A durable authorization survives a crash during recovery, but never grants a
	// new recreation on a later invocation without the explicit CLI option again.
	authPath := dir + "/MISSING_AUTHORIZED"
	if s.exists(authPath) {
		var old missingAuthorization
		if err := read(s, authPath, &old); err != nil {
			return nil, wrap(5, "invalid missing-file authorization", err)
		}
		if old.Plan != planHash(p) || !old.MissingFileRecoveryExplicit {
			return nil, fail(5, "invalid missing-file authorization")
		}
		for _, v := range old.Paths {
			if profile.Target(v) != nil {
				return nil, fail(5, "invalid authorization target")
			}
			allowedExisting[profile.Key(v)] = true
		}
		if len(auth) > 0 && !equal(auth, allowedExisting) {
			return nil, fail(5, "recovery authorization already recorded; reuse the same --restore-missing targets")
		}
	}
	restored := []string{}
	matched := map[string]bool{}
	for i, c := range p.Changes {
		intent, err := s.hasMarker(dir, fmt.Sprintf("APPLY_INTENT_%04d", i), p, i)
		if err != nil {
			return nil, wrap(5, "invalid intent", err)
		}
		if c.Before != "" {
			if err = s.expect(blob(dir, c.Before), c.Before); err != nil {
				return nil, wrap(5, "recovery snapshot invalid", err)
			}
		}
		stageSurvives := false
		if c.After != "" {
			stage := fmt.Sprintf("%s/stage/apply-%04d", dir, i)
			if s.exists(stage) {
				if err := s.expect(stage, c.After); err != nil {
					return nil, wrap(5, "recovery apply stage invalid", err)
				}
				stageSurvives = true
			}
		}
		exists := s.exists(c.Path)
		if !exists {
			if c.Before == "" {
				continue
			}
			if !c.Game {
				continue
			}
			if c.After == "" && intent {
				continue
			}
			key := profile.Key(c.Path)
			if !auth[key] || !intent {
				return nil, fail(5, "missing target is ambiguous; recover --restore-missing requires valid apply intent", c.Path)
			}
			matched[key] = true
			continue
		}
		hash, _, err := s.hash(c.Path)
		if err != nil {
			return nil, wrap(4, "unsafe recovery target", err)
		}
		if auth[profile.Key(c.Path)] {
			if !allowedExisting[profile.Key(c.Path)] || hash != c.Before {
				return nil, fail(4, "--restore-missing cannot overwrite an existing target", c.Path)
			}
			matched[profile.Key(c.Path)] = true
		}
		// Successful replacement consumes the exact transaction-owned stage.
		// If it survives, matching destination bytes came from elsewhere and
		// must not be removed or overwritten during rollback.
		if stageSurvives && hash == c.After && c.After != c.Before {
			return nil, fail(4, "unowned target appeared before staged promotion", c.Path)
		}
		if hash != c.Before && hash != c.After {
			return nil, fail(4, "external drift blocks recovery", c.Path)
		}
		if hash == c.After && c.After != c.Before && !intent {
			return nil, fail(5, "changed target has no apply intent", c.Path)
		}
	}
	for key := range auth {
		if !matched[key] {
			return nil, fail(5, "missing-file target is not authorized by this operation", key)
		}
	}
	if opt.DryRun {
		return nil, nil
	}
	guardProcesses := false
	for _, c := range p.Changes {
		if !c.Game {
			continue
		}
		if !s.exists(c.Path) {
			if c.Before != "" {
				guardProcesses = true
				break
			}
			continue
		}
		hash, _, err := s.hash(c.Path)
		if err != nil {
			return nil, err
		}
		if hash != c.Before {
			guardProcesses = true
			break
		}
	}
	if guardProcesses {
		if err := s.m.processes(); err != nil {
			return nil, err
		}
	}
	if len(auth) > 0 && !s.exists(authPath) {
		paths := append([]string{}, opt.RestoreMissing...)
		sort.Strings(paths)
		if err := s.write(authPath, missingAuthorization{planHash(p), paths, true}); err != nil {
			return nil, err
		}
	}
	for i, c := range p.Changes {
		missing := !s.exists(c.Path)
		if !s.exists(c.Path) && !c.Game {
			continue
		}
		if c.Before == "" {
			if missing {
				continue
			}
			if err := s.expect(c.Path, c.After); err != nil {
				return nil, err
			}
			if err := s.root.Remove(c.Path); err != nil {
				return nil, err
			}
			continue
		}
		intent, err := s.hasMarker(dir, fmt.Sprintf("APPLY_INTENT_%04d", i), p, i)
		if err != nil {
			return nil, wrap(5, "invalid intent", err)
		}
		if s.exists(c.Path) {
			hash, _, err := s.hash(c.Path)
			if err != nil {
				return nil, err
			}
			if hash == c.Before {
				continue
			}
			if hash != c.After {
				return nil, fail(4, "target changed during recovery", c.Path)
			}
		}
		stage := fmt.Sprintf("%s/stage/rollback-%04d", dir, i)
		if s.exists(stage) {
			if err := s.removeOwnedStage(stage); err != nil {
				return nil, err
			}
		}
		if err := s.copy(blob(dir, c.Before), stage, c.Before); err != nil {
			return nil, err
		}
		if err := s.m.point(fmt.Sprintf("rollback-staged-%d", i)); err != nil {
			return nil, err
		}
		if s.exists(c.Path) {
			if err := s.expect(c.Path, c.After); err != nil {
				return nil, err
			}
		} else if !(c.After == "" && intent) && !auth[profile.Key(c.Path)] {
			return nil, fail(5, "target disappeared during recovery", c.Path)
		}
		var promoteErr error
		if missing {
			promoteErr = s.promoteNew(stage, c.Path, c.Before)
		} else {
			promoteErr = s.replace(stage, c.Path, c.Before)
		}
		if promoteErr != nil {
			return nil, promoteErr
		}
		if missing {
			restored = append(restored, c.Path)
		}
		if err := s.m.point(fmt.Sprintf("recovered-%d", i)); err != nil {
			return nil, err
		}
	}
	for _, c := range p.Changes {
		if err := s.expectState(c.Path, c.Before); err != nil {
			return nil, err
		}
	}
	if err := s.pruneImport(p); err != nil {
		return nil, err
	}
	if s.exists(path.Join(dir, commitStage)) {
		if err := s.removeOwnedStage(path.Join(dir, commitStage)); err != nil {
			return nil, err
		}
	}
	if err := s.mark(dir, "ROLLED_BACK", p, -1); err != nil {
		return nil, err
	}
	if err := s.cache(p.Before); err != nil {
		return nil, err
	}
	return restored, nil
}

func (s *session) removeOwnedStage(name string) error {
	if err := s.guard(name, false); err != nil {
		return wrapPath(5, "unsafe transaction staging", name, err)
	}
	info, err := s.root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fail(5, "transaction staging is not a regular file", name)
	}
	return s.root.Remove(name)
}

// PREPARED owns these exact temporary names, including incomplete copies. It
// never owns arbitrary descendants. Promoted baseline objects are retained.
func (s *session) abandonPreparation(p plan, dry bool) error {
	dir := txPath(p.Sequence)
	allowed := map[string]bool{}
	for i, c := range p.Changes {
		if c.Before != "" {
			allowed[blob(dir, c.Before)] = true
		}
		if c.After != "" {
			allowed[blob(dir, c.After)] = true
			allowed[fmt.Sprintf("%s/stage/apply-%04d", dir, i)] = true
		}
	}
	for _, b := range p.Baselines {
		allowed[blob(dir, b.SHA256)] = true
		temp := b.Temporary
		if p.Schema == 1 {
			temp = dir + "/stage/baseline-" + b.SHA256
		}
		allowed[temp] = true
		if s.exists(b.Path) {
			if err := s.expect(b.Path, b.SHA256); err != nil {
				return wrap(5, "promoted baseline invalid", err)
			}
		}
	}
	files := []string{}
	for _, folder := range []string{dir + "/blobs", dir + "/stage"} {
		if !s.exists(folder) {
			continue
		}
		if err := winfs.CheckDir(pathJoin(s.m.config.Root, folder)); err != nil {
			return wrapPath(5, "unsafe preparation directory", folder, err)
		}
		f, err := s.root.Open(folder)
		if err != nil {
			return err
		}
		entries, err := f.ReadDir(-1)
		f.Close()
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := folder + "/" + entry.Name()
			if !allowed[name] || entry.IsDir() {
				return fail(5, "unexpected preparation artifact", name)
			}
			if err := s.guard(name, false); err != nil {
				return wrapPath(5, "unsafe preparation artifact", name, err)
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Size() > 64<<20 {
				return fail(5, "unexpected preparation artifact size", name)
			}
			files = append(files, name)
		}
	}
	sort.Strings(files)
	if dry {
		return nil
	}
	for i, name := range files {
		if err := s.guard(name, false); err != nil {
			return err
		}
		if err := s.root.Remove(name); err != nil {
			return err
		}
		if err := s.m.point(fmt.Sprintf("abort-cleanup-%d", i)); err != nil {
			return err
		}
	}
	for _, folder := range []string{dir + "/blobs", dir + "/stage"} {
		if s.exists(folder) {
			if err := s.root.Remove(folder); err != nil {
				return err
			}
		}
	}
	if err := s.pruneImport(p); err != nil {
		return err
	}
	return s.mark(dir, "ABORTED", p, -1)
}
