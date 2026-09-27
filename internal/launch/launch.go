// Package launch contains pure launch selections, profiles, and native argument construction.
package launch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
)

type Region string

const (
	RegionUS Region = "us"
	RegionJP Region = "jp"
	RegionEU Region = "eu"
)

type Language string

const (
	LanguageEnglish  Language = "en"
	LanguageJapanese Language = "jp"
	LanguageFrench   Language = "fr"
	LanguageItalian  Language = "it"
	LanguageGerman   Language = "gr"
	LanguageSpanish  Language = "sp"
)

type Controller string

const (
	ControllerKeyboard Controller = "kbd"
	ControllerXbox     Controller = "xbox"
	ControllerPS4      Controller = "ps4"
	ControllerPS5      Controller = "ps5"
	ControllerNX       Controller = "nx"
)

type Destination string

const (
	DestinationStartup Destination = "startup"
	DestinationMenu    Destination = "menu"
)

type Selection struct {
	Region      Region      `json:"region"`
	Language    Language    `json:"language"`
	Controller  Controller  `json:"controller"`
	Destination Destination `json:"destination"`
}
type Profile struct{ Selection }
type Config struct {
	Schema           int                `json:"schema"`
	DefaultProfileID string             `json:"defaultProfileID"`
	Profiles         map[string]Profile `json:"profiles"`
}

const SchemaVersion = 1
const SeedProfileID = "na-startup"

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

func Seed() Config {
	return Config{Schema: SchemaVersion, DefaultProfileID: SeedProfileID, Profiles: map[string]Profile{SeedProfileID: {Selection: Selection{Region: RegionUS, Language: LanguageEnglish, Controller: ControllerKeyboard, Destination: DestinationStartup}}}}
}
func (s Selection) Validate() error {
	if !validRegion(s.Region) {
		return fmt.Errorf("invalid region %q", s.Region)
	}
	if !validLanguage(s.Language) {
		return fmt.Errorf("invalid language %q", s.Language)
	}
	if !validController(s.Controller) {
		return fmt.Errorf("invalid controller %q", s.Controller)
	}
	if s.Destination != DestinationStartup {
		return fmt.Errorf("unsupported destination %q", s.Destination)
	}
	if !regionLanguage(s.Region, s.Language) {
		return fmt.Errorf("unsupported launch combination: region=%q language=%q", s.Region, s.Language)
	}
	return nil
}

// RegionLanguages lists the languages the official launcher offers for each
// game region (the language lists in GameLanguageSelectMGS3::FrameUpdate).
// Japan is left out: the launcher offers it only when a Steam download is
// installed (RegionSelectMGS3::FrameUpdate, DefManager::SetDLInfo), and the
// manager cannot tell that download apart yet.
var RegionLanguages = map[Region][]Language{
	RegionUS: {LanguageEnglish, LanguageFrench, LanguageSpanish},
	RegionEU: {LanguageEnglish, LanguageFrench, LanguageItalian, LanguageGerman, LanguageSpanish},
}

func regionLanguage(r Region, l Language) bool {
	for _, allowed := range RegionLanguages[r] {
		if allowed == l {
			return true
		}
	}
	return false
}

// TestedSelections were played through a manager launch on a real install:
// North America, English, keyboard prompts (2026-09-21) and Europe, French,
// Xbox prompts (2026-09-27, French text and Xbox prompts seen in game).
var TestedSelections = []Selection{
	{Region: RegionUS, Language: LanguageEnglish, Controller: ControllerKeyboard, Destination: DestinationStartup},
	{Region: RegionEU, Language: LanguageFrench, Controller: ControllerXbox, Destination: DestinationStartup},
}

// Tested reports whether this exact selection was played. The other
// selections use the launcher's own arguments but were never run.
func (s Selection) Tested() bool {
	for _, t := range TestedSelections {
		if s == t {
			return true
		}
	}
	return false
}

// ctrlTypeArguments are the launcher's -ctrltype values (Def::.cctor).
var ctrlTypeArguments = map[Controller]string{
	ControllerXbox: "XBOX", ControllerPS4: "PS4", ControllerPS5: "PS5", ControllerNX: "NX", ControllerKeyboard: "KBD",
}

func validRegion(v Region) bool { return v == RegionUS || v == RegionJP || v == RegionEU }
func validLanguage(v Language) bool {
	switch v {
	case LanguageEnglish, LanguageJapanese, LanguageFrench, LanguageItalian, LanguageGerman, LanguageSpanish:
		return true
	}
	return false
}
func validController(v Controller) bool {
	switch v {
	case ControllerKeyboard, ControllerXbox, ControllerPS4, ControllerPS5, ControllerNX:
		return true
	}
	return false
}
func BuildArguments(s Selection) ([]string, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	// The launcher build always passes its own region as EU (Def::regionLauncher)
	// and the button prompts of the controller it detected.
	return []string{"-region", string(s.Region), "-lan", string(s.Language), "-selfregion", "EU", "-launcherpath", "launcher.exe", "-ctrltype", ctrlTypeArguments[s.Controller]}, nil
}
func (c Config) Validate() error {
	if c.Schema != SchemaVersion {
		return fmt.Errorf("unsupported config schema %d", c.Schema)
	}
	if len(c.Profiles) == 0 {
		return errors.New("profiles must not be empty")
	}
	if !idPattern.MatchString(c.DefaultProfileID) {
		return fmt.Errorf("invalid default profile ID %q", c.DefaultProfileID)
	}
	if _, ok := c.Profiles[c.DefaultProfileID]; !ok {
		return fmt.Errorf("default profile %q does not exist", c.DefaultProfileID)
	}
	for id, p := range c.Profiles {
		if !idPattern.MatchString(id) {
			return fmt.Errorf("invalid profile ID %q", id)
		}
		if err := p.Selection.Validate(); err != nil {
			return fmt.Errorf("profile %q: %w", id, err)
		}
	}
	return nil
}
func Parse(data []byte) (Config, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return Config{}, errors.New("empty launch configuration")
	}
	if err := rejectDuplicateFields(data); err != nil {
		return Config{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var c Config
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("decode launch configuration: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, errors.New("trailing JSON data")
		}
		return Config{}, fmt.Errorf("trailing JSON data: %w", err)
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}
func (c Config) Marshal() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(c)
}
func rejectDuplicateFields(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := scanValue(dec); err != nil {
		return fmt.Errorf("invalid launch configuration: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON data")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}
func scanValue(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := t.(json.Delim); ok {
		switch d {
		case '{':
			seen := map[string]bool{}
			for dec.More() {
				key, e := dec.Token()
				if e != nil {
					return e
				}
				name, ok := key.(string)
				if !ok {
					return errors.New("object key is not string")
				}
				if seen[name] {
					return fmt.Errorf("duplicate JSON field %q", name)
				}
				seen[name] = true
				if e := scanValue(dec); e != nil {
					return e
				}
			}
			_, err = dec.Token()
		case '[':
			for dec.More() {
				if e := scanValue(dec); e != nil {
					return e
				}
			}
			_, err = dec.Token()
		}
	}
	return err
}
func (c Config) ProfileIDs() []string {
	ids := make([]string, 0, len(c.Profiles))
	for id := range c.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
func (c Config) Profile(id string) (Profile, error) {
	p, ok := c.Profiles[id]
	if !ok {
		return Profile{}, fmt.Errorf("profile %q does not exist", id)
	}
	return p, nil
}
