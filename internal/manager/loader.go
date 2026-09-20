package manager

import (
	"errors"
	"fmt"
	"mgs3mod/internal/profile"
	"os"
)

var loaderConfigPaths = []string{
	"wininet.ini",
	"global.ini",
	"scripts/global.ini",
	"plugins/global.ini",
	"update/global.ini",
}

func isASI(mod Mod) bool {
	for _, file := range mod.Manifest.Files {
		if profile.PluginTarget(file.Target) == nil {
			return true
		}
	}
	return false
}

func isLoader(mod Mod) bool {
	return mod.Manifest.SchemaVersion == 3 && mod.Manifest.ID == profile.LoaderID && mod.Manifest.Version == profile.LoaderVersion
}

func enabledASI(st State) bool {
	for _, mod := range st.Mods {
		if mod.Enabled && isASI(mod) {
			return true
		}
	}
	return false
}

func loaderTargetKeys() map[string]bool {
	keys := make(map[string]bool, len(profile.LoaderFiles))
	for _, file := range profile.LoaderFiles {
		keys[profile.Key(file.Path)] = true
	}
	return keys
}

func (m *Manager) defaultLoaderConfig() error {
	for _, name := range loaderConfigPaths {
		_, err := os.Lstat(pathJoin(m.config.Root, name))
		if err == nil {
			return fail(4, "managed ASI loader supports only its built-in default configuration; remove or rename loader configuration before enabling", name)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return wrapPath(1, "cannot inspect ASI loader configuration", name, err)
		}
	}
	return nil
}

// asiLoader requires manager ownership before checking the exact pinned files.
// CheckASILoader remains an internal test seam; production never configures it.
func (m *Manager) asiLoader(st State) error {
	if m.config.CheckASILoader != nil {
		return wrap(3, "ASI loader prerequisite", m.config.CheckASILoader(m.config.Root))
	}
	loader, ok := st.Mods[profile.LoaderID]
	if !ok || !loader.Enabled || !isLoader(loader) {
		return fail(3, fmt.Sprintf("ASI plugins require enabled managed %s %s", profile.LoaderID, profile.LoaderVersion))
	}
	if err := m.defaultLoaderConfig(); err != nil {
		return err
	}
	return wrap(3, "managed ASI loader files differ from the pinned release", profile.Check(m.config.Root, profile.LoaderFiles))
}
