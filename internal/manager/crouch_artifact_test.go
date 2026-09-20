package manager

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
)

// These tests copy verified local baseline bytes into isolated fixtures. They
// never mutate the source installation and do not redistribute its assets.
func crouchFixture(t *testing.T) (*Manager, string, string, *packagefmt.Package) {
	t.Helper()
	archive := os.Getenv("MGS3MOD_CROUCH_PACKAGE")
	originalRoot := os.Getenv("MGS3MOD_CROUCH_BASELINE_ROOT")
	if archive == "" || originalRoot == "" {
		t.Skip("set MGS3MOD_CROUCH_PACKAGE and MGS3MOD_CROUCH_BASELINE_ROOT for real-artifact checks")
	}
	pkg, err := packagefmt.Load(archive, false)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.ID != profile.CrouchID || len(pkg.Manifest.Files) != 119 {
		t.Fatal("wrong Crouch Walk package")
	}
	m, root := fixture(t)
	allowASILoader(m)
	for _, f := range pkg.Manifest.Files {
		parent := filepath.Dir(filepath.FromSlash(f.Target))
		info, err := os.Stat(filepath.Join(originalRoot, parent))
		if err != nil || !info.IsDir() {
			t.Fatalf("required original parent directory %s: %v", parent, err)
		}
		if err := os.MkdirAll(filepath.Join(root, parent), 0700); err != nil {
			t.Fatal(err)
		}
		if f.OriginalAbsent {
			continue
		}
		b, err := os.ReadFile(filepath.Join(originalRoot, filepath.FromSlash(f.Target)))
		if err != nil || transaction.Hash(b) != f.OriginalSHA256 {
			t.Fatalf("baseline mismatch %s: %v", f.Target, err)
		}
		put(t, root, f.Target, b)
	}
	return m, root, archive, pkg
}

func assertCrouchState(t *testing.T, root string, pkg *packagefmt.Package, enabled bool) {
	t.Helper()
	for _, f := range pkg.Manifest.Files {
		want := f.OriginalSHA256
		if enabled {
			want = f.PayloadSHA256
		}
		if want == "" {
			assertAbsent(t, root, f.Target)
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Target)))
		if err != nil || transaction.Hash(b) != want {
			t.Fatalf("wrong target state %s: %v", f.Target, err)
		}
	}
	assertAbsent(t, root, "d3d11.dll")
}

func TestCrouchArtifactLifecycle(t *testing.T) {
	m, root, archive, pkg := crouchFixture(t)
	invoke(t, m, "init", "")
	invoke(t, m, "add", makePluginPackage(t, root, "qcamo"))
	invoke(t, m, "enable", "qcamo")
	invoke(t, m, "add", archive)
	before := snapshot(t, root)
	if _, err := m.Run("enable", profile.CrouchID, Options{DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("dry run mutated fixture")
	}
	m.config.CheckASILoader = func(string) error { return errors.New("missing loader") }
	if _, err := m.Run("enable", profile.CrouchID, Options{}); ExitCode(err) != 3 {
		t.Fatalf("loader gate: %v", err)
	}
	allowASILoader(m)
	invoke(t, m, "enable", profile.CrouchID)
	assertCrouchState(t, root, pkg, true)
	assertPlugin(t, root)
	invoke(t, m, "verify", "")
	invoke(t, m, "disable", profile.CrouchID)
	assertCrouchState(t, root, pkg, false)
	assertPlugin(t, root)
	invoke(t, m, "enable", profile.CrouchID)
	invoke(t, m, "remove", profile.CrouchID)
	assertCrouchState(t, root, pkg, false)
	assertPlugin(t, root)
	invoke(t, m, "verify", "")
}

func TestCrouchArtifactDriftAndForeignFiles(t *testing.T) {
	m, root, archive, pkg := crouchFixture(t)
	invoke(t, m, "init", "")
	put(t, root, "mgs3crouchwalk.ini", []byte("foreign"))
	before := snapshot(t, root)
	if _, err := m.Run("add", archive, Options{}); ExitCode(err) != 4 {
		t.Fatalf("foreign config accepted: %v", err)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("foreign rejection mutated fixture")
	}
	if err := os.Remove(filepath.Join(root, "mgs3crouchwalk.ini")); err != nil {
		t.Fatal(err)
	}
	invoke(t, m, "add", archive)
	invoke(t, m, "enable", profile.CrouchID)
	put(t, root, "mgs3crouchwalk.ini", []byte("edited"))
	before = snapshot(t, root)
	if _, err := m.Run("disable", profile.CrouchID, Options{}); ExitCode(err) != 4 {
		t.Fatalf("config drift accepted: %v", err)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("drift rejection mutated fixture")
	}
	put(t, root, "mgs3crouchwalk.ini", pkg.Payloads["payload/mgs3crouchwalk.ini"])
	invoke(t, m, "restore", "")
	assertCrouchState(t, root, pkg, false)
}

func TestCrouchArtifactMixedRecovery(t *testing.T) {
	for _, operation := range []string{"enable", "disable"} {
		for _, point := range []string{"ready", "replaced-60", "committed"} {
			t.Run(operation+"/"+point, func(t *testing.T) {
				m, root, archive, pkg := crouchFixture(t)
				invoke(t, m, "init", "")
				invoke(t, m, "add", archive)
				if operation == "disable" {
					invoke(t, m, "enable", profile.CrouchID)
				}
				killAt(t, m, operation, profile.CrouchID, point)
				invoke(t, m, "recover", "")
				enabled := operation == "enable" && point == "committed" || operation == "disable" && point != "committed"
				assertCrouchState(t, root, pkg, enabled)
				invoke(t, m, "verify", "")
			})
		}
	}
}

func TestCrouchArtifactExplicitMissingRecovery(t *testing.T) {
	m, root, archive, pkg := crouchFixture(t)
	invoke(t, m, "init", "")
	invoke(t, m, "add", archive)
	killAt(t, m, "enable", profile.CrouchID, "replaced-0")
	// The first sorted target is a present-origin French animation.
	target := "assets/mtar/fr/00023291.mtar"
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(target))); err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, root)
	if _, err := m.Run("recover", "", Options{}); err == nil {
		t.Fatal("plain recovery recreated externally missing animation")
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("blocked recovery mutated fixture")
	}
	if _, err := m.Run("recover", "", Options{RestoreMissing: []string{target}}); err != nil {
		t.Fatal(err)
	}
	assertCrouchState(t, root, pkg, false)
	invoke(t, m, "verify", "")
}
