// Package gui is the logic behind the window app (cmd/mgs3mod-gui): finding
// the game folder, the one-click Delta controls flows and plain-language
// errors. It calls the manager directly and does not import Wails, so every
// check the command line has (fingerprints, running game, lock, journal,
// backups) applies unchanged.
package gui

import (
	"errors"
	"fmt"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/packagefmt"
	"mgs3mod/internal/profile"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	fpvID   = "fpv-move"
	faceID  = "qcamo-face"
	qcamoID = "qcamo"
)

var pluginIDs = []string{fpvID, faceID}
var kitIDs = []string{profile.LoaderID, fpvID, faceID}

var displayNames = map[string]string{fpvID: "fpv-move", faceID: "QCamo face paint"}

// Opener returns a manager bound to one game folder.
type Opener func(root string) (*manager.Manager, error)

// Service runs one action at a time; the manager's own lock also guards
// against a command line run at the same time.
type Service struct {
	open      Opener
	kitDir    string
	settings  string
	steamPath func() (string, error)
	progress  func(Event)
	busy      sync.Mutex
	running   atomic.Bool
}

// New returns a service that installs the packages found in kitDir and keeps
// its settings in settingsPath ("" keeps nothing).
func New(open Opener, kitDir, settingsPath string) *Service {
	return &Service{open: open, kitDir: kitDir, settings: settingsPath, steamPath: registrySteamPath}
}

// Production binds the service to the compiled game profile.
func Production(kitDir, settingsPath string) *Service {
	return New(func(root string) (*manager.Manager, error) {
		return manager.Production().WithRoot(root)
	}, kitDir, settingsPath)
}

// Detection lists game folders found through Steam plus the remembered one.
type Detection struct {
	Folders  []string `json:"folders"`
	Selected string   `json:"selected"`
	Note     string   `json:"note,omitempty"`
}

// Detect finds game folders. The remembered folder wins when it still holds
// the game; otherwise a single Steam match is selected.
func (s *Service) Detect() Detection {
	d := Detection{Folders: []string{}}
	saved := LoadSettings(s.settings).GameRoot
	steam, err := s.steamPath()
	var found []string
	if err == nil {
		found, err = SteamGames(steam)
	}
	if err != nil {
		d.Note = "Steam libraries could not be read: " + err.Error()
	}
	seen := map[string]bool{}
	add := func(p string) {
		key := strings.ToLower(filepath.Clean(p))
		if !seen[key] {
			seen[key] = true
			d.Folders = append(d.Folders, filepath.Clean(p))
		}
	}
	if saved != "" && HasGameExe(saved) {
		add(saved)
		d.Selected = filepath.Clean(saved)
	}
	for _, p := range found {
		add(p)
	}
	if d.Selected == "" && len(d.Folders) == 1 {
		d.Selected = d.Folders[0]
	}
	return d
}

// Remember stores the chosen game folder.
func (s *Service) Remember(root string) error {
	if s.settings == "" {
		return nil
	}
	settings := LoadSettings(s.settings)
	settings.GameRoot = filepath.Clean(root)
	return SaveSettings(s.settings, settings)
}

// markQCamo records (on) or forgets (off) that Install turned QCamo off in
// root, so that Uninstall can turn it back on.
func (s *Service) markQCamo(root string, on bool) error {
	if s.settings == "" {
		return nil
	}
	settings := LoadSettings(s.settings)
	key := folderKey(root)
	kept := []string{}
	for _, r := range settings.QCamoTurnedOff {
		if folderKey(r) != key {
			kept = append(kept, r)
		}
	}
	if on {
		kept = append(kept, filepath.Clean(root))
	}
	if len(kept) == len(settings.QCamoTurnedOff) && !on {
		return nil
	}
	settings.QCamoTurnedOff = kept
	return SaveSettings(s.settings, settings)
}

func (s *Service) qcamoMarked(root string) bool {
	for _, r := range LoadSettings(s.settings).QCamoTurnedOff {
		if folderKey(r) == folderKey(root) {
			return true
		}
	}
	return false
}

func folderKey(root string) string {
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	return strings.ToLower(filepath.Clean(root))
}

// Check is the one-line result of doctor.
type Check struct {
	OK      bool   `json:"ok"`
	Summary string `json:"summary"`
	Details string `json:"details"`
}

// Doctor runs the manager's read-only doctor.
func (s *Service) Doctor(root string) Check {
	m, err := s.manager(root)
	if err == nil {
		_, err = m.Run("doctor", "", manager.Options{})
	}
	if err != nil {
		p := Explain(err)
		return Check{Summary: p.Summary, Details: "game folder: " + root + "\n" + p.Details}
	}
	return Check{OK: true, Summary: "Supported game version", Details: "game folder: " + root + "\ndoctor: installation inspected"}
}

// ModRow is one stored package.
type ModRow struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
	Kit     bool   `json:"kit"` // part of the Delta controls kit
}

// Overview is the Delta controls line and the mod list.
type Overview struct {
	Initialized bool     `json:"initialized"`
	Delta       string   `json:"delta"`
	DeltaState  string   `json:"deltaState"` // installed, missing, partial, outdated, unknown
	Kit         string   `json:"kit"`
	Mods        []ModRow `json:"mods"`
	Problem     *Problem `json:"problem,omitempty"`
}

// Status reads the manager's status and compares it with the kit packages.
func (s *Service) Status(root string) Overview {
	o := Overview{Mods: []ModRow{}, DeltaState: "unknown"}
	if strings.TrimSpace(root) == "" {
		o.Delta = "Choose the game folder first"
		return o
	}
	kit, kitErr := s.kit()
	if kitErr != nil {
		o.Kit = kitErr.Error()
	} else {
		o.Kit = "Packages found: " + describe(kit)
	}
	m, err := s.manager(root)
	var result manager.Result
	if err == nil {
		result, err = m.Run("status", "", manager.Options{})
	}
	if result.State == nil {
		if err != nil && !notInitialized(root) {
			p := Explain(err)
			o.Problem = &p
			o.Delta = "Status unavailable"
			return o
		}
		o.Delta, o.DeltaState = "Not installed", "missing"
		return o
	}
	o.Initialized = result.Initialized
	if err != nil {
		p := Explain(err)
		o.Problem = &p
	}
	for _, issue := range result.Issues {
		if o.Problem == nil {
			o.Problem = &Problem{Summary: summary(3, issue, nil), Details: issue}
		}
	}
	st := result.State
	ids := make([]string, 0, len(st.Mods))
	for id := range st.Mods {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		mod := st.Mods[id]
		o.Mods = append(o.Mods, ModRow{ID: id, Name: mod.Manifest.Name, Version: mod.Manifest.Version, Enabled: mod.Enabled, Kit: id == profile.LoaderID || id == fpvID || id == faceID})
	}
	o.Delta, o.DeltaState = deltaLine(st, kit)
	return o
}

// notInitialized reports whether the game folder has no manager state folder.
func notInitialized(root string) bool {
	_, err := os.Lstat(filepath.Join(root, ".mgs3mod"))
	return errors.Is(err, os.ErrNotExist)
}

func deltaLine(st *manager.State, kit map[string]kitPackage) (string, string) {
	installed := []string{}
	missing := []string{}
	outdated := false
	for _, id := range pluginIDs {
		mod, ok := st.Mods[id]
		if !ok || !mod.Enabled {
			missing = append(missing, displayNames[id])
			continue
		}
		installed = append(installed, displayNames[id]+" "+mod.Manifest.Version)
		if k, ok := kit[id]; ok && k.Digest != mod.Digest {
			outdated = true
		}
	}
	if loader, ok := st.Mods[profile.LoaderID]; len(installed) > 0 && (!ok || !loader.Enabled) {
		return "ASI loader missing: " + strings.Join(installed, ", ") + " cannot load. Use Install.", "partial"
	}
	switch {
	case len(installed) == 0:
		return "Not installed", "missing"
	case len(missing) > 0:
		return "Partly installed: " + strings.Join(installed, ", ") + "; missing " + strings.Join(missing, ", "), "partial"
	case outdated:
		return "Other version installed: " + strings.Join(installed, ", ") + ". Install / Update replaces it with " + describe(kit) + ".", "outdated"
	}
	return "Installed: " + strings.Join(installed, ", "), "installed"
}

func (s *Service) manager(root string) (*manager.Manager, error) {
	if strings.TrimSpace(root) == "" {
		return nil, &manager.Error{Code: 2, Message: "choose the game folder first"}
	}
	return s.open(root)
}

type kitPackage struct {
	Path    string
	Version string
	Digest  string
}

// kit loads the kit packages beside the executable: <id>-*.mgs3mod.zip for
// the loader and both plugins. Two candidates for one ID are refused.
func (s *Service) kit() (map[string]kitPackage, error) {
	kit := map[string]kitPackage{}
	if s.kitDir == "" {
		return kit, nil
	}
	for _, id := range kitIDs {
		matches, err := filepath.Glob(filepath.Join(s.kitDir, id+"-*.mgs3mod.zip"))
		if err != nil {
			return nil, err
		}
		candidates := matches
		if len(candidates) == 0 {
			continue
		}
		if len(candidates) > 1 {
			return nil, fmt.Errorf("several %s packages beside mgs3mod-gui.exe; keep one: %s", id, strings.Join(baseNames(candidates), ", "))
		}
		pkg, err := packagefmt.Load(candidates[0], false)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(candidates[0]), err)
		}
		if pkg.Manifest.ID != id {
			return nil, fmt.Errorf("%s holds %s, not %s", filepath.Base(candidates[0]), pkg.Manifest.ID, id)
		}
		kit[id] = kitPackage{Path: candidates[0], Version: pkg.Manifest.Version, Digest: packagefmt.Digest(pkg)}
	}
	return kit, nil
}

func baseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func describe(kit map[string]kitPackage) string {
	parts := []string{}
	for _, id := range pluginIDs {
		if k, ok := kit[id]; ok {
			parts = append(parts, displayNames[id]+" "+k.Version)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

// ExeDir returns the folder of the running executable.
func ExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}
