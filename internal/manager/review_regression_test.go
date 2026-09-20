package manager

import (
	"bytes"
	"errors"
	"fmt"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/transaction"
	"os"
	"path/filepath"
	"testing"
)

func TestBudgetIncludesEveryRepeatedPayloadStage(t *testing.T) {
	m, _ := fixture(t)
	invoke(t, m, "init", "")
	s, err := m.open(false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	data := bytes.Repeat([]byte("x"), 4096)
	hash := transaction.Hash(data)
	p := plan{}
	for i := 0; i < 256; i++ {
		p.Changes = append(p.Changes, change{Path: fmt.Sprintf("textures/flatlist/_win/%d.ctxr", i), After: hash})
	}
	budget, err := s.requiredSpace(p, map[string]input{hash: {Data: data}})
	if err != nil {
		t.Fatal(err)
	}
	minimum := uint64(64<<20) + 256*uint64(len(data))
	if budget < minimum {
		t.Fatalf("budget %d does not cover %d distinct staged copies plus headroom (%d)", budget, 256, minimum)
	}
}

func TestConditionalRestoreIgnoresUnaffectedUpdatedTargets(t *testing.T) {
	for _, alreadyOriginal := range []bool{false, true} {
		t.Run(fmt.Sprint(alreadyOriginal), func(t *testing.T) {
			m, root := fixture(t)
			a := makePackage(t, root, "one", targetA)
			b := makePackage(t, root, "two", targetB)
			invoke(t, m, "init", "")
			invoke(t, m, "add", a)
			invoke(t, m, "add", b)
			invoke(t, m, "enable", "two")
			invoke(t, m, "disable", "two")
			invoke(t, m, "enable", "one")
			put(t, root, "game.exe", []byte("updated"))
			put(t, root, targetB, []byte("updated unrelated texture"))
			if alreadyOriginal {
				put(t, root, targetA, []byte("original:"+targetA))
			}
			result := invoke(t, m, "restore", "")
			if result.Compatible || result.State.Mods["one"].Enabled {
				t.Fatal("conditional restore state wrong")
			}
			assertBytes(t, root, targetA, "original:"+targetA)
			assertBytes(t, root, targetB, "updated unrelated texture")
		})
	}
}

func TestConditionalRestoreAlreadyBaselineSkipsPhysicalAccess(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "one")
	put(t, root, "game.exe", []byte("updated"))
	put(t, root, targetA, []byte("original:"+targetA))
	target := filepath.Join(root, filepath.FromSlash(targetA))
	if err := os.Chmod(target, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(target, 0600)
	m.config.CheckProcesses = func(string) error { return errors.New("game is running") }
	result := invoke(t, m, "restore", "")
	if result.State.Mods["one"].Enabled || len(result.Paths) != 0 || !result.Committed {
		t.Fatalf("logical-only restore result: %+v", result)
	}
	assertBytes(t, root, targetA, "original:"+targetA)
}

func TestFailedRecoveryHasNoSuccessMessageOrPrematurePaths(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA, targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "replaced-0")
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(targetB))); err != nil {
		t.Fatal(err)
	}
	put(t, root, targetA, []byte("external drift"))
	result, err := m.Run("recover", "", Options{RestoreMissing: []string{targetB}})
	if ExitCode(err) != 4 || result.Message != "recovery failed" || !result.RecoveryRequired || len(result.Paths) != 0 {
		t.Fatalf("failed recovery result: %+v err=%v", result, err)
	}

	put(t, root, targetA, []byte("original:"+targetA))
	dry, err := m.Run("recover", "", Options{DryRun: true, RestoreMissing: []string{targetB}})
	if err != nil || dry.Message != "recovery preflight complete" || len(dry.Paths) != 0 {
		t.Fatalf("dry-run recovery result: %+v err=%v", dry, err)
	}
	result, err = m.Run("recover", "", Options{RestoreMissing: []string{targetB}})
	if err != nil || len(result.Paths) != 1 || result.Paths[0] != targetB || result.Message != "recovery resolved" {
		t.Fatalf("successful recovery result: %+v err=%v", result, err)
	}
}

func TestInterruptedLibraryAddRecoveryDoesNotCheckProcesses(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	killAt(t, m, "add", pkg, "replaced-0")
	m.config.CheckProcesses = func(string) error { return errors.New("game is running") }
	invoke(t, m, "recover", "")
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(stateDir+"/library/one"))); !os.IsNotExist(err) {
		t.Fatalf("interrupted library import remains: %v", err)
	}
}

func TestPreparationSchemaTwoAndSchemaOneCompatibility(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)

	s, err := m.open(false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	h, err := s.load()
	if err != nil {
		t.Fatal(err)
	}
	p, inputs, err := s.build(h, "enable", "one")
	if err != nil {
		t.Fatal(err)
	}
	if p.Schema != 2 || len(p.Baselines) != 1 {
		t.Fatalf("new preparation schema: %+v", p)
	}
	b := p.Baselines[0]
	if b.Target != targetA || b.Path != baseline(b.SHA256) || b.Temporary == "" || b.VerifiedObjectExisted == nil || *b.VerifiedObjectExisted {
		t.Fatalf("incomplete schema-2 baseline inventory: %+v", b)
	}

	p.Schema = 1
	for i := range p.Baselines {
		p.Baselines[i].Target = ""
		p.Baselines[i].Temporary = ""
		p.Baselines[i].VerifiedObjectExisted = nil
	}
	legacyInventory := canonical(p.Baselines[0])
	for _, field := range [][]byte{[]byte(`"target"`), []byte(`"temporary"`), []byte(`"verifiedObjectExisted"`)} {
		if bytes.Contains(legacyInventory, field) {
			t.Fatalf("schema-1 inventory contains schema-2 field %s: %s", field, legacyInventory)
		}
	}
	wantBytes := append([]byte(nil), canonical(p)...)
	wantHash := planHash(*p)
	if err := validatePlan(*p, h.state, p.Sequence); err != nil {
		t.Fatalf("schema-1 plan rejected: %v", err)
	}
	if err := s.execute(*p, inputs); err != nil {
		t.Fatal(err)
	}
	var decoded plan
	if err := read(s, txPath(p.Sequence)+"/PREPARED", &decoded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical(decoded), wantBytes) || planHash(decoded) != wantHash {
		t.Fatalf("schema-1 canonical plan changed: hash=%s want=%s", planHash(decoded), wantHash)
	}
	if _, err := s.load(); err != nil {
		t.Fatalf("schema-1 history rejected: %v", err)
	}
}

func TestSchemaTwoReusesSamePlanBaselineObject(t *testing.T) {
	m, root := fixture(t)
	pkgDir := makePackage(t, root, "one", targetA, targetB)
	pkg, err := packagefmt.Load(pkgDir, false)
	if err != nil {
		t.Fatal(err)
	}
	original := []byte("shared original")
	originalHash := transaction.Hash(original)
	for i := range pkg.Manifest.Files {
		pkg.Manifest.Files[i].OriginalSHA256 = originalHash
		put(t, root, pkg.Manifest.Files[i].Target, original)
	}
	put(t, pkgDir, "manifest.json", canonical(pkg.Manifest))
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkgDir)
	invoke(t, m, "enable", "one")
	invoke(t, m, "disable", "one")
	for _, target := range []string{targetA, targetB} {
		assertBytes(t, root, target, string(original))
	}
}

func TestCommittedRecoveryIgnoresUnaffectedUpdatedTargets(t *testing.T) {
	m, root := fixture(t)
	a := makePackage(t, root, "one", targetA)
	b := makePackage(t, root, "two", targetB)
	invoke(t, m, "init", "")
	invoke(t, m, "add", a)
	invoke(t, m, "add", b)
	invoke(t, m, "enable", "two")
	invoke(t, m, "disable", "two")
	invoke(t, m, "enable", "one")
	killAt(t, m, "remove", "one", "committed")
	put(t, root, "game.exe", []byte("updated"))
	put(t, root, targetB, []byte("updated unrelated texture"))
	invoke(t, m, "recover", "")
	assertBytes(t, root, targetA, "original:"+targetA)
	assertBytes(t, root, targetB, "updated unrelated texture")
}

func TestProfileRecheckedUnderLock(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	m.config.Fault = func(point string) error {
		if point == "lock-acquired" {
			return os.WriteFile(filepath.Join(root, "game.exe"), []byte("updated"), 0600)
		}
		return nil
	}
	if _, err := m.Run("enable", "one", Options{}); ExitCode(err) != 3 {
		t.Fatalf("stale pre-lock profile accepted: %v", err)
	}
	assertBytes(t, root, targetA, "original:"+targetA)
}

func TestMissingTargetErrorIncludesPath(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "one")
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(targetA))); err != nil {
		t.Fatal(err)
	}
	_, err := m.Run("verify", "", Options{})
	var detail *Error
	if !errors.As(err, &detail) || len(detail.Paths) != 1 || detail.Paths[0] != targetA {
		t.Fatalf("missing structured path: %v", err)
	}
}

func TestIncompleteInitRejectsUnexpectedCache(t *testing.T) {
	m, root := fixture(t)
	killAt(t, m, "init", "", "init-prepared")
	put(t, root, stateDir+"/state.json", []byte("unknown cache"))
	before := snapshot(t, root)
	if _, err := m.Run("recover", "", Options{}); ExitCode(err) != 5 {
		t.Fatalf("unexpected cache adopted: %v", err)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("invalid init recovery changed files")
	}
}

func TestReadOnlyTargetFailsBeforeJournal(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	target := filepath.Join(root, filepath.FromSlash(targetA))
	if err := os.Chmod(target, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(target, 0600)
	before := snapshot(t, root)
	for _, dry := range []bool{true, false} {
		result, err := m.Run("enable", "one", Options{DryRun: dry})
		if ExitCode(err) != 1 || result.RecoveryRequired {
			t.Fatalf("permission failure category: %+v %v", result, err)
		}
		if !equal(before, snapshot(t, root)) {
			t.Fatal("preflight permission failure wrote files")
		}
	}
}

func TestAbandonedPreparationCleanupResumes(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	killAt(t, m, "enable", "one", "baseline-staged")
	killAt(t, m, "recover", "", "abort-cleanup-0")
	invoke(t, m, "recover", "")
	for _, folder := range []string{"blobs", "stage"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(txPath(2)), folder)); !os.IsNotExist(err) {
			t.Fatalf("abandoned %s remains: %v", folder, err)
		}
	}
	assertBytes(t, root, targetA, "original:"+targetA)
	invoke(t, m, "enable", "one")
	invoke(t, m, "verify", "")
}

func TestStatusShowsOwnershipAndPendingRecovery(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "one")
	result := invoke(t, m, "status", "")
	if len(result.Targets) != 1 || result.Targets[0].Owner != "one" || result.Targets[0].Path != targetA {
		t.Fatalf("ownership omitted: %+v", result)
	}
	killAt(t, m, "disable", "one", "ready")
	result, err := m.Run("status", "", Options{})
	if ExitCode(err) != 5 || !result.RecoveryRequired || len(result.Targets) != 1 {
		t.Fatalf("pending state omitted: %+v %v", result, err)
	}
}

func TestCommittedFailureReportsCommittedState(t *testing.T) {
	m, root := fixture(t)
	pkg := makePackage(t, root, "one", targetA)
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "one")
	m.config.Fault = func(point string) error {
		if point == "committed" {
			return errors.New("injected failure after durable commit")
		}
		return nil
	}
	result, err := m.Run("remove", "one", Options{})
	if ExitCode(err) != 5 || !result.Committed || !result.RecoveryRequired || result.State == nil {
		t.Fatalf("incorrect committed failure result: %+v %v", result, err)
	}
	if _, ok := result.State.Mods["one"]; ok {
		t.Fatal("failure reported removed mod as still stored")
	}
	assertBytes(t, root, targetA, "original:"+targetA)
	m.config.Fault = nil
	invoke(t, m, "recover", "")
	invoke(t, m, "verify", "")
}
