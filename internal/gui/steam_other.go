//go:build !windows

package gui

func registrySteamPath() (string, error) { return "", nil }
