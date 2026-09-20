package manager

import (
	"errors"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"os"
	"path/filepath"
	"testing"
)

func loaderArtifact(t *testing.T) string {
	t.Helper()
	archive := os.Getenv("MGS3MOD_ASI_LOADER_PACKAGE")
	if archive == "" {
		t.Skip("set MGS3MOD_ASI_LOADER_PACKAGE to the generated release package")
	}
	pkg, err := packagefmt.Load(archive, false)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.ID != profile.LoaderID || pkg.Manifest.Version != profile.LoaderVersion || len(pkg.Manifest.Files) != len(profile.LoaderFiles) {
		t.Fatal("not the pinned ASI loader release package")
	}
	return archive
}

func assertLoaderFiles(t *testing.T, root string, present bool) {
	t.Helper()
	for _, file := range profile.LoaderFiles {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if !present {
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("%s exists or cannot be inspected: %v", file.Path, err)
			}
			continue
		}
		if err != nil || transaction.Hash(data) != file.SHA256 {
			t.Fatalf("%s differs from pinned loader: %v", file.Path, err)
		}
	}
}

func assertQCamoRelease(t *testing.T, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, pluginTarget))
	if err != nil || transaction.Hash(data) != "4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193" {
		t.Fatalf("QCamo release payload mismatch: %v", err)
	}
}

func TestLoaderReleaseArtifactLifecycleAndDependency(t *testing.T) {
	loader := loaderArtifact(t)
	qcamo := filepath.Join(filepath.Dir(loader), "qcamo-1.0.4.mgs3mod.zip")
	if _, err := packagefmt.Load(qcamo, false); err != nil {
		t.Fatalf("load companion QCamo package: %v", err)
	}
	m, root := fixture(t)
	invoke(t, m, "init", "")
	invoke(t, m, "add", loader)
	invoke(t, m, "add", qcamo)
	assertLoaderFiles(t, root, false)
	if _, err := m.Run("enable", "qcamo", Options{}); ExitCode(err) != 3 {
		t.Fatalf("ASI enabled before loader: %v", err)
	}
	invoke(t, m, "enable", profile.LoaderID)
	assertLoaderFiles(t, root, true)
	invoke(t, m, "enable", "qcamo")
	invoke(t, m, "verify", "")
	for _, operation := range []string{"disable", "remove"} {
		if _, err := m.Run(operation, profile.LoaderID, Options{}); ExitCode(err) != 4 {
			t.Fatalf("%s loader with enabled ASI: %v", operation, err)
		}
	}
	invoke(t, m, "restore", "")
	assertLoaderFiles(t, root, false)
	assertAbsent(t, root, pluginTarget)
	invoke(t, m, "enable", profile.LoaderID)
	invoke(t, m, "enable", "qcamo")
	invoke(t, m, "disable", "qcamo")
	invoke(t, m, "remove", "qcamo")
	assertLoaderFiles(t, root, true)
	invoke(t, m, "disable", profile.LoaderID)
	invoke(t, m, "remove", profile.LoaderID)
	assertLoaderFiles(t, root, false)
	if len(invoke(t, m, "list", "").State.Mods) != 0 {
		t.Fatal("package removal retained managed packages")
	}
}

func TestLoaderAndASICrashRecoveryDuringRestore(t *testing.T) {
	loader := loaderArtifact(t)
	qcamo := filepath.Join(filepath.Dir(loader), "qcamo-1.0.4.mgs3mod.zip")
	points := []string{"prepared", "ready", "intent-0", "replaced-0", "applied-0", "intent-1", "replaced-1", "applied-1", "committed", "cached", "cleaned"}
	for _, point := range points {
		t.Run(point, func(t *testing.T) {
			m, root := fixture(t)
			invoke(t, m, "init", "")
			invoke(t, m, "add", loader)
			invoke(t, m, "add", qcamo)
			invoke(t, m, "enable", profile.LoaderID)
			invoke(t, m, "enable", "qcamo")
			killAt(t, m, "restore", "", point)
			invoke(t, m, "recover", "")
			committed := point == "committed" || point == "cached" || point == "cleaned"
			assertLoaderFiles(t, root, !committed)
			if committed {
				assertAbsent(t, root, pluginTarget)
			} else {
				assertQCamoRelease(t, root)
			}
			invoke(t, m, "verify", "")
		})
	}
}

func TestLoaderReleaseArtifactRefusesPreexistingExactFile(t *testing.T) {
	archive := loaderArtifact(t)
	pkg, err := packagefmt.Load(archive, false)
	if err != nil {
		t.Fatal(err)
	}
	m, root := fixture(t)
	invoke(t, m, "init", "")
	for _, file := range pkg.Manifest.Files {
		put(t, root, file.Target, pkg.Payloads[file.Source])
	}
	if _, err := m.Run("add", archive, Options{}); ExitCode(err) != 4 {
		t.Fatalf("preexisting exact loader file was adopted: %v", err)
	}
	assertLoaderFiles(t, root, true)
}

func TestMissingLoaderDoesNotBlockASIDisableOrRemove(t *testing.T) {
	loader := loaderArtifact(t)
	qcamo := filepath.Join(filepath.Dir(loader), "qcamo-1.0.4.mgs3mod.zip")
	for _, operation := range []string{"disable", "remove"} {
		t.Run(operation, func(t *testing.T) {
			m, root := fixture(t)
			invoke(t, m, "init", "")
			invoke(t, m, "add", loader)
			invoke(t, m, "add", qcamo)
			invoke(t, m, "enable", profile.LoaderID)
			invoke(t, m, "enable", "qcamo")
			for _, file := range profile.LoaderFiles {
				if err := os.Remove(filepath.Join(root, filepath.FromSlash(file.Path))); err != nil {
					t.Fatal(err)
				}
			}
			invoke(t, m, operation, "qcamo")
			assertAbsent(t, root, pluginTarget)
		})
	}
}

func TestLoaderReleaseArtifactCrashRecovery(t *testing.T) {
	archive := loaderArtifact(t)
	points := []string{"prepared", "ready", "intent-0", "replaced-0", "applied-0", "committed", "cached", "cleaned"}
	for _, operation := range []string{"enable", "disable"} {
		for _, point := range points {
			t.Run(operation+"/"+point, func(t *testing.T) {
				m, root := fixture(t)
				invoke(t, m, "init", "")
				invoke(t, m, "add", archive)
				if operation == "disable" {
					invoke(t, m, "enable", profile.LoaderID)
				}
				killAt(t, m, operation, profile.LoaderID, point)
				invoke(t, m, "recover", "")
				committed := point == "committed" || point == "cached" || point == "cleaned"
				wantPresent := operation == "enable" && committed || operation == "disable" && !committed
				assertLoaderFiles(t, root, wantPresent)
				invoke(t, m, "verify", "")
			})
		}
	}
}
