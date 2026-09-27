package gui

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mgs3mod/internal/launch"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/profile"
	"os"
	"path/filepath"
)

var regionNames = map[launch.Region]string{launch.RegionUS: "North America", launch.RegionJP: "Japan", launch.RegionEU: "Europe"}
var languageNames = map[launch.Language]string{
	launch.LanguageEnglish: "English", launch.LanguageJapanese: "Japanese", launch.LanguageFrench: "French",
	launch.LanguageItalian: "Italian", launch.LanguageGerman: "German", launch.LanguageSpanish: "Spanish",
}
var controllerNames = map[launch.Controller]string{
	launch.ControllerKeyboard: "keyboard button prompts", launch.ControllerXbox: "Xbox button prompts",
	launch.ControllerPS4: "PlayStation 4 button prompts", launch.ControllerPS5: "PlayStation 5 button prompts",
	launch.ControllerNX: "Nintendo Switch button prompts",
}
var destinationNames = map[launch.Destination]string{launch.DestinationStartup: "game startup", launch.DestinationMenu: "main menu"}

func name[K comparable](names map[K]string, key K) string {
	if n, ok := names[key]; ok {
		return n
	}
	return fmt.Sprint(key)
}

// Describe puts a launch selection in words.
func Describe(sel launch.Selection) string {
	return name(regionNames, sel.Region) + " · " + name(languageNames, sel.Language) + " · " +
		name(controllerNames, sel.Controller) + " · " + name(destinationNames, sel.Destination)
}

// LaunchProfile is one saved (or the built-in) launch profile.
type LaunchProfile struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Region      string `json:"region"`
	Language    string `json:"language"`
	Controller  string `json:"controller"`
	Default     bool   `json:"default"`
	Supported   bool   `json:"supported"`
	Tested      bool   `json:"tested"`
	Reason      string `json:"reason,omitempty"`
}

// LaunchInfo is what the Play card shows.
type LaunchInfo struct {
	Available bool            `json:"available"`
	Reason    string          `json:"reason,omitempty"`
	Saved     bool            `json:"saved"` // profiles come from mgs3mod-launch.json
	Profiles  []LaunchProfile `json:"profiles"`
	SteamCopy bool            `json:"steamCopy"`
}

// LaunchInfo lists the launch profiles. It changes nothing.
func (s *Service) LaunchInfo(root string) LaunchInfo {
	info := LaunchInfo{Profiles: []LaunchProfile{}}
	if root == "" {
		info.Reason = "Choose the game folder first."
		return info
	}
	info.SteamCopy = isSteamCopy(root)
	if notInitialized(root) {
		info.Reason = "Use Install / Update first: quick launch needs the manager set up in this game folder."
		return info
	}
	m, err := s.manager(root)
	var config manager.LaunchConfig
	if err == nil {
		config, err = m.LoadLaunchConfig()
	}
	if err != nil {
		info.Reason = Explain(err).Summary
		return info
	}
	info.Saved = config.Persisted
	for _, id := range config.Config.ProfileIDs() {
		p, _ := config.Config.Profile(id)
		lp := LaunchProfile{ID: id, Description: Describe(p.Selection), Region: string(p.Region), Language: string(p.Language),
			Controller: string(p.Controller), Default: id == config.Config.DefaultProfileID, Supported: true, Tested: p.Selection.Tested()}
		if err := p.Selection.Validate(); err != nil {
			lp.Supported, lp.Reason = false, "The manager refuses it: "+err.Error()
		}
		info.Profiles = append(info.Profiles, lp)
	}
	for _, p := range info.Profiles {
		if p.Supported {
			info.Available = true
		}
	}
	if !info.Available {
		info.Reason = "No saved launch profile can be started."
	}
	return info
}

// steamWrappedBytes is the size of the Steam-wrapped executable, so other
// files are told apart without hashing them.
const steamWrappedBytes = 12948040

// isSteamCopy reports whether the game executable is the Steam-wrapped file.
func isSteamCopy(root string) bool {
	path := filepath.Join(root, GameExe)
	if info, err := os.Stat(path); err != nil || info.Size() != steamWrappedBytes {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return false
	}
	return hex.EncodeToString(h.Sum(nil)) == profile.SteamWrappedExecutable
}

// Launch starts the game with one launch profile, skipping the Master
// Collection selector, through the manager's own launch checks.
func (s *Service) Launch(root, profileID string) Outcome {
	f, ok := s.begin("launch", root)
	if !ok {
		return f.out
	}
	defer s.end()
	var sel launch.Selection
	pid := 0
	f.plan(
		action{label: "Check the game folder and launch settings", before: func() error {
			config, err := f.m.LoadLaunchConfig()
			if err != nil {
				return err
			}
			p, err := config.Config.Profile(profileID)
			if err != nil {
				return &manager.Error{Code: 2, Message: err.Error()}
			}
			sel = p.Selection
			_, err = f.m.Launch(sel, true)
			return err
		}},
		action{label: "Start the game", before: func() error {
			receipt, err := f.m.Launch(sel, false)
			pid = receipt.PID
			return err
		}},
	)
	if _, ok := f.runFrom(0); !ok {
		return f.finish("")
	}
	return f.finish(fmt.Sprintf("Game started (%s; process %d).", Describe(sel), pid))
}
