package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mgs3mod/internal/launch"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/profile"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLaunchFormsAndRestrictions(t *testing.T) {
	valid := [][]string{
		{"launch", "--region", "us", "--language", "en", "--controller", "kbd", "--target", "startup", "--dry-run"},
		{"--dry-run", "--target", "startup", "launch", "--controller", "kbd", "--language", "en", "--region", "us"},
		{"launch", "--profile", "na-startup", "--region", "us"},
	}
	for _, args := range valid {
		if _, err := parse(args); err != nil {
			t.Fatalf("parse valid %v: %v", args, err)
		}
	}
	invalid := [][]string{
		{"launch", "--region"},
		{"launch", "--profile", ""},
		{"launch", "--region", "", "--language", "en", "--controller", "kbd", "--target", "startup"},
		{"launch", "--profile", "one", "--profile", "two"},
		{"launch", "--region", "us", "--language", "en"},
		{"launch", "--region", "us", "--language", "en", "--controller", "kbd", "--target", "menu"},
		{"launch", "--select", "--dry-run"},
		{"launch", "--select", "--profile", "na-startup"},
		{"doctor", "--profile", "na-startup"},
		{"doctor", "--profile", ""},
		{"status", "--region", "us"},
	}
	for _, args := range invalid {
		if _, err := parse(args); err == nil {
			t.Fatalf("parse accepted %v", args)
		}
	}
}

func TestValidLaunchProfileIDMatchesDomain(t *testing.T) {
	for _, id := range []string{"a", "na-startup", "x.y_z-9", strings.Repeat("a", 64)} {
		if !validLaunchProfileID(id) {
			t.Fatalf("rejected valid profile ID %q", id)
		}
	}
	for _, id := range []string{"", "Upper", ".leading", "-leading", "bad space", "bad/sep", strings.Repeat("a", 65)} {
		if validLaunchProfileID(id) {
			t.Fatalf("accepted invalid profile ID %q", id)
		}
	}
}

func TestLaunchExplicitJSONDryRunUsesSelectedRoot(t *testing.T) {
	root, m := launchFixture(t)
	if err := m.SaveLaunchConfig(launch.Seed()); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := Run([]string{"--json", "--game-root", root, "launch", "--profile", "na-startup", "--dry-run"}, &out, &diagnostics, m)
	if code != 0 || diagnostics.Len() != 0 {
		t.Fatalf("launch: code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	var response struct {
		OK     bool                  `json:"ok"`
		Result manager.LaunchReceipt `json:"result"`
		Error  any                   `json:"error"`
	}
	if err := json.Unmarshal(out.Bytes(), &response); err != nil || !response.OK || response.Error != nil {
		t.Fatalf("JSON response: %s (%v)", out.Bytes(), err)
	}
	wantArgs := []string{"-region", "us", "-lan", "en", "-selfregion", "EU", "-launcherpath", "launcher.exe", "-ctrltype", "KBD"}
	if response.Result.Executable != filepath.Join(root, "METAL GEAR SOLID3.exe") || response.Result.WorkingDir != root || !response.Result.DryRun || !equalStrings(response.Result.Arguments, wantArgs) {
		t.Fatalf("receipt: %+v", response.Result)
	}
}

func TestLaunchDirectOverridesDoNotPersist(t *testing.T) {
	root, m := launchFixture(t)
	var out, diagnostics bytes.Buffer
	code := Run([]string{"launch", "--region", "us", "--language", "en", "--controller", "kbd", "--target", "startup", "--dry-run"}, &out, &diagnostics, m)
	if code != 0 || diagnostics.Len() != 0 {
		t.Fatalf("direct dry-run: code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	if _, err := os.Stat(filepath.Join(root, "mgs3mod-launch.json")); !os.IsNotExist(err) {
		t.Fatalf("direct override persisted config: %v", err)
	}
}

func TestLaunchDirectOverridesIgnoreButPreserveCorruptPreferences(t *testing.T) {
	root, m := launchFixture(t)
	path := filepath.Join(root, "mgs3mod-launch.json")
	corrupt := []byte("{")
	if err := os.WriteFile(path, corrupt, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := Run([]string{"launch", "--region", "us", "--language", "en", "--controller", "kbd", "--target", "startup", "--dry-run"}, &out, &diagnostics, m)
	if code != 0 || diagnostics.Len() != 0 {
		t.Fatalf("direct dry-run: code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, corrupt) {
		t.Fatalf("corrupt preferences changed: %q (%v)", after, err)
	}
}

func TestLaunchMissingNamedProfileIsUsageError(t *testing.T) {
	for _, jsonMode := range []bool{false, true} {
		t.Run(fmt.Sprintf("json=%t", jsonMode), func(t *testing.T) {
			root, m := launchFixture(t)
			args := []string{"launch", "--profile", "missing", "--dry-run"}
			if jsonMode {
				args = append(args, "--json")
			}
			var out, diagnostics bytes.Buffer
			code := Run(args, &out, &diagnostics, m)
			if code != 2 {
				t.Fatalf("missing profile code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
			}
			if jsonMode {
				var response struct {
					OK    bool           `json:"ok"`
					Error *manager.Error `json:"error"`
				}
				if err := json.Unmarshal(out.Bytes(), &response); err != nil || response.OK || response.Error == nil || !strings.Contains(response.Error.Message, "does not exist") || diagnostics.Len() != 0 {
					t.Fatalf("JSON missing profile: stdout=%q stderr=%q error=%v", out.String(), diagnostics.String(), err)
				}
			} else if out.Len() != 0 || !strings.Contains(diagnostics.String(), "does not exist") {
				t.Fatalf("missing profile: stdout=%q stderr=%q", out.String(), diagnostics.String())
			}
			if _, err := os.Stat(filepath.Join(root, "mgs3mod-launch.json")); !os.IsNotExist(err) {
				t.Fatalf("missing profile wrote config: %v", err)
			}
		})
	}
}

func TestNonLaunchJSONParseErrorKeepsResultObject(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := RunWithInput([]string{"doctor", "extra", "--json"}, &out, &diagnostics, nil, strings.NewReader(""), false)
	if code != 2 || diagnostics.Len() != 0 {
		t.Fatalf("parse error: code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	var response struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &response); err != nil || bytes.Equal(response.Result, []byte("null")) || len(response.Result) == 0 {
		t.Fatalf("legacy JSON result object lost: %s (%v)", out.Bytes(), err)
	}
}

func TestLaunchInputErrorsAreUsageAndDoNotPersist(t *testing.T) {
	root, m := launchFixture(t)
	for _, args := range [][]string{{"launch"}, {"--json", "launch"}, {"launch", "--select"}} {
		var out, diagnostics bytes.Buffer
		code := RunWithInput(args, &out, &diagnostics, m, strings.NewReader(""), false)
		if code != 2 {
			t.Fatalf("code for %v: %d stdout=%q stderr=%q", args, code, out.String(), diagnostics.String())
		}
		if contains(args, "--json") {
			var value any
			if err := json.Unmarshal(out.Bytes(), &value); err != nil || diagnostics.Len() != 0 {
				t.Fatalf("JSON usage error: %q (%v), stderr=%q", out.String(), err, diagnostics.String())
			}
		}
	}
	if _, err := os.Stat(filepath.Join(root, "mgs3mod-launch.json")); !os.IsNotExist(err) {
		t.Fatalf("input error wrote config: %v", err)
	}
}

func TestLaunchJSONAndDryRunNeverOpenImplicitSelector(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		json bool
	}{
		{name: "json", args: []string{"launch", "--json"}, json: true},
		{name: "dry-run", args: []string{"launch", "--dry-run"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, m := launchFixture(t)
			var out, diagnostics bytes.Buffer
			code := RunWithInput(tc.args, &out, &diagnostics, m, strings.NewReader("4\nshould-not-save\n"), true)
			if code != 2 || strings.Contains(out.String(), "Choice [1]") {
				t.Fatalf("implicit selector: code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
			}
			if tc.json {
				var response struct {
					OK bool `json:"ok"`
				}
				if err := json.Unmarshal(out.Bytes(), &response); err != nil || response.OK || diagnostics.Len() != 0 {
					t.Fatalf("JSON error: stdout=%q stderr=%q error=%v", out.String(), diagnostics.String(), err)
				}
			} else if out.Len() != 0 || !strings.Contains(diagnostics.String(), "no saved launch profile") {
				t.Fatalf("dry-run error: stdout=%q stderr=%q", out.String(), diagnostics.String())
			}
			if _, err := os.Stat(filepath.Join(root, "mgs3mod-launch.json")); !os.IsNotExist(err) {
				t.Fatalf("implicit selector wrote config: %v", err)
			}
		})
	}
}

func TestLaunchJSONErrorPreservesManagerPaths(t *testing.T) {
	var out, diagnostics bytes.Buffer
	want := &manager.Error{Code: 5, Message: "synthetic", Paths: []string{"mgs3mod-launch.json"}}
	if code := writeCLIError(&out, &diagnostics, true, want); code != want.Code || diagnostics.Len() != 0 {
		t.Fatalf("writeCLIError code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	var response struct {
		Error *manager.Error `json:"error"`
	}
	if err := json.Unmarshal(out.Bytes(), &response); err != nil || response.Error == nil || !equalStrings(response.Error.Paths, want.Paths) {
		t.Fatalf("JSON error = %#v (%v)", response.Error, err)
	}
}

func TestLaunchSelectorCancelAndEOFDoNotWrite(t *testing.T) {
	for _, input := range []string{"6\n", ""} {
		root, m := launchFixture(t)
		var out, diagnostics bytes.Buffer
		if code := RunWithInput([]string{"launch", "--select"}, &out, &diagnostics, m, strings.NewReader(input), true); code != 0 {
			t.Fatalf("selector %q code=%d stdout=%q stderr=%q", input, code, out.String(), diagnostics.String())
		}
		if _, err := os.Stat(filepath.Join(root, "mgs3mod-launch.json")); !os.IsNotExist(err) {
			t.Fatalf("selector %q wrote config: %v", input, err)
		}
	}
}

func TestLaunchSelectorSavesProfileAndDefault(t *testing.T) {
	root, m := launchFixture(t)
	input := "3\nus\nen\nkbd\nstartup\n4\nmy-profile\n5\nmy-profile\n6\n"
	var out, diagnostics bytes.Buffer
	if code := RunWithInput([]string{"launch", "--select"}, &out, &diagnostics, m, strings.NewReader(input), true); code != 0 {
		t.Fatalf("selector code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	if !strings.Contains(out.String(), "Profiles:") || !strings.Contains(out.String(), "Default profile ID") {
		t.Fatalf("missing selector prompts: %q", out.String())
	}
	loaded, err := m.WithRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	config, err := loaded.LoadLaunchConfig()
	if err != nil || !config.Persisted || config.Config.DefaultProfileID != "my-profile" {
		t.Fatalf("saved config: %+v (%v)", config, err)
	}
	if _, err = config.Config.Profile("my-profile"); err != nil {
		t.Fatalf("missing saved profile: %v", err)
	}
}

func TestLaunchSelectorRejectsDuplicateProfileWithoutOverwrite(t *testing.T) {
	root, m := launchFixture(t)
	selected, err := m.WithRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = selected.SaveLaunchConfig(launch.Seed()); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "mgs3mod-launch.json"))
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := RunWithInput([]string{"launch", "--select"}, &out, &diagnostics, m, strings.NewReader("4\nna-startup\n"), true)
	if code != 2 || !strings.Contains(diagnostics.String(), "already exists") {
		t.Fatalf("duplicate profile: code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	after, err := os.ReadFile(filepath.Join(root, "mgs3mod-launch.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("duplicate profile changed config: %v", err)
	}
}

func launchFixture(t *testing.T) (string, *manager.Manager) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "game root with spaces")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "METAL GEAR SOLID3.exe"), []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "launcher.exe"), []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	m := manager.New(manager.Config{Root: "not-selected", Core: []profile.Fingerprint{}, CheckProcesses: func(string) error { return nil }})
	if code := Run([]string{"init", "--game-root", root, "--json"}, io.Discard, io.Discard, m); code != 0 {
		t.Fatalf("initialize fixture: %d", code)
	}
	selected, err := m.WithRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	return root, selected
}

func equalStrings(got, want []string) bool {
	return len(got) == len(want) && func() bool {
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}()
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
