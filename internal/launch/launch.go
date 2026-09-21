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
	if s.Region != RegionUS || s.Language != LanguageEnglish || s.Controller != ControllerKeyboard {
		return fmt.Errorf("unsupported launch combination: region=%q language=%q controller=%q", s.Region, s.Language, s.Controller)
	}
	return nil
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
	return []string{"-region", string(s.Region), "-lan", string(s.Language), "-selfregion", "EU", "-launcherpath", "launcher.exe", "-ctrltype", "KBD"}, nil
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
