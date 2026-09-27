package gui

import (
	"errors"

	"golang.org/x/sys/windows/registry"
)

// registrySteamPath returns HKCU\Software\Valve\Steam SteamPath, or "" without Steam.
func registrySteamPath() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer key.Close()
	value, _, err := key.GetStringValue("SteamPath")
	if errors.Is(err, registry.ErrNotExist) {
		return "", nil
	}
	return value, err
}
