package gui

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Settings is what the window app remembers between runs. It lives in
// %LOCALAPPDATA%\mgs3mod\gui.json, never inside the game folder.
type Settings struct {
	GameRoot string `json:"gameRoot"`
	// QCamoTurnedOff lists the game folders where Install turned QCamo 1.0.4
	// off to make room for the face paint fork; Uninstall turns it back on.
	QCamoTurnedOff []string `json:"qcamoTurnedOff,omitempty"`
}

// DefaultSettingsPath returns %LOCALAPPDATA%\mgs3mod\gui.json.
func DefaultSettingsPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mgs3mod", "gui.json"), nil
}

// LoadSettings returns empty settings when the file is absent or unreadable:
// the only setting is a convenience, so a bad file must not block the app.
func LoadSettings(path string) Settings {
	var s Settings
	if path == "" {
		return s
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) > 64<<10 {
		return Settings{}
	}
	if json.Unmarshal(data, &s) != nil {
		return Settings{}
	}
	return s
}

// SaveSettings writes the file through a temporary file and a rename.
func SaveSettings(path string, s Settings) error {
	if path == "" {
		return errors.New("no settings path")
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "gui-*.json.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, writeErr := tmp.Write(append(data, '\n'))
	syncErr := tmp.Sync()
	closeErr := tmp.Close()
	if err = errors.Join(writeErr, syncErr, closeErr); err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		os.Remove(name)
	}
	return err
}
