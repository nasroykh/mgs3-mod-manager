package gui

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// The manager's path checks reject 8.3 TEMP aliases; use a repository-local
	// temporary folder like the manager tests do.
	dir, _ := filepath.Abs("../../.cache/gui-tests")
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(err)
	}
	os.Setenv("TMP", dir)
	os.Setenv("TEMP", dir)
	os.Exit(m.Run())
}

func put(t *testing.T, root, p string, b []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, b, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestParseLibraryFolders(t *testing.T) {
	current := `"libraryfolders"
{
	"0"
	{
		"path"		"C:\\Program Files (x86)\\Steam"
		"label"		""
		"apps"
		{
			"228980"		"418053312"
		}
	}
	// a comment
	"1"
	{
		"path"		"D:\\SteamLibrary"
		"apps" { "2131630" "12345" }
	}
	"2" { "path" "E:\\Games\\Steam \"Lib\"" }
}`
	got, err := ParseLibraryFolders(current)
	want := []string{`C:\Program Files (x86)\Steam`, `D:\SteamLibrary`, `E:\Games\Steam "Lib"`}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("current format: %q %v", got, err)
	}
	old := "\"LibraryFolders\"\n{\n\t\"TimeNextStatsReport\"\t\t\"1700000000\"\n\t\"ContentStatsID\"\t\t\"-123\"\n\t\"1\"\t\t\"D:\\\\Steam Library\"\n\t\"2\"\t\t\"F:\\\\SL\"\n}\n"
	got, err = ParseLibraryFolders(old)
	if err != nil || !reflect.DeepEqual(got, []string{`D:\Steam Library`, `F:\SL`}) {
		t.Fatalf("old format: %q %v", got, err)
	}
	for _, bad := range []string{`"libraryfolders" {`, `"libraryfolders" { "0" { "path" "C:\\x }}`, `}`} {
		if _, err := ParseLibraryFolders(bad); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

func TestSteamGamesFindsEveryLibrary(t *testing.T) {
	base := t.TempDir()
	steam := filepath.Join(base, "Steam")
	lib := filepath.Join(base, "Lib Two")
	put(t, steam, "steamapps/common/MGS3 MC/"+GameExe, []byte("x"))
	put(t, steam, "steamapps/common/Other Game/game.exe", []byte("x"))
	put(t, lib, "steamapps/common/METAL GEAR SOLID 3 MASTER COLLECTION/"+GameExe, []byte("x"))
	if err := os.MkdirAll(filepath.Join(lib, "steamapps/common/Empty/"+GameExe), 0700); err != nil {
		t.Fatal(err) // a folder named like the executable is not a game
	}
	vdf := "\"libraryfolders\"\n{\n\"0\" { \"path\" \"" + strings.ReplaceAll(steam, `\`, `\\`) + "\" }\n\"1\" { \"path\" \"" + strings.ReplaceAll(lib, `\`, `\\`) + "\" }\n\"2\" { \"path\" \"" + strings.ReplaceAll(filepath.Join(base, "gone"), `\`, `\\`) + "\" }\n}\n"
	put(t, steam, "steamapps/libraryfolders.vdf", []byte(vdf))
	got, err := SteamGames(filepath.ToSlash(steam))
	want := []string{filepath.Join(lib, "steamapps", "common", "METAL GEAR SOLID 3 MASTER COLLECTION"), filepath.Join(steam, "steamapps", "common", "MGS3 MC")}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q %v, want %q", got, err, want)
	}
	put(t, steam, "steamapps/libraryfolders.vdf", []byte(`"libraryfolders" {`))
	got, err = SteamGames(steam)
	if err == nil || len(got) != 1 {
		t.Fatalf("broken vdf: %q %v", got, err)
	}
}

func TestSettingsRoundTripAndBadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mgs3mod", "gui.json")
	if s := LoadSettings(path); s.GameRoot != "" {
		t.Fatal("absent file gave a folder")
	}
	if err := SaveSettings(path, Settings{GameRoot: `D:\Games\MGS3`}); err != nil {
		t.Fatal(err)
	}
	if s := LoadSettings(path); s.GameRoot != `D:\Games\MGS3` {
		t.Fatalf("round trip: %+v", s)
	}
	put(t, filepath.Dir(path), "gui.json", []byte("{not json"))
	if s := LoadSettings(path); s.GameRoot != "" {
		t.Fatal("bad file gave a folder")
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temporary files left: %d entries", len(entries))
	}
}

func TestExplain(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{&manager.Error{Code: 3, Message: "close the game and launcher before continuing: installation process running: X (pid 1)"}, "Close the game and the Master Collection launcher first"},
		{&manager.Error{Code: 3, Message: `installation fingerprints: core hash mismatch for "METAL GEAR SOLID3.exe": got a want b`, Paths: []string{`D:\g`}}, "Not supported: METAL GEAR SOLID3.exe differs"},
		{&manager.Error{Code: 3, Message: `installation fingerprints: open core "Engine.dll": missing`}, "Engine.dll is missing"},
		{&manager.Error{Code: 4, Message: "target owned by qcamo-face", Paths: []string{"qcamo.asi"}}, "Conflict: qcamo.asi is already changed by the mod qcamo-face. Turn qcamo-face off first."},
		{&manager.Error{Code: 6, Message: "another manager holds the lock"}, "Another copy of the manager"},
		{&manager.Error{Code: 5, Message: "unresolved transaction; run recover"}, "recover"},
		{&manager.Error{Code: 3, Message: "ASI plugins require enabled managed asi-loader 9.7.4"}, "needs the ASI loader"},
		{&manager.Error{Code: 4, Message: "expected file to be absent", Paths: []string{"wininet.dll"}}, "wininet.dll already exists"},
		{&manager.Error{Code: 4, Message: "unexpected file contents", Paths: []string{"fpvmove.asi"}}, "changed outside the manager: fpvmove.asi"},
		{&manager.Error{Code: 5, Message: "baseline corrupted: unexpected file contents"}, "The manager stopped: baseline corrupted"},
		{errors.New("plain"), "The manager stopped: plain"},
		// An interrupted change wraps its cause; the advice must be recovery.
		{&manager.Error{Code: 5, Message: "operation interrupted; files may be partially changed; run recover: close the game and launcher before continuing: running"}, "An earlier change was interrupted"},
		{&manager.Error{Code: 5, Message: "operation interrupted; files may be partially changed; run recover: expected file to be absent", Paths: []string{"qcamo.asi"}}, "An earlier change was interrupted"},
		{&manager.Error{Code: 3, Message: "installation fingerprints: core hash mismatch for \"Engine.dll\": got a want b"}, "Not supported: Engine.dll differs"},
	}
	for _, c := range cases {
		p := Explain(c.err)
		if !strings.Contains(p.Summary, c.want) {
			t.Errorf("%v: summary %q lacks %q", c.err, p.Summary, c.want)
		}
		if !strings.Contains(p.Details, c.err.Error()) {
			t.Errorf("%v: details %q lack the original message", c.err, p.Details)
		}
	}
}

// Manager fixtures, as in internal/manager tests: a fake core file, a stubbed
// process check and ASI loader check, and small plugin packages.

type fixture struct {
	t       *testing.T
	root    string
	kit     string
	running error
	core    []byte
}

func newFixture(t *testing.T) *fixture {
	base := t.TempDir()
	f := &fixture{t: t, root: filepath.Join(base, "game"), kit: filepath.Join(base, "kit"), core: []byte("core")}
	put(t, f.root, "game.exe", []byte("core"))
	if err := os.MkdirAll(f.kit, 0700); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) open(root string) (*manager.Manager, error) {
	return manager.New(manager.Config{
		Root:           root,
		Core:           []profile.Fingerprint{{Path: "game.exe", SHA256: transaction.Hash(f.core)}},
		CheckProcesses: func(string) error { return f.running },
		CheckASILoader: func(string) error { return nil },
	}), nil
}

func (f *fixture) service() *Service {
	s := New(f.open, f.kit, filepath.Join(filepath.Dir(f.root), "settings", "gui.json"))
	s.steamPath = func() (string, error) { return "", nil } // not this machine's Steam
	return s
}

func (f *fixture) run(command, arg string) manager.Result {
	f.t.Helper()
	m, _ := f.open(f.root)
	r, err := m.Run(command, arg, manager.Options{})
	if err != nil {
		f.t.Fatalf("%s %s: %v", command, arg, err)
	}
	return r
}

// pkg packs a one-plugin package into dir and returns the zip path.
func (f *fixture) pkg(dir, id, version, target string, marker byte) string {
	f.t.Helper()
	src := filepath.Join(filepath.Dir(f.root), "authoring", id+"-"+version)
	payload := testPluginPE(marker)
	put(f.t, src, "payload/"+target, payload)
	manifest := packagefmt.Manifest{SchemaVersion: 2, ID: id, Version: version, Name: id, Profile: profile.ID, Files: []packagefmt.File{{
		Source: "payload/" + target, Target: target, OriginalAbsent: true,
		PayloadSHA256: transaction.Hash(payload), PayloadBytes: int64(len(payload)),
	}}}
	b, _ := json.Marshal(manifest)
	put(f.t, src, "manifest.json", b)
	out := filepath.Join(dir, id+"-"+version+".mgs3mod.zip")
	if _, err := packagefmt.Pack(src, out); err != nil {
		f.t.Fatal(err)
	}
	return out
}

// kitFiles writes the three kit packages. The loader is a stand-in plugin:
// the real loader's pinned files are covered by the manager's artifact tests.
func (f *fixture) kitFiles() {
	f.pkg(f.kit, profile.LoaderID, "9.7.4", "loader.asi", 1)
	f.pkg(f.kit, fpvID, "0.8.4", "fpvmove.asi", 2)
	f.pkg(f.kit, faceID, "1.0.4-face.4", "qcamo.asi", 3)
}

func testPluginPE(marker byte) []byte {
	const peOffset = 0x80
	const optionalSize = 240
	b := make([]byte, 0x400)
	b[0], b[1] = 'M', 'Z'
	binary.LittleEndian.PutUint32(b[0x3c:], peOffset)
	copy(b[peOffset:], []byte{'P', 'E', 0, 0})
	coff := peOffset + 4
	binary.LittleEndian.PutUint16(b[coff:], 0x8664)
	binary.LittleEndian.PutUint16(b[coff+2:], 1)
	binary.LittleEndian.PutUint16(b[coff+16:], optionalSize)
	binary.LittleEndian.PutUint16(b[coff+18:], 0x2002)
	optional := coff + 20
	binary.LittleEndian.PutUint16(b[optional:], 0x20b)
	binary.LittleEndian.PutUint32(b[optional+4:], 0x200)
	binary.LittleEndian.PutUint32(b[optional+16:], 0x1000)
	binary.LittleEndian.PutUint32(b[optional+20:], 0x1000)
	binary.LittleEndian.PutUint32(b[optional+32:], 0x1000)
	binary.LittleEndian.PutUint32(b[optional+36:], 0x200)
	binary.LittleEndian.PutUint32(b[optional+56:], 0x2000)
	binary.LittleEndian.PutUint32(b[optional+60:], 0x200)
	binary.LittleEndian.PutUint16(b[optional+68:], 2)
	binary.LittleEndian.PutUint32(b[optional+108:], 16)
	section := optional + optionalSize
	copy(b[section:], []byte(".text\x00\x00\x00"))
	binary.LittleEndian.PutUint32(b[section+8:], 0x200)
	binary.LittleEndian.PutUint32(b[section+12:], 0x1000)
	binary.LittleEndian.PutUint32(b[section+16:], 0x200)
	binary.LittleEndian.PutUint32(b[section+20:], 0x200)
	binary.LittleEndian.PutUint32(b[section+36:], 0x60000020)
	b[0x200] = 0xc3
	b[0x201] = marker
	return b
}

func (f *fixture) exists(target string) bool {
	_, err := os.Lstat(filepath.Join(f.root, target))
	return err == nil
}

func (f *fixture) mods() map[string]manager.Mod {
	return f.run("list", "").State.Mods
}

func mustOK(t *testing.T, out Outcome) {
	t.Helper()
	if !out.OK {
		t.Fatalf("failed at %q: %s\n%s", out.Failed, out.Summary, out.Details)
	}
}

func TestInstallFreshThenAgainThenUninstall(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	if o := s.Status(f.root); o.DeltaState != "missing" || o.Initialized || o.Problem != nil {
		t.Fatalf("fresh status: %+v", o)
	}
	out := s.Install(f.root)
	mustOK(t, out)
	mods := f.mods()
	for _, id := range kitIDs {
		if !mods[id].Enabled {
			t.Fatalf("%s not enabled: %+v", id, mods)
		}
	}
	for _, target := range []string{"loader.asi", "fpvmove.asi", "qcamo.asi"} {
		if !f.exists(target) {
			t.Fatalf("%s missing", target)
		}
	}
	o := s.Status(f.root)
	if o.DeltaState != "installed" || o.Delta != "Installed: fpv-move 0.8.4, QCamo face paint 1.0.4-face.4" || len(o.Mods) != 3 {
		t.Fatalf("status after install: %+v", o)
	}

	generation := f.run("list", "").State.Generation
	mustOK(t, s.Install(f.root))
	if f.run("list", "").State.Generation != generation {
		t.Fatal("a second Install changed the state")
	}

	out = s.Uninstall(f.root)
	mustOK(t, out)
	if len(f.mods()) != 0 {
		t.Fatalf("mods left: %+v", f.mods())
	}
	for _, target := range []string{"loader.asi", "fpvmove.asi", "qcamo.asi"} {
		if f.exists(target) {
			t.Fatalf("%s left behind", target)
		}
	}
	if strings.Contains(out.Summary, "stays") {
		t.Fatalf("loader kept without users: %s", out.Summary)
	}
}

func TestInstallReplacesOtherVersionAndLeavesOtherModsAlone(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	old := t.TempDir()
	f.run("init", "")
	f.run("add", f.pkg(old, profile.LoaderID, "9.7.4", "loader.asi", 1))
	f.run("enable", profile.LoaderID)
	f.run("add", f.pkg(old, qcamoID, "1.0.4", "qcamo.asi", 4))
	f.run("enable", qcamoID)
	f.run("add", f.pkg(old, fpvID, "0.8.3", "fpvmove.asi", 5))
	f.run("enable", fpvID)
	f.run("add", f.pkg(old, "crouch-walk", "0.2.1", "crouch.asi", 6))
	f.run("enable", "crouch-walk")
	f.run("add", f.pkg(old, "camo-test", "0.1.0", "camo.asi", 7))
	s := f.service()
	if o := s.Status(f.root); o.DeltaState != "partial" {
		t.Fatalf("status before: %+v", o)
	}

	out := s.Install(f.root)
	mustOK(t, out)
	joined := strings.Join(out.Steps, "\n")
	for _, step := range []string{"Turn off QCamo 1.0.4", "Remove fpv-move 0.8.3", "Turn on fpv-move 0.8.4", "Turn on QCamo face paint 1.0.4-face.4"} {
		if !strings.Contains(joined, step) {
			t.Fatalf("missing step %q in:\n%s", step, joined)
		}
	}
	mods := f.mods()
	if mods[fpvID].Manifest.Version != "0.8.4" || !mods[fpvID].Enabled || !mods[faceID].Enabled {
		t.Fatalf("plugins: %+v", mods)
	}
	if q, ok := mods[qcamoID]; !ok || q.Enabled {
		t.Fatalf("qcamo must stay stored and off: %+v", q)
	}
	if !mods["crouch-walk"].Enabled || mods["camo-test"].Enabled {
		t.Fatalf("other mods changed: %+v", mods)
	}

	out = s.Uninstall(f.root)
	mustOK(t, out)
	mods = f.mods()
	if _, ok := mods[fpvID]; ok {
		t.Fatal("fpv-move left")
	}
	if _, ok := mods[faceID]; ok {
		t.Fatal("qcamo-face left")
	}
	if !mods[profile.LoaderID].Enabled || !mods["crouch-walk"].Enabled || !strings.Contains(out.Summary, "crouch-walk, qcamo still use it") {
		t.Fatalf("loader or crouch-walk changed: %s %+v", out.Summary, mods)
	}
	if !mods[qcamoID].Enabled || !strings.Contains(out.Summary, "QCamo 1.0.4 is on again") {
		t.Fatalf("uninstall did not turn qcamo 1.0.4 back on: %s %+v", out.Summary, mods[qcamoID])
	}
}

func TestInstallRefusesAndChangesNothing(t *testing.T) {
	t.Run("game running", func(t *testing.T) {
		f := newFixture(t)
		f.kitFiles()
		f.running = errors.New("installation process running: METAL GEAR SOLID3.exe (pid 42)")
		out := f.service().Install(f.root)
		if out.OK || out.Failed != "Check the game folder" || !strings.HasPrefix(out.Summary, "Close the game") {
			t.Fatalf("%+v", out)
		}
		if f.exists(".mgs3mod") {
			t.Fatal("state created while the game runs")
		}
	})
	t.Run("other game version", func(t *testing.T) {
		f := newFixture(t)
		f.kitFiles()
		f.core = []byte("other build")
		out := f.service().Install(f.root)
		if out.OK || !strings.HasPrefix(out.Summary, "Not supported: game.exe differs") || f.exists(".mgs3mod") {
			t.Fatalf("%+v", out)
		}
		if c := f.service().Doctor(f.root); c.OK || !strings.HasPrefix(c.Summary, "Not supported") {
			t.Fatalf("doctor: %+v", c)
		}
	})
	t.Run("kit incomplete", func(t *testing.T) {
		f := newFixture(t)
		f.pkg(f.kit, fpvID, "0.8.4", "fpvmove.asi", 2)
		out := f.service().Install(f.root)
		if out.OK || out.Failed != "Find the kit packages" || !strings.Contains(out.Summary, "asi-loader package not found") || f.exists(".mgs3mod") {
			t.Fatalf("%+v", out)
		}
	})
	t.Run("two builds of one plugin", func(t *testing.T) {
		f := newFixture(t)
		f.kitFiles()
		f.pkg(f.kit, fpvID, "0.8.3", "fpvmove.asi", 5)
		out := f.service().Install(f.root)
		if out.OK || !strings.Contains(out.Summary, "several fpv-move packages") || f.exists(".mgs3mod") {
			t.Fatalf("%+v", out)
		}
	})
	t.Run("unmanaged file in the way", func(t *testing.T) {
		f := newFixture(t)
		f.kitFiles()
		put(t, f.root, "fpvmove.asi", []byte("someone else's"))
		out := f.service().Install(f.root)
		if out.OK || !strings.Contains(out.Summary, "fpvmove.asi already exists") || !strings.HasPrefix(out.Failed, "Store fpv-move") {
			t.Fatalf("%+v", out)
		}
		if f.mods()[profile.LoaderID].Enabled != true {
			t.Fatal("steps before the failure were not kept")
		}
	})
	t.Run("busy", func(t *testing.T) {
		f := newFixture(t)
		f.kitFiles()
		s := f.service()
		s.busy.Lock()
		out := s.Install(f.root)
		s.busy.Unlock()
		if out.OK || !strings.Contains(out.Summary, "another action is still running") || f.exists(".mgs3mod") {
			t.Fatalf("%+v", out)
		}
	})
}

func TestModsSwitchesAndAdd(t *testing.T) {
	f := newFixture(t)
	f.kitFiles()
	s := f.service()
	extra := f.pkg(t.TempDir(), qcamoID, "1.0.4", "qcamo.asi", 4)
	out := s.AddPackage(f.root, extra)
	mustOK(t, out)
	if f.mods()[qcamoID].Enabled {
		t.Fatal("added package is on")
	}
	mustOK(t, s.AddPackage(f.root, extra))
	mustOK(t, s.Install(f.root))
	out = s.SetEnabled(f.root, qcamoID, true)
	if out.OK || !strings.Contains(out.Summary, "Conflict: qcamo.asi is already changed by the mod qcamo-face") {
		t.Fatalf("conflict: %+v", out)
	}
	mustOK(t, s.SetEnabled(f.root, faceID, false))
	mustOK(t, s.SetEnabled(f.root, qcamoID, true))
	if o := s.Status(f.root); o.DeltaState != "partial" || !strings.Contains(o.Delta, "missing QCamo face paint") {
		t.Fatalf("status: %+v", o)
	}
	mustOK(t, s.Verify(f.root))
	out = s.AddPackage(f.root, filepath.Join(f.kit, "missing.zip"))
	if out.OK || out.Summary != "This file is not a valid mod package." {
		t.Fatalf("bad package: %+v", out)
	}
}

func TestDetectPrefersRememberedFolder(t *testing.T) {
	f := newFixture(t)
	s := f.service()
	if d := s.Detect(); d.Selected != "" && !HasGameExe(d.Selected) {
		t.Fatalf("detect selected a folder without the game: %+v", d)
	}
	put(t, f.root, GameExe, []byte("x"))
	if err := s.Remember(f.root); err != nil {
		t.Fatal(err)
	}
	if d := s.Detect(); d.Selected != f.root || d.Folders[0] != f.root {
		t.Fatalf("remembered folder not selected: %+v", d)
	}
	os.Remove(filepath.Join(f.root, GameExe))
	if d := s.Detect(); d.Selected == f.root {
		t.Fatalf("selected a folder without the game: %+v", d)
	}
}
