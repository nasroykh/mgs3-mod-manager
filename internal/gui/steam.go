package gui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GameExe is the executable that marks an MGS3 game folder.
const GameExe = "METAL GEAR SOLID3.exe"

// ParseLibraryFolders returns every library path in Steam's
// steamapps/libraryfolders.vdf. It accepts the current format
// ("0" { "path" "D:\\SteamLibrary" ... }) and the old one ("1" "D:\\SteamLibrary").
func ParseLibraryFolders(text string) ([]string, error) {
	tokens, err := vdfTokens(text)
	if err != nil {
		return nil, err
	}
	paths := []string{}
	depth := 0
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch {
		case t.brace == '{':
			depth++
		case t.brace == '}':
			if depth == 0 {
				return nil, errors.New("libraryfolders.vdf: unbalanced braces")
			}
			depth--
		case i+1 < len(tokens) && tokens[i+1].brace == 0:
			key, value := t.text, tokens[i+1].text
			i++
			if depth == 2 && strings.EqualFold(key, "path") || depth == 1 && isNumber(key) {
				if value != "" {
					paths = append(paths, value)
				}
			}
		}
	}
	if depth != 0 {
		return nil, errors.New("libraryfolders.vdf: unbalanced braces")
	}
	return paths, nil
}

type vdfToken struct {
	text  string
	brace byte
}

func vdfTokens(text string) ([]vdfToken, error) {
	tokens := []vdfToken{}
	for i := 0; i < len(text); {
		c := text[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			i++
		case c == '/' && i+1 < len(text) && text[i+1] == '/':
			for i < len(text) && text[i] != '\n' {
				i++
			}
		case c == '{' || c == '}':
			tokens = append(tokens, vdfToken{brace: c})
			i++
		case c == '"':
			var b strings.Builder
			i++
			closed := false
			for i < len(text) {
				c = text[i]
				if c == '"' {
					closed = true
					i++
					break
				}
				if c == '\\' && i+1 < len(text) {
					i++
					switch text[i] {
					case 'n':
						b.WriteByte('\n')
					case 't':
						b.WriteByte('\t')
					default:
						b.WriteByte(text[i])
					}
					i++
					continue
				}
				b.WriteByte(c)
				i++
			}
			if !closed {
				return nil, errors.New("libraryfolders.vdf: unterminated string")
			}
			tokens = append(tokens, vdfToken{text: b.String()})
		default:
			start := i
			for i < len(text) && !strings.ContainsRune(" \t\r\n{}\"", rune(text[i])) {
				i++
			}
			tokens = append(tokens, vdfToken{text: text[start:i]})
		}
	}
	return tokens, nil
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// FindGames looks in each library's steamapps/common folders for the game
// executable. The game's folder name is not assumed.
func FindGames(libraries []string) []string {
	found := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		common := filepath.Join(library, "steamapps", "common")
		entries, err := os.ReadDir(common)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			folder := filepath.Join(common, entry.Name())
			if !HasGameExe(folder) {
				continue
			}
			key := strings.ToLower(filepath.Clean(folder))
			if !seen[key] {
				seen[key] = true
				found = append(found, filepath.Clean(folder))
			}
		}
	}
	sort.Strings(found)
	return found
}

// HasGameExe reports whether folder holds the game executable as a file.
func HasGameExe(folder string) bool {
	info, err := os.Stat(filepath.Join(folder, GameExe))
	return err == nil && info.Mode().IsRegular()
}

// SteamGames reads the Steam libraries listed under steamPath and returns
// every MGS3 folder in them.
func SteamGames(steamPath string) ([]string, error) {
	if steamPath == "" {
		return nil, nil
	}
	steamPath = filepath.Clean(filepath.FromSlash(steamPath))
	libraries := []string{steamPath}
	data, err := os.ReadFile(filepath.Join(steamPath, "steamapps", "libraryfolders.vdf"))
	if err == nil {
		more, parseErr := ParseLibraryFolders(string(data))
		if parseErr != nil {
			return FindGames(libraries), fmt.Errorf("read Steam libraries: %w", parseErr)
		}
		libraries = append(libraries, more...)
	} else if !errors.Is(err, os.ErrNotExist) {
		return FindGames(libraries), fmt.Errorf("read Steam libraries: %w", err)
	}
	return FindGames(libraries), nil
}
