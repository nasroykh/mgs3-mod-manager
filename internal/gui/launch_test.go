package gui

import (
	"mgs3mod/internal/launch"
	"strings"
	"testing"
)

func TestDescribe(t *testing.T) {
	got := Describe(launch.Seed().Profiles[launch.SeedProfileID].Selection)
	if got != "North America · English · keyboard button prompts · game startup" {
		t.Fatalf("%q", got)
	}
	if got = Describe(launch.Selection{Region: "xx", Language: launch.LanguageGerman, Controller: launch.ControllerPS5, Destination: launch.DestinationMenu}); got != "xx · German · PlayStation 5 button prompts · main menu" {
		t.Fatalf("%q", got)
	}
}

func TestLaunchInfo(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	if info := s.LaunchInfo(""); info.Available || info.Reason == "" {
		t.Fatalf("no folder: %+v", info)
	}
	if info := s.LaunchInfo(f.root); info.Available || !strings.Contains(info.Reason, "Install / Update first") || info.SteamCopy {
		t.Fatalf("not set up: %+v", info)
	}
	mustOK(t, s.Install(f.root))
	info := s.LaunchInfo(f.root)
	if !info.Available || info.Saved || len(info.Profiles) != 1 {
		t.Fatalf("seed: %+v", info)
	}
	p := info.Profiles[0]
	if p.ID != launch.SeedProfileID || !p.Default || !p.Supported || p.Description != "North America · English · keyboard button prompts · game startup" {
		t.Fatalf("seed profile: %+v", p)
	}
}

// The fixture has no game executable, so the launch stops at its first
// check and no process is started.
func TestLaunchStopsAtThePreflight(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	mustOK(t, s.Install(f.root))
	r := &recorder{}
	s.SetProgress(r.add)
	out := s.Launch(f.root, launch.SeedProfileID)
	steps := r.check(t, "launch", out)
	if out.OK || states(steps) != "failed,waiting" || !strings.Contains(out.Summary, "missing or unsafe to start") {
		t.Fatalf("%s %+v", states(steps), out)
	}
	out = s.Launch(f.root, "no-such-profile")
	if out.OK || !strings.Contains(out.Details, `profile "no-such-profile" does not exist`) {
		t.Fatalf("unknown profile: %+v", out)
	}
}
