package gui

import (
	"mgs3mod/internal/launch"
	"mgs3mod/internal/manager"
	"strings"
)

// Choice is one value of a launch option, with its label.
type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// RegionChoice is a game region and the languages the launcher allows there.
type RegionChoice struct {
	Choice
	Languages []Choice `json:"languages"`
}

// LaunchOptions lists what a launch profile can hold.
type LaunchOptions struct {
	Regions     []RegionChoice `json:"regions"`
	Controllers []Choice       `json:"controllers"`
	Tested      []TestedChoice `json:"tested"` // the combinations played
}

// TestedChoice names the tested region, language and button prompts.
type TestedChoice struct {
	Region     string `json:"region"`
	Language   string `json:"language"`
	Controller string `json:"controller"`
}

// Options returns the launcher's regions, languages and button prompts.
func (s *Service) Options() LaunchOptions {
	o := LaunchOptions{Tested: []TestedChoice{}}
	for _, t := range launch.TestedSelections {
		o.Tested = append(o.Tested, TestedChoice{string(t.Region), string(t.Language), string(t.Controller)})
	}
	for _, r := range []launch.Region{launch.RegionUS, launch.RegionEU} {
		rc := RegionChoice{Choice: Choice{Value: string(r), Label: regionNames[r]}}
		for _, l := range launch.RegionLanguages[r] {
			rc.Languages = append(rc.Languages, Choice{Value: string(l), Label: languageNames[l]})
		}
		o.Regions = append(o.Regions, rc)
	}
	for _, c := range []launch.Controller{launch.ControllerKeyboard, launch.ControllerXbox, launch.ControllerPS4, launch.ControllerPS5, launch.ControllerNX} {
		label := strings.TrimSuffix(controllerNames[c], " button prompts")
		o.Controllers = append(o.Controllers, Choice{Value: string(c), Label: strings.ToUpper(label[:1]) + label[1:]})
	}
	return o
}

// ProfileID turns a name into a launch profile ID: lower case letters,
// digits, dots, dashes and underscores, starting with a letter or digit.
func ProfileID(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_':
			b.WriteRune(r)
			dash = false
		case r == '-' || r == ' ' || r == '/':
			if b.Len() > 0 && !dash {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	id := strings.Trim(b.String(), "-._")
	if len(id) > 64 {
		id = strings.TrimRight(id[:64], "-._")
	}
	return id
}

// ProfileResult is the outcome of a profile change.
type ProfileResult struct {
	OK      bool   `json:"ok"`
	ID      string `json:"id,omitempty"`
	Summary string `json:"summary"`
}

func profileFail(err error) ProfileResult {
	return ProfileResult{Summary: Explain(err).Summary}
}

// profileAction runs one profile change under the one-action lock; the
// manager's lock guards the file against the command line.
func (s *Service) profileAction(root string, change func(*manager.Manager) error) error {
	if !s.busy.TryLock() {
		return errBusy
	}
	s.running.Store(true) // the close guard also covers profile writes
	defer s.end()
	m, err := s.manager(root)
	if err != nil {
		return err
	}
	return change(m)
}

// SaveProfile creates a profile (original == "") or changes one; a new name
// renames it. makeDefault also makes it the default.
func (s *Service) SaveProfile(root, original, name, region, language, controller string, makeDefault bool) ProfileResult {
	id := ProfileID(name)
	if original != "" && strings.TrimSpace(name) == original {
		id = original // keep IDs from the command line that ProfileID would change
	}
	if id == "" {
		return ProfileResult{Summary: "Give the profile a name with letters or numbers."}
	}
	profile := launch.Profile{Selection: launch.Selection{Region: launch.Region(region), Language: launch.Language(language),
		Controller: launch.Controller(controller), Destination: launch.DestinationStartup}}
	if err := profile.Selection.Validate(); err != nil {
		return profileFail(&manager.Error{Code: 2, Message: err.Error()})
	}
	err := s.profileAction(root, func(m *manager.Manager) error {
		_, err := m.PutLaunchProfile(original, id, profile, makeDefault)
		return err
	})
	if err != nil {
		return profileFail(err)
	}
	return ProfileResult{OK: true, ID: id, Summary: "Saved launch profile " + id + "."}
}

// DeleteProfile removes a launch profile.
func (s *Service) DeleteProfile(root, id string) ProfileResult {
	err := s.profileAction(root, func(m *manager.Manager) error {
		_, err := m.RemoveLaunchProfile(id)
		return err
	})
	if err != nil {
		return profileFail(err)
	}
	return ProfileResult{OK: true, Summary: "Deleted launch profile " + id + "."}
}

// SetDefaultProfile makes a profile the one Launch game selects first.
func (s *Service) SetDefaultProfile(root, id string) ProfileResult {
	err := s.profileAction(root, func(m *manager.Manager) error {
		_, err := m.SetDefaultLaunchProfile(id)
		return err
	})
	if err != nil {
		return profileFail(err)
	}
	return ProfileResult{OK: true, ID: id, Summary: id + " is now the default launch profile."}
}
