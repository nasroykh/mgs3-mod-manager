//go:build windows

package winfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveNewNeverOverwritesDestination(t *testing.T) {
	root := fixtureDir(t)
	src, dst := filepath.Join(root, "stage"), filepath.Join(root, "qcamo.asi")
	if err := os.WriteFile(src, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := MoveNew(src, dst); err == nil {
		t.Fatal("replaced destination")
	}
	for path, want := range map[string]string{src: "new", dst: "foreign"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatalf("%s: %q %v", path, got, err)
		}
	}
	if err := os.Remove(dst); err != nil {
		t.Fatal(err)
	}
	if err := MoveNew(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil || string(got) != "new" {
		t.Fatalf("promoted: %q %v", got, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("source still present")
	}
}
