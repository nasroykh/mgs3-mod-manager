package gui

import (
	"errors"
	"mgs3mod/internal/profile"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A QCamo 1.0.4 user with no other ASI mod: Install turns QCamo off,
// Uninstall turns it back on, so the loader stays for it.
func TestQCamoRoundTrip(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	old := t.TempDir()
	f.run("init", "")
	f.run("add", f.pkg(old, profile.LoaderID, "9.7.4", "loader.asi", 1))
	f.run("enable", profile.LoaderID)
	f.run("add", f.pkg(old, qcamoID, "1.0.4", "qcamo.asi", 4))
	f.run("enable", qcamoID)
	s := f.service()

	mustOK(t, s.Install(f.root))
	if f.mods()[qcamoID].Enabled || !s.qcamoMarked(f.root) {
		t.Fatal("install did not turn qcamo off and remember it")
	}
	out := s.Uninstall(f.root)
	mustOK(t, out)
	mods := f.mods()
	if !mods[qcamoID].Enabled || !mods[profile.LoaderID].Enabled || s.qcamoMarked(f.root) {
		t.Fatalf("round trip: %s %+v", out.Summary, mods)
	}
	if !strings.Contains(out.Summary, "QCamo 1.0.4 is on again") || !strings.Contains(out.Summary, "qcamo still uses it") {
		t.Fatalf("summary: %s", out.Summary)
	}
}

// QCamo that the user had turned off stays off, and the loader goes.
func TestUninstallLeavesQCamoTheUserTurnedOff(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	f.run("init", "")
	f.run("add", f.pkg(t.TempDir(), qcamoID, "1.0.4", "qcamo.asi", 4))
	s := f.service()
	mustOK(t, s.Install(f.root))
	if s.qcamoMarked(f.root) {
		t.Fatal("remembered a qcamo that was off")
	}
	mustOK(t, s.Uninstall(f.root))
	mods := f.mods()
	if mods[qcamoID].Enabled {
		t.Fatal("uninstall turned on a qcamo the user had off")
	}
	if _, ok := mods[profile.LoaderID]; ok {
		t.Fatal("loader kept with no enabled ASI mod")
	}
}

// Once the user switches QCamo in the Mods list, Uninstall leaves it alone.
func TestQCamoSwitchForgetsTheMark(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	f.run("init", "")
	f.run("add", f.pkg(t.TempDir(), profile.LoaderID, "9.7.4", "loader.asi", 1))
	f.run("enable", profile.LoaderID)
	f.run("add", f.pkg(t.TempDir(), qcamoID, "1.0.4", "qcamo.asi", 4))
	f.run("enable", qcamoID)
	s := f.service()
	mustOK(t, s.Install(f.root))
	mustOK(t, s.SetEnabled(f.root, faceID, false))
	mustOK(t, s.SetEnabled(f.root, qcamoID, true))
	mustOK(t, s.SetEnabled(f.root, qcamoID, false))
	if s.qcamoMarked(f.root) {
		t.Fatal("mark kept after the user switched qcamo")
	}
	mustOK(t, s.Uninstall(f.root))
	if f.mods()[qcamoID].Enabled {
		t.Fatal("uninstall overrode the user's choice")
	}
}

// A failure before the face paint step leaves QCamo as it was.
func TestInstallFailureKeepsQCamo(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	f.run("init", "")
	f.run("add", f.pkg(t.TempDir(), profile.LoaderID, "9.7.4", "loader.asi", 1))
	f.run("enable", profile.LoaderID)
	f.run("add", f.pkg(t.TempDir(), qcamoID, "1.0.4", "qcamo.asi", 4))
	f.run("enable", qcamoID)
	put(t, f.root, "fpvmove.asi", []byte("someone else's"))
	s := f.service()
	out := s.Install(f.root)
	if out.OK || !strings.HasPrefix(out.Failed, "Store fpv-move") {
		t.Fatalf("%+v", out)
	}
	if !f.mods()[qcamoID].Enabled || s.qcamoMarked(f.root) {
		t.Fatal("qcamo turned off before a failed step")
	}
}

func TestUninstallChecksWithoutDoctor(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	if out := s.Uninstall(f.root); !out.OK || !strings.HasPrefix(out.Summary, "Nothing to remove") || f.exists(".mgs3mod") {
		t.Fatalf("uninitialized: %+v", out)
	}
	mustOK(t, s.Install(f.root))
	f.running = errors.New("installation process running: METAL GEAR SOLID3.exe (pid 42)")
	out := s.Uninstall(f.root)
	if out.OK || !strings.HasPrefix(out.Failed, "Turn off QCamo face paint") || !strings.HasPrefix(out.Summary, "Close the game") {
		t.Fatalf("game running: %+v", out)
	}
	if !f.mods()[faceID].Enabled {
		t.Fatal("changed while the game runs")
	}
}

func TestStatusOutdatedAndKitMismatch(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	mustOK(t, s.Install(f.root))
	if err := os.Remove(filepath.Join(f.kit, "fpv-move-0.8.4.mgs3mod.zip")); err != nil {
		t.Fatal(err)
	}
	f.pkg(f.kit, fpvID, "0.8.5", "fpvmove.asi", 9)
	o := s.Status(f.root)
	if o.DeltaState != "outdated" || !strings.Contains(o.Delta, "replaces it with fpv-move 0.8.5") {
		t.Fatalf("outdated: %+v", o)
	}
	mustOK(t, s.Install(f.root))
	if v := f.mods()[fpvID].Manifest.Version; v != "0.8.5" {
		t.Fatalf("update installed %s", v)
	}

	// A kit file whose name and manifest ID disagree is refused.
	wrong := f.pkg(t.TempDir(), "camo-test", "0.1.0", "camo.asi", 7)
	if err := os.Rename(wrong, filepath.Join(f.kit, "qcamo-face-9.mgs3mod.zip")); err != nil {
		t.Fatal(err)
	}
	out := s.Install(f.root)
	if out.OK || !strings.Contains(out.Summary, "several qcamo-face packages") {
		t.Fatalf("two face packages: %+v", out)
	}
	os.Remove(filepath.Join(f.kit, "qcamo-face-1.0.4-face.4.mgs3mod.zip"))
	out = s.Install(f.root)
	if out.OK || !strings.Contains(out.Summary, "holds camo-test, not qcamo-face") {
		t.Fatalf("wrong ID: %+v", out)
	}
}

// A failed QCamo disable forgets the mark; Busy holds during every event.
func TestQCamoMarkUndoneWhenTheDisableFails(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	f.run("init", "")
	f.run("add", f.pkg(t.TempDir(), profile.LoaderID, "9.7.4", "loader.asi", 1))
	f.run("enable", profile.LoaderID)
	f.run("add", f.pkg(t.TempDir(), qcamoID, "1.0.4", "qcamo.asi", 4))
	f.run("enable", qcamoID)
	s := f.service()
	notBusy := 0
	var plan []Step
	s.SetProgress(func(e Event) {
		if e.Kind == "plan" {
			plan = e.Steps
		}
		if e.Kind != "finished" && !s.Busy() {
			notBusy++
		}
		// The game "starts" just as QCamo is being turned off.
		if e.Kind == "step" && e.State == StepRunning && strings.HasPrefix(plan[e.Index].Label, "Turn off QCamo") {
			f.running = errors.New("installation process running: METAL GEAR SOLID3.exe (pid 7)")
		}
	})
	out := s.Install(f.root)
	if out.OK || !strings.HasPrefix(out.Failed, "Turn off QCamo 1.0.4") {
		t.Fatalf("%+v", out)
	}
	if s.qcamoMarked(f.root) || !f.mods()[qcamoID].Enabled {
		t.Fatal("mark kept after a failed disable, or qcamo changed")
	}
	if notBusy != 0 {
		t.Fatalf("%d events while not busy", notBusy)
	}
	if s.Busy() {
		t.Fatal("still busy")
	}
}
