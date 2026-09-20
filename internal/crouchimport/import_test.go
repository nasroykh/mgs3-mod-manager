package crouchimport

import (
	"bytes"
	"mgs3mod/internal/packagefmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertRejectsWrongSizeAndHash(t *testing.T) {
	root := t.TempDir()
	wrongSize := filepath.Join(root, "wrong-size.zip")
	if err := os.WriteFile(wrongSize, []byte("not the pinned archive"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Convert(wrongSize, filepath.Join(root, "out.zip"), true); err == nil {
		t.Fatal("accepted wrong-size archive")
	}
	wrongHash := filepath.Join(root, "wrong-hash.zip")
	f, err := os.Create(wrongHash)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(UpstreamBytes); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Convert(wrongHash, filepath.Join(root, "out2.zip"), true); err == nil {
		t.Fatal("accepted correct-size wrong-hash archive")
	}
}

func TestConvertPreservesExistingOutput(t *testing.T) {
	archive := os.Getenv("MGS3MOD_CROUCH_ARCHIVE")
	if archive == "" {
		t.Skip("MGS3MOD_CROUCH_ARCHIVE not set")
	}
	root := t.TempDir()
	out := filepath.Join(root, "existing.zip")
	original := []byte("keep me")
	if err := os.WriteFile(out, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Convert(archive, out, false); err == nil {
		t.Fatal("accepted existing output")
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatal("existing output changed")
	}
}

func TestRealArchiveDryRunDeterministicAndComplete(t *testing.T) {
	archive := os.Getenv("MGS3MOD_CROUCH_ARCHIVE")
	if archive == "" {
		t.Skip("MGS3MOD_CROUCH_ARCHIVE not set")
	}
	root := t.TempDir()
	dry := filepath.Join(root, "dry.zip")
	if _, err := Convert(archive, dry, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dry); !os.IsNotExist(err) {
		t.Fatalf("dry-run output exists: %v", err)
	}
	if matches, err := filepath.Glob(filepath.Join(root, ".crouchimport-*")); err != nil || len(matches) != 0 {
		t.Fatalf("dry-run staging artifacts: %v %v", matches, err)
	}
	out1 := filepath.Join(root, "one.zip")
	out2 := filepath.Join(root, "two.zip")
	if _, err := Convert(archive, out1, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Convert(archive, out2, false); err != nil {
		t.Fatal(err)
	}
	b1, err := os.ReadFile(out1)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := os.ReadFile(out2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b1, b2) {
		t.Fatal("repeated conversions differ")
	}
	pkg, err := packagefmt.Load(out1, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkg.Manifest.Files) != 119 || len(pkg.Payloads) != 119 {
		t.Fatalf("got %d manifest files and %d payloads, want 119 each", len(pkg.Manifest.Files), len(pkg.Payloads))
	}
	for source := range pkg.Payloads {
		if filepath.Base(source) == "d3d11.dll" {
			t.Fatal("d3d11.dll included")
		}
	}
}
