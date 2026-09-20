package manager

import (
	"os"
	"path/filepath"
	"testing"

	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/transaction"
)

// This opt-in integration test manages the real release payload in a synthetic
// installation. It never loads the DLL or touches the real game installation.
func TestQCamoReleaseArtifactLifecycle(t *testing.T) {
	archive := os.Getenv("MGS3MOD_QCAMO_PACKAGE")
	if archive == "" {
		t.Skip("set MGS3MOD_QCAMO_PACKAGE to the generated release package")
	}
	pkg, err := packagefmt.Load(archive, false)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.ID != "qcamo" || len(pkg.Manifest.Files) != 1 {
		t.Fatal("not the QCamo release package")
	}
	want := "4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193"
	if pkg.Manifest.Files[0].PayloadSHA256 != want {
		t.Fatal("not the pinned official QCamo payload")
	}
	m, root := fixture(t)
	allowASILoader(m)
	invoke(t, m, "init", "")
	invoke(t, m, "add", archive)
	assertAbsent(t, root, "qcamo.asi")
	for i := 0; i < 2; i++ {
		invoke(t, m, "enable", "qcamo")
		data, err := os.ReadFile(filepath.Join(root, "qcamo.asi"))
		if err != nil || transaction.Hash(data) != want {
			t.Fatalf("release payload mismatch: %v", err)
		}
		invoke(t, m, "verify", "")
		if i == 0 {
			invoke(t, m, "disable", "qcamo")
		} else {
			invoke(t, m, "remove", "qcamo")
		}
		assertAbsent(t, root, "qcamo.asi")
	}
	invoke(t, m, "verify", "")
}
