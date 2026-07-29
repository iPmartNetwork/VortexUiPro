//go:build !linux && !darwin

package service

import "fmt"

// LoadPlugin is a stub for platforms that don't support Go plugins (e.g. Windows).
// The real implementation is in plugin_loader.go (Linux/Darwin only).
func (s *PluginService) LoadPlugin(path, id, name, version string) error {
	return fmt.Errorf("plugin loading not supported on this platform")
}
