package profile

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"mgs3mod/internal/winfs"
)

const (
	Root = `C:\Games\METAL GEAR SOLID 3 - MCV`
	ID   = "mgs3-mcv-local-0d585dcc6a67"
)

// Fingerprint identifies one immutable installation identity file.
type Fingerprint struct {
	Path   string
	SHA256 string
}

var Core = []Fingerprint{
	{Path: `METAL GEAR SOLID3.exe`, SHA256: "0d585dcc6a671be5d64d3d0a856c53f9ee0e58e7e4993f76dff29772c7a4bc80"},
	{Path: `Engine.dll`, SHA256: "4067774bd2945dfab1a81ee0f657b3b6c9414b1b363d29830652e6d93c516996"},
	{Path: `Renderer.dll`, SHA256: "663199bce1a252861369710d62219a73d2855043ec955ab7e8a926ea13986ac1"},
	{Path: `launcher.exe`, SHA256: "e061111cef605bdbf0ea7bc9cf686a1d29e317bf1da923e3c92c52f523479784"},
	{Path: `launcher_Data/Managed/Assembly-CSharp.dll`, SHA256: "3f4de01b2de9e9efc31001c0ad376b294eaeec5dd092f7e681204e93684af945"},
}

var targetPrefixes = [...]string{
	"textures/flatlist/_win/",
	"hqtex/flatlist/_win/",
}

// Target permits only existing .ctxr texture targets beneath approved trees.
func Target(path string) error {
	if err := Path(path); err != nil {
		return err
	}
	if !strings.HasSuffix(path, ".ctxr") {
		return fmt.Errorf("target must be a .ctxr file: %q", path)
	}
	for _, prefix := range targetPrefixes {
		if strings.HasPrefix(path, prefix) {
			return nil
		}
	}
	return fmt.Errorf("target is outside approved texture prefixes: %q", path)
}

// PluginTarget permits only a root-level lowercase .asi plugin filename.
func PluginTarget(path string) error {
	if err := Path(path); err != nil {
		return err
	}
	if strings.Contains(path, "/") || path != strings.ToLower(path) || !strings.HasSuffix(path, ".asi") || len(path) <= len(".asi") {
		return fmt.Errorf("plugin target must be a root-level lowercase .asi filename: %q", path)
	}
	return nil
}

// Path validates canonical printable-ASCII Windows-relative syntax.
func Path(path string) error {
	if path == "" || strings.ContainsAny(path, `\\:`) || strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return fmt.Errorf("path is not canonical: %q", path)
	}
	if len(path) >= 2 && path[1] == ':' {
		return fmt.Errorf("path is not relative: %q", path)
	}
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.HasSuffix(part, " ") || strings.HasSuffix(part, ".") {
			return fmt.Errorf("path is not canonical: %q", path)
		}
		for i := 0; i < len(part); i++ {
			if part[i] < 0x20 || part[i] > 0x7e || strings.ContainsRune(`*?[]<>|"`, rune(part[i])) {
				return fmt.Errorf("path contains unsafe character: %q", path)
			}
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		switch base {
		case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
			return fmt.Errorf("path contains reserved Windows name: %q", path)
		}
	}
	return nil
}

// Key returns the case-insensitive ownership key for a canonical path.
func Key(path string) string {
	return strings.ToLower(strings.ReplaceAll(path, `\`, "/"))
}

// Check verifies root and all supplied core hashes. A nil core uses compiled
// Core, allowing callers to use Check(root, nil) for the fixed profile.
func Check(root string, core []Fingerprint) error {
	if core == nil {
		core = Core
	}
	if err := winfs.CheckDir(root); err != nil {
		return fmt.Errorf("installation root: %w", err)
	}
	seen := make(map[string]struct{}, len(core))
	for _, expected := range core {
		if err := Path(expected.Path); err != nil {
			return fmt.Errorf("core path %q: %w", expected.Path, err)
		}
		key := Key(expected.Path)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate core path: %q", expected.Path)
		}
		seen[key] = struct{}{}
		if len(expected.SHA256) != sha256.Size*2 {
			return fmt.Errorf("core hash for %q is not SHA-256", expected.Path)
		}
		if expected.SHA256 != strings.ToLower(expected.SHA256) {
			return fmt.Errorf("core hash for %q is not lowercase hex", expected.Path)
		}
		if _, err := hex.DecodeString(expected.SHA256); err != nil {
			return fmt.Errorf("core hash for %q is not lowercase hex: %w", expected.Path, err)
		}
		if err := winfs.CheckPath(root, expected.Path, false); err != nil {
			return fmt.Errorf("core file %q: %w", expected.Path, err)
		}
		file, err := os.Open(root + `\` + strings.ReplaceAll(expected.Path, "/", `\`))
		if err != nil {
			return fmt.Errorf("open core %q: %w", expected.Path, err)
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return fmt.Errorf("hash core %q: %w", expected.Path, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close core %q: %w", expected.Path, closeErr)
		}
		actual := hex.EncodeToString(hash.Sum(nil))
		if actual != strings.ToLower(expected.SHA256) {
			return fmt.Errorf("core hash mismatch for %q: got %s want %s", expected.Path, actual, expected.SHA256)
		}
	}
	return nil
}
