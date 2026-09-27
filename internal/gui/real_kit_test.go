package gui

import (
	"mgs3mod/internal/profile"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRealKitOnGameCopy runs the window app's flows with the compiled game
// profile, the real process check and the real kit packages. Set
// MGS3MOD_GUI_GAME_COPY to a disposable copy of a supported game folder (never
// the installed game) and MGS3MOD_GUI_KIT to an extracted kit folder.
func TestRealKitOnGameCopy(t *testing.T) {
	root, kitDir := os.Getenv("MGS3MOD_GUI_GAME_COPY"), os.Getenv("MGS3MOD_GUI_KIT")
	if root == "" || kitDir == "" {
		t.Skip("set MGS3MOD_GUI_GAME_COPY and MGS3MOD_GUI_KIT")
	}
	if strings.EqualFold(filepath.Clean(root), filepath.Clean(profile.Root)) {
		t.Fatal("refusing the installed game folder; use a copy")
	}
	s := Production(kitDir, "")
	if c := s.Doctor(root); !c.OK {
		t.Fatalf("doctor: %s\n%s", c.Summary, c.Details)
	}
	out := s.Install(root)
	mustOK(t, out)
	t.Logf("install: %s\n  %s", out.Summary, strings.Join(out.Steps, "\n  "))
	o := s.Status(root)
	if o.DeltaState != "installed" || o.Problem != nil {
		t.Fatalf("status after install: %+v", o)
	}
	t.Logf("status: %s; %s; mods %+v", o.Delta, o.Kit, o.Mods)
	for _, target := range []string{"wininet.dll", "fpvmove.asi", "qcamo.asi"} {
		if _, err := os.Stat(filepath.Join(root, target)); err != nil {
			t.Fatalf("%s: %v", target, err)
		}
	}
	mustOK(t, s.Install(root))
	mustOK(t, s.SetEnabled(root, faceID, false))
	mustOK(t, s.SetEnabled(root, faceID, true))
	mustOK(t, s.Verify(root))
	out = s.Uninstall(root)
	mustOK(t, out)
	t.Logf("uninstall: %s\n  %s", out.Summary, strings.Join(out.Steps, "\n  "))
	for _, target := range []string{"wininet.dll", "fpvmove.asi", "qcamo.asi"} {
		if _, err := os.Lstat(filepath.Join(root, target)); !os.IsNotExist(err) {
			t.Fatalf("%s left behind: %v", target, err)
		}
	}
	if o = s.Status(root); o.DeltaState != "missing" || len(o.Mods) != 0 {
		t.Fatalf("status after uninstall: %+v", o)
	}
}
