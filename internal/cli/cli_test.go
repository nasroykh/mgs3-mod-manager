package cli

import (
	"bytes"
	"encoding/json"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/profile"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsingContract(t *testing.T) {
	for _, args := range [][]string{{"enable", "camo-test", "--json", "--dry-run"}, {"pack", "authoring", "--out", "result.zip"}, {"recover", "--restore-missing", "textures/flatlist/_win/a.ctxr", "--restore-missing", "hqtex/flatlist/_win/b.ctxr"}, {"restore", "--baseline"}} {
		if _, err := parse(args); err != nil {
			t.Errorf("%v: %v", args, err)
		}
	}
	for _, args := range [][]string{{"enable"}, {"restore"}, {"restore", "--baseline", "extra"}, {"pack", "a"}, {"list", "--dry-run"}, {"enable", "x", "--force"}, {"recover", "--out", "x"}, {"enable", "x", "--restore-missing", "x"}, {"list", "--json", "--json"}, {"init", "--baseline"}} {
		if _, err := parse(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
	for _, args := range [][]string{{"import-crouch"}, {"import-crouch", "archive.zip"}, {"import-crouch", "a", "b", "--out", "x.zip"}} {
		if _, err := parse(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
	if p, err := parse([]string{"import-crouch", "archive.zip", "--out", "out.zip", "--dry-run", "--json"}); err != nil || p.command != "import-crouch" || !p.options.DryRun || p.out != "out.zip" {
		t.Fatalf("import-crouch parse: %#v %v", p, err)
	}
}

func TestPackDryRunChecksOutputWithoutWrites(t *testing.T) {
	parent, err := filepath.Abs("../../.cache/cli-tests")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp(parent, "pack-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	author := filepath.Join(root, "author")
	if err = os.MkdirAll(filepath.Join(author, "payload"), 0700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schemaVersion":1,"id":"test","version":"0.1.0","name":"test","profile":"` + profile.ID + `","files":[{"source":"payload/a.ctxr","target":"textures/flatlist/_win/a.ctxr","originalSha256":"` + strings.Repeat("0", 64) + `"}]}`
	if err = os.WriteFile(filepath.Join(author, "manifest.json"), []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(author, "payload", "a.ctxr"), []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "out.zip")
	var out, diagnostics bytes.Buffer
	args := []string{"pack", author, "--out", output, "--dry-run", "--json"}
	if code := Run(args, &out, &diagnostics, nil); code != 0 {
		t.Fatalf("dry pack: %d %s", code, out.String())
	}
	if _, err = os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("dry pack created output")
	}
	if err = os.WriteFile(output, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if code := Run(args, &out, &diagnostics, nil); code != 2 {
		t.Fatalf("existing output accepted: %d %s", code, out.String())
	}
	b, err := os.ReadFile(output)
	if err != nil || string(b) != "existing" {
		t.Fatal("existing output changed")
	}
	args[3] = filepath.Join(root, "missing", "out.zip")
	out.Reset()
	if code := Run(args, &out, &diagnostics, nil); code != 2 {
		t.Fatalf("missing parent accepted: %d %s", code, out.String())
	}
}

func TestJSONErrorsAndHelpHaveNoManagerSideEffects(t *testing.T) {
	// Nil manager proves parsing errors and help do not call the filesystem layer.
	for _, args := range [][]string{{"bad", "--json"}, {"enable", "--json"}, {"--help", "--json"}, {"recover", "--help", "--json"}} {
		var out, errOut bytes.Buffer
		code := Run(args, &out, &errOut, nil)
		var value map[string]any
		if err := json.Unmarshal(out.Bytes(), &value); err != nil {
			t.Fatalf("invalid JSON: %s (%v)", out.Bytes(), err)
		}
		if errOut.Len() != 0 {
			t.Fatalf("JSON mixed diagnostics: %q", errOut.String())
		}
		if code != 0 && code != 2 {
			t.Fatalf("wrong code %d", code)
		}
	}
	var out, errOut bytes.Buffer
	if code := Run([]string{"--help"}, &out, &errOut, manager.Production()); code != 0 || out.Len() == 0 {
		t.Fatal("help failed")
	}
}
