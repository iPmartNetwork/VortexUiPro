package service

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type HookPoint string

const (
	HookAfterLogin         HookPoint = "after_login"
	HookBeforeUserCreate   HookPoint = "before_user_create"
	HookAfterUserCreate    HookPoint = "after_user_create"
	HookBeforeUserDelete   HookPoint = "before_user_delete"
	HookAfterTrafficReset  HookPoint = "after_traffic_reset"
	HookAfterInboundCreate HookPoint = "after_inbound_create"
	HookOnNodeHeartbeat    HookPoint = "on_node_heartbeat"
	HookOnTicketCreate     HookPoint = "on_ticket_create"
	HookDailyCron          HookPoint = "daily_cron"
	HookHourlyCron         HookPoint = "hourly_cron"
)

type PluginInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description,omitempty"`
	Author      string     `json:"author,omitempty"`
	Hooks       []HookPoint `json:"hooks"`
	Enabled     bool       `json:"enabled"`
	Path        string     `json:"path,omitempty"`
	Status      string     `json:"status"`
	LoadedAt    int64      `json:"loaded_at"`
	Error       string     `json:"error,omitempty"`
}

type PluginContext struct {
	PluginID string
	Data     map[string]any
}

type PluginInterface interface {
	Init(cfg map[string]any) error
	Execute(hook HookPoint, ctx *PluginContext) error
	Stop() error
}

type PluginService struct {
	plugins   map[string]*PluginInfo
	instances map[string]PluginInterface
	mu        sync.RWMutex
	hooks     map[HookPoint][]string
	pluginDir string
}

func NewPluginService(pluginDir string) *PluginService {
	return &PluginService{
		plugins:   make(map[string]*PluginInfo),
		instances: make(map[string]PluginInterface),
		hooks:     make(map[HookPoint][]string),
		pluginDir: pluginDir,
	}
}

func (s *PluginService) ListPlugins() []*PluginInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*PluginInfo, 0, len(s.plugins))
	for _, p := range s.plugins {
		list = append(list, p)
	}
	return list
}

func (s *PluginService) GetPlugin(id string) (*PluginInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.plugins[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("plugin not found: %s", id)
}

// LoadPlugin is defined in plugin_loader.go (Linux/Darwin only).
// On Windows, it returns a "not supported" error.
// See: plugin_stub.go for the Windows/other platform stub.

func (s *PluginService) UnloadPlugin(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inst, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("plugin not loaded: %s", id)
	}
	if err := inst.Stop(); err != nil {
		log.Printf("[Plugin] Stop error for %s: %v", id, err)
	}
	delete(s.instances, id)
	delete(s.plugins, id)
	for hook := range s.hooks {
		plugins := s.hooks[hook]
		for i, pid := range plugins {
			if pid == id {
				s.hooks[hook] = append(plugins[:i], plugins[i+1:]...)
				break
			}
		}
	}
	log.Printf("[Plugin] Unloaded: %s", id)
	return nil
}

func (s *PluginService) EnablePlugin(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.plugins[id]; ok {
		p.Enabled = enabled
		if enabled { p.Status = "loaded" } else { p.Status = "disabled" }
		return nil
	}
	return fmt.Errorf("plugin not found: %s", id)
}

func (s *PluginService) TriggerHook(hook HookPoint, ctx *PluginContext) {
	s.mu.RLock()
	pluginIDs := s.hooks[hook]
	instances := make(map[string]PluginInterface)
	for _, id := range pluginIDs {
		if p, ok := s.plugins[id]; ok && p.Enabled {
			if inst, ok2 := s.instances[id]; ok2 {
				instances[id] = inst
			}
		}
	}
	s.mu.RUnlock()
	for id, inst := range instances {
		func(pid string, pluginInst PluginInterface) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Plugin] Panic in %s hook %s: %v", pid, hook, r)
				}
			}()
			if err := pluginInst.Execute(hook, ctx); err != nil {
				log.Printf("[Plugin] Error in %s hook %s: %v", pid, hook, err)
			}
		}(id, inst)
	}
}

func (s *PluginService) RegisterStaticPlugin(id, name, version string, inst PluginInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	info := &PluginInfo{
		ID: id, Name: name, Version: version,
		Enabled: true, Status: "loaded",
		LoadedAt: time.Now().UnixMilli(),
	}
	s.plugins[id] = info
	s.instances[id] = inst
	defaultHooks := []HookPoint{HookAfterLogin, HookAfterUserCreate, HookHourlyCron}
	for _, h := range defaultHooks {
		s.hooks[h] = append(s.hooks[h], id)
	}
	info.Hooks = defaultHooks
	log.Printf("[Plugin] Static plugin registered: %s v%s", name, version)
}

func (s *PluginService) GetPluginDir() string {
	return s.pluginDir
}
