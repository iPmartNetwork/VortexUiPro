package mtproto

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SecretEntry defines one named FakeTLS secret served by mtg-multi.
type SecretEntry struct {
	Name        string `json:"name"`
	Secret      string `json:"secret"`
	AdTag       string `json:"ad_tag,omitempty"`
	QuotaBytes  int64  `json:"quota_bytes,omitempty"`
	ExpiresUnix int64  `json:"expires_unix,omitempty"`
}

// Instance defines the runtime config for one mtproto inbound.
type Instance struct {
	ID                    int64          `json:"id"`
	Tag                   string         `json:"tag"`
	Listen                string         `json:"listen"`
	Port                  int            `json:"port"`
	Secrets               []SecretEntry  `json:"secrets"`
	Debug                 bool           `json:"debug"`
	ProxyProtocolListener bool           `json:"proxy_protocol_listener"`
	PreferIP              string         `json:"prefer_ip,omitempty"`
	FrontingIP            string         `json:"fronting_ip,omitempty"`
	FrontingPort          int            `json:"fronting_port,omitempty"`
	ThrottleMaxConnections int           `json:"throttle_max_connections,omitempty"`
	RouteThroughXray      bool           `json:"route_through_xray,omitempty"`
	XrayRoutePort         int            `json:"xray_route_port,omitempty"`
	PublicIPv4            string         `json:"public_ipv4,omitempty"`
	PublicIPv6            string         `json:"public_ipv6,omitempty"`
}

func (inst Instance) bindTo() string {
	listen := inst.Listen
	if listen == "" {
		listen = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", listen, inst.Port)
}

func (inst Instance) structuralFP() string {
	parts := []string{
		inst.bindTo(),
		strconv.FormatBool(inst.Debug),
		strconv.FormatBool(inst.ProxyProtocolListener),
		inst.PreferIP,
		inst.FrontingIP,
		strconv.Itoa(inst.FrontingPort),
		strconv.Itoa(inst.ThrottleMaxConnections),
		strconv.FormatBool(inst.RouteThroughXray),
		strconv.Itoa(inst.XrayRoutePort),
		inst.PublicIPv4,
		inst.PublicIPv6,
	}
	return strings.Join(parts, "|")
}

func (inst Instance) secretsFP() string {
	pairs := make([]string, 0, len(inst.Secrets))
	for _, e := range inst.Secrets {
		pairs = append(pairs, fmt.Sprintf("%s=%s;tag=%s;q=%d;exp=%d", e.Name, e.Secret, e.AdTag, e.QuotaBytes, e.ExpiresUnix))
	}
	slices.Sort(pairs)
	return strings.Join(pairs, "|")
}

// Traffic is per-client byte delta from mtg-multi stats.
type Traffic struct {
	Tag   string `json:"tag"`
	Email string `json:"email"`
	Up    int64  `json:"up"`
	Down  int64  `json:"down"`
}

type clientCounters struct {
	up   int64
	down int64
}

type managed struct {
	proc         *Process
	tag          string
	structuralFP string
	secretsFP    string
	apiPort      int
	apiToken     string
	last         map[string]clientCounters
}

// Manager owns the set of running mtg-multi processes.
type Manager struct {
	mu      sync.Mutex
	procs   map[int64]*managed
	swept   bool
	binPath string
	dataDir string
}

// NewManager creates a new MTProto process manager.
func NewManager(binPath, dataDir string) *Manager {
	if binPath == "" {
		binPath = "/usr/local/bin/mtg-multi"
	}
	if dataDir == "" {
		dataDir = "/etc/vortex/mtproto"
	}
	return &Manager{
		procs:   make(map[int64]*managed),
		binPath: binPath,
		dataDir: dataDir,
	}
}

// Ensure starts or updates a mtg-multi process for the given instance.
func (m *Manager) Ensure(inst Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepOrphans()
	return m.ensureLocked(inst)
}

func (m *Manager) sweepOrphans() {
	if m.swept {
		return
	}
	m.swept = true
	if n := killStrayMtgProcesses(m.binPath); n > 0 {
		fmt.Printf("mtproto: terminated %d orphaned mtg process(es)\n", n)
	}
}

func (m *Manager) ensureLocked(inst Instance) error {
	structFP := inst.structuralFP()
	secFP := inst.secretsFP()

	if cur, ok := m.procs[inst.ID]; ok {
		switch ensureActionFor(cur.proc.IsRunning(), cur.structuralFP, cur.secretsFP, structFP, secFP) {
		case actionNoop:
			cur.tag = inst.Tag
			return nil
		case actionReload:
			if err := writeConfig(m.configPath(inst.ID), inst, cur.apiPort, cur.apiToken); err != nil {
				return err
			}
			if applySecrets(cur.apiPort, cur.apiToken, inst) {
				cur.tag = inst.Tag
				cur.secretsFP = secFP
				return nil
			}
			fmt.Printf("mtproto: live update unavailable for inbound %d, restarting\n", inst.ID)
			fallthrough
		case actionRestart:
			_ = cur.proc.Stop()
			delete(m.procs, inst.ID)
		}
	}

	apiPort, err := freeLocalPort()
	if err != nil {
		return err
	}
	apiToken, err := newAPIToken()
	if err != nil {
		return err
	}

	cfgPath := m.configPath(inst.ID)
	if err := os.MkdirAll(m.dataDir, 0750); err != nil {
		return err
	}
	if err := writeConfig(cfgPath, inst, apiPort, apiToken); err != nil {
		return err
	}

	proc := newProcess(m.binPath, cfgPath, fmt.Sprintf("inbound %d", inst.ID))
	if err := proc.Start(); err != nil {
		return err
	}

	m.procs[inst.ID] = &managed{
		proc:         proc,
		tag:          inst.Tag,
		structuralFP: structFP,
		secretsFP:    secFP,
		apiPort:      apiPort,
		apiToken:     apiToken,
		last:         make(map[string]clientCounters),
	}
	return nil
}

func (m *Manager) configPath(id int64) string {
	return filepath.Join(m.dataDir, fmt.Sprintf("mtg-%d.toml", id))
}

// Remove stops and forgets the mtg process for an inbound.
func (m *Manager) Remove(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur, ok := m.procs[id]; ok {
		_ = cur.proc.Stop()
		delete(m.procs, id)
		_ = os.Remove(m.configPath(id))
	}
}

// Reconcile drives the running set toward the desired instances.
func (m *Manager) Reconcile(desired []Instance) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepOrphans()

	want := make(map[int64]struct{}, len(desired))
	for _, inst := range desired {
		want[inst.ID] = struct{}{}
	}
	for id, cur := range m.procs {
		if _, ok := want[id]; !ok {
			_ = cur.proc.Stop()
			delete(m.procs, id)
			_ = os.Remove(m.configPath(id))
		}
	}
	for _, inst := range desired {
		if err := m.ensureLocked(inst); err != nil {
			fmt.Printf("mtproto: reconcile failed for inbound %d: %v\n", inst.ID, err)
		}
	}
}

// StopAll stops every managed mtg process.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, cur := range m.procs {
		_ = cur.proc.Stop()
		_ = os.Remove(m.configPath(id))
		delete(m.procs, id)
	}
}

// CollectTraffic scrapes each running mtg-multi /stats endpoint.
func (m *Manager) CollectTraffic() ([]Traffic, []string) {
	type snap struct {
		id       int64
		apiPort  int
		apiToken string
		tag      string
		last     map[string]clientCounters
	}

	m.mu.Lock()
	snaps := make([]snap, 0, len(m.procs))
	for id, cur := range m.procs {
		if cur.proc == nil || !cur.proc.IsRunning() {
			continue
		}
		lastCopy := make(map[string]clientCounters, len(cur.last))
		maps.Copy(lastCopy, cur.last)
		snaps = append(snaps, snap{id: id, apiPort: cur.apiPort, apiToken: cur.apiToken, tag: cur.tag, last: lastCopy})
	}
	m.mu.Unlock()

	var out []Traffic
	var online []string
	for _, s := range snaps {
		users, ok := scrapeStats(s.apiPort, s.apiToken)
		if !ok {
			continue
		}
		newLast := make(map[string]clientCounters, len(users))
		for email, u := range users {
			up := u.BytesIn
			down := u.BytesOut
			newLast[email] = clientCounters{up: up, down: down}
			if u.Connections > 0 {
				online = append(online, email)
			}
			prev, had := s.last[email]
			if !had {
				continue
			}
			du := up - prev.up
			dd := down - prev.down
			if du < 0 {
				du = 0
			}
			if dd < 0 {
				dd = 0
			}
			if du > 0 || dd > 0 {
				out = append(out, Traffic{Tag: s.tag, Email: email, Up: du, Down: dd})
			}
		}

		m.mu.Lock()
		if cur, ok := m.procs[s.id]; ok {
			cur.last = newLast
		}
		m.mu.Unlock()
	}
	return out, online
}

// HasRunning checks if any mtg processes are running.
func (m *Manager) HasRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cur := range m.procs {
		if cur.proc != nil && cur.proc.IsRunning() {
			return true
		}
	}
	return false
}

type action int

const (
	actionNoop action = iota
	actionReload
	actionRestart
)

func ensureActionFor(running bool, curStructFP, curSecretsFP, newStructFP, newSecretsFP string) action {
	if !running || curStructFP != newStructFP {
		return actionRestart
	}
	if curSecretsFP != newSecretsFP {
		return actionReload
	}
	return actionNoop
}

// renderConfig builds the mtg-multi TOML config.
func renderConfig(inst Instance, apiPort int, apiToken string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "bind-to = %q\n", inst.bindTo())
	if inst.Debug {
		b.WriteString("debug = true\n")
	}
	if inst.ProxyProtocolListener {
		b.WriteString("proxy-protocol-listener = true\n")
	}
	if inst.PreferIP != "" {
		fmt.Fprintf(&b, "prefer-ip = %q\n", inst.PreferIP)
	}
	fmt.Fprintf(&b, "api-bind-to = \"127.0.0.1:%d\"\n", apiPort)
	if apiToken != "" {
		fmt.Fprintf(&b, "api-token = %q\n", apiToken)
	}
	if inst.PublicIPv4 != "" {
		fmt.Fprintf(&b, "public-ipv4 = %q\n", inst.PublicIPv4)
	}
	if inst.PublicIPv6 != "" {
		fmt.Fprintf(&b, "public-ipv6 = %q\n", inst.PublicIPv6)
	}
	if inst.FrontingIP != "" || inst.FrontingPort > 0 {
		b.WriteString("\n[domain-fronting]\n")
		if inst.FrontingIP != "" {
			fmt.Fprintf(&b, "host = %q\n", inst.FrontingIP)
		}
		if inst.FrontingPort > 0 {
			fmt.Fprintf(&b, "port = %d\n", inst.FrontingPort)
		}
	}
	if inst.RouteThroughXray && inst.XrayRoutePort > 0 {
		fmt.Fprintf(&b, "\n[network]\nproxies = [\"socks5://127.0.0.1:%d\"]\n", inst.XrayRoutePort)
	}
	if inst.ThrottleMaxConnections > 0 {
		fmt.Fprintf(&b, "\n[throttle]\nmax-connections = %d\n", inst.ThrottleMaxConnections)
	}

	// Secret ad tags (only for clients that have them)
	tagged := false
	for _, e := range inst.Secrets {
		if e.AdTag == "" {
			continue
		}
		if !tagged {
			b.WriteString("\n[secret-ad-tags]\n")
			tagged = true
		}
		fmt.Fprintf(&b, "%q = %q\n", e.Name, e.AdTag)
	}

	// Secret limits (quota/expiry per client)
	for _, e := range inst.Secrets {
		if e.QuotaBytes <= 0 && e.ExpiresUnix <= 0 {
			continue
		}
		fmt.Fprintf(&b, "\n[secret-limits.%q]\n", e.Name)
		if e.QuotaBytes > 0 {
			fmt.Fprintf(&b, "quota = %q\n", strconv.FormatInt(e.QuotaBytes, 10)+"B")
		}
		if e.ExpiresUnix > 0 {
			fmt.Fprintf(&b, "expires = %q\n", time.Unix(e.ExpiresUnix, 0).UTC().Format(time.RFC3339))
		}
	}

	// Secrets section
	b.WriteString("\n[secrets]\n")
	for _, e := range inst.Secrets {
		fmt.Fprintf(&b, "%q = %q\n", e.Name, e.Secret)
	}

	return b.String()
}

func writeConfig(path string, inst Instance, apiPort int, apiToken string) error {
	return os.WriteFile(path, []byte(renderConfig(inst, apiPort, apiToken)), 0640)
}

func applySecrets(port int, token string, inst Instance) bool {
	type secretPutEntry struct {
		Secret  string `json:"secret"`
		AdTag   string `json:"ad_tag,omitempty"`
		Quota   string `json:"quota,omitempty"`
		Expires string `json:"expires,omitempty"`
	}

	secrets := make(map[string]secretPutEntry, len(inst.Secrets))
	for _, e := range inst.Secrets {
		entry := secretPutEntry{Secret: e.Secret, AdTag: e.AdTag}
		if e.QuotaBytes > 0 {
			entry.Quota = strconv.FormatInt(e.QuotaBytes, 10) + "B"
		}
		if e.ExpiresUnix > 0 {
			entry.Expires = time.Unix(e.ExpiresUnix, 0).UTC().Format(time.RFC3339)
		}
		secrets[e.Name] = entry
	}

	body, _ := json.Marshal(map[string]any{"secrets": secrets})
	client := http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut,
		fmt.Sprintf("http://127.0.0.1:%d/secrets", port), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

type statsUser struct {
	Connections int64 `json:"connections"`
	BytesIn     int64 `json:"bytes_in"`
	BytesOut    int64 `json:"bytes_out"`
}

func scrapeStats(port int, token string) (map[string]statsUser, bool) {
	client := http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
		fmt.Sprintf("http://127.0.0.1:%d/stats", port), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	var parsed struct {
		Users map[string]statsUser `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, false
	}
	return parsed.Users, true
}

func newAPIToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func freeLocalPort() (int, error) {
	l, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
