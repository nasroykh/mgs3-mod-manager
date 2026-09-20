package manager

import (
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testLoaderMod(enabled bool) Mod {
	files := make([]packagefmt.File, 0, len(profile.LoaderFiles))
	for _, file := range profile.LoaderFiles {
		files = append(files, packagefmt.File{Source: "payload/" + file.Path, Target: file.Path, OriginalAbsent: true, PayloadSHA256: file.SHA256, PayloadBytes: profile.LoaderBytes})
	}
	return Mod{Manifest: packagefmt.Manifest{SchemaVersion: 3, ID: profile.LoaderID, Version: profile.LoaderVersion, Name: "Ultimate ASI Loader", Profile: profile.ID, Files: files}, Digest: transaction.Hash([]byte("loader-package")), Enabled: enabled}
}

func testAbsentBaseline(target string) Baseline {
	return Baseline{Target: target, Captured: time.Now().UTC().Format(time.RFC3339Nano), Profile: profile.ID, Absent: true}
}

func TestLoaderRejectsMissingAndArbitraryDLLs(t *testing.T) {
	root := t.TempDir()
	m := New(Config{Root: root})
	if err := m.asiLoader(emptyState()); ExitCode(err) != 3 {
		t.Fatalf("missing loader: %v", err)
	}
	for _, file := range profile.LoaderFiles {
		if err := os.WriteFile(filepath.Join(root, file.Path), []byte("not a trusted loader"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.asiLoader(emptyState()); ExitCode(err) != 3 {
		t.Fatalf("arbitrary DLLs accepted: %v", err)
	}
}

func TestLoaderUsesInternalTestSeam(t *testing.T) {
	called := false
	m := New(Config{Root: "fixture", CheckASILoader: func(root string) error { called = root == "fixture"; return nil }})
	if err := m.asiLoader(emptyState()); err != nil || !called {
		t.Fatalf("loader seam: %v %v", called, err)
	}
}

func TestLoaderEnableDoesNotDependOnItself(t *testing.T) {
	m, _ := fixture(t)
	state := emptyState()
	state.Mods[profile.LoaderID] = testLoaderMod(false)
	p, _, err := m.openBuildForTest(state, "enable", profile.LoaderID)
	if err != nil {
		t.Fatalf("loader enable required loader: %v", err)
	}
	if !p.After.Mods[profile.LoaderID].Enabled || len(p.Changes) != len(profile.LoaderFiles) {
		t.Fatalf("wrong loader enable plan: %+v", p)
	}
}

func TestSchema2ASIUsingLoaderIDCannotBypassDependency(t *testing.T) {
	m, _ := fixture(t)
	state := emptyState()
	payload := testPluginPE()
	manifest := packagefmt.Manifest{SchemaVersion: 2, ID: profile.LoaderID, Version: "1.0.0", Name: "spoof", Profile: profile.ID, Files: []packagefmt.File{{Source: "payload/spoof.asi", Target: "spoof.asi", OriginalAbsent: true, PayloadSHA256: transaction.Hash(payload), PayloadBytes: int64(len(payload))}}}
	state.Mods[profile.LoaderID] = Mod{Manifest: manifest, Digest: transaction.Hash([]byte("spoof-package"))}
	if _, _, err := m.openBuildForTest(state, "enable", profile.LoaderID); ExitCode(err) != 3 {
		t.Fatalf("schema-2 ASI using loader ID bypassed dependency: %v", err)
	}
}

func TestLoaderEnableRejectsConfigurationOverrides(t *testing.T) {
	for _, config := range loaderConfigPaths {
		t.Run(config, func(t *testing.T) {
			m, root := fixture(t)
			put(t, root, config, []byte("[GlobalSets]"))
			state := emptyState()
			state.Mods[profile.LoaderID] = testLoaderMod(false)
			if _, _, err := m.openBuildForTest(state, "enable", profile.LoaderID); ExitCode(err) != 4 {
				t.Fatalf("configuration override accepted: %v", err)
			}
		})
	}
}

func TestLoaderDisableAndRemoveBlockedByEnabledASI(t *testing.T) {
	m, _ := fixture(t)
	for _, operation := range []string{"disable", "remove"} {
		t.Run(operation, func(t *testing.T) {
			state := emptyState()
			state.Mods[profile.LoaderID] = testLoaderMod(true)
			plugin := packagefmt.Manifest{SchemaVersion: 2, ID: "qcamo", Version: "1.0.0", Name: "QCamo", Profile: profile.ID, Files: []packagefmt.File{{Source: "payload/qcamo.asi", Target: "qcamo.asi", OriginalAbsent: true, PayloadSHA256: transaction.Hash(testPluginPE()), PayloadBytes: int64(len(testPluginPE()))}}}
			state.Mods["qcamo"] = Mod{Manifest: plugin, Digest: transaction.Hash([]byte("qcamo-package")), Enabled: true}
			if _, _, err := m.openBuildForTest(state, operation, profile.LoaderID); ExitCode(err) != 4 {
				t.Fatalf("%s with enabled ASI: %v", operation, err)
			}
		})
	}
}

func TestRestoreDisablesLoaderAndASIAtomically(t *testing.T) {
	m, _ := fixture(t)
	state := emptyState()
	state.Mods[profile.LoaderID] = testLoaderMod(true)
	plugin := packagefmt.Manifest{SchemaVersion: 2, ID: "qcamo", Version: "1.0.0", Name: "QCamo", Profile: profile.ID, Files: []packagefmt.File{{Source: "payload/qcamo.asi", Target: "qcamo.asi", OriginalAbsent: true, PayloadSHA256: transaction.Hash(testPluginPE()), PayloadBytes: int64(len(testPluginPE()))}}}
	state.Mods["qcamo"] = Mod{Manifest: plugin, Digest: transaction.Hash([]byte("qcamo-package")), Enabled: true}
	state.Baselines[profile.Key("qcamo.asi")] = testAbsentBaseline("qcamo.asi")
	for _, file := range profile.LoaderFiles {
		state.Baselines[profile.Key(file.Path)] = testAbsentBaseline(file.Path)
	}
	p, _, err := m.openBuildForTest(state, "restore", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.After.Mods[profile.LoaderID].Enabled || p.After.Mods["qcamo"].Enabled {
		t.Fatalf("restore did not disable both packages: %+v", p.After.Mods)
	}
}

func (m *Manager) openBuildForTest(state State, command, arg string) (*plan, map[string]input, error) {
	if _, err := m.Run("init", "", Options{}); err != nil {
		return nil, nil, err
	}
	s, err := m.open(false)
	if err != nil {
		return nil, nil, err
	}
	defer s.close()
	return s.build(history{state: state}, command, arg)
}
