package manager

import "path/filepath"

// WithRoot selects a different installation without weakening the compiled
// fingerprint or process checks. Existing state remains bound to its root.
func (m *Manager) WithRoot(root string) (*Manager, error) {
	if root == "" {
		return nil, fail(2, "game root must not be empty")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, wrap(2, "game root", err)
	}
	c := m.config
	c.Root = filepath.Clean(absolute)
	return New(c), nil
}
