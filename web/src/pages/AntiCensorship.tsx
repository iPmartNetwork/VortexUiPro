import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '../api/client'

// ─── Types ───────────────────────────────────────────────────────────

interface TLSTrick {
  name: string
  description: string
  config: Record<string, any>
  enabled: boolean
}

interface TLSFingerprint {
  name: string
  client: string
  security: string
}

interface RealityScanResult {
  target: string
  port: number
  reachable: boolean
  latency_ms: number
  tls_version: string
  server_name: string
  error?: string
}

interface GeneratedCert {
  certificate: string
  private_key: string
  domain: string
}

interface AntiDPIConfig {
  fragment: Record<string, any>
  padding: Record<string, any>
  tls_fingerprint: string
  allow_insecure: boolean
  mix_https_target: string
}

interface MTProtoConfig {
  enabled: boolean
  port: number
  secret: string
  fake_tls: boolean
  fake_tls_domain: string
  tag: string
}

interface WarpConfig {
  enabled: boolean
  mode: string
  wireguard_address_v4: string
  wireguard_address_v6: string
  endpoint: string
}

// ─── Component ───────────────────────────────────────────────────────

export function AntiCensorshipPage() {
  const [activeTab, setActiveTab] = useState('tricks')
  const [tricks, setTricks] = useState<TLSTrick[]>([])
  const [fingerprints, setFingerprints] = useState<TLSFingerprint[]>([])
  const [scanning, setScanning] = useState(false)
  const [scanTarget, setScanTarget] = useState('')
  const [scanResult, setScanResult] = useState<RealityScanResult | null>(null)
  const [certDomain, setCertDomain] = useState('vortex.local')
  const [certificate, setCertificate] = useState<GeneratedCert | null>(null)
  const [decoyDomain, setDecoyDomain] = useState('')
  const [decoyConfig, setDecoyConfig] = useState<any>(null)
  const [genConfig, setGenConfig] = useState<any>(null)
  const [genType, setGenType] = useState<'fragment' | 'padding' | 'mix'>('fragment')
  const [copied, setCopied] = useState(false)

  // AntiDPI
  const [antiDPIConfig, setAntiDPIConfig] = useState<AntiDPIConfig | null>(null)
  const [antiDPITransport, setAntiDPITransport] = useState('tcp')
  const [loadingAntiDPI, setLoadingAntiDPI] = useState(false)

  // MTProto
  const [mtprotoConfig, setMTProtoConfig] = useState<MTProtoConfig | null>(null)
  const [loadingMTProto, setLoadingMTProto] = useState(false)

  // WARP
  const [warpConfig, setWarpConfig] = useState<WarpConfig | null>(null)
  const [loadingWARP, setLoadingWARP] = useState(false)

  // Fetch tricks & fingerprints on mount
  useEffect(() => {
    apiClient.get('/api/v1/anticensor/tricks').then(r => setTricks(r.data.tricks || [])).catch(() => {})
    apiClient.get('/api/v1/anticensor/fingerprints').then(r => setFingerprints(r.data.fingerprints || [])).catch(() => {})
  }, [])

  const handleScan = useCallback(async () => {
    if (!scanTarget) return
    setScanning(true)
    setScanResult(null)
    try {
      const res = await apiClient.get('/api/v1/anticensor/scan', { params: { target: scanTarget } })
      setScanResult(res.data)
    } catch { setScanResult({ target: scanTarget, port: 443, reachable: false, latency_ms: 0, tls_version: '', server_name: '', error: 'Scan failed' }) }
    finally { setScanning(false) }
  }, [scanTarget])

  const handleGenerateCert = useCallback(async () => {
    try {
      const res = await apiClient.get('/api/v1/anticensor/cert', { params: { domain: certDomain } })
      setCertificate(res.data)
    } catch { setCertificate(null) }
  }, [certDomain])

  const handleGenerateDecoy = useCallback(async () => {
    try {
      const res = await apiClient.get('/api/v1/anticensor/decoy', { params: { domain: decoyDomain || undefined } })
      setDecoyConfig(res.data)
    } catch { setDecoyConfig(null) }
  }, [decoyDomain])

  const handleGenerateConfig = useCallback(async (type: 'fragment' | 'padding' | 'mix') => {
    setGenType(type)
    try {
      const endpoint = type === 'fragment' ? '/api/v1/anticensor/fragment' 
        : type === 'padding' ? '/api/v1/anticensor/padding' 
        : '/api/v1/anticensor/mix'
      const res = await apiClient.get(endpoint)
      setGenConfig(res.data.config)
    } catch { setGenConfig(null) }
  }, [])

  const handleGenerateAntiDPI = useCallback(async () => {
    setLoadingAntiDPI(true)
    try {
      const res = await apiClient.get('/api/v1/anticensor/anti-dpi', { params: { transport: antiDPITransport } })
      setAntiDPIConfig(res.data)
    } catch { setAntiDPIConfig(null) }
    finally { setLoadingAntiDPI(false) }
  }, [antiDPITransport])

  const handleGenerateMTProto = useCallback(async () => {
    setLoadingMTProto(true)
    try {
      const res = await apiClient.get('/api/v1/anticensor/mtproto')
      setMTProtoConfig(res.data)
    } catch { setMTProtoConfig(null) }
    finally { setLoadingMTProto(false) }
  }, [])

  const handleGenerateWARP = useCallback(async () => {
    setLoadingWARP(true)
    try {
      const res = await apiClient.get('/api/v1/anticensor/warp')
      setWarpConfig(res.data)
    } catch { setWarpConfig(null) }
    finally { setLoadingWARP(false) }
  }, [])

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const tabs = [
    { id: 'tricks', label: 'TLS Tricks', icon: '🛡️' },
    { id: 'fingerprints', label: 'Fingerprints', icon: '🔑' },
    { id: 'reality', label: 'Reality Scanner', icon: '📡' },
    { id: 'cert', label: 'SSL Cert', icon: '🔐' },
    { id: 'decoy', label: 'Decoy', icon: '🎭' },
    { id: 'config', label: 'Config Gen', icon: '⚙️' },
    { id: 'antidpi', label: 'Anti-DPI', icon: '🧬' },
    { id: 'mtproto', label: 'MTProto', icon: '📱' },
    { id: 'warp', label: 'WARP', icon: '☁️' },
  ]

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-[var(--text-primary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Anti-Censorship Suite</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Advanced tools to bypass internet censorship and DPI</p>
            </div>
          </div>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-[rgba(139,92,246,0.1)] border border-[rgba(139,92,246,0.2)]">
            <span className="w-2 h-2 rounded-full bg-[var(--success)] animate-pulse" />
            <span className="text-xs text-[var(--accent-light)] font-semibold">ACTIVE</span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap gap-2">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              activeTab === tab.id
                ? 'bg-gradient-to-r from-purple-600 to-cyan-500 text-[var(--text-primary)] shadow-lg shadow-purple-600/20'
                : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] hover:text-[var(--text-primary)] border border-[var(--border-light)]'
            }`}
          >
            <span>{tab.icon}</span>
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      {/* ─── TLS Tricks Tab ─────────────────────────────────────────── */}
      {activeTab === 'tricks' && (
        <div className="space-y-4">
          <p className="text-sm text-[var(--text-muted)]">Toggle anti-censorship TLS obfuscation techniques. These techniques help bypass Deep Packet Inspection (DPI) in restricted networks.</p>
          <p className="text-sm text-[var(--text-muted)] mb-4">Toggle anti-censorship TLS obfuscation techniques. These techniques help bypass Deep Packet Inspection (DPI) in restricted networks.</p>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {tricks.map(trick => (
              <div key={trick.name} className="glass-card p-5 hover:border-purple-500/30 transition-all group">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-[var(--text-primary)] font-semibold text-sm uppercase tracking-wide">{trick.name.replace(/_/g, ' ')}</h3>
                      <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${
                        trick.enabled ? 'bg-green-500/10 text-green-400' : 'bg-[var(--bg-elevated)] text-[var(--text-muted)]'
                      }`}>
                        {trick.enabled ? 'ON' : 'OFF'}
                      </span>
                    </div>
                    <p className="text-[var(--text-secondary)] text-xs leading-relaxed">{trick.description}</p>
                    {trick.config && Object.keys(trick.config).length > 0 && (
                      <div className="mt-3 flex flex-wrap gap-1.5">
                        {Object.entries(trick.config).map(([key, val]) => (
                          <span key={key} className="px-2 py-0.5 rounded bg-[var(--bg-elevated)] text-[var(--text-secondary)] text-[10px] font-mono border border-[var(--border-light)]">
                            {key}: <span className="text-[var(--accent-light)]">{typeof val === 'boolean' ? (val ? '✓' : '✗') : String(val)}</span>
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer shrink-0">
                    <input type="checkbox" checked={trick.enabled} readOnly className="sr-only peer" />
                    <div className="w-10 h-5 bg-[var(--bg-elevated)] rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-purple-600 group-hover:peer-checked:bg-purple-500 transition-colors"></div>
                  </label>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ─── Fingerprints Tab ───────────────────────────────────────── */}
      {activeTab === 'fingerprints' && (
        <div className="glass-card p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-8 h-8 rounded-lg bg-[rgba(139,92,246,0.1)] border border-[rgba(139,92,246,0.1)] flex items-center justify-center">
              <span className="text-sm">🔑</span>
            </div>
            <h3 className="text-base font-bold text-[var(--text-primary)]">Available TLS Fingerprints</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="table-modern">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Client</th>
                  <th>Security</th>
                </tr>
              </thead>
              <tbody>
                {fingerprints.map(fp => (
                  <tr key={fp.name}>
                    <td className="text-sm font-medium text-[var(--text-primary)]">{fp.name}</td>
                    <td><span className="badge badge-purple">{fp.client}</span></td>
                    <td><span className="badge badge-cyan">{fp.security}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="text-xs text-[var(--text-muted)] mt-4">💡 Use different fingerprints to avoid TLS fingerprint-based blocking. Chrome fingerprints work best in most environments.</p>
        </div>
      )}

      {/* ─── Reality Scanner Tab ────────────────────────────────────── */}
      {activeTab === 'reality' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">📡 Reality Target Scanner</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Scan a target host to check if it's reachable and REALITY-compatible. Results show TLS version, latency, and server name.</p>
            <div className="flex gap-3">
              <input
                type="text"
                value={scanTarget}
                onChange={e => setScanTarget(e.target.value)}
                placeholder="e.g. www.yahoo.com"
                className="flex-1 px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
                onKeyDown={e => e.key === 'Enter' && handleScan()}
              />
              <button
                onClick={handleScan}
                disabled={scanning || !scanTarget}
                className="px-5 py-2.5 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 disabled:opacity-50 transition text-sm font-medium whitespace-nowrap"
              >
                {scanning ? (
                  <span className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Scanning...
                  </span>
                ) : 'Scan'}
              </button>
            </div>
          </div>

          {scanResult && (
            <div className="glass-card p-6">
              <h3 className="text-[var(--text-primary)] font-semibold mb-4">📊 Scan Results</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                  <span className="text-[var(--text-secondary)] text-sm">Target</span>
                  <span className="text-[var(--text-primary)] text-sm font-mono">{scanResult.target}:{scanResult.port}</span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                  <span className="text-[var(--text-secondary)] text-sm">Reachable</span>
                  <span className={`flex items-center gap-1.5 text-sm font-medium ${scanResult.reachable ? 'text-green-400' : 'text-red-400'}`}>
                    <span className={`w-2 h-2 rounded-full ${scanResult.reachable ? 'bg-green-500' : 'bg-red-500'}`} />
                    {scanResult.reachable ? 'YES' : 'NO'}
                  </span>
                </div>
                {scanResult.reachable && (
                  <>
                    <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                      <span className="text-[var(--text-secondary)] text-sm">Latency</span>
                      <span className="text-purple-300 text-sm font-mono">{scanResult.latency_ms}ms</span>
                    </div>
                    <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                      <span className="text-[var(--text-secondary)] text-sm">TLS Version</span>
                      <span className="text-blue-300 text-sm font-mono">{scanResult.tls_version}</span>
                    </div>
                    <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                      <span className="text-[var(--text-secondary)] text-sm">Server Name</span>
                      <span className="text-[var(--text-primary)] text-sm font-mono">{scanResult.server_name || 'N/A'}</span>
                    </div>
                  </>
                )}
                {scanResult.error && (
                  <div className="p-3 rounded-lg bg-red-500/5 border border-red-500/20">
                    <span className="text-red-400 text-xs">{scanResult.error}</span>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ─── SSL Cert Tab ───────────────────────────────────────────── */}
      {activeTab === 'cert' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">🔐 Generate Self-Signed Certificate</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Create a self-signed TLS certificate for testing and development. You can use this certificate to enable TLS on your proxy inbounds.</p>
            <div className="flex gap-3 mb-4">
              <input
                type="text"
                value={certDomain}
                onChange={e => setCertDomain(e.target.value)}
                placeholder="Domain (e.g. vortex.local)"
                className="flex-1 px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
              />
              <button
                onClick={handleGenerateCert}
                className="px-5 py-2.5 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 transition text-sm font-medium"
              >
                Generate
              </button>
            </div>
            <div className="p-3 rounded-lg bg-yellow-500/5 border border-yellow-500/20">
              <p className="text-xs text-yellow-400">⚠️ Self-signed certificates are not trusted by browsers. Use a proper CA (e.g., Let's Encrypt) for production.</p>
            </div>
          </div>

          {certificate && (
            <div className="glass-card p-6">
              <h3 className="text-[var(--text-primary)] font-semibold mb-4">📜 Generated Certificate</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                  <span className="text-[var(--text-secondary)] text-sm">Domain</span>
                  <span className="text-purple-300 text-sm font-mono">{certificate.domain}</span>
                </div>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-[var(--text-secondary)] text-xs font-medium uppercase">Certificate (PEM)</span>
                    <button onClick={() => copyToClipboard(certificate.certificate)} className="text-xs text-purple-400 hover:text-purple-300 transition">
                      {copied ? '✅ Copied!' : '📋 Copy'}
                    </button>
                  </div>
                  <pre className="p-3 rounded-lg bg-[var(--bg-deep)] text-green-300 text-[10px] font-mono overflow-x-auto max-h-32 border border-[var(--border-light)]">{certificate.certificate.slice(0, 500)}...</pre>
                </div>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-[var(--text-secondary)] text-xs font-medium uppercase">Private Key (PEM)</span>
                    <button onClick={() => copyToClipboard(certificate.private_key)} className="text-xs text-purple-400 hover:text-purple-300 transition">📋 Copy</button>
                  </div>
                  <pre className="p-3 rounded-lg bg-[var(--bg-deep)] text-yellow-300 text-[10px] font-mono overflow-x-auto max-h-32 border border-[var(--border-light)]">{certificate.private_key.slice(0, 500)}...</pre>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* ─── Decoy Config Tab ───────────────────────────────────────── */}
      {activeTab === 'decoy' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">🎭 Decoy Site Configuration</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Generate a decoy/masquerade configuration to blend proxy traffic with legitimate HTTPS traffic using REALITY technology.</p>
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1.5">Decoy Domain</label>
                <input
                  type="text"
                  value={decoyDomain}
                  onChange={e => setDecoyDomain(e.target.value)}
                  placeholder="cdn.example.com"
                  className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
                />
              </div>
              <button
                onClick={handleGenerateDecoy}
                className="px-5 py-2.5 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 transition text-sm font-medium"
              >
                Generate Decoy Config
              </button>
            </div>
          </div>

          {decoyConfig && (
            <div className="glass-card p-6">
              <h3 className="text-[var(--text-primary)] font-semibold mb-4">📋 Decoy Configuration</h3>
              <div className="space-y-3">
                {Object.entries(decoyConfig).map(([key, val]) => (
                  <div key={key} className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                    <span className="text-[var(--text-secondary)] text-xs uppercase tracking-wide">{key.replace(/_/g, ' ')}</span>
                    <span className="text-[var(--text-primary)] text-sm font-mono">
                      {typeof val === 'boolean' ? (val ? '✓' : '✗') : String(val)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ─── Config Generator Tab ───────────────────────────────────── */}
      {activeTab === 'config' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">⚙️ Stream Config Generator</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Generate ready-to-use xray stream settings for anti-censorship techniques. Choose fragmentation, padding, or mixed HTTPS configuration.</p>
            <div className="flex gap-2 mb-4">
              {(['fragment', 'padding', 'mix'] as const).map(type => (
                <button
                  key={type}
                  onClick={() => handleGenerateConfig(type)}
                  className={`flex-1 px-4 py-2.5 rounded-lg text-sm font-medium transition-all ${
                    genType === type
                      ? 'bg-purple-600 text-[var(--text-primary)]'
                      : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] border border-[var(--border-light)]'
                  }`}
                >
                  {type === 'fragment' ? 'Fragment' : type === 'padding' ? 'Padding' : 'Mix HTTPS'}
                </button>
              ))}
            </div>
            <button
              onClick={() => handleGenerateConfig(genType)}
              className="px-5 py-2.5 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 transition text-sm font-medium"
            >
              Generate {genType === 'fragment' ? 'Fragment' : genType === 'padding' ? 'Padding' : 'Mix'} Config
            </button>
          </div>

          {genConfig && (
            <div className="glass-card p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-[var(--text-primary)] font-semibold">📄 Generated Config</h3>
                <button onClick={() => copyToClipboard(JSON.stringify(genConfig, null, 2))} className="text-xs text-purple-400 hover:text-purple-300 transition">
                  {copied ? '✅ Copied!' : '📋 Copy JSON'}
                </button>
              </div>
              <pre className="p-4 rounded-lg bg-[var(--bg-deep)] text-green-300 text-xs font-mono overflow-x-auto max-h-80 border border-[var(--border-light)]">
                {typeof genConfig === 'string' ? genConfig : JSON.stringify(genConfig, null, 2)}
              </pre>
              <div className="mt-4 p-3 rounded-lg bg-blue-500/5 border border-blue-500/20">
                <p className="text-xs text-blue-300">💡 Paste this into the <span className="font-mono">streamSettings</span> section of your xray inbound/outbound configuration.</p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* ─── Anti-DPI Tab ───────────────────────────────────────────── */}
      {activeTab === 'antidpi' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">🧬 Anti-DPI Bundle</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Generate a bundled anti-DPI configuration combining fragment, padding, TLS fingerprint, and mix HTTPS settings. One-click config to bypass Deep Packet Inspection.</p>
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1.5">Transport Protocol</label>
                <div className="flex gap-2">
                  {['tcp', 'grpc', 'websocket'].map(t => (
                    <button
                      key={t}
                      onClick={() => setAntiDPITransport(t)}
                      className={`flex-1 px-4 py-2.5 rounded-lg text-sm font-medium transition-all ${
                        antiDPITransport === t
                          ? 'bg-purple-600 text-[var(--text-primary)]'
                          : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] border border-[var(--border-light)]'
                      }`}
                    >
                      {t.toUpperCase()}
                    </button>
                  ))}
                </div>
              </div>
              <button
                onClick={handleGenerateAntiDPI}
                disabled={loadingAntiDPI}
                className="w-full px-5 py-3 bg-gradient-to-r from-purple-600 to-pink-600 text-[var(--text-primary)] rounded-lg hover:from-purple-500 hover:to-pink-500 disabled:opacity-50 transition text-sm font-medium"
              >
                {loadingAntiDPI ? (
                  <span className="flex items-center justify-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Generating...
                  </span>
                ) : '🧬 Generate Anti-DPI Bundle'}
              </button>
            </div>
            <div className="mt-4 p-3 rounded-lg bg-green-500/5 border border-green-500/20">
              <p className="text-xs text-green-300">✓ Combines Fragment + Padding + TLS Fingerprint + Mix HTTPS in one config. Works with xray and sing-box.</p>
            </div>
          </div>

          {antiDPIConfig && (
            <div className="glass-card p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-[var(--text-primary)] font-semibold">📄 Anti-DPI Config</h3>
                <button onClick={() => copyToClipboard(JSON.stringify(antiDPIConfig, null, 2))} className="text-xs text-purple-400 hover:text-purple-300 transition">
                  {copied ? '✅ Copied!' : '📋 Copy JSON'}
                </button>
              </div>
              <div className="space-y-3 mb-4">
                <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                  <span className="text-[var(--text-secondary)] text-sm">TLS Fingerprint</span>
                  <span className="text-purple-300 text-sm font-mono">{antiDPIConfig.tls_fingerprint}</span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                  <span className="text-[var(--text-secondary)] text-sm">Allow Insecure</span>
                  <span className={`text-sm font-medium ${antiDPIConfig.allow_insecure ? 'text-yellow-400' : 'text-green-400'}`}>
                    {antiDPIConfig.allow_insecure ? 'YES' : 'NO'}
                  </span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                  <span className="text-[var(--text-secondary)] text-sm">Mix HTTPS Target</span>
                  <span className="text-[var(--text-primary)] text-sm font-mono">{antiDPIConfig.mix_https_target}</span>
                </div>
              </div>
              <pre className="p-4 rounded-lg bg-[var(--bg-deep)] text-green-300 text-xs font-mono overflow-x-auto max-h-64 border border-[var(--border-light)]">
                {JSON.stringify(antiDPIConfig, null, 2)}
              </pre>
            </div>
          )}
        </div>
      )}

      {/* ─── MTProto Tab ────────────────────────────────────────────── */}
      {activeTab === 'mtproto' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">📱 MTProto Proxy Config</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Generate a Telegram MTProto proxy configuration. MTProto proxies help users connect to Telegram in restricted regions using the official Telegram apps.</p>
            <ul className="space-y-2 mb-4">
              {[
                'No additional software needed on client side',
                'Works with official Telegram apps',
                'Supports Fake TLS for traffic obfuscation',
                'High performance and low latency',
              ].map((tip, i) => (
                <li key={i} className="flex items-start gap-2 text-xs text-[var(--text-secondary)]">
                  <span className="text-green-400 mt-0.5">✓</span>
                  <span>{tip}</span>
                </li>
              ))}
            </ul>
            <button
              onClick={handleGenerateMTProto}
              disabled={loadingMTProto}
              className="w-full px-5 py-3 bg-gradient-to-r from-blue-600 to-cyan-600 text-[var(--text-primary)] rounded-lg hover:from-blue-500 hover:to-cyan-500 disabled:opacity-50 transition text-sm font-medium"
            >
              {loadingMTProto ? (
                <span className="flex items-center justify-center gap-2">
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Generating...
                </span>
              ) : '📱 Generate MTProto Config'}
            </button>
          </div>

          {mtprotoConfig && (
            <div className="glass-card p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-[var(--text-primary)] font-semibold">📋 MTProto Configuration</h3>
                <button onClick={() => copyToClipboard(JSON.stringify(mtprotoConfig, null, 2))} className="text-xs text-purple-400 hover:text-purple-300 transition">
                  {copied ? '✅ Copied!' : '📋 Copy JSON'}
                </button>
              </div>
              <div className="space-y-3">
                {Object.entries(mtprotoConfig).map(([key, val]) => (
                  <div key={key} className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                    <span className="text-[var(--text-secondary)] text-xs uppercase tracking-wide">{key.replace(/_/g, ' ')}</span>
                    <span className="text-[var(--text-primary)] text-sm font-mono">
                      {typeof val === 'boolean' ? (val ? '✓' : '✗') : String(val)}
                    </span>
                  </div>
                ))}
              </div>
              <div className="mt-4 p-3 rounded-lg bg-blue-500/5 border border-blue-500/20">
                <p className="text-xs text-blue-300">💡 Share the proxy with <span className="font-mono">t.me/proxy?server=YOUR_IP&port={mtprotoConfig.port}&secret={mtprotoConfig.secret}</span></p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* ─── WARP Tab ───────────────────────────────────────────────── */}
      {activeTab === 'warp' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6">
            <h3 className="text-[var(--text-primary)] font-semibold mb-4">☁️ Cloudflare WARP Integration</h3>
            <p className="text-xs text-[var(--text-muted)] mb-4">Configure Cloudflare WARP to route proxy traffic through Cloudflare's network. This masks your server IP and provides DDoS protection.</p>
            <ul className="space-y-2 mb-4">
              {[
                'Routes traffic through Cloudflare global network',
                'Hides your server IP address (proxy traffic origin)',
                'Includes DDoS protection',
                'Works with xray and sing-box routing',
              ].map((tip, i) => (
                <li key={i} className="flex items-start gap-2 text-xs text-[var(--text-secondary)]">
                  <span className="text-green-400 mt-0.5">✓</span>
                  <span>{tip}</span>
                </li>
              ))}
            </ul>
            <button
              onClick={handleGenerateWARP}
              disabled={loadingWARP}
              className="w-full px-5 py-3 bg-gradient-to-r from-orange-600 to-yellow-600 text-[var(--text-primary)] rounded-lg hover:from-orange-500 hover:to-yellow-500 disabled:opacity-50 transition text-sm font-medium"
            >
              {loadingWARP ? (
                <span className="flex items-center justify-center gap-2">
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Generating...
                </span>
              ) : '☁️ Generate WARP Config'}
            </button>
          </div>

          {warpConfig && (
            <div className="glass-card p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-[var(--text-primary)] font-semibold">📋 WARP Configuration</h3>
                <button onClick={() => copyToClipboard(JSON.stringify(warpConfig, null, 2))} className="text-xs text-purple-400 hover:text-purple-300 transition">
                  {copied ? '✅ Copied!' : '📋 Copy JSON'}
                </button>
              </div>
              <div className="space-y-3">
                {Object.entries(warpConfig).map(([key, val]) => (
                  <div key={key} className="flex items-center justify-between p-3 rounded-lg bg-[var(--bg-elevated)]/50">
                    <span className="text-[var(--text-secondary)] text-xs uppercase tracking-wide">{key.replace(/_/g, ' ')}</span>
                    <span className="text-[var(--text-primary)] text-sm font-mono">
                      {typeof val === 'boolean' ? (val ? '✓' : '✗') : String(val)}
                    </span>
                  </div>
                ))}
              </div>
              <div className="mt-4 p-3 rounded-lg bg-yellow-500/5 border border-yellow-500/20">
                <p className="text-xs text-yellow-400">⚠️ WARP requires a Cloudflare WARP Zero Trust account and valid WireGuard keys. Replace the placeholder keys with your actual WARP credentials.</p>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
