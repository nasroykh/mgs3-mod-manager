// Package cli contains the MGS3 command interface.
package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mgs3mod/internal/manager"
	"mgs3mod/internal/packagefmt"
	"sort"
	"strings"
)

const Help = `mgs3mod — MGS3 texture and ASI mod manager

Commands:
  doctor                         Inspect this installation (read only)
  init                           Initialize manager state
  pack <folder> --out <zip>       Create a validated local package
  add <zip-or-folder>             Store a package, disabled
  list | status | verify         Inspect packages, state, or integrity
  enable <id> | disable <id>      Apply a mod or restore its original files
  remove <id>                    Disable and remove its stored package
  restore --baseline             Restore all managed local originals
  recover                        Resolve an interrupted operation
  recover --restore-missing <relative-target> [--restore-missing <target> ...]

Options:
  --game-root <folder>  Select an installation of the supported game build
  --json       Emit one JSON result, including failures
  --dry-run    Validate mutations without writing any files
  --help       Show help without changing state

Missing files may have been deleted externally. Plain recover never recreates
them. Each --restore-missing target requires valid unresolved apply intent and
matching core fingerprints; the flag cannot overwrite existing files.
Baseline means captured local bytes, not factory settings. Close MGS3 and its
launcher before changing managed files. Game fingerprints remain mandatory.
Without --game-root, the original local installation path is used.
`

type parsed struct {
	command, arg, out, root string
	json, help, baseline    bool
	options                 manager.Options
}

type repeatedStrings []string

func (v *repeatedStrings) String() string         { return strings.Join(*v, ",") }
func (v *repeatedStrings) Set(value string) error { *v = append(*v, value); return nil }

func parse(args []string) (parsed, error) {
	p := parsed{}
	flags := flag.NewFlagSet("mgs3mod", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&p.json, "json", false, "emit JSON")
	flags.BoolVar(&p.help, "help", false, "show help")
	flags.BoolVar(&p.options.DryRun, "dry-run", false, "validate without writes")
	flags.BoolVar(&p.baseline, "baseline", false, "restore the local baseline")
	flags.StringVar(&p.out, "out", "", "output package")
	flags.StringVar(&p.root, "game-root", "", "game installation directory")
	flags.Var((*repeatedStrings)(&p.options.RestoreMissing), "restore-missing", "explicit missing target")
	pos := []string{}
	options := []string{}
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		v := args[i]
		if !strings.HasPrefix(v, "--") {
			pos = append(pos, v)
			continue
		}
		name := strings.TrimPrefix(v, "--")
		if flags.Lookup(name) == nil {
			return p, fmt.Errorf("unknown option %s", v)
		}
		if seen[v] && v != "--restore-missing" {
			return p, fmt.Errorf("duplicate option %s", v)
		}
		seen[v] = true
		options = append(options, v)
		if v == "--out" || v == "--restore-missing" || v == "--game-root" {
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "--") {
				return p, fmt.Errorf("%s needs a value", v)
			}
			i++
			options = append(options, args[i])
		}
	}
	if err := flags.Parse(options); err != nil {
		return p, err
	}
	if seen["--game-root"] && strings.TrimSpace(p.root) == "" {
		return p, fmt.Errorf("--game-root needs a nonempty directory")
	}
	if len(pos) == 0 {
		if p.help || len(args) == 0 {
			p.help = true
			return p, nil
		}
		return p, fmt.Errorf("command required")
	}
	p.command = pos[0]
	arity := map[string]int{"doctor": 0, "init": 0, "pack": 1, "add": 1, "list": 0, "status": 0, "verify": 0, "enable": 1, "disable": 1, "remove": 1, "restore": 0, "recover": 0}
	n, ok := arity[p.command]
	if !ok {
		return p, fmt.Errorf("unknown command %s", p.command)
	}
	if p.help {
		return p, nil
	}
	if len(pos) != n+1 {
		return p, fmt.Errorf("%s expects %d argument(s)", p.command, n)
	}
	if n == 1 {
		p.arg = pos[1]
	}
	if (p.command == "restore") != p.baseline {
		return p, fmt.Errorf("restore requires --baseline; no other command accepts it")
	}
	if (p.command == "pack") != (p.out != "") {
		return p, fmt.Errorf("pack requires --out; no other command accepts it")
	}
	if len(p.options.RestoreMissing) > 0 && p.command != "recover" {
		return p, fmt.Errorf("--restore-missing requires recover")
	}
	if p.options.DryRun && (p.command == "doctor" || p.command == "list" || p.command == "status" || p.command == "verify") {
		return p, fmt.Errorf("--dry-run is only for mutation commands")
	}
	return p, nil
}

func Run(args []string, stdout, stderr io.Writer, m *manager.Manager) int {
	wantsJSON := false
	for _, v := range args {
		if v == "--json" {
			wantsJSON = true
		}
	}
	p, err := parse(args)
	result := manager.Result{Command: p.command, DryRun: p.options.DryRun}
	if err != nil {
		err = &manager.Error{Code: 2, Message: err.Error()}
	} else if p.help {
		if wantsJSON {
			json.NewEncoder(stdout).Encode(map[string]any{"ok": true, "help": Help})
		} else {
			fmt.Fprint(stdout, Help)
		}
		return 0
	} else if p.command == "pack" {
		var pkg *packagefmt.Package
		if p.options.DryRun {
			pkg, err = packagefmt.Load(p.arg, true)
			if err == nil {
				err = packagefmt.CheckOutput(p.out)
			}
		} else {
			pkg, err = packagefmt.Pack(p.arg, p.out)
		}
		if err != nil {
			err = &manager.Error{Code: packagefmt.ErrorCode(err), Message: err.Error()}
		} else {
			result.Message = "package validated"
			result.Paths = []string{p.out}
			if pkg != nil {
				result.Message += ": " + pkg.Manifest.ID
			}
		}
	} else {
		if p.root != "" {
			m, err = m.WithRoot(p.root)
		}
		if err == nil {
			result, err = m.Run(p.command, p.arg, p.options)
		}
	}
	code := 0
	var detail *manager.Error
	if err != nil {
		code = manager.ExitCode(err)
		if !errors.As(err, &detail) {
			detail = &manager.Error{Code: code, Message: err.Error()}
		}
	}
	if wantsJSON {
		response := struct {
			OK     bool           `json:"ok"`
			Result manager.Result `json:"result"`
			Error  *manager.Error `json:"error,omitempty"`
		}{err == nil, result, detail}
		if e := json.NewEncoder(stdout).Encode(response); e != nil {
			fmt.Fprintln(stderr, e)
			return 1
		}
	} else {
		if result.Message != "" {
			fmt.Fprintln(stdout, result.Message)
		}
		if result.State != nil {
			ids := []string{}
			for id := range result.State.Mods {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				mod := result.State.Mods[id]
				state := "disabled"
				if mod.Enabled {
					state = "enabled"
				}
				fmt.Fprintf(stdout, "%s  %s  %s  %d target(s)\n", id, mod.Manifest.Version, state, len(mod.Manifest.Files))
			}
		}
		for _, p := range result.Paths {
			fmt.Fprintln(stdout, p)
		}
		for _, target := range result.Targets {
			owner := target.Owner
			if owner == "" {
				owner = "baseline"
			}
			status := "matches"
			if target.Issue != "" {
				status = target.Issue
			}
			fmt.Fprintf(stdout, "%s  owner=%s  %s\n", target.Path, owner, status)
		}
		for _, issue := range result.Issues {
			fmt.Fprintln(stdout, issue)
		}
		if result.DryRun {
			fmt.Fprintln(stdout, "Dry run: no files changed.")
		}
		if err != nil {
			fmt.Fprintf(stderr, "error [%d]: %s\n", code, err)
		}
	}
	return code
}
