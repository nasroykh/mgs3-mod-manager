package profile

import (
	_ "embed"
	"encoding/json"
)

const (
	CrouchID      = "crouch-walk"
	CrouchVersion = "0.2.1"
)

// CrouchFile describes one pinned crouch-walk replacement target.
type CrouchFile struct {
	Target         string `json:"target"`
	OriginalSHA256 string `json:"originalSha256"`
	OriginalAbsent bool   `json:"originalAbsent"`
	PayloadSHA256  string `json:"payloadSha256"`
	PayloadBytes   int64  `json:"payloadBytes"`
}

//go:embed crouch-files.json
var crouchFilesJSON []byte

var crouchFiles = loadCrouchFiles()

func loadCrouchFiles() []CrouchFile {
	var files []CrouchFile
	if err := json.Unmarshal(crouchFilesJSON, &files); err != nil || len(files) == 0 {
		panic("invalid embedded crouch-walk metadata")
	}
	return files
}

// CrouchFiles returns a copy of pinned crouch-walk metadata.
func CrouchFiles() []CrouchFile {
	return append([]CrouchFile(nil), crouchFiles...)
}

// CrouchFileFor returns pinned metadata for target.
func CrouchFileFor(target string) (CrouchFile, bool) {
	for _, file := range crouchFiles {
		if file.Target == target {
			return file, true
		}
	}
	return CrouchFile{}, false
}

// ReplacementTarget permits existing textures and present-origin crouch files.
func ReplacementTarget(path string) error {
	if Target(path) == nil {
		return nil
	}
	if file, ok := CrouchFileFor(path); ok && !file.OriginalAbsent {
		return nil
	}
	return Target(path)
}
