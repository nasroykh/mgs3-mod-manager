package gui

import (
	"errors"
	"fmt"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/profile"
	"path/filepath"
	"sort"
	"strings"
)

// Step states, in the order a step goes through them.
const (
	StepWaiting = "waiting"
	StepRunning = "running"
	StepDone    = "done"
	StepSkipped = "skipped" // the manager found nothing to do
	StepFailed  = "failed"
)

// Step is one line of an action's step list.
type Step struct {
	Label string `json:"label"`
	State string `json:"state"`
}

// Event reports an action's progress: "plan" with the whole step list
// (sent again when the list grows), "step" when one step changes state, and
// "finished" last with the outcome.
type Event struct {
	Action  string   `json:"action"`
	Kind    string   `json:"kind"`
	Steps   []Step   `json:"steps,omitempty"`
	Index   int      `json:"index"`
	State   string   `json:"state,omitempty"`
	Outcome *Outcome `json:"outcome,omitempty"`
}

// Outcome is the result of an action: the steps done, and the failed step.
type Outcome struct {
	OK      bool     `json:"ok"`
	Steps   []string `json:"steps"`
	Failed  string   `json:"failed,omitempty"`
	Summary string   `json:"summary"`
	Details string   `json:"details,omitempty"`
}

// action is one planned step: a manager command, or only a local effect.
type action struct {
	label   string
	command string
	arg     string
	dryRun  bool
	before  func() error               // runs first; an error fails the step
	after   func(manager.Result) error // runs after the command succeeded
	undo    func()                     // runs when the step fails
}

// SetProgress sets the callback that receives every Event. It is called on
// the goroutine running the action.
func (s *Service) SetProgress(progress func(Event)) { s.progress = progress }

// Busy reports whether an action is running.
func (s *Service) Busy() bool { return s.running.Load() }

// flow runs one action and reports its progress.
type flow struct {
	s      *Service
	name   string
	m      *manager.Manager
	steps  []Step
	acts   []action
	out    Outcome
	failed bool
}

func (f *flow) emit(e Event) {
	if f.s.progress != nil {
		e.Action = f.name
		f.s.progress(e)
	}
}

// plan appends steps and sends the whole list.
func (f *flow) plan(acts ...action) {
	for _, a := range acts {
		f.acts = append(f.acts, a)
		f.steps = append(f.steps, Step{Label: a.label, State: StepWaiting})
	}
	f.emit(Event{Kind: "plan", Steps: append([]Step(nil), f.steps...)})
}

func (f *flow) set(i int, state string) {
	f.steps[i].State = state
	f.emit(Event{Kind: "step", Index: i, State: state})
}

// runFrom runs the planned steps from index i and stops at the first error.
// It returns the last manager result.
func (f *flow) runFrom(i int) (manager.Result, bool) {
	var last manager.Result
	for ; i < len(f.acts); i++ {
		a := f.acts[i]
		f.set(i, StepRunning)
		var err error
		if a.before != nil {
			err = a.before()
		}
		r := manager.Result{}
		if err == nil && a.command != "" {
			r, err = f.m.Run(a.command, a.arg, manager.Options{DryRun: a.dryRun})
		}
		if err == nil && a.after != nil {
			err = a.after(r)
		}
		if err != nil {
			if a.undo != nil {
				a.undo()
			}
			f.set(i, StepFailed)
			f.fail(a, err)
			return r, false
		}
		last = r
		if a.command != "" && !a.dryRun && r.Message == alreadyDone {
			f.set(i, StepSkipped)
			continue
		}
		f.set(i, StepDone)
		if a.command != "list" && a.command != "doctor" && !a.dryRun {
			f.out.Steps = append(f.out.Steps, a.label)
		}
	}
	return last, true
}

func (f *flow) fail(a action, err error) {
	f.failed = true
	f.out.Failed = a.label
	p := Explain(err)
	f.out.Summary = p.Summary
	where := a.label
	if a.command != "" {
		where += " (" + strings.TrimSpace(a.command+" "+a.arg) + ")"
	}
	f.out.Details = "step: " + where + "\n" + p.Details
}

// stop fails the action before any step list exists.
func (f *flow) stop(label string, err error) Outcome {
	f.plan(action{label: label})
	f.set(len(f.steps)-1, StepRunning)
	f.set(len(f.steps)-1, StepFailed)
	f.fail(action{label: label}, err)
	return f.finish("")
}

func (f *flow) finish(summary string) Outcome {
	if !f.failed {
		f.out.OK = true
		f.out.Summary = summary
	}
	out := f.out
	f.emit(Event{Kind: "finished", Outcome: &out})
	return out
}

var errBusy = errors.New("another action is still running")

// alreadyDone is the manager's message for a command with nothing to do.
const alreadyDone = "already in requested state"

// begin takes the one-action lock and opens the manager.
func (s *Service) begin(name, root string) (*flow, bool) {
	f := &flow{s: s, name: name, out: Outcome{Steps: []string{}}}
	if !s.busy.TryLock() {
		f.stop("Start", errBusy)
		return f, false
	}
	s.running.Store(true)
	m, err := s.manager(root)
	if err != nil {
		s.end()
		f.stop("Open the game folder", err)
		return f, false
	}
	f.m = m
	return f, true
}

func (s *Service) end() {
	s.running.Store(false)
	s.busy.Unlock()
}

// readState returns the stored state, or nil when the manager is not set up.
func readState(m *manager.Manager, root string) (*manager.State, error) {
	if notInitialized(root) {
		return nil, nil
	}
	r, err := m.Run("list", "", manager.Options{})
	if err != nil {
		return nil, err
	}
	return r.State, nil
}

// planInstall lists what Install does from the stored state (nil when the
// manager is not set up) and the kit. Steps with nothing to do are left out.
func planInstall(st *manager.State, kit map[string]kitPackage) []action {
	acts := []action{}
	mods := map[string]manager.Mod{}
	if st == nil {
		acts = append(acts, action{label: "Set up the manager in the game folder", command: "init"})
	} else {
		mods = st.Mods
	}
	loader, stored := mods[profile.LoaderID]
	if !stored {
		acts = append(acts, action{label: "Store the ASI loader " + kit[profile.LoaderID].Version, command: "add", arg: kit[profile.LoaderID].Path})
	}
	if !stored || !loader.Enabled {
		acts = append(acts, action{label: "Turn on the ASI loader", command: "enable", arg: profile.LoaderID})
	}
	for _, id := range pluginIDs {
		k := kit[id]
		name := displayNames[id]
		if q, ok := mods[qcamoID]; id == faceID && ok && q.Enabled {
			acts = append(acts, action{label: "Turn off QCamo " + q.Manifest.Version + " (kept in the library)", command: "disable", arg: qcamoID})
		}
		mod, stored := mods[id]
		enabled := stored && mod.Enabled
		if stored && mod.Digest != k.Digest {
			if mod.Enabled {
				acts = append(acts, action{label: "Turn off " + name + " " + mod.Manifest.Version, command: "disable", arg: id})
			}
			acts = append(acts, action{label: "Remove " + name + " " + mod.Manifest.Version, command: "remove", arg: id})
			stored, enabled = false, false
		}
		if !stored {
			acts = append(acts, action{label: "Store " + name + " " + k.Version, command: "add", arg: k.Path})
		}
		if !enabled {
			acts = append(acts,
				action{label: "Check " + name + " " + k.Version, command: "enable", arg: id, dryRun: true},
				action{label: "Turn on " + name + " " + k.Version, command: "enable", arg: id})
		}
	}
	return append(acts, action{label: "Verify", command: "verify"})
}

// Install installs or updates the Delta controls kit: the ASI loader,
// fpv-move and the QCamo face paint fork. It stops at the first error; each
// manager command is its own transaction, so what was done stays consistent.
func (s *Service) Install(root string) Outcome {
	f, ok := s.begin("install", root)
	if !ok {
		return f.out
	}
	defer s.end()
	f.plan(action{label: "Find the kit packages"}, action{label: "Check the game folder", command: "doctor"})
	f.set(0, StepRunning)
	kit, err := s.kit()
	if err == nil {
		for _, id := range kitIDs {
			if _, ok := kit[id]; !ok {
				err = fmt.Errorf("%s package not found beside mgs3mod-gui.exe; extract the whole kit zip into one folder", id)
				break
			}
		}
	}
	if err != nil {
		f.set(0, StepFailed)
		f.fail(f.acts[0], err)
		return f.finish("")
	}
	f.set(0, StepDone)
	if _, ok := f.runFrom(1); !ok {
		return f.finish("")
	}
	st, err := readState(f.m, root)
	if err != nil {
		return f.stop("Read the installed mods", err)
	}
	acts := planInstall(st, kit)
	for i := range acts {
		if acts[i].command == "disable" && acts[i].arg == qcamoID {
			// Remember it first, so that Uninstall turns QCamo back on; forget
			// it again when the disable failed or found QCamo already off.
			acts[i].before = func() error { return s.markQCamo(root, true) }
			acts[i].after = func(r manager.Result) error {
				if r.Message == alreadyDone {
					return s.markQCamo(root, false)
				}
				return nil
			}
			acts[i].undo = func() { s.markQCamo(root, false) }
		}
	}
	first := len(f.acts)
	f.plan(acts...)
	if _, ok := f.runFrom(first); !ok {
		return f.finish("")
	}
	if len(f.out.Steps) == 1 {
		return f.finish("Already installed: " + describe(map[string]kitPackage{fpvID: kit[fpvID], faceID: kit[faceID]}) + ". Every file checked.")
	}
	return f.finish("Installed: " + describe(map[string]kitPackage{fpvID: kit[fpvID], faceID: kit[faceID]}) + ". Start the game as you normally do.")
}

// planUninstall lists what Uninstall does: remove both plugins, turn QCamo
// back on when Install turned it off (restoreQCamo), then remove the ASI
// loader unless another enabled mod still needs it. note says why the loader
// stays or that QCamo comes back.
func planUninstall(st *manager.State, restoreQCamo bool) ([]action, string) {
	if st == nil {
		return nil, ""
	}
	acts := []action{}
	after := map[string]manager.Mod{}
	for id, mod := range st.Mods {
		after[id] = mod
	}
	for _, id := range []string{faceID, fpvID} {
		mod, stored := st.Mods[id]
		if !stored {
			continue
		}
		name := displayNames[id] + " " + mod.Manifest.Version
		if mod.Enabled {
			acts = append(acts, action{label: "Turn off " + name, command: "disable", arg: id})
		}
		acts = append(acts, action{label: "Remove " + name, command: "remove", arg: id})
		delete(after, id)
	}
	note := ""
	if q, ok := after[qcamoID]; restoreQCamo && ok && !q.Enabled {
		acts = append(acts, action{label: "Turn QCamo " + q.Manifest.Version + " back on", command: "enable", arg: qcamoID})
		q.Enabled = true
		after[qcamoID] = q
		note = " QCamo " + q.Manifest.Version + " is on again."
	}
	if loader, stored := after[profile.LoaderID]; stored {
		if users := loaderUsers(&manager.State{Mods: after}); len(users) > 0 {
			verb := " still uses it."
			if len(users) > 1 {
				verb = " still use it."
			}
			note += " The ASI loader stays: " + strings.Join(users, ", ") + verb
		} else {
			if loader.Enabled {
				acts = append(acts, action{label: "Turn off the ASI loader", command: "disable", arg: profile.LoaderID})
			}
			acts = append(acts, action{label: "Remove the ASI loader", command: "remove", arg: profile.LoaderID})
		}
	}
	return append(acts, action{label: "Verify", command: "verify"}), note
}

// Preview is what Uninstall would do, for the confirmation dialog.
type Preview struct {
	Steps   []string `json:"steps"`
	Note    string   `json:"note,omitempty"`
	Problem *Problem `json:"problem,omitempty"`
}

// PlanUninstall returns Uninstall's steps without changing anything.
func (s *Service) PlanUninstall(root string) Preview {
	p := Preview{Steps: []string{}}
	m, err := s.manager(root)
	var st *manager.State
	if err == nil {
		st, err = readState(m, root)
	}
	if err != nil {
		problem := Explain(err)
		p.Problem = &problem
		return p
	}
	acts, note := planUninstall(st, s.qcamoMarked(root))
	for _, a := range acts {
		p.Steps = append(p.Steps, a.label)
	}
	p.Note = strings.TrimSpace(note)
	return p
}

// Uninstall removes fpv-move and the face paint fork, turns QCamo back on
// when Install turned it off, then removes the ASI loader unless another mod
// that is turned on still needs it. It does not run doctor first: each
// manager command makes its own checks, and disabling an ASI plugin is
// allowed even when the loader files drifted.
func (s *Service) Uninstall(root string) Outcome {
	f, ok := s.begin("uninstall", root)
	if !ok {
		return f.out
	}
	defer s.end()
	f.plan(action{label: "Read the installed mods", command: "list"})
	if notInitialized(root) {
		f.set(0, StepSkipped)
		return f.finish("Nothing to remove: the manager is not set up in this game folder.")
	}
	r, ok := f.runFrom(0)
	if !ok {
		return f.finish("")
	}
	marked := s.qcamoMarked(root)
	acts, note := planUninstall(r.State, marked)
	if marked {
		// Forget the mark once QCamo is back on, or now when there is nothing
		// to turn back on (QCamo was removed or switched on meanwhile).
		forget := func(manager.Result) error { return s.markQCamo(root, false) }
		step := -1
		for i, a := range acts {
			if a.command == "enable" && a.arg == qcamoID {
				step = i
			}
		}
		if step >= 0 {
			acts[step].after = forget
		} else if err := forget(manager.Result{}); err != nil {
			return f.stop("Forget that QCamo was on", err)
		}
	}
	f.plan(acts...)
	if _, ok := f.runFrom(1); !ok {
		return f.finish("")
	}
	if len(acts) == 1 {
		return f.finish("Nothing of Delta controls to remove; every managed file checked." + note)
	}
	return f.finish("Delta controls removed; the original files are back." + note)
}

// loaderUsers lists the other enabled mods with ASI plugins.
func loaderUsers(st *manager.State) []string {
	users := []string{}
	for id, mod := range st.Mods {
		if id == profile.LoaderID || !mod.Enabled {
			continue
		}
		for _, file := range mod.Manifest.Files {
			if profile.PluginTarget(file.Target) == nil {
				users = append(users, id)
				break
			}
		}
	}
	sort.Strings(users)
	return users
}

// SetEnabled turns one stored mod on or off.
func (s *Service) SetEnabled(root, id string, on bool) Outcome {
	name := "disable"
	a := action{label: "Turn off " + id, command: "disable", arg: id}
	if on {
		name = "enable"
		a = action{label: "Turn on " + id, command: "enable", arg: id}
	}
	if id == qcamoID {
		// The user decides about QCamo from now on; Uninstall leaves it.
		a.after = func(manager.Result) error { s.markQCamo(root, false); return nil }
	}
	f, ok := s.begin(name, root)
	if !ok {
		return f.out
	}
	defer s.end()
	f.plan(a)
	if _, ok := f.runFrom(0); !ok {
		return f.finish("")
	}
	return f.finish(a.label + ": done.")
}

// AddPackage stores a package, turned off. It sets up the manager first when
// the game folder has never been used with it.
func (s *Service) AddPackage(root, path string) Outcome {
	f, ok := s.begin("add", root)
	if !ok {
		return f.out
	}
	defer s.end()
	base := filepath.Base(path)
	f.plan(action{label: "Check the game folder", command: "doctor"})
	r, ok := f.runFrom(0)
	if !ok {
		return f.finish("")
	}
	acts := []action{}
	if !r.Initialized {
		acts = append(acts, action{label: "Set up the manager in the game folder", command: "init"})
	}
	acts = append(acts, action{label: "Store " + base, command: "add", arg: path})
	f.plan(acts...)
	if _, ok := f.runFrom(1); !ok {
		return f.finish("")
	}
	if f.steps[len(f.steps)-1].State == StepSkipped {
		return f.finish(base + " is already stored.")
	}
	return f.finish("Stored " + base + ", turned off. Use its switch to turn it on.")
}

// Verify runs the manager's verify.
func (s *Service) Verify(root string) Outcome {
	f, ok := s.begin("verify", root)
	if !ok {
		return f.out
	}
	defer s.end()
	f.plan(action{label: "Verify", command: "verify"})
	if _, ok := f.runFrom(0); !ok {
		return f.finish("")
	}
	return f.finish("Verify passed: every managed file matches.")
}
