package gui

import (
	"strings"
	"testing"
)

func TestProfileID(t *testing.T) {
	for name, want := range map[string]string{
		"EU PS5":                      "eu-ps5",
		"  Europe / German ":          "europe-german",
		"-x-":                         "x",
		"Été!!":                       "t",
		"***":                         "",
		"a" + strings.Repeat("b", 80): "a" + strings.Repeat("b", 63),
	} {
		if got := ProfileID(name); got != want {
			t.Errorf("ProfileID(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestOptionsMatchTheLauncher(t *testing.T) {
	o := (&Service{}).Options()
	got := []string{}
	for _, r := range o.Regions {
		langs := []string{}
		for _, l := range r.Languages {
			langs = append(langs, l.Value)
		}
		got = append(got, r.Value+":"+strings.Join(langs, ","))
	}
	if strings.Join(got, " ") != "us:en,fr,sp eu:en,fr,it,gr,sp" || len(o.Controllers) != 5 || o.Controllers[0].Label != "Keyboard" || len(o.Tested) != 2 || o.Tested[0] != (TestedChoice{"us", "en", "kbd"}) || o.Tested[1] != (TestedChoice{"eu", "fr", "xbox"}) {
		t.Fatalf("%q %+v", got, o.Controllers)
	}
}

func TestProfilesCreateEditRenameDefaultDelete(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	mustOK(t, s.Install(f.root))

	if r := s.SaveProfile(f.root, "", "EU pad", "eu", "gr", "ps5", false); !r.OK || r.ID != "eu-pad" {
		t.Fatalf("create: %+v", r)
	}
	if r := s.SaveProfile(f.root, "", "eu pad", "eu", "en", "kbd", false); r.OK || !strings.Contains(r.Summary, "already exists") {
		t.Fatalf("duplicate: %+v", r)
	}
	if r := s.SaveProfile(f.root, "", "bad", "us", "gr", "kbd", false); r.OK || !strings.Contains(r.Summary, "does not offer this language") {
		t.Fatalf("combination: %+v", r)
	}
	if r := s.SaveProfile(f.root, "", "!!!", "us", "en", "kbd", false); r.OK {
		t.Fatalf("empty name: %+v", r)
	}
	info := s.LaunchInfo(f.root)
	if !info.Saved || len(info.Profiles) != 2 {
		t.Fatalf("after create: %+v", info)
	}
	for _, p := range info.Profiles {
		if p.ID == "eu-pad" && (p.Tested || !p.Supported || p.Region != "eu" || p.Language != "gr" || p.Controller != "ps5" || p.Description != "Europe · German · PlayStation 5 button prompts · game startup") {
			t.Fatalf("eu-pad: %+v", p)
		}
		if p.ID == "na-startup" && (!p.Tested || !p.Default) {
			t.Fatalf("seed: %+v", p)
		}
	}

	// Edit in place and make it the default, then rename it.
	if r := s.SaveProfile(f.root, "eu-pad", "EU pad", "eu", "it", "xbox", true); !r.OK {
		t.Fatalf("edit: %+v", r)
	}
	if r := s.SaveProfile(f.root, "eu-pad", "Italy Xbox", "eu", "it", "xbox", false); !r.OK || r.ID != "italy-xbox" {
		t.Fatalf("rename: %+v", r)
	}
	info = s.LaunchInfo(f.root)
	ids := map[string]LaunchProfile{}
	for _, p := range info.Profiles {
		ids[p.ID] = p
	}
	if _, old := ids["eu-pad"]; old || !ids["italy-xbox"].Default || ids["italy-xbox"].Language != "it" {
		t.Fatalf("after rename: %+v", info.Profiles)
	}

	// A name the slug would change keeps its ID when it is not edited.
	if r := s.SaveProfile(f.root, "italy-xbox", "italy-xbox", "eu", "it", "ps4", false); !r.OK || r.ID != "italy-xbox" {
		t.Fatalf("edit keeps ID: %+v", r)
	}
	if r := s.SaveProfile(f.root, "missing", "other", "eu", "it", "ps4", false); r.OK {
		t.Fatalf("renamed a missing profile: %+v", r)
	}
	if r := s.SetDefaultProfile(f.root, "na-startup"); !r.OK {
		t.Fatalf("default: %+v", r)
	}
	if r := s.DeleteProfile(f.root, "italy-xbox"); !r.OK {
		t.Fatalf("delete: %+v", r)
	}
	if r := s.DeleteProfile(f.root, "na-startup"); r.OK || !strings.Contains(r.Summary, "only launch profile") {
		t.Fatalf("delete last: %+v", r)
	}
	s.busy.Lock()
	r := s.SetDefaultProfile(f.root, "na-startup")
	s.busy.Unlock()
	if r.OK {
		t.Fatal("changed a profile while busy")
	}
}
