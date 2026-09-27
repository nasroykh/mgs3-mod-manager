package gui

import (
	"errors"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func testMod(id, version, digest, target string, enabled bool) manager.Mod {
	return manager.Mod{Manifest: packagefmt.Manifest{SchemaVersion: 2, ID: id, Version: version, Files: []packagefmt.File{{Target: target}}}, Digest: digest, Enabled: enabled}
}

var testKit = map[string]kitPackage{
	profile.LoaderID: {Path: "kit/asi-loader.zip", Version: "9.7.4", Digest: "L"},
	fpvID:            {Path: "kit/fpv.zip", Version: "0.8.4", Digest: "F"},
	faceID:           {Path: "kit/face.zip", Version: "1.0.4-face.4", Digest: "Q"},
}

func labels(acts []action) []string {
	out := []string{}
	for _, a := range acts {
		out = append(out, a.label)
	}
	return out
}

func TestPlanInstall(t *testing.T) {
	fresh := labels(planInstall(nil, testKit))
	want := []string{
		"Set up the manager in the game folder",
		"Store the ASI loader 9.7.4", "Turn on the ASI loader",
		"Store fpv-move 0.8.4", "Check fpv-move 0.8.4", "Turn on fpv-move 0.8.4",
		"Store QCamo face paint 1.0.4-face.4", "Check QCamo face paint 1.0.4-face.4", "Turn on QCamo face paint 1.0.4-face.4",
		"Verify",
	}
	if !reflect.DeepEqual(fresh, want) {
		t.Fatalf("fresh:\n%q\nwant\n%q", fresh, want)
	}

	current := &manager.State{Mods: map[string]manager.Mod{
		profile.LoaderID: testMod(profile.LoaderID, "9.7.4", "L", "wininet.dll", true),
		fpvID:            testMod(fpvID, "0.8.4", "F", "fpvmove.asi", true),
		faceID:           testMod(faceID, "1.0.4-face.4", "Q", "qcamo.asi", true),
		"crouch-walk":    testMod("crouch-walk", "0.2.1", "C", "mgs3crouchwalk.asi", true),
	}}
	if got := labels(planInstall(current, testKit)); !reflect.DeepEqual(got, []string{"Verify"}) {
		t.Fatalf("up to date: %q", got)
	}

	older := &manager.State{Mods: map[string]manager.Mod{
		profile.LoaderID: testMod(profile.LoaderID, "9.7.4", "other loader build", "wininet.dll", false),
		fpvID:            testMod(fpvID, "0.8.3", "old", "fpvmove.asi", true),
		faceID:           testMod(faceID, "1.0.4-face.4", "Q", "qcamo.asi", false),
		qcamoID:          testMod(qcamoID, "1.0.4", "Z", "qcamo.asi", true),
	}}
	got := labels(planInstall(older, testKit))
	want = []string{
		"Turn on the ASI loader",
		"Turn off fpv-move 0.8.3", "Remove fpv-move 0.8.3", "Store fpv-move 0.8.4", "Check fpv-move 0.8.4", "Turn on fpv-move 0.8.4",
		"Turn off QCamo 1.0.4 (kept in the library)", "Check QCamo face paint 1.0.4-face.4", "Turn on QCamo face paint 1.0.4-face.4",
		"Verify",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("older:\n%q\nwant\n%q", got, want)
	}
}

func TestPlanUninstall(t *testing.T) {
	if acts, _ := planUninstall(nil, true); acts != nil {
		t.Fatal("plan for a folder without the manager")
	}
	st := func(qcamoOn, crouch bool) *manager.State {
		mods := map[string]manager.Mod{
			profile.LoaderID: testMod(profile.LoaderID, "9.7.4", "L", "wininet.dll", true),
			fpvID:            testMod(fpvID, "0.8.4", "F", "fpvmove.asi", true),
			faceID:           testMod(faceID, "1.0.4-face.4", "Q", "qcamo.asi", true),
			qcamoID:          testMod(qcamoID, "1.0.4", "Z", "qcamo.asi", qcamoOn),
			"camo-test":      testMod("camo-test", "0.1.0", "T", "camo.asi", false),
		}
		if crouch {
			mods["crouch-walk"] = testMod("crouch-walk", "0.2.1", "C", "mgs3crouchwalk.asi", true)
		}
		return &manager.State{Mods: mods}
	}
	removeBoth := []string{"Turn off QCamo face paint 1.0.4-face.4", "Remove QCamo face paint 1.0.4-face.4", "Turn off fpv-move 0.8.4", "Remove fpv-move 0.8.4"}
	cases := []struct {
		name    string
		st      *manager.State
		restore bool
		tail    []string
		note    string
	}{
		{"loader alone", st(false, false), false, []string{"Turn off the ASI loader", "Remove the ASI loader", "Verify"}, ""},
		{"crouch-walk keeps the loader", st(false, true), false, []string{"Verify"}, " The ASI loader stays: crouch-walk still uses it."},
		{"qcamo comes back and keeps the loader", st(false, false), true, []string{"Turn QCamo 1.0.4 back on", "Verify"}, " QCamo 1.0.4 is on again. The ASI loader stays: qcamo still uses it."},
		{"two users", st(false, true), true, []string{"Turn QCamo 1.0.4 back on", "Verify"}, " QCamo 1.0.4 is on again. The ASI loader stays: crouch-walk, qcamo still use it."},
	}
	for _, c := range cases {
		acts, note := planUninstall(c.st, c.restore)
		if got, want := labels(acts), append(append([]string{}, removeBoth...), c.tail...); !reflect.DeepEqual(got, want) || note != c.note {
			t.Errorf("%s:\n%q %q\nwant\n%q %q", c.name, got, note, want, c.note)
		}
	}
}

type recorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *recorder) add(e Event) { r.mu.Lock(); r.events = append(r.events, e); r.mu.Unlock() }

// check replays the events: every step goes waiting, running, then one end
// state; plans only grow; finished comes last and matches the outcome.
func (r *recorder) check(t *testing.T, action string, out Outcome) []Step {
	t.Helper()
	steps := []Step{}
	for i, e := range r.events {
		if e.Action != action {
			t.Fatalf("event %d for %q, want %q", i, e.Action, action)
		}
		switch e.Kind {
		case "plan":
			if len(e.Steps) < len(steps) {
				t.Fatalf("plan shrank at event %d", i)
			}
			for j := range steps {
				if e.Steps[j] != steps[j] {
					t.Fatalf("plan rewrote step %d at event %d: %+v vs %+v", j, i, e.Steps[j], steps[j])
				}
			}
			steps = append([]Step(nil), e.Steps...)
		case "step":
			prev := steps[e.Index].State
			ok := prev == StepWaiting && e.State == StepRunning || prev == StepRunning && (e.State == StepDone || e.State == StepSkipped || e.State == StepFailed) || prev == StepWaiting && e.State == StepSkipped
			if !ok {
				t.Fatalf("step %d went %s -> %s", e.Index, prev, e.State)
			}
			steps[e.Index].State = e.State
		case "finished":
			if i != len(r.events)-1 || e.Outcome == nil || !reflect.DeepEqual(*e.Outcome, out) {
				t.Fatalf("finished event %d of %d: %+v vs %+v", i, len(r.events), e.Outcome, out)
			}
		default:
			t.Fatalf("unknown event kind %q", e.Kind)
		}
	}
	if len(r.events) == 0 || r.events[len(r.events)-1].Kind != "finished" {
		t.Fatal("no finished event")
	}
	return steps
}

func states(steps []Step) string {
	out := []string{}
	for _, s := range steps {
		out = append(out, s.State)
	}
	return strings.Join(out, ",")
}

func TestProgressEvents(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()

	r := &recorder{}
	s.SetProgress(r.add)
	out := s.Install(f.root)
	mustOK(t, out)
	steps := r.check(t, "install", out)
	if len(steps) != 12 || strings.Contains(states(steps), StepWaiting) || strings.Contains(states(steps), StepFailed) {
		t.Fatalf("install steps: %+v", steps)
	}
	if s.Busy() {
		t.Fatal("busy after the action")
	}

	r = &recorder{}
	s.SetProgress(r.add)
	out = s.Install(f.root)
	mustOK(t, out)
	if steps = r.check(t, "install", out); len(steps) != 3 || !strings.HasPrefix(out.Summary, "Already installed") {
		t.Fatalf("second install: %+v %s", steps, out.Summary)
	}

	// A failure leaves the later steps waiting.
	r = &recorder{}
	s.SetProgress(r.add)
	f.running = errors.New("installation process running: METAL GEAR SOLID3.exe (pid 42)")
	out = s.Uninstall(f.root)
	steps = r.check(t, "uninstall", out)
	if out.OK || states(steps) != "done,failed,waiting,waiting,waiting,waiting,waiting,waiting" {
		t.Fatalf("failed uninstall: %s %+v", states(steps), out)
	}
	f.running = nil

	preview := s.PlanUninstall(f.root)
	if preview.Problem != nil || len(preview.Steps) != 7 || preview.Steps[0] != "Turn off QCamo face paint 1.0.4-face.4" || preview.Note != "" {
		t.Fatalf("preview: %+v", preview)
	}
	r = &recorder{}
	s.SetProgress(r.add)
	out = s.Uninstall(f.root)
	mustOK(t, out)
	steps = r.check(t, "uninstall", out)
	got := []string{}
	for _, st := range steps[1:] {
		got = append(got, st.Label)
	}
	if !reflect.DeepEqual(got, preview.Steps) {
		t.Fatalf("uninstall ran %q, preview said %q", got, preview.Steps)
	}

	// The busy lock refuses a second action and reports it as an event too.
	r = &recorder{}
	s.SetProgress(r.add)
	s.busy.Lock()
	out = s.Verify(f.root)
	s.busy.Unlock()
	if steps = r.check(t, "verify", out); out.OK || states(steps) != StepFailed {
		t.Fatalf("busy: %+v", out)
	}
}

func TestPlanUninstallPreviewWithoutManager(t *testing.T) {
	f := newFixture(t)
	if p := f.service().PlanUninstall(f.root); p.Problem != nil || len(p.Steps) != 0 {
		t.Fatalf("%+v", p)
	}
	if p := f.service().PlanUninstall(""); p.Problem == nil {
		t.Fatal("no problem for an empty folder")
	}
}
