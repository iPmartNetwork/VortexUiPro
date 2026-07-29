//go:build linux || darwin

package service

import (
	"fmt"
	"log"
	"plugin"
	"time"
)

// LoadPlugin loads a Go plugin from a .so file (Linux/Darwin only).
func (s *PluginService) LoadPlugin(path, id, name, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("open plugin %s: %w", path, err)
	}

	symPlugin, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("lookup Plugin symbol: %w", err)
	}

	pluginInstance, ok := symPlugin.(PluginInterface)
	if !ok {
		return fmt.Errorf("plugin does not implement PluginInterface")
	}

	if err := pluginInstance.Init(map[string]any{}); err != nil {
		return fmt.Errorf("plugin init: %w", err)
	}

	info := &PluginInfo{
		ID: id, Name: name, Version: version,
		Enabled: true, Path: path, Status: "loaded",
		LoadedAt: time.Now().UnixMilli(),
	}

	s.plugins[id] = info
	s.instances[id] = pluginInstance

	defaultHooks := []HookPoint{HookAfterLogin, HookAfterUserCreate, HookHourlyCron}
	for _, h := range defaultHooks {
		s.hooks[h] = append(s.hooks[h], id)
	}
	info.Hooks = defaultHooks

	log.Printf("[Plugin] Loaded: %s v%s", name, version)
	return nil
}
