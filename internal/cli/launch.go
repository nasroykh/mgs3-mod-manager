package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mgs3mod/internal/launch"
	"mgs3mod/internal/manager"
	"regexp"
	"strings"
)

var launchProfileIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

func Run(args []string, stdout, stderr io.Writer, m *manager.Manager) int {
	return RunWithInput(args, stdout, stderr, m, strings.NewReader(""), false)
}

func validLaunchProfileID(id string) bool {
	return launchProfileIDPattern.MatchString(id)
}

func launchSelection(region, language, controller, target string) (launch.Selection, error) {
	s := launch.Selection{Region: launch.Region(region), Language: launch.Language(language), Controller: launch.Controller(controller), Destination: launch.Destination(target)}
	if err := s.Validate(); err != nil {
		return s, err
	}
	return s, nil
}

func RunWithInput(args []string, stdout, stderr io.Writer, m *manager.Manager, stdin io.Reader, interactive bool) int {
	p, err := parse(args)
	wantsJSON := false
	for _, a := range args {
		if a == "--json" {
			wantsJSON = true
		}
	}
	if err != nil {
		return runExisting(args, stdout, stderr, m)
	}
	if p.help {
		if wantsJSON {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"ok": true, "help": Help})
		} else {
			_, _ = io.WriteString(stdout, Help)
		}
		return 0
	}
	if p.command != "launch" {
		return runExisting(args, stdout, stderr, m)
	}
	if m == nil {
		return writeCLIError(stdout, stderr, wantsJSON, fmt.Errorf("manager unavailable"))
	}
	if p.root != "" {
		m, err = m.WithRoot(p.root)
		if err != nil {
			return writeCLIError(stdout, stderr, wantsJSON, err)
		}
	}
	if p.selectLaunch {
		s, cfg, ok, e := selectLaunch(stdin, stdout, m, interactive)
		if e != nil {
			return writeCLIError(stdout, stderr, wantsJSON, selectionInputError(e))
		}
		if !ok {
			return 0
		}
		return launchResolved(stdout, stderr, m, s, p.options.DryRun, wantsJSON, cfg)
	}
	if p.profile == "" && p.region != "" {
		selection, e := launchSelection(p.region, p.language, p.controller, p.target)
		if e != nil {
			return writeCLIError(stdout, stderr, wantsJSON, &manager.Error{Code: 2, Message: e.Error()})
		}
		return launchResolved(stdout, stderr, m, selection, p.options.DryRun, wantsJSON, launch.Config{})
	}
	loaded, err := m.LoadLaunchConfig()
	if err != nil {
		return writeCLIError(stdout, stderr, wantsJSON, err)
	}
	var selection launch.Selection
	var cfg launch.Config
	if p.profile != "" {
		cfg = loaded.Config
		profile, e := cfg.Profile(p.profile)
		if e != nil {
			return writeCLIError(stdout, stderr, wantsJSON, &manager.Error{Code: 2, Message: e.Error()})
		}
		selection = profile.Selection
		if p.region != "" {
			selection.Region = launch.Region(p.region)
		}
		if p.language != "" {
			selection.Language = launch.Language(p.language)
		}
		if p.controller != "" {
			selection.Controller = launch.Controller(p.controller)
		}
		if p.target != "" {
			selection.Destination = launch.Destination(p.target)
		}
	} else {
		cfg = loaded.Config
		profile, e := cfg.Profile(cfg.DefaultProfileID)
		if e != nil {
			return writeCLIError(stdout, stderr, wantsJSON, &manager.Error{Code: 2, Message: e.Error()})
		}
		selection = profile.Selection
		if !loaded.Persisted && (!interactive || p.json || p.options.DryRun) {
			return writeCLIError(stdout, stderr, wantsJSON, &manager.Error{Code: 2, Message: "no saved launch profile; provide --profile or complete launch overrides"})
		}
		if !loaded.Persisted && interactive {
			s, cfg2, ok, e := selectLaunch(stdin, stdout, m, interactive)
			if e != nil {
				return writeCLIError(stdout, stderr, wantsJSON, selectionInputError(e))
			}
			if !ok {
				return 0
			}
			selection, cfg = s, cfg2
		}
	}
	return launchResolved(stdout, stderr, m, selection, p.options.DryRun, wantsJSON, cfg)
}

func selectionInputError(err error) error {
	var detail *manager.Error
	if errors.As(err, &detail) {
		return err
	}
	return &manager.Error{Code: 2, Message: err.Error()}
}

func runExisting(args []string, stdout, stderr io.Writer, m *manager.Manager) int {
	return runOriginal(args, stdout, stderr, m)
}

func writeCLIError(stdout, stderr io.Writer, jsonMode bool, err error) int {
	code := manager.ExitCode(err)
	if code == 0 {
		code = 2
	}
	detail := launchErrorDetail(err, code)
	if jsonMode {
		_ = json.NewEncoder(stdout).Encode(map[string]any{"ok": false, "result": nil, "error": detail})
		return code
	}
	_, _ = fmt.Fprintf(stderr, "error [%d]: %s\n", code, err)
	return code
}

func launchErrorDetail(err error, code int) *manager.Error {
	var detail *manager.Error
	if errors.As(err, &detail) {
		return detail
	}
	return &manager.Error{Code: code, Message: err.Error()}
}

func launchResolved(stdout, stderr io.Writer, m *manager.Manager, s launch.Selection, dry, jsonMode bool, _ launch.Config) int {
	receipt, err := m.Launch(s, dry)
	code := 0
	if err != nil {
		code = manager.ExitCode(err)
		if code == 0 {
			code = 1
		}
	}
	if jsonMode {
		var detail *manager.Error
		if err != nil {
			detail = launchErrorDetail(err, code)
		}
		_ = json.NewEncoder(stdout).Encode(map[string]any{"ok": err == nil, "result": receipt, "error": detail})
		return code
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error [%d]: %s\n", code, err)
		return code
	}
	if dry {
		fmt.Fprintf(stdout, "Executable: %s\nWorking directory: %s\nArguments: %s\nSelection: %s/%s/%s/%s\n%s\n", receipt.Executable, receipt.WorkingDir, strings.Join(receipt.Arguments, " "), s.Region, s.Language, s.Controller, s.Destination, receipt.Message)
	} else {
		fmt.Fprintf(stdout, "%s (PID %d)\n", receipt.Message, receipt.PID)
	}
	return 0
}

func selectLaunch(in io.Reader, out io.Writer, m *manager.Manager, interactive bool) (launch.Selection, launch.Config, bool, error) {
	if !interactive {
		return launch.Selection{}, launch.Config{}, false, fmt.Errorf("launch selector requires interactive input")
	}
	loaded, err := m.LoadLaunchConfig()
	if err != nil {
		return launch.Selection{}, launch.Config{}, false, err
	}
	cfg := loaded.Config
	current := cfg.Profiles[cfg.DefaultProfileID].Selection
	currentID := cfg.DefaultProfileID
	r := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "Current: profile=%s region=%s language=%s controller=%s destination=%s (%s)\nProfiles: %s\n1) launch current\n2) select existing profile\n3) edit current choices\n4) save as named profile\n5) choose default\n6) cancel\nChoice [1]: ", currentID, current.Region, current.Language, current.Controller, current.Destination, persistedLabel(loaded.Persisted), strings.Join(cfg.ProfileIDs(), ", "))
		line, e := r.ReadString(10)
		line = strings.TrimSpace(line)
		if e != nil && line == "" {
			return launch.Selection{}, cfg, false, nil
		}
		if line == "" {
			line = "1"
		}
		if strings.EqualFold(line, "q") || line == "6" || line == "\x1b" {
			return launch.Selection{}, cfg, false, nil
		}
		switch line {
		case "1":
			if err := current.Validate(); err != nil {
				return launch.Selection{}, cfg, false, err
			}
			return current, cfg, true, nil
		case "2":
			id, e := readLine(out, r, "Profile ID", "")
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
			p, e := cfg.Profile(id)
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
			current = p.Selection
			currentID = id
		case "3":
			current, e = editSelection(out, r, current)
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
		case "4":
			id, e := readLine(out, r, "Profile ID", "")
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
			if !validLaunchProfileID(id) {
				return launch.Selection{}, cfg, false, fmt.Errorf("invalid profile ID %q", id)
			}
			if e = current.Validate(); e != nil {
				return launch.Selection{}, cfg, false, e
			}
			cfg, e = m.AddLaunchProfile(id, launch.Profile{Selection: current})
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
			currentID = id
			loaded.Persisted = true
		case "5":
			id, e := readLine(out, r, "Default profile ID", cfg.DefaultProfileID)
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
			cfg, e = m.SetDefaultLaunchProfile(id)
			if e != nil {
				return launch.Selection{}, cfg, false, e
			}
			currentID = id
			current = cfg.Profiles[id].Selection
			loaded.Persisted = true
		default:
			fmt.Fprintln(out, "Invalid choice")
		}
	}
}

func persistedLabel(v bool) string {
	if v {
		return "persisted"
	}
	return "in-memory default"
}
func readLine(out io.Writer, r *bufio.Reader, label, def string) (string, error) {
	if def == "" {
		fmt.Fprintf(out, "%s: ", label)
	} else {
		fmt.Fprintf(out, "%s [%s]: ", label, def)
	}
	line, e := r.ReadString(10)
	if e != nil && line == "" {
		return "", fmt.Errorf("input ended")
	}
	line = strings.TrimSpace(line)
	if line == "" {
		line = def
	}
	return line, nil
}
func editSelection(out io.Writer, r *bufio.Reader, s launch.Selection) (launch.Selection, error) {
	var e error
	if v, x := readLine(out, r, "region", string(s.Region)); x != nil {
		return s, x
	} else {
		s.Region = launch.Region(v)
	}
	if v, x := readLine(out, r, "language", string(s.Language)); x != nil {
		return s, x
	} else {
		s.Language = launch.Language(v)
	}
	if v, x := readLine(out, r, "controller", string(s.Controller)); x != nil {
		return s, x
	} else {
		s.Controller = launch.Controller(v)
	}
	if v, x := readLine(out, r, "target", string(s.Destination)); x != nil {
		return s, x
	} else {
		s.Destination = launch.Destination(v)
	}
	e = s.Validate()
	return s, e
}
