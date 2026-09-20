package packagefmt

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestZIPRejectsEveryUndeployableWindowsCharacter(t *testing.T) {
	root := packageFixtureDir(t)
	for i, char := range []rune(`[]<>"|`) {
		source := "payload/bad" + string(char) + ".ctxr"
		var manifest Manifest
		if err := json.Unmarshal(testManifest("", 0), &manifest); err != nil {
			t.Fatal(err)
		}
		manifest.Files[0].Source = source
		data, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.Create(filepath.Join(root, fmt.Sprintf("case-%d.zip", i)))
		if err != nil {
			t.Fatal(err)
		}
		writer := zip.NewWriter(file)
		for name, body := range map[string][]byte{"manifest.json": data, source: []byte("payload")} {
			entry, e := writer.Create(name)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = entry.Write(body); e != nil {
				t.Fatal(e)
			}
		}
		if err = writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err = Load(file.Name(), true); err == nil {
			t.Fatalf("ZIP accepted undeployable character %q", char)
		}
	}
}

func TestPackageIOErrorCategory(t *testing.T) {
	err := fmt.Errorf("write package ZIP: %w", &os.PathError{Op: "write", Path: "output.zip", Err: syscall.Errno(112)})
	if ErrorCode(err) != 1 {
		t.Fatal("disk-full I/O error mapped as package validation")
	}
	if ErrorCode(&validationError{err}) != 1 {
		t.Fatal("operating-system failure hidden by validation wrapper")
	}
}

func FuzzPackagePath(f *testing.F) {
	for _, seed := range []string{"payload/a.ctxr", "../x", "payload/a[.ctxr", "CON", "payload/a:stream", "payload/a/../b"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if err := validatePackagePath(value); err == nil {
			if filepath.IsAbs(value) || filepath.ToSlash(filepath.Clean(value)) != value {
				t.Fatalf("noncanonical path accepted: %q", value)
			}
		}
	})
}
