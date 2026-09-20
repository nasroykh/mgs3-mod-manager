package manager

import (
	"errors"
	"fmt"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	dir, _ := filepath.Abs("../../.cache/manager-tests")
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(err)
	}
	os.Setenv("TMP", dir)
	os.Setenv("TEMP", dir)
	os.Exit(m.Run())
}
func put(t *testing.T, root, p string, b []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, b, 0600); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) (*Manager, string) {
	t.Helper()
	root := t.TempDir()
	put(t, root, "game.exe", []byte("core"))
	m := New(Config{Root: root, Core: []profile.Fingerprint{{Path: "game.exe", SHA256: transaction.Hash([]byte("core"))}}, CheckProcesses: func(string) error { return nil }})
	return m, root
}
func invoke(t *testing.T, m *Manager, command, arg string) Result {
	t.Helper()
	r, e := m.Run(command, arg, Options{})
	if e != nil {
		t.Fatalf("%s %s: %v", command, arg, e)
	}
	return r
}
func makePackage(t *testing.T, root, id string, targets ...string) string {
	t.Helper()
	dir := filepath.Join(root, "authoring", id)
	manifest := packagefmt.Manifest{SchemaVersion: 1, ID: id, Version: "0.1.0", Name: id, Profile: profile.ID, Files: []packagefmt.File{}}
	for i, target := range targets {
		original := []byte("original:" + target)
		payload := []byte("mod:" + id + ":" + target)
		put(t, root, target, original)
		source := fmt.Sprintf("payload/%d.ctxr", i)
		put(t, dir, source, payload)
		manifest.Files = append(manifest.Files, packagefmt.File{Source: source, Target: target, OriginalSHA256: transaction.Hash(original), PayloadSHA256: transaction.Hash(payload), PayloadBytes: int64(len(payload))})
	}
	put(t, dir, "manifest.json", canonical(manifest))
	return dir
}

const targetA = "textures/flatlist/_win/a.ctxr"
const targetB = "hqtex/flatlist/_win/b.ctxr"

func assertBytes(t *testing.T, root, p, want string) {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
	if e != nil || string(b) != want {
		t.Fatalf("%s got %q (%v), want %q", p, b, e, want)
	}
}
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			out[rel] = "dir"
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		out[rel] = transaction.Hash(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestLifecycleAndDryRun(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA, targetB)
	before := snapshot(t, root)
	if _, e := m.Run("init", "", Options{DryRun: true}); e != nil {
		t.Fatal(e)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("init dry run wrote files")
	}
	invoke(t, m, "init", "")
	invoke(t, m, "init", "")
	for _, op := range []struct{ c, a string }{{"add", pkg}, {"enable", "one"}, {"disable", "one"}, {"enable", "one"}, {"restore", ""}, {"enable", "one"}, {"remove", "one"}} {
		before = snapshot(t, root)
		if _, e := m.Run(op.c, op.a, Options{DryRun: true}); e != nil {
			t.Fatalf("dry %s: %v", op.c, e)
		}
		if !equal(before, snapshot(t, root)) {
			t.Fatalf("%s dry-run wrote", op.c)
		}
		invoke(t, m, op.c, op.a)
		invoke(t, m, "verify", "")
		if op.c == "enable" {
			assertBytes(t, root, targetA, "mod:one:"+targetA)
		} else {
			assertBytes(t, root, targetA, "original:"+targetA)
		}
	}
	r := invoke(t, m, "list", "")
	if len(r.State.Mods) != 0 || len(r.State.Baselines) != 2 {
		t.Fatal("remove lost originals or retained mod")
	}
	for _, b := range r.State.Baselines {
		assertBytes(t, root, baseline(b.SHA256), "original:"+b.Target)
	}
	invoke(t, m, "add", pkg)
	invoke(t, m, "add", pkg)
	invoke(t, m, "disable", "one")
}

func TestConflictsDriftAndCoreMismatch(t *testing.T) {
	m, root := fixture(t)
	a := makePackage(t, root, "one", targetA)
	b := makePackage(t, root, "two", targetA)
	c := makePackage(t, root, "three", targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", a)
	invoke(t, m, "add", b)
	invoke(t, m, "add", c)
	invoke(t, m, "enable", "one")
	before := snapshot(t, root)
	if _, e := m.Run("enable", "two", Options{}); ExitCode(e) != 4 {
		t.Fatalf("want conflict 4 got %v", e)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("conflict wrote")
	}
	invoke(t, m, "enable", "three")
	put(t, root, "game.exe", []byte("updated"))
	if _, e := m.Run("disable", "one", Options{}); ExitCode(e) != 3 {
		t.Fatalf("core mismatch %v", e)
	}
	invoke(t, m, "restore", "")
	assertBytes(t, root, targetA, "original:"+targetA)
	put(t, root, targetA, []byte("external"))
	before = snapshot(t, root)
	if _, e := m.Run("verify", "", Options{}); ExitCode(e) != 4 {
		t.Fatalf("drift %v", e)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("drift wrote")
	}
}

func TestCrashChild(t *testing.T) {
	if os.Getenv("MGS3_TEST_CHILD") != "1" {
		return
	}
	root := os.Getenv("MGS3_TEST_ROOT")
	m := New(Config{Root: root, Core: []profile.Fingerprint{{Path: "game.exe", SHA256: transaction.Hash([]byte("core"))}}, CheckProcesses: func(string) error { return nil }, Fault: func(point string) error {
		if point == os.Getenv("MGS3_TEST_POINT") {
			os.Exit(91)
		}
		return nil
	}})
	_, err := m.Run(os.Getenv("MGS3_TEST_COMMAND"), os.Getenv("MGS3_TEST_ARG"), Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(92)
	}
	os.Exit(93)
}
func killAt(t *testing.T, m *Manager, command, arg, point string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCrashChild$")
	cmd.Env = append(os.Environ(), "MGS3_TEST_CHILD=1", "MGS3_TEST_ROOT="+m.config.Root, "MGS3_TEST_COMMAND="+command, "MGS3_TEST_ARG="+arg, "MGS3_TEST_POINT="+point)
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 91 {
		t.Fatalf("fault %s not reached: %v %s", point, err, out)
	}
}

func TestProcessCrashMatrix(t *testing.T) {
	for _, op := range []string{"enable", "disable", "remove", "restore"} {
		for _, point := range []string{"prepared", "ready", "intent-0", "replaced-0", "applied-0", "intent-1", "replaced-1", "applied-1", "committed", "cached", "cleaned"} {
			t.Run(op+"/"+point, func(t *testing.T) {
				m, root := fixture(t)
				pkg := makePackage(t, root, "one", targetA, targetB)
				invoke(t, m, "init", "")
				invoke(t, m, "add", pkg)
				before := "original:" + targetA
				after := "mod:one:" + targetA
				if op != "enable" {
					invoke(t, m, "enable", "one")
					before, after = after, before
				}
				arg := "one"
				if op == "restore" {
					arg = ""
				}
				killAt(t, m, op, arg, point)
				invoke(t, m, "recover", "")
				committed := point == "committed" || point == "cached" || point == "cleaned"
				want := before
				if committed {
					want = after
				}
				assertBytes(t, root, targetA, want)
				invoke(t, m, "verify", "")
				invoke(t, m, "recover", "")
			})
		}
	}
	for _, point := range []string{"baseline-staged", "baseline-promoted"} {
		t.Run(point, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			killAt(t, m, "enable", "one", point)
			invoke(t, m, "recover", "")
			assertBytes(t, root, targetA, "original:"+targetA)
			invoke(t, m, "enable", "one")
		})
	}
}

func TestInitializationCrashMatrix(t *testing.T) {
	for _, point := range []string{"init-prepared", "init-directory-baseline", "init-directory-library", "init-directory-transactions", "init-installation", "init-committed"} {
		t.Run(point, func(t *testing.T) {
			m, _ := fixture(t)
			killAt(t, m, "init", "", point)
			invoke(t, m, "recover", "")
			invoke(t, m, "verify", "")
		})
	}
}

func TestMissingAndCorruptRecovery(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA, targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "replaced-0")
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(targetB))); err != nil {
		t.Fatal(err)
	} // sorted targetB was first, with an apply intent
	before := snapshot(t, root)
	if _, e := m.Run("recover", "", Options{}); ExitCode(e) != 5 {
		t.Fatalf("missing should block: %v", e)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("blocked recovery wrote")
	}
	r, e := m.Run("recover", "", Options{RestoreMissing: []string{targetB}})
	if e != nil {
		t.Fatal(e)
	}
	if !r.MissingFileRecoveryExplicit || len(r.Paths) != 1 {
		t.Fatal("missing recovery not reported")
	}
	assertBytes(t, root, targetB, "original:"+targetB)
	invoke(t, m, "verify", "")
	// Cache damage is reconstructible; journal damage is not.
	put(t, root, stateDir+"/state.json", []byte("torn"))
	invoke(t, m, "recover", "")
	put(t, root, txPath(2)+"/READY", []byte("torn"))
	before = snapshot(t, root)
	if _, e = m.Run("recover", "", Options{}); ExitCode(e) != 5 {
		t.Fatalf("torn READY accepted: %v", e)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("corrupt recovery wrote")
	}
}

func TestRecoveryInterruptedAgain(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA, targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "applied-1")
	killAt(t, m, "recover", "", "recovered-0")
	invoke(t, m, "recover", "")
	assertBytes(t, root, targetA, "original:"+targetA)
	assertBytes(t, root, targetB, "original:"+targetB)
}

func TestImportAndCleanupCrashes(t *testing.T) {
	for _, point := range []string{"prepared", "ready", "replaced-0", "applied-1", "committed", "cached"} {
		t.Run("add/"+point, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA)
			invoke(t, m, "init", "")
			killAt(t, m, "add", pkg, point)
			invoke(t, m, "recover", "")
			invoke(t, m, "add", pkg)
			invoke(t, m, "verify", "")
		})
	}
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("cleanup-%d", i), func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA, targetB)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			invoke(t, m, "enable", "one")
			killAt(t, m, "remove", "one", fmt.Sprintf("cleanup-%d", i))
			invoke(t, m, "recover", "")
			invoke(t, m, "add", pkg)
			invoke(t, m, "verify", "")
		})
	}
}

func TestTamperingAndMultiFilePreflight(t *testing.T) {
	for _, kind := range []string{"target", "payload", "baseline"} {
		t.Run(kind, func(t *testing.T) {
			m, root := fixture(t)
			pkg := makePackage(t, root, "one", targetA, targetB)
			invoke(t, m, "init", "")
			invoke(t, m, "add", pkg)
			invoke(t, m, "enable", "one")
			invoke(t, m, "disable", "one")
			switch kind {
			case "target":
				put(t, root, targetB, []byte("drift"))
			case "payload":
				put(t, root, stateDir+"/library/one/payload/1.ctxr", []byte("drift"))
			case "baseline":
				put(t, root, baseline(transaction.Hash([]byte("original:"+targetA))), []byte("drift"))
			}
			before := snapshot(t, root)
			if _, e := m.Run("enable", "one", Options{}); e == nil {
				t.Fatal("tamper accepted")
			}
			if !equal(before, snapshot(t, root)) {
				t.Fatal("preflight wrote")
			}
			assertBytes(t, root, targetA, "original:"+targetA)
		})
	}
}

func TestCanonicalRecordsRejectAmbiguity(t *testing.T) {
	for _, data := range []string{`{"generation":0,"generation":0,"mods":{},"baselines":{}}`, `{"generation":0,"mods":{},"baselines":{},"extra":1}`} {
		var st State
		if transaction.Decode([]byte(data), &st) == nil {
			t.Fatal("ambiguous record accepted")
		}
	}
	b := canonical(emptyState())
	var st State
	if e := transaction.Decode(b, &st); e != nil {
		t.Fatal(e)
	}
}
