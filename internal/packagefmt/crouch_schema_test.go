package packagefmt

import (
	"encoding/json"
	"strings"
	"testing"

	"mgs3mod/internal/profile"
)

func crouchManifestForTest() Manifest {
	files := profile.CrouchFiles()
	m := Manifest{SchemaVersion: 4, ID: profile.CrouchID, Version: profile.CrouchVersion, Name: "Crouch walk", Profile: profile.ID}
	for _, file := range files {
		m.Files = append(m.Files, File{Source: "payload/" + file.Target, Target: file.Target, OriginalSHA256: file.OriginalSHA256, OriginalAbsent: file.OriginalAbsent, PayloadSHA256: file.PayloadSHA256, PayloadBytes: file.PayloadBytes})
	}
	return m
}

func TestSchema4CrouchManifestRejectsIdentityAndSetTampering(t *testing.T) {
	valid := crouchManifestForTest()
	if _, err := Parse(mustJSON(valid), true); err != nil {
		t.Fatalf("valid schema4 manifest rejected: %v", err)
	}
	cases := map[string]func(*Manifest){
		"wrong id":       func(m *Manifest) { m.ID = "spoof" },
		"wrong version":  func(m *Manifest) { m.Version = "0.2.2" },
		"wrong profile":  func(m *Manifest) { m.Profile = "other-profile" },
		"missing":        func(m *Manifest) { m.Files = m.Files[:len(m.Files)-1] },
		"extra":          func(m *Manifest) { m.Files = append(m.Files, m.Files[0]) },
		"duplicate":      func(m *Manifest) { m.Files[1].Target = m.Files[0].Target; m.Files[1].Source = m.Files[0].Source },
		"target tamper":  func(m *Manifest) { m.Files[0].Target = "textures/flatlist/_win/fake.ctxr" },
		"source tamper":  func(m *Manifest) { m.Files[0].Source = "payload/other.mtar" },
		"origin tamper":  func(m *Manifest) { m.Files[0].OriginalSHA256 = strings.Repeat("0", 64) },
		"payload tamper": func(m *Manifest) { m.Files[0].PayloadSHA256 = strings.Repeat("0", 64) },
		"size tamper":    func(m *Manifest) { m.Files[0].PayloadBytes++ },
	}
	for name, mutate := range cases {
		m := valid
		m.Files = append([]File(nil), valid.Files...)
		mutate(&m)
		if _, err := Parse(mustJSON(m), true); err == nil {
			t.Errorf("Parse accepted %s", name)
		}
	}
}

func TestSchema4DoesNotDowngradeToLegacySchemas(t *testing.T) {
	valid := crouchManifestForTest()
	for _, schema := range []int{1, 2, 3} {
		m := valid
		m.SchemaVersion = schema
		if _, err := Parse(mustJSON(m), true); err == nil {
			t.Errorf("schema %d accepted crouch asset package", schema)
		}
	}
}

func TestCrouchAssetsSkipPEValidationButASIDoesNot(t *testing.T) {
	file := profile.CrouchFiles()[0]
	asset := &File{Target: file.Target}
	if err := fillOrCheckPayload(asset, []byte("not a PE"), true); err != nil {
		t.Fatalf("non-PE crouch asset rejected: %v", err)
	}
	plugin := &File{Target: "dinput8.asi"}
	if err := fillOrCheckPayload(plugin, []byte("not a PE"), true); err == nil {
		t.Fatal("non-PE ASI payload accepted")
	}
}

func mustJSON(m Manifest) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}
