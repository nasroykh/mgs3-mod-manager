package manager

import (
	"encoding/json"
	"fmt"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"mgs3mod/internal/winfs"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var seqPattern = regexp.MustCompile(`^[0-9]{20}$`)

func canonical(v any) []byte { b, _ := json.Marshal(v); return b }
func equal(a, b any) bool    { return string(canonical(a)) == string(canonical(b)) }
func planHash(p plan) string { return transaction.Hash(canonical(p)) }
func validState(st State) error {
	if st.Mods == nil || st.Baselines == nil {
		return fmt.Errorf("missing state maps")
	}
	owners := map[string]string{}
	for id, m := range st.Mods {
		if id != m.Manifest.ID || !hashPattern.MatchString(m.Digest) {
			return fmt.Errorf("invalid mod identity")
		}
		if _, err := packagefmt.Parse(canonical(m.Manifest), false); err != nil {
			return err
		}
		if m.Enabled {
			for _, f := range m.Manifest.Files {
				k := profile.Key(f.Target)
				if owners[k] != "" {
					return fmt.Errorf("duplicate owner")
				}
				owners[k] = id
				b, ok := st.Baselines[k]
				if !ok || b.Absent != f.OriginalAbsent || !b.Absent && b.SHA256 != f.OriginalSHA256 {
					return fmt.Errorf("enabled mod lacks correct baseline")
				}
			}
		}
	}
	for key, b := range st.Baselines {
		targetErr := profile.Target(b.Target)
		if b.Absent {
			targetErr = profile.AddedTarget(b.Target)
		}
		validContents := b.Absent && b.SHA256 == "" && b.Bytes == 0 || !b.Absent && hashPattern.MatchString(b.SHA256) && b.Bytes >= 0 && b.Bytes <= 64<<20
		if targetErr != nil || key != profile.Key(b.Target) || !validContents || b.Profile != profile.ID || b.Captured == "" {
			return fmt.Errorf("invalid baseline")
		}
	}
	return nil
}
func expected(st State) map[string]string {
	result := map[string]string{}
	for k, b := range st.Baselines {
		result[k] = b.SHA256
	}
	for _, m := range st.Mods {
		if m.Enabled {
			for _, f := range m.Manifest.Files {
				result[profile.Key(f.Target)] = f.PayloadSHA256
			}
		}
	}
	return result
}
func libraryItems(mod Mod) []inventory {
	base := stateDir + "/library/" + mod.Manifest.ID + "/"
	b := canonical(mod.Manifest)
	items := []inventory{{Path: base + "manifest.json", SHA256: transaction.Hash(b), Bytes: int64(len(b))}}
	for _, f := range mod.Manifest.Files {
		items = append(items, inventory{Path: base + f.Source, SHA256: f.PayloadSHA256, Bytes: f.PayloadBytes})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items
}
func validatePlan(p plan, previous State, seq uint64) error {
	if p.ConditionalRestore && p.Operation != "restore" {
		return fmt.Errorf("conditional restore flag on another operation")
	}
	if (p.Schema != 1 && p.Schema != 2) || p.Sequence != seq || !equal(previous, p.Before) || p.After.Generation != p.Before.Generation+1 {
		return fmt.Errorf("invalid generation chain")
	}
	if err := validState(p.Before); err != nil {
		return err
	}
	if err := validState(p.After); err != nil {
		return err
	}
	added, removed, toggled := 0, 0, 0
	var addedMod, removedMod Mod
	for id, old := range p.Before.Mods {
		next, ok := p.After.Mods[id]
		if !ok {
			removed++
			removedMod = old
			continue
		}
		if old.Digest != next.Digest || !equal(old.Manifest, next.Manifest) {
			return fmt.Errorf("changed immutable package")
		}
		if old.Enabled != next.Enabled {
			toggled++
			if p.Operation == "enable" && !next.Enabled || p.Operation != "enable" && next.Enabled {
				return fmt.Errorf("wrong activation transition")
			}
		}
	}
	for id, next := range p.After.Mods {
		if _, ok := p.Before.Mods[id]; !ok {
			added++
			addedMod = next
			if next.Enabled {
				return fmt.Errorf("import cannot enable")
			}
		}
	}
	switch p.Operation {
	case "add":
		if added != 1 || removed != 0 || toggled != 0 {
			return fmt.Errorf("invalid import")
		}
	case "remove":
		if removed != 1 || added != 0 || toggled != 0 {
			return fmt.Errorf("invalid removal")
		}
	case "enable", "disable":
		if toggled != 1 || added != 0 || removed != 0 {
			return fmt.Errorf("invalid activation")
		}
	case "restore":
		if toggled < 1 || added != 0 || removed != 0 {
			return fmt.Errorf("invalid restore")
		}
	default:
		return fmt.Errorf("unknown operation")
	}
	newBaselines := []inventory{}
	for k, b := range p.Before.Baselines {
		if !equal(b, p.After.Baselines[k]) {
			return fmt.Errorf("baseline changed")
		}
	}
	for k, b := range p.After.Baselines {
		if _, ok := p.Before.Baselines[k]; !ok {
			if p.Operation != "enable" {
				return fmt.Errorf("baseline outside enable")
			}
			found := false
			for _, m := range p.After.Mods {
				if m.Enabled {
					for _, f := range m.Manifest.Files {
						if profile.Key(f.Target) == k && f.OriginalAbsent == b.Absent && f.OriginalSHA256 == b.SHA256 {
							found = true
						}
					}
				}
			}
			if !found {
				return fmt.Errorf("unreferenced new baseline")
			}
			if b.Absent {
				continue
			}
			item := inventory{Path: baseline(b.SHA256), SHA256: b.SHA256, Bytes: b.Bytes}
			if p.Schema == 2 {
				item.Target = b.Target
				item.Temporary = fmt.Sprintf("%s/stage/baseline-%04d", txPath(p.Sequence), len(newBaselines))
				for _, recorded := range p.Baselines {
					if recorded.Target == b.Target {
						item.VerifiedObjectExisted = recorded.VerifiedObjectExisted
						break
					}
				}
			}
			newBaselines = append(newBaselines, item)
		}
	}
	if p.Schema == 1 {
		sort.Slice(newBaselines, func(i, j int) bool { return newBaselines[i].Path < newBaselines[j].Path })
		newBaselines = uniqueInventory(newBaselines)
	} else {
		sort.Slice(newBaselines, func(i, j int) bool { return newBaselines[i].Target < newBaselines[j].Target })
		for i := range newBaselines {
			newBaselines[i].Temporary = fmt.Sprintf("%s/stage/baseline-%04d", txPath(p.Sequence), i)
			if newBaselines[i].VerifiedObjectExisted == nil {
				return fmt.Errorf("missing verified-object state")
			}
		}
	}
	if !equal(newBaselines, p.Baselines) {
		return fmt.Errorf("invalid baseline preparation inventory")
	}
	before, after := expected(p.Before), expected(p.After)
	changes := []change{}
	for k, h := range after {
		old, ok := before[k]
		if !ok {
			old = p.After.Baselines[k].SHA256
		}
		if old != h {
			changes = append(changes, change{p.After.Baselines[k].Target, old, h, true})
		}
	}
	if added == 1 {
		for _, item := range libraryItems(addedMod) {
			changes = append(changes, change{item.Path, "", item.SHA256, false})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	if p.ConditionalRestore {
		expectedChanges := map[string]change{}
		for _, c := range changes {
			expectedChanges[c.Path] = c
		}
		for _, c := range p.Changes {
			want, ok := expectedChanges[c.Path]
			if !ok || !equal(c, want) {
				return fmt.Errorf("invalid conditional restore inventory")
			}
			delete(expectedChanges, c.Path)
		}
	} else if !equal(changes, p.Changes) {
		return fmt.Errorf("invalid change inventory")
	}
	cleanup := []inventory{}
	if removed == 1 {
		cleanup = libraryItems(removedMod)
	}
	if !equal(cleanup, p.Cleanup) {
		return fmt.Errorf("invalid removal inventory")
	}
	return nil
}
func uniqueInventory(items []inventory) []inventory {
	out := []inventory{}
	for _, v := range items {
		if len(out) == 0 || out[len(out)-1].Path != v.Path {
			out = append(out, v)
		}
	}
	return out
}

type pending struct {
	plan             plan
	ready, committed bool
}
type history struct {
	state    State
	sequence uint64
	tail     *pending
}

func (s *session) mark(dir, kind string, p plan, index int) error {
	return s.write(path.Join(dir, kind), marker{planHash(p), kind, index})
}
func (s *session) hasMarker(dir, kind string, p plan, index int) (bool, error) {
	name := path.Join(dir, kind)
	if !s.exists(name) {
		return false, nil
	}
	var got marker
	if err := read(s, name, &got); err != nil {
		return false, err
	}
	if got != (marker{planHash(p), kind, index}) {
		return false, fmt.Errorf("invalid %s marker", kind)
	}
	return true, nil
}
func (s *session) load() (history, error) {
	h := history{state: emptyState()}
	if err := s.checkInit(); err != nil {
		return h, err
	}
	dir := stateDir + "/transactions"
	if err := winfs.CheckDir(pathJoin(s.m.config.Root, dir)); err != nil {
		return h, wrap(5, "unsafe transaction directory", err)
	}
	f, err := s.root.Open(dir)
	if err != nil {
		return h, err
	}
	entries, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		return h, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if !entry.IsDir() || !seqPattern.MatchString(entry.Name()) {
			return h, fail(5, "unexpected transaction entry", entry.Name())
		}
		seq, err := strconv.ParseUint(entry.Name(), 10, 64)
		if err != nil || seq != h.sequence+1 {
			return h, fail(5, "invalid transaction sequence")
		}
		if h.tail != nil {
			return h, fail(5, "transaction after unresolved operation")
		}
		dir = txPath(seq)
		if err := winfs.CheckDir(pathJoin(s.m.config.Root, dir)); err != nil {
			return h, wrap(5, "unsafe transaction", err)
		}
		var p plan
		if err = read(s, dir+"/PREPARED", &p); err != nil {
			return h, wrap(5, "invalid preparation record", err)
		}
		if err = validatePlan(p, h.state, seq); err != nil {
			return h, wrap(5, "invalid plan", err)
		}
		flags := map[string]bool{}
		for _, kind := range []string{"READY", "COMMITTED", "ROLLED_BACK", "ABORTED", "CLEANED"} {
			ok, e := s.hasMarker(dir, kind, p, -1)
			if e != nil {
				return h, wrap(5, "invalid journal", e)
			}
			flags[kind] = ok
		}
		if flags["COMMITTED"] && s.exists(dir+"/"+commitStage) {
			return h, fail(5, "commit marker and staging both exist")
		}
		if flags["COMMITTED"] && flags["ROLLED_BACK"] || flags["ABORTED"] && (flags["READY"] || flags["COMMITTED"] || flags["ROLLED_BACK"]) || flags["COMMITTED"] && !flags["READY"] || flags["ROLLED_BACK"] && !flags["READY"] || flags["CLEANED"] && !flags["COMMITTED"] {
			return h, fail(5, "impossible transaction transition")
		}
		known := map[string]bool{"PREPARED": true, "blobs": true, "stage": true, commitStage: true}
		for k := range flags {
			known[k] = true
		}
		known["MISSING_AUTHORIZED"] = true
		for i := range p.Changes {
			applied := false
			for _, prefix := range []string{"APPLY_INTENT", "APPLIED"} {
				name := fmt.Sprintf("%s_%04d", prefix, i)
				known[name] = true
				ok, e := s.hasMarker(dir, name, p, i)
				if e != nil {
					return h, wrap(5, "invalid apply marker", e)
				}
				if ok && !flags["READY"] {
					return h, fail(5, "apply before READY")
				}
				if prefix == "APPLIED" && ok {
					applied = true
					intent, _ := s.hasMarker(dir, fmt.Sprintf("APPLY_INTENT_%04d", i), p, i)
					if !intent {
						return h, fail(5, "applied without intent")
					}
				}
				if flags["COMMITTED"] && !ok {
					return h, fail(5, "commit lacks completed apply records")
				}
			}
			if applied && p.Changes[i].After != "" && s.exists(fmt.Sprintf("%s/stage/apply-%04d", dir, i)) {
				return h, fail(5, "applied change retains its apply stage")
			}
		}
		if err := s.validateAuthorization(dir, p, flags["READY"], flags["COMMITTED"]); err != nil {
			return h, wrap(5, "invalid authorization record", err)
		}
		f, e := s.root.Open(dir)
		if e != nil {
			return h, e
		}
		inside, e := f.ReadDir(-1)
		f.Close()
		if e != nil {
			return h, e
		}
		for _, v := range inside {
			if !known[v.Name()] {
				return h, fail(5, "unknown journal entry", v.Name())
			}
			if v.Name() == commitStage {
				name := dir + "/" + commitStage
				if err := s.guard(name, false); err != nil {
					return h, wrapPath(5, "unsafe commit staging", name, err)
				}
			}
		}
		h.sequence = seq
		if flags["COMMITTED"] {
			h.state = p.After
			if !flags["CLEANED"] {
				h.tail = &pending{p, true, true}
			}
		} else if !flags["ROLLED_BACK"] && !flags["ABORTED"] {
			h.tail = &pending{p, flags["READY"], false}
		}
	}
	return h, nil
}

func (s *session) validateAuthorization(dir string, p plan, ready, committed bool) error {
	file := dir + "/MISSING_AUTHORIZED"
	if !s.exists(file) {
		return nil
	}
	if !ready || committed {
		return fmt.Errorf("authorization outside unresolved rollback")
	}
	var record missingAuthorization
	if err := read(s, file, &record); err != nil {
		return err
	}
	if record.Plan != planHash(p) || !record.MissingFileRecoveryExplicit || len(record.Paths) == 0 || !sort.StringsAreSorted(record.Paths) {
		return fmt.Errorf("invalid missing-file authorization")
	}
	seen := map[string]bool{}
	for _, target := range record.Paths {
		key := profile.Key(target)
		if profile.Target(target) != nil || seen[key] {
			return fmt.Errorf("invalid or duplicate authorized target")
		}
		seen[key] = true
		found := false
		for i, c := range p.Changes {
			if c.Game && profile.Key(c.Path) == key {
				intent, err := s.hasMarker(dir, fmt.Sprintf("APPLY_INTENT_%04d", i), p, i)
				if err != nil {
					return err
				}
				found = intent
				break
			}
		}
		if !found {
			return fmt.Errorf("authorization lacks target apply intent")
		}
	}
	return nil
}
func pathJoin(root, rel string) string {
	return root + string(os.PathSeparator) + strings.ReplaceAll(rel, "/", string(os.PathSeparator))
}
func clone(st State) State { var next State; json.Unmarshal(canonical(st), &next); return next }

// State cache is never read as authority. It may be absent or torn after a crash.
func (s *session) cache(st State) error {
	p := stateDir + "/state.json"
	if s.exists(p) {
		if err := s.guard(p, false); err != nil {
			return err
		}
		if err := s.root.Remove(p); err != nil {
			return err
		}
	}
	return s.write(p, st)
}
