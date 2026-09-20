package packagefmt

import (
	"encoding/json"
	"os"
	"testing"

	"mgs3mod/internal/profile"
)

func pinnedLoaderManifest() Manifest {
	m := Manifest{SchemaVersion: 3, ID: profile.LoaderID, Version: profile.LoaderVersion, Name: "Ultimate ASI Loader", Profile: profile.ID}
	for _, f := range profile.LoaderFiles {
		m.Files = append(m.Files, File{Source: "payload/" + f.Path, Target: f.Path, OriginalAbsent: true, PayloadSHA256: profile.LoaderSHA256, PayloadBytes: profile.LoaderBytes})
	}
	return m
}

func TestLoaderManifestIsPinnedAndIsolated(t *testing.T) {
	valid := pinnedLoaderManifest()
	encoded, _ := json.Marshal(valid)
	if _, err := Parse(encoded, false); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Manifest){
		"schema2 DLL":            func(m *Manifest) { m.SchemaVersion = 2 },
		"wrong ID":               func(m *Manifest) { m.ID = "untrusted" },
		"wrong version":          func(m *Manifest) { m.Version = "9.7.5" },
		"incomplete":             func(m *Manifest) { m.Files = nil },
		"arbitrary DLL":          func(m *Manifest) { m.Files[0].Target = "engine.dll" },
		"ASI in loader":          func(m *Manifest) { m.Files[0].Target = "other.asi" },
		"replace existing":       func(m *Manifest) { m.Files[0].OriginalAbsent = false },
		"has original hash":      func(m *Manifest) { m.Files[0].OriginalSHA256 = profile.LoaderSHA256 },
		"wrong payload":          func(m *Manifest) { m.Files[0].PayloadSHA256 = profile.Core[0].SHA256 },
		"missing authoring hash": func(m *Manifest) { m.Files[0].PayloadSHA256 = "" },
		"wrong size":             func(m *Manifest) { m.Files[0].PayloadBytes-- },
		"duplicate target":       func(m *Manifest) { m.Files = append(m.Files, m.Files[0]) },
		"case alias":             func(m *Manifest) { m.Files[0].Target = "WININET.dll" },
	} {
		t.Run(name, func(t *testing.T) {
			m := pinnedLoaderManifest()
			mutate(&m)
			data, _ := json.Marshal(m)
			for _, authoring := range []bool{false, true} {
				if _, err := Parse(data, authoring); err == nil {
					t.Fatalf("accepted %s authoring=%v", name, authoring)
				}
			}
		})
	}
}

func TestOfficialLoaderArtifact(t *testing.T) {
	archive := os.Getenv("MGS3MOD_ASI_LOADER_PACKAGE")
	if archive == "" {
		t.Skip("set MGS3MOD_ASI_LOADER_PACKAGE to generated pinned loader package")
	}
	pkg, err := Load(archive, false)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.SchemaVersion != 3 || pkg.Manifest.ID != profile.LoaderID {
		t.Fatal("not a loader package")
	}
	for _, f := range pkg.Manifest.Files {
		data := append([]byte(nil), pkg.Payloads[f.Source]...)
		data[len(data)-1] ^= 1
		if err := fillOrCheckPayload(&f, data, false); err == nil {
			t.Fatal("modified loader passed digest validation")
		}
	}
}
