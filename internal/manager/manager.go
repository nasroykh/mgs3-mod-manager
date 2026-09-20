package manager

import (
	"errors"
	"fmt"
	"math"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"mgs3mod/internal/winfs"
	"os"
	"sort"
	"time"
)

func (s *session) verifyState(st State, targets bool) error {
	if err := validState(st); err != nil {
		return wrap(5, "invalid state", err)
	}
	for _, b := range st.Baselines {
		if err := s.expect(baseline(b.SHA256), b.SHA256); err != nil {
			return wrap(5, "baseline corrupted", err)
		}
	}
	for id, mod := range st.Mods {
		pkg, err := packagefmt.Load(pathJoin(s.m.config.Root, stateDir+"/library/"+id), false)
		if err != nil {
			return wrapPath(5, "stored package corrupted", stateDir+"/library/"+id, err)
		}
		if packagefmt.Digest(pkg) != mod.Digest || !equal(pkg.Manifest, mod.Manifest) {
			return fail(5, "stored package identity mismatch", id)
		}
	}
	if targets {
		for key, hash := range expected(st) {
			if err := s.expect(st.Baselines[key].Target, hash); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Run(command, arg string, opt Options) (Result, error) {
	result := Result{Command: command, DryRun: opt.DryRun}
	for _, target := range opt.RestoreMissing {
		if profile.Target(target) != nil {
			return result, fail(2, "invalid missing-file target", target)
		}
	}
	if command == "init" {
		return m.init(opt.DryRun)
	}
	compatibleErr := m.compatible()
	result.Compatible = compatibleErr == nil
	_, statErr := os.Lstat(pathJoin(m.config.Root, stateDir))
	if command == "doctor" && errors.Is(statErr, os.ErrNotExist) {
		result.Message = "installation inspected; manager not initialized"
		if compatibleErr != nil {
			return result, compatibleErr
		}
		return result, m.processes()
	}
	s, err := m.open(false)
	if err != nil {
		return result, err
	}
	defer s.close()
	compatibleErr = m.compatible()
	result.Compatible = compatibleErr == nil
	if command == "recover" && !s.exists(stateDir+"/INIT_COMMITTED") {
		if compatibleErr != nil {
			return result, compatibleErr
		}
		if len(opt.RestoreMissing) > 0 {
			return result, fail(5, "initialization has no target apply intent")
		}
		var got initRecord
		if err = read(s, stateDir+"/INIT_PREPARED", &got); err != nil {
			return result, wrap(5, "invalid initialization evidence", err)
		}
		if !equal(got, s.initSpec()) {
			return result, fail(5, "initialization identity differs")
		}
		if !opt.DryRun {
			err = s.finishInit()
		}
		result.Message = "finish interrupted initialization"
		result.Initialized = err == nil && !opt.DryRun
		return result, err
	}
	h, err := s.load()
	if err != nil {
		return result, err
	}
	result.Initialized = true
	result.State = &h.state
	if command == "recover" {
		restored, e := s.recover(h, opt, result.Compatible)
		result.MissingFileRecoveryExplicit = len(opt.RestoreMissing) > 0
		if e != nil {
			result.Message = "recovery failed"
			current, readErr := s.load()
			if readErr != nil {
				result.State = nil
				result.RecoveryRequired = true
				return result, wrap(5, result.Message, e)
			}
			result.State = &current.state
			result.RecoveryRequired = current.tail != nil
			if h.tail != nil && h.tail.committed && current.state.Generation == h.tail.plan.After.Generation {
				result.Committed = true
			}
			if !result.RecoveryRequired {
				result.Message = "recovery failed without an unresolved journal"
				return result, wrap(1, result.Message, e)
			}
			var detail *Error
			if !errors.As(e, &detail) {
				e = wrap(5, result.Message, e)
			}
			return result, e
		}
		if opt.DryRun {
			result.Message = "recovery preflight complete"
		} else {
			result.Paths = restored
			result.Message = "recovery resolved"
		}
		return result, nil
	}
	if h.tail != nil {
		result.RecoveryRequired = true
		if command == "status" {
			result.Targets, _ = s.inspectTargets(h.state)
		}
		result.Message = "unresolved transaction; run recover"
		return result, fail(5, result.Message)
	}
	switch command {
	case "list":
		result.Message = "stored mods"
		return result, nil
	case "doctor", "status", "verify":
		result.Message = "installation and managed state inspected"
		if err = s.verifyState(h.state, command != "status"); err != nil {
			return result, err
		}
		if command == "status" {
			result.Targets, err = s.inspectTargets(h.state)
			if err != nil {
				return result, err
			}
		}
		if compatibleErr != nil {
			result.Issues = append(result.Issues, compatibleErr.Error())
			if command != "status" {
				return result, compatibleErr
			}
		}
		if command == "doctor" {
			return result, m.processes()
		}
		return result, nil
	}
	if command != "restore" && compatibleErr != nil {
		return result, compatibleErr
	}
	conditional := command == "restore" && compatibleErr != nil
	if !conditional {
		if err = s.verifyState(h.state, true); err != nil {
			return result, err
		}
	}
	p, inputs, err := s.build(h, command, arg)
	if err != nil {
		return result, err
	}
	if p == nil {
		result.Message = "already in requested state"
		return result, nil
	}
	if conditional {
		result.Issues = append(result.Issues, compatibleErr.Error())
		p.ConditionalRestore = true
		inputs = map[string]input{}
		changes := make([]change, 0, len(p.Changes))
		for _, c := range p.Changes {
			actual, _, e := s.hash(c.Path)
			if e != nil {
				return result, wrapPath(4, "conditional restore target", c.Path, e)
			}
			if actual != c.Before && actual != c.After {
				return result, fail(4, "updated target cannot be restored", c.Path)
			}
			if actual == c.After {
				continue
			}
			changes = append(changes, c)
			inputs[c.Before] = input{Path: c.Path}
			inputs[c.After] = input{Path: baseline(c.After)}
		}
		p.Changes = changes
		if err = validatePlan(*p, h.state, p.Sequence); err != nil {
			return result, wrap(5, "conditional restore plan", err)
		}
	}
	for _, c := range p.Changes {
		if c.Game {
			if err = m.processes(); err != nil {
				return result, err
			}
			break
		}
	}
	result.Message = "operation preflight complete"
	for _, c := range p.Changes {
		result.Paths = append(result.Paths, c.Path)
	}
	for _, c := range p.Cleanup {
		result.Paths = append(result.Paths, c.Path)
	}
	budget, err := s.requiredSpace(*p, inputs)
	if err != nil {
		return result, err
	}
	result.RequiredBytes = budget
	available, e := winfs.FreeBytes(m.config.Root)
	if e != nil {
		return result, e
	}
	if available < budget {
		return result, fail(1, "insufficient free space for backups and staging")
	}
	if err = s.preflightAccess(*p); err != nil {
		return result, err
	}
	if opt.DryRun {
		return result, nil
	}
	if err = s.execute(*p, inputs); err != nil {
		result.Message = "operation interrupted; files may be partially changed; run recover"
		current, readErr := s.load()
		if readErr != nil {
			result.State = nil
			result.RecoveryRequired = true
		} else {
			result.State = &current.state
			result.RecoveryRequired = current.tail != nil
			if current.state.Generation == p.After.Generation {
				result.Committed = true
				result.Message = "operation committed; finalization incomplete; run recover"
			}
		}
		if !result.RecoveryRequired {
			result.Message = "operation failed without an unresolved journal"
			if result.Committed {
				result.Message = "operation committed and finalized; final reporting failed"
			}
			return result, wrap(1, result.Message, err)
		}
		return result, wrap(5, result.Message, err)
	}
	result.State = &p.After
	result.Committed = true
	result.Message = "operation complete"
	return result, nil
}

func (s *session) preflightAccess(p plan) error {
	dirs := []string{stateDir, stateDir + "/transactions"}
	if len(p.Baselines) > 0 {
		dirs = append(dirs, stateDir+"/baseline")
	}
	if p.Operation == "add" {
		dirs = append(dirs, stateDir+"/library")
	}
	for _, dir := range dirs {
		if err := winfs.CheckCreateAccess(pathJoin(s.m.config.Root, dir)); err != nil {
			return wrapPath(1, "cannot create manager files", dir, err)
		}
	}
	if s.exists(stateDir + "/state.json") {
		if err := winfs.CheckReplaceAccess(s.m.config.Root, stateDir+"/state.json"); err != nil {
			return wrapPath(1, "cannot refresh state cache", stateDir+"/state.json", err)
		}
	}
	for _, c := range p.Changes {
		if c.Game {
			if err := winfs.CheckReplaceAccess(s.m.config.Root, c.Path); err != nil {
				return wrapPath(1, "target replacement denied", c.Path, err)
			}
		}
	}
	for _, item := range p.Cleanup {
		if err := winfs.CheckReplaceAccess(s.m.config.Root, item.Path); err != nil {
			return wrapPath(1, "package removal denied", item.Path, err)
		}
	}
	return nil
}

func (s *session) inspectTargets(st State) ([]TargetStatus, error) {
	targets := []TargetStatus{}
	var first error
	owners := map[string]string{}
	for id, mod := range st.Mods {
		if mod.Enabled {
			for _, f := range mod.Manifest.Files {
				owners[profile.Key(f.Target)] = id
			}
		}
	}
	hashes := expected(st)
	keys := []string{}
	for key := range hashes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := st.Baselines[key].Target
		actual, _, err := s.hash(path)
		item := TargetStatus{Path: path, Owner: owners[key], ExpectedSHA256: hashes[key], ActualSHA256: actual}
		if err != nil {
			item.Issue = err.Error()
		} else if actual != hashes[key] {
			item.Issue = "external drift"
		}
		if item.Issue != "" && first == nil {
			first = fail(4, item.Issue, path)
		}
		targets = append(targets, item)
	}
	return targets, first
}

func (s *session) requiredSpace(p plan, inputs map[string]input) (uint64, error) {
	budget := uint64(64 << 20)
	sizes := map[string]uint64{}
	add := func(n uint64) error {
		if math.MaxUint64-budget < n {
			return fail(2, "storage budget overflow")
		}
		budget += n
		return nil
	}
	for hash, src := range inputs {
		var n int64
		if src.Path != "" {
			actual, size, err := s.hash(src.Path)
			if err != nil {
				return 0, err
			}
			if actual != hash {
				return 0, fail(4, "preparation source contents differ", src.Path)
			}
			n = size
		} else {
			n = int64(len(src.Data))
		}
		if n < 0 || n > 64<<20 {
			return 0, fail(2, "file exceeds texture limit")
		}
		sizes[hash] = uint64(n)
		if err := add(uint64(n)); err != nil {
			return 0, err
		} // Immutable transaction blob.
	}
	for _, c := range p.Changes {
		size, ok := sizes[c.After]
		if !ok {
			return 0, fmt.Errorf("missing staged source")
		}
		if err := add(size); err != nil {
			return 0, err
		} // One apply stage per path, even for duplicate content.
		if c.Before != "" {
			size, ok = sizes[c.Before]
			if !ok {
				return 0, fmt.Errorf("missing recovery source")
			}
			if err := add(size); err != nil {
				return 0, err
			}
		}
	}
	for _, b := range p.Baselines {
		if !s.exists(b.Path) {
			if err := add(uint64(b.Bytes)); err != nil {
				return 0, err
			}
		}
	}
	return budget, nil
}

func (s *session) build(h history, command, arg string) (*plan, map[string]input, error) {
	next := clone(h.state)
	inputs := map[string]input{}
	cleanup := []inventory{}
	newBaselines := []inventory{}
	changes := []change{}
	switch command {
	case "add":
		pkg, err := packagefmt.Load(arg, false)
		if err != nil {
			return nil, nil, wrapPath(packagefmt.ErrorCode(err), "cannot load package", arg, err)
		}
		digest := packagefmt.Digest(pkg)
		if old, ok := next.Mods[pkg.Manifest.ID]; ok {
			if old.Digest == digest {
				return nil, nil, nil
			}
			return nil, nil, fail(4, "ID already stores different content; remove it before adding")
		}
		if s.exists(stateDir + "/library/" + pkg.Manifest.ID) {
			return nil, nil, fail(5, "library ID has uncommitted or unexpected files")
		}
		for _, f := range pkg.Manifest.Files {
			key := profile.Key(f.Target)
			if b, ok := next.Baselines[key]; ok {
				if b.SHA256 != f.OriginalSHA256 {
					return nil, nil, fail(4, "package original differs from captured baseline", f.Target)
				}
			} else {
				if err = s.expect(f.Target, f.OriginalSHA256); err != nil {
					return nil, nil, err
				}
			}
		}
		mod := Mod{Manifest: pkg.Manifest, Digest: digest}
		next.Mods[pkg.Manifest.ID] = mod
		for _, item := range libraryItems(mod) {
			changes = append(changes, change{item.Path, "", item.SHA256, false})
		}
		b := canonical(pkg.Manifest)
		inputs[transaction.Hash(b)] = input{Data: b}
		for _, f := range pkg.Manifest.Files {
			inputs[f.PayloadSHA256] = input{Data: pkg.Payloads[f.Source]}
		}
	case "enable", "disable", "remove":
		mod, ok := next.Mods[arg]
		if !ok {
			return nil, nil, fail(2, "unknown mod ID", arg)
		}
		if command == "enable" && mod.Enabled || command == "disable" && !mod.Enabled {
			return nil, nil, nil
		}
		if command == "enable" {
			owners := map[string]string{}
			for id, m := range next.Mods {
				if m.Enabled {
					for _, f := range m.Manifest.Files {
						owners[profile.Key(f.Target)] = id
					}
				}
			}
			for _, f := range mod.Manifest.Files {
				key := profile.Key(f.Target)
				if owner := owners[key]; owner != "" {
					return nil, nil, fail(4, "target owned by "+owner, f.Target)
				}
				if err := s.expect(f.Target, f.OriginalSHA256); err != nil {
					return nil, nil, err
				}
				if b, ok := next.Baselines[key]; ok {
					if b.SHA256 != f.OriginalSHA256 {
						return nil, nil, fail(4, "baseline mismatch", f.Target)
					}
				} else {
					_, n, err := s.hash(f.Target)
					if err != nil {
						return nil, nil, err
					}
					b := Baseline{f.Target, f.OriginalSHA256, n, time.Now().UTC().Format(time.RFC3339Nano), profile.ID}
					next.Baselines[key] = b
					newBaselines = append(newBaselines, inventory{Path: baseline(b.SHA256), SHA256: b.SHA256, Bytes: b.Bytes, Target: b.Target})
					inputs[b.SHA256] = input{Path: b.Target}
				}
			}
			mod.Enabled = true
			next.Mods[arg] = mod
		} else if command == "disable" {
			mod.Enabled = false
			next.Mods[arg] = mod
		} else {
			cleanup = libraryItems(mod)
			delete(next.Mods, arg)
		}
	case "restore":
		changed := false
		for id, m := range next.Mods {
			if m.Enabled {
				m.Enabled = false
				next.Mods[id] = m
				changed = true
			}
		}
		if !changed {
			return nil, nil, nil
		}
	default:
		return nil, nil, fail(2, "unknown command")
	}
	before, after := expected(h.state), expected(next)
	for key, hash := range after {
		old, ok := before[key]
		if !ok {
			old = next.Baselines[key].SHA256
		}
		if old == hash {
			continue
		}
		target := next.Baselines[key].Target
		changes = append(changes, change{target, old, hash, true})
		inputs[old] = input{Path: target}
		if hash == next.Baselines[key].SHA256 {
			inputs[hash] = input{Path: baseline(hash)}
		} else {
			found := false
			for id, mod := range next.Mods {
				if mod.Enabled {
					for _, f := range mod.Manifest.Files {
						if profile.Key(f.Target) == key {
							inputs[hash] = input{Path: stateDir + "/library/" + id + "/" + f.Source}
							found = true
						}
					}
				}
			}
			if !found {
				return nil, nil, fmt.Errorf("no replacement source")
			}
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	sort.Slice(newBaselines, func(i, j int) bool { return newBaselines[i].Target < newBaselines[j].Target })
	next.Generation++
	p := &plan{Schema: 2, Sequence: h.sequence + 1, Operation: command, Before: h.state, After: next, Changes: changes, Cleanup: cleanup, Baselines: newBaselines}
	for i := range p.Baselines {
		item := &p.Baselines[i]
		item.Temporary = fmt.Sprintf("%s/stage/baseline-%04d", txPath(p.Sequence), i)
		existed := s.exists(item.Path)
		if existed {
			if err := s.expect(item.Path, item.SHA256); err != nil {
				return nil, nil, wrap(5, "existing baseline is corrupt", err)
			}
		}
		item.VerifiedObjectExisted = &existed
	}
	if len(canonical(p)) > transaction.MaxRecord-256 {
		return nil, nil, fail(2, "operation exceeds the 8 MiB journal record limit")
	}
	if err := validatePlan(*p, h.state, p.Sequence); err != nil {
		return nil, nil, wrap(1, "generated plan invalid", err)
	}
	return p, inputs, nil
}
