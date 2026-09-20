package packagefmt

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testOriginal = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func testManifest(payloadHash string, payloadBytes int64) []byte {
	m := Manifest{
		SchemaVersion: 1,
		ID:            "camo-test",
		Version:       "0.1.0",
		Name:          "Camo test",
		Profile:       "mgs3-mcv-local-0d585dcc6a67",
		Files: []File{{
			Source:         "payload/camo.ctxr",
			Target:         "textures/flatlist/_win/camo.ctxr",
			OriginalSHA256: testOriginal,
			PayloadSHA256:  payloadHash,
			PayloadBytes:   payloadBytes,
		}},
	}
	b, _ := json.Marshal(m)
	return b
}

func TestParseRejectsNestedDuplicateAndUnknownKeys(t *testing.T) {
	for _, input := range []string{
		`{"schemaVersion":1,"id":"camo-test","version":"0.1.0","name":"x","profile":"mgs3-mcv-local-0d585dcc6a67","files":[{"source":"payload/a","source":"payload/b","target":"textures/flatlist/_win/a.ctxr","originalSha256":"` + testOriginal + `"}]}`,
		`{"schemaVersion":1,"id":"camo-test","version":"0.1.0","name":"x","profile":"mgs3-mcv-local-0d585dcc6a67","files":[],"extra":true}`,
		`{"schemaVersion":1,"ID":"camo-test","version":"0.1.0","name":"x","profile":"mgs3-mcv-local-0d585dcc6a67","files":[]}`,
	} {
		if _, err := Parse([]byte(input), true); err == nil {
			t.Fatalf("Parse accepted invalid manifest: %s", input)
		}
	}
}

func schema2Manifest(original string, absent bool, target string) []byte {
	m := Manifest{SchemaVersion: 2, ID: "plugin-test", Version: "1.0.0", Name: "Plugin test", Profile: "mgs3-mcv-local-0d585dcc6a67", Files: []File{{Source: "payload/plugin.asi", Target: target, OriginalSHA256: original, OriginalAbsent: absent}}}
	b, _ := json.Marshal(m)
	return b
}

func TestSchema1OriginalAbsentCompatibility(t *testing.T) {
	data := testManifest("", 0)
	if strings.Contains(string(data), "originalAbsent") {
		t.Fatal("schema1 marshal emitted originalAbsent")
	}
	withFalse := strings.Replace(string(data), `"originalSha256"`, `"originalAbsent":false,"originalSha256"`, 1)
	if _, err := Parse([]byte(withFalse), true); err == nil {
		t.Fatal("schema1 accepted explicit originalAbsent=false")
	}
}

func TestSchema2OriginalCombinations(t *testing.T) {
	texture := "textures/flatlist/_win/a.ctxr"
	for name, data := range map[string][]byte{
		"plugin hash":         schema2Manifest(testOriginal, true, "dinput8.asi"),
		"plugin absent false": schema2Manifest("", false, "dinput8.asi"),
		"texture absent":      schema2Manifest("", true, texture),
	} {
		if _, err := Parse(data, true); err == nil {
			t.Errorf("Parse accepted invalid %s", name)
		}
	}
	if _, err := Parse(schema2Manifest(testOriginal, false, texture), true); err != nil {
		t.Fatalf("Parse rejected valid schema2 texture: %v", err)
	}
}

func TestSchema2FolderZIPRoundTrip(t *testing.T) {
	fixture := packageFixtureDir(t)
	folder := filepath.Join(fixture, "authoring")
	if err := os.MkdirAll(filepath.Join(folder, "payload"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := syntheticPE(0x8664, true, 0x20b)
	if err := os.WriteFile(filepath.Join(folder, "payload", "plugin.asi"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "manifest.json"), schema2Manifest("", true, "dinput8.asi"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(fixture, "plugin.zip")
	p, err := Pack(folder, out)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(out, false)
	if err != nil {
		t.Fatal(err)
	}
	if Digest(p) != Digest(loaded) {
		t.Fatalf("folder/ZIP digest mismatch: %s != %s", Digest(p), Digest(loaded))
	}
	if _, err := Load(folder, true); err != nil {
		t.Fatalf("folder reload failed: %v", err)
	}
}

func TestPluginPEValidation(t *testing.T) {
	file := &File{Target: "dinput8.asi"}
	nonExecutable := syntheticPE(0x8664, true, 0x20b)
	binary.LittleEndian.PutUint32(nonExecutable[0x188+36:], 0x40000040)
	virtualTail := syntheticPE(0x8664, true, 0x20b)
	binary.LittleEndian.PutUint32(virtualTail[0x188+8:], 0x1000)
	binary.LittleEndian.PutUint32(virtualTail[0x98+16:], 0x1300)
	for name, payload := range map[string][]byte{
		"non-executable entry":  nonExecutable,
		"entry in virtual tail": virtualTail,
		"empty":                 nil,
		"malformed":             []byte("not a PE"),
		"wrong machine":         syntheticPE(0x14c, true, 0x20b),
		"not DLL":               syntheticPE(0x8664, false, 0x20b),
		"PE32":                  syntheticPE(0x8664, true, 0x10b),
	} {
		if err := fillOrCheckPayload(file, payload, true); err == nil {
			t.Errorf("accepted invalid plugin payload %s", name)
		}
	}
	if err := fillOrCheckPayload(file, syntheticPEWithBSS(), true); err != nil {
		t.Fatalf("rejected valid synthetic PE32+ DLL: %v", err)
	}
}

func syntheticPEWithBSS() []byte {
	b := syntheticPE(0x8664, true, 0x20b)
	coff := 0x84
	binary.LittleEndian.PutUint16(b[coff+2:], 2)
	second := 0x188 + 40
	copy(b[second:], []byte(".bss\x00\x00\x00\x00"))
	binary.LittleEndian.PutUint32(b[second+8:], 0x1000)
	binary.LittleEndian.PutUint32(b[second+12:], 0x2000)
	return b
}

func syntheticPE(machine uint16, dll bool, magic uint16) []byte {
	const peOffset = 0x80
	const optionalSize = 240
	b := make([]byte, 0x400)
	b[0], b[1] = 'M', 'Z'
	binary.LittleEndian.PutUint32(b[0x3c:], peOffset)
	copy(b[peOffset:], []byte{'P', 'E', 0, 0})
	coff := peOffset + 4
	binary.LittleEndian.PutUint16(b[coff:], machine)
	binary.LittleEndian.PutUint16(b[coff+2:], 1)
	binary.LittleEndian.PutUint16(b[coff+16:], optionalSize)
	characteristics := uint16(0x0002)
	if dll {
		characteristics |= 0x2000
	}
	binary.LittleEndian.PutUint16(b[coff+18:], characteristics)
	optional := coff + 20
	binary.LittleEndian.PutUint16(b[optional:], magic)
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
	b[0x200] = 0xC3
	return b
}

func TestFolderPackLoadRoundTripAndNoOverwrite(t *testing.T) {
	fixture := packageFixtureDir(t)
	folder := filepath.Join(fixture, "authoring")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(folder, "payload"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("synthetic ctxr payload")
	if err := os.WriteFile(filepath.Join(folder, "payload", "camo.ctxr"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "manifest.json"), testManifest("", 0), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(fixture, "camo.zip")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relFolder, err := filepath.Rel(cwd, folder)
	if err != nil {
		t.Fatal(err)
	}
	relOut, err := filepath.Rel(cwd, out)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Pack(relFolder, relOut)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Manifest.Files[0].PayloadBytes; got != int64(len(payload)) {
		t.Fatalf("payload size = %d, want %d", got, len(payload))
	}
	if _, err := Pack(relFolder, relOut); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second Pack error = %v, want existing-output error", err)
	}
	relOut2 := filepath.Join(filepath.Dir(relOut), "camo-2.zip")
	if _, err := Pack(relFolder, relOut2); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(filepath.Dir(out), "camo-2.zip"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated Pack calls produced different ZIP bytes")
	}
	archive, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	wantTime := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	if len(archive.File) != 2 {
		t.Fatalf("ZIP entry count = %d, want 2", len(archive.File))
	}
	for _, entry := range archive.File {
		if !entry.Modified.Equal(wantTime) {
			t.Fatalf("ZIP entry %q timestamp = %s, want %s", entry.Name, entry.Modified, wantTime)
		}
	}
	loaded, err := Load(relOut, false)
	if err != nil {
		t.Fatal(err)
	}
	if Digest(p) != Digest(loaded) {
		t.Fatalf("Pack and Load digests differ: %s != %s", Digest(p), Digest(loaded))
	}
}

func packageFixtureDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(filepath.Join("..", "..", ".cache"), "packagefmt-test-")
	if err != nil {
		t.Fatal(err)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestDigestIndependentOfPayloadMapOrder(t *testing.T) {
	a := &Package{Manifest: Manifest{SchemaVersion: 1, ID: "camo-test", Version: "0.1.0"}, Payloads: map[string][]byte{"payload/b": []byte("b"), "payload/a": []byte("a")}}
	b := &Package{Manifest: a.Manifest, Payloads: map[string][]byte{"payload/a": []byte("a"), "payload/b": []byte("b")}}
	if Digest(a) != Digest(b) {
		t.Fatal("Digest depends on map insertion order")
	}
}

func TestZIPRejectsTraversalAndCaseCollision(t *testing.T) {
	fixture := packageFixtureDir(t)
	manifest := testManifest("", 0)
	for name, entries := range map[string][]string{
		"traversal.zip":      {"manifest.json", "../outside"},
		"case-collision.zip": {"manifest.json", "payload/camo.ctxr", "payload/CAMO.ctxr"},
	} {
		path := filepath.Join(fixture, name)
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		writer := zip.NewWriter(file)
		for _, entry := range entries {
			data := []byte("x")
			if entry == "manifest.json" {
				data = manifest
			}
			part, err := writer.Create(entry)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := part.Write(data); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path, true); err == nil {
			t.Fatalf("Load accepted malicious archive %s", name)
		}
	}
	if _, err := Parse(bytes.Repeat([]byte("x"), maxManifestBytes+1), true); err == nil {
		t.Fatal("Parse accepted oversized manifest")
	}
}

func FuzzParse(f *testing.F) {
	f.Add(testManifest("", 0))
	f.Add([]byte(`{"schemaVersion":1,"files":[{"source":"payload/a","target":"../x","originalSha256":"bad"}]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = Parse(data, true)
	})
}
