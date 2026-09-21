package launch

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildArguments(t *testing.T) {
	got, err := BuildArguments(Seed().Profiles[SeedProfileID].Selection)
	want := []string{"-region", "us", "-lan", "en", "-selfregion", "EU", "-launcherpath", "launcher.exe", "-ctrltype", "KBD"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildArguments() = %#v, %v; want %#v", got, err, want)
	}
}

func TestSelectionRejectsInvalidAndUnsupported(t *testing.T) {
	base := Seed().Profiles[SeedProfileID].Selection
	for name, change := range map[string]func(*Selection){
		"region":      func(s *Selection) { s.Region = Region("xx") },
		"language":    func(s *Selection) { s.Language = Language("xx") },
		"controller":  func(s *Selection) { s.Controller = Controller("xx") },
		"destination": func(s *Selection) { s.Destination = DestinationMenu },
		"combination": func(s *Selection) { s.Region = RegionEU },
	} {
		t.Run(name, func(t *testing.T) {
			s := base
			change(&s)
			if _, err := BuildArguments(s); err == nil {
				t.Fatal("accepted unsupported selection")
			}
		})
	}
}

func TestStrictParseAndStableRoundTrip(t *testing.T) {
	cases := []string{
		`{"schema":1,"defaultProfileID":"na-startup","profiles":{"na-startup":{"region":"us","language":"en","controller":"kbd","destination":"startup"}},"extra":1}`,
		`{"schema":1,"schema":1,"defaultProfileID":"na-startup","profiles":{"na-startup":{"region":"us","language":"en","controller":"kbd","destination":"startup"}}}`,
		`{"schema":1,"defaultProfileID":"na-startup","profiles":{"na-startup":{"region":"us","language":"en","controller":"kbd","destination":"startup"}}} trailing`,
		`{"schema":2,"defaultProfileID":"na-startup","profiles":{"na-startup":{"region":"us","language":"en","controller":"kbd","destination":"startup"}}}`,
		`{"schema":1,"defaultProfileID":"missing","profiles":{"na-startup":{"region":"us","language":"en","controller":"kbd","destination":"startup"}}}`,
	}
	for _, input := range cases {
		if _, err := Parse([]byte(input)); err == nil {
			t.Errorf("Parse accepted %s", input)
		}
	}
	c, err := Parse(mustMarshal(Seed()))
	if err != nil {
		t.Fatal(err)
	}
	a, err := c.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.Marshal()
	if err != nil || string(a) != string(b) {
		t.Fatalf("unstable marshal: %s / %s (%v)", a, b, err)
	}
}

func TestSeedIsInMemoryAndCanonical(t *testing.T) {
	c := Seed()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.DefaultProfileID != SeedProfileID {
		t.Fatal("wrong seed default")
	}
	data, err := c.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "path") || strings.Contains(string(data), "save") {
		t.Fatalf("seed contains forbidden launch state: %s", data)
	}
}

func mustMarshal(c Config) []byte {
	b, err := c.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}
