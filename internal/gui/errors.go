package gui

import (
	"errors"
	"fmt"
	"mgs3mod/internal/manager"
	"regexp"
	"strings"
)

// Problem is a manager error in one plain sentence, with the original
// message kept for reports.
type Problem struct {
	Summary string `json:"summary"`
	Details string `json:"details"`
}

var coreMismatch = regexp.MustCompile(`core hash mismatch for "([^"]+)"`)
var coreMissing = regexp.MustCompile(`(?:open core|core file) "([^"]+)"`)

// Explain maps the common manager errors to one sentence. Anything it does
// not recognize keeps the manager's own message.
func Explain(err error) Problem {
	if err == nil {
		return Problem{}
	}
	code := manager.ExitCode(err)
	message := err.Error()
	var paths []string
	var detail *manager.Error
	if errors.As(err, &detail) {
		paths = detail.Paths
	}
	details := fmt.Sprintf("error [%d]: %s", code, message)
	for _, p := range paths {
		details += "\n" + p
	}
	return Problem{Summary: summary(code, message, paths), Details: details}
}

func summary(code int, message string, paths []string) string {
	first := ""
	if len(paths) > 0 {
		first = paths[0]
	}
	has := func(s string) bool { return strings.Contains(message, s) }
	switch {
	// An interrupted change wraps its cause; recovery comes first.
	case has("unresolved transaction"), has("run recover"), has("use recover"):
		return "An earlier change was interrupted. Run \"mgs3mod.exe recover\" from the kit folder (see the guide, \"Advanced: the command line\") before anything else."
	case has("close the game and launcher"):
		return "Close the game and the Master Collection launcher first, then try again."
	case has("another manager holds the lock"):
		return "Another copy of the manager is changing this game folder. Wait for it to finish, then try again."
	case code == 3 && has("installation fingerprints"):
		if m := coreMismatch.FindStringSubmatch(message); m != nil {
			return "Not supported: " + m[1] + " differs from the supported game version. Do not replace game files to make it pass."
		}
		if m := coreMissing.FindStringSubmatch(message); m != nil {
			return "This is not a supported MGS3 game folder: " + m[1] + " is missing or unreadable."
		}
		return "This folder is not a supported MGS3 game folder."
	case has("ASI plugins require enabled managed"):
		return "This mod needs the ASI loader. Use Install in Delta controls, or add and turn on asi-loader first."
	case has("enabled ASI packages depend on the managed loader"):
		return "The ASI loader is still needed by another mod that is turned on."
	case has("managed ASI loader files differ"):
		return "The ASI loader file (wininet.dll) was changed outside the manager."
	case has("managed ASI loader supports only its built-in default configuration"):
		return "Remove or rename " + first + " first: it is another ASI loader's configuration."
	case has("target owned by "):
		owner := message[strings.Index(message, "target owned by ")+len("target owned by "):]
		return "Conflict: " + first + " is already changed by the mod " + owner + ". Turn " + owner + " off first."
	case has("expected file to be absent"):
		return first + " already exists and does not belong to the manager (another loader or mod?). Remove it first."
	case code == 4 && (has("unexpected file contents") || has("external drift")):
		return "A game file was changed outside the manager: " + first + "."
	case has("ID already stores different content"):
		return "A different build of this mod is already stored. Remove it first."
	case has("cannot load package"):
		return "This file is not a valid mod package."
	case has("insufficient free space"):
		return "Not enough free disk space on the game drive."
	case has("unsupported launch combination: region=\"jp\""):
		return "The Japan region is not offered yet: the launcher shows it only with a Steam download installed, and it was never tested."
	case has("unsupported launch combination"):
		return "The launcher does not offer this language for this region."
	case has("unsupported destination"):
		return "Only the game startup destination is supported."
	case has("launch profile ") && has(" already exists"):
		return "A launch profile with this name already exists. Choose another name."
	case has("is the only profile and cannot be removed"):
		return "This is the only launch profile; it cannot be deleted."
	case has("game process exited before launch verification"):
		return "The game closed right after starting. On a Steam copy, start the game from Steam."
	case has("launch protection remained active until child exit"):
		return "The game started but could not be confirmed, so the manager kept its lock until the game closed."
	case has("unsafe or missing game executable"), has("unsafe or missing launcher executable"):
		return "The game or launcher executable is missing or unsafe to start: " + first + "."
	case has("manager not initialized"):
		return "The manager is not set up in this game folder yet, or its .mgs3mod folder is damaged."
	case has("unknown mod ID"):
		return "That mod is not stored: " + first + "."
	}
	return "The manager stopped: " + message
}
