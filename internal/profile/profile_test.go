package profile

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathPolicy(t *testing.T) {
	valid := []string{
		"textures/flatlist/_win/sna_def_olive.bmp.ctxr",
		"hqtex/flatlist/_win/sna_def_olive.bmp.ctxr",
	}
	for _, path := range valid {
		if err := Target(path); err != nil {
			t.Fatalf("Target(%q): %v", path, err)
		}
	}
	invalid := []string{
		"textures\\flatlist\\_win\\x.ctxr",
		"textures/flatlist/_win/../x.ctxr",
		"textures/flatlist/_win/x.ctxr:evil",
		"textures/flatlist/_win/CON.ctxr",
		"textures/other/x.ctxr",
		"textures/flatlist/_win/x.dds",
	}
	for _, path := range invalid {
		if err := Target(path); err == nil {
			t.Errorf("Target(%q) unexpectedly succeeded", path)
		}
	}
	if got := Key(`HQTEX\FLATLIST\_WIN\X.CTX`); got != "hqtex/flatlist/_win/x.ctx" {
		t.Fatalf("Key returned %q", got)
	}
}

func TestPluginTargetPolicy(t *testing.T) {
	for _, path := range []string{"dinput8.asi", "plugin-1.asi"} {
		if err := PluginTarget(path); err != nil {
			t.Errorf("PluginTarget(%q): %v", path, err)
		}
	}
	for _, path := range []string{".asi", "DINPUT8.asi", "dinput8.ASI", "plugins/dinput8.asi", "../dinput8.asi", "dinput8.asi/"} {
		if err := PluginTarget(path); err == nil {
			t.Errorf("PluginTarget(%q) unexpectedly succeeded", path)
		}
	}
}

func TestCheckAcceptsAlternateHashes(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "game.exe"), []byte("wrapped"), 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("wrapped"))
	wrapped := hex.EncodeToString(sum[:])
	other := strings.Repeat("0", 64)
	if err := Check(root, []Fingerprint{{Path: "game.exe", SHA256: other, Alternates: []string{wrapped}}}); err != nil {
		t.Fatalf("alternate rejected: %v", err)
	}
	if err := Check(root, []Fingerprint{{Path: "game.exe", SHA256: wrapped}}); err != nil {
		t.Fatalf("primary rejected: %v", err)
	}
	err := Check(root, []Fingerprint{{Path: "game.exe", SHA256: other, Alternates: []string{strings.Repeat("1", 64)}}})
	if err == nil || !strings.Contains(err.Error(), " or ") {
		t.Fatalf("mismatch accepted or unreported: %v", err)
	}
	for _, bad := range []string{"abc", strings.Repeat("A", 64), strings.Repeat("g", 64)} {
		if err := Check(root, []Fingerprint{{Path: "game.exe", SHA256: wrapped, Alternates: []string{bad}}}); err == nil {
			t.Errorf("malformed alternate %q accepted", bad)
		}
	}
}

func TestCoreExecutableAcceptsSteamWrappedBuild(t *testing.T) {
	exe := Core[0]
	if exe.Path != "METAL GEAR SOLID3.exe" || !exe.Matches(SteamWrappedExecutable) || !exe.Matches(exe.SHA256) {
		t.Fatalf("executable fingerprint: %+v", exe)
	}
	for _, f := range Core[1:] {
		if len(f.Alternates) != 0 {
			t.Errorf("%s has alternates: %v", f.Path, f.Alternates)
		}
	}
}
