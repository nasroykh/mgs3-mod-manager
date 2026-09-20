package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mgs3mod/internal/manager"
	"mgs3mod/internal/profile"
)

func TestRootOption(t *testing.T) {
	for _, args := range [][]string{
		{"doctor", "--game-root", `D:\SteamLibrary\MGS3`, "--json"},
		{"--game-root", `D:\SteamLibrary\MGS3`, "doctor"},
	} {
		p, err := parse(args)
		if err != nil || p.root != `D:\SteamLibrary\MGS3` {
			t.Fatalf("root parse: %+v %v", p, err)
		}
	}
	for _, args := range [][]string{{"doctor", "--game-root"}, {"doctor", "--game-root", ""}, {"doctor", "--game-root", "one", "--game-root", "two"}} {
		if _, err := parse(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestSelectedRootPreservesFingerprintGuard(t *testing.T) {
	root := t.TempDir()
	m := manager.New(manager.Config{Root: "not-selected", Core: []profile.Fingerprint{{Path: "game.exe", SHA256: strings.Repeat("0", 64)}}})
	var out, diagnostics bytes.Buffer
	code := Run([]string{"init", "--game-root", root, "--json"}, &out, &diagnostics, m)
	if code != 3 || !strings.Contains(out.String(), "game.exe") {
		t.Fatalf("fingerprint guard: %d %s", code, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".mgs3mod")); !os.IsNotExist(err) {
		t.Fatal("failed compatibility check created state")
	}
}

func TestSelectedRootCanInitializeCompatibleInstallation(t *testing.T) {
	root := t.TempDir()
	m := manager.New(manager.Config{Root: "not-selected", Core: []profile.Fingerprint{}, CheckProcesses: func(string) error { return nil }})
	var out, diagnostics bytes.Buffer
	if code := Run([]string{"init", "--game-root", root, "--json"}, &out, &diagnostics, m); code != 0 {
		t.Fatalf("init: %d %s", code, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".mgs3mod", "INIT_COMMITTED")); err != nil {
		t.Fatal(err)
	}
}
