package manager

import (
	"encoding/binary"
	"errors"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"mgs3mod/internal/transaction"
	"os"
	"path/filepath"
	"testing"
)

const pluginTarget = "qcamo.asi"

func makePluginPackage(t *testing.T, root, id string) string {
	t.Helper()
	dir := filepath.Join(root, "authoring", id)
	payload := testPluginPE()
	source := "payload/qcamo.asi"
	put(t, dir, source, payload)
	manifest := packagefmt.Manifest{
		SchemaVersion: 2,
		ID:            id,
		Version:       "1.0.0",
		Name:          id,
		Profile:       profile.ID,
		Files: []packagefmt.File{{
			Source:         source,
			Target:         pluginTarget,
			OriginalAbsent: true,
			PayloadSHA256:  transaction.Hash(payload),
			PayloadBytes:   int64(len(payload)),
		}},
	}
	put(t, dir, "manifest.json", canonical(manifest))
	return dir
}

func testPluginPE() []byte {
	const peOffset = 0x80
	const optionalSize = 240
	b := make([]byte, 0x400)
	b[0], b[1] = 'M', 'Z'
	binary.LittleEndian.PutUint32(b[0x3c:], peOffset)
	copy(b[peOffset:], []byte{'P', 'E', 0, 0})
	coff := peOffset + 4
	binary.LittleEndian.PutUint16(b[coff:], 0x8664)
	binary.LittleEndian.PutUint16(b[coff+2:], 1)
	binary.LittleEndian.PutUint16(b[coff+16:], optionalSize)
	binary.LittleEndian.PutUint16(b[coff+18:], 0x2002)
	optional := coff + 20
	binary.LittleEndian.PutUint16(b[optional:], 0x20b)
	binary.LittleEndian.PutUint32(b[optional+4:], 0x200)
	binary.LittleEndian.PutUint32(b[optional+16:], 0x1000)
	binary.LittleEndian.PutUint32(b[optional+20:], 0x1000)
	binary.LittleEndian.PutUint32(b[optional+32:], 0x1000)
	binary.LittleEndian.PutUint32(b[optional+36:], 0x200)
	binary.LittleEndian.PutUint32(b[optional+56:], 0x2000)
	binary.LittleEndian.PutUint32(b[optional+60:], 0x200)
	binary.LittleEndian.PutUint16(b[optional+68:], 2)
	binary.LittleEndian.PutUint32(b[optional+108:], 16)
	section := optional + optionalSize
	copy(b[section:], []byte(".text\x00\x00\x00"))
	binary.LittleEndian.PutUint32(b[section+8:], 0x200)
	binary.LittleEndian.PutUint32(b[section+12:], 0x1000)
	binary.LittleEndian.PutUint32(b[section+16:], 0x200)
	binary.LittleEndian.PutUint32(b[section+20:], 0x200)
	binary.LittleEndian.PutUint32(b[section+36:], 0x60000020)
	b[0x200] = 0xc3
	return b
}

func assertPlugin(t *testing.T, root string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, pluginTarget))
	if err != nil || transaction.Hash(b) != transaction.Hash(testPluginPE()) {
		t.Fatalf("plugin mismatch: %v", err)
	}
}

func allowASILoader(m *Manager) {
	m.config.CheckASILoader = func(string) error { return nil }
}

func assertAbsent(t *testing.T, root, path string) {
	t.Helper()
	_, err := os.Lstat(filepath.Join(root, filepath.FromSlash(path)))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s exists or cannot be inspected: %v", path, err)
	}
}

func TestPluginAbsentBaselineLifecycle(t *testing.T) {
	m, root := fixture(t)
	allowASILoader(m)
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	assertAbsent(t, root, pluginTarget)

	invoke(t, m, "enable", "qcamo")
	assertPlugin(t, root)
	state := invoke(t, m, "list", "").State
	baselineState := state.Baselines[profile.Key(pluginTarget)]
	if !baselineState.Absent || baselineState.SHA256 != "" || baselineState.Bytes != 0 {
		t.Fatalf("wrong absent baseline: %+v", baselineState)
	}

	invoke(t, m, "disable", "qcamo")
	assertAbsent(t, root, pluginTarget)
	invoke(t, m, "enable", "qcamo")
	invoke(t, m, "remove", "qcamo")
	assertAbsent(t, root, pluginTarget)
	if len(invoke(t, m, "list", "").State.Mods) != 0 {
		t.Fatal("remove retained plugin package")
	}
}

func TestPluginRejectsPreexistingUnownedTarget(t *testing.T) {
	m, root := fixture(t)
	allowASILoader(m)
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	put(t, root, pluginTarget, []byte("foreign"))
	before := snapshot(t, root)
	if _, err := m.Run("add", pkg, Options{}); ExitCode(err) != 4 {
		t.Fatalf("preexisting plugin accepted: %v", err)
	}
	if !equal(before, snapshot(t, root)) {
		t.Fatal("rejected add changed installation")
	}
	assertBytes(t, root, pluginTarget, "foreign")
}

func TestPluginRequiresLoaderOnlyForEnableAndEnabledVerification(t *testing.T) {
	m, root := fixture(t)
	m.config.CheckASILoader = func(string) error { return errors.New("missing loader") }
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	if _, err := m.Run("enable", "qcamo", Options{}); ExitCode(err) != 3 {
		t.Fatalf("enable without loader got %v", err)
	}
	assertAbsent(t, root, pluginTarget)
	allowASILoader(m)
	invoke(t, m, "enable", "qcamo")
	m.config.CheckASILoader = func(string) error { return errors.New("missing loader") }
	if _, err := m.Run("verify", "", Options{}); ExitCode(err) != 3 {
		t.Fatalf("verify without loader got %v", err)
	}
	result, err := m.Run("status", "", Options{})
	if err != nil || len(result.Issues) == 0 {
		t.Fatalf("status did not report missing loader: %+v %v", result, err)
	}
	invoke(t, m, "disable", "qcamo")
	assertAbsent(t, root, pluginTarget)
}

func TestPluginConditionalRestoreDeletesManagedFile(t *testing.T) {
	m, root := fixture(t)
	allowASILoader(m)
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "qcamo")
	put(t, root, "game.exe", []byte("updated"))
	invoke(t, m, "restore", "")
	assertAbsent(t, root, pluginTarget)
}

func TestPluginCrashRecoveryCreateAndDelete(t *testing.T) {
	points := []string{"prepared", "ready", "intent-0", "replaced-0", "applied-0", "committed", "cached", "cleaned"}
	for _, operation := range []string{"enable", "disable"} {
		for _, point := range points {
			t.Run(operation+"/"+point, func(t *testing.T) {
				m, root := fixture(t)
				allowASILoader(m)
				pkg := makePluginPackage(t, root, "qcamo")
				invoke(t, m, "init", "")
				invoke(t, m, "add", pkg)
				if operation == "disable" {
					invoke(t, m, "enable", "qcamo")
				}
				killAt(t, m, operation, "qcamo", point)
				invoke(t, m, "recover", "")
				committed := point == "committed" || point == "cached" || point == "cleaned"
				wantPresent := operation == "enable" && committed || operation == "disable" && !committed
				if wantPresent {
					assertPlugin(t, root)
				} else {
					assertAbsent(t, root, pluginTarget)
				}
				invoke(t, m, "verify", "")
			})
		}
	}
}

func TestPluginCreateRacePreservesForeignFile(t *testing.T) {
	m, root := fixture(t)
	allowASILoader(m)
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	checks := 0
	m.config.CheckProcesses = func(string) error {
		checks++
		if checks == 2 {
			put(t, root, pluginTarget, []byte("foreign"))
		}
		return nil
	}
	if _, err := m.Run("enable", "qcamo", Options{}); err == nil {
		t.Fatal("enable overwrote raced file")
	}
	assertBytes(t, root, pluginTarget, "foreign")
	if _, err := m.Run("recover", "", Options{}); ExitCode(err) != 4 {
		t.Fatalf("recovery did not preserve mismatched foreign file: %v", err)
	}
	assertBytes(t, root, pluginTarget, "foreign")
}

func TestPluginExactPayloadRaceRecoveryPreservesUnownedFile(t *testing.T) {
	m, root := fixture(t)
	allowASILoader(m)
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	checks := 0
	m.config.CheckProcesses = func(string) error {
		checks++
		if checks == 2 {
			put(t, root, pluginTarget, testPluginPE())
		}
		return nil
	}
	if _, err := m.Run("enable", "qcamo", Options{}); ExitCode(err) != 5 {
		t.Fatalf("exact-byte race did not leave recoverable operation: %v", err)
	}
	assertPlugin(t, root)
	if _, err := m.Run("recover", "", Options{}); ExitCode(err) != 4 {
		t.Fatalf("recovery claimed unowned exact-byte file: %v", err)
	}
	assertPlugin(t, root)
}

func TestPluginDeleteProcessCheckDriftPreservesForeignFile(t *testing.T) {
	m, root := fixture(t)
	allowASILoader(m)
	pkg := makePluginPackage(t, root, "qcamo")
	invoke(t, m, "init", "")
	invoke(t, m, "add", pkg)
	invoke(t, m, "enable", "qcamo")
	checks := 0
	m.config.CheckProcesses = func(string) error {
		checks++
		if checks == 2 {
			put(t, root, pluginTarget, []byte("foreign"))
		}
		return nil
	}
	if _, err := m.Run("disable", "qcamo", Options{}); ExitCode(err) != 5 {
		t.Fatalf("disable did not reject process-check drift: %v", err)
	}
	assertBytes(t, root, pluginTarget, "foreign")
}
