import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '../api/client'

interface KnownProvider {
  provider: string
  domain: string
  ip: string
}

interface ScanResult {
  domain: string
  provider: string
  frontable: boolean
  reachable: boolean
  latency_ms: number
  tls_version: string
  server_name: string
  error?: string
}

interface CDNDomain {
  id: number
  domain: string
  cdn_provider: string
  status: string
  reachable: boolean
  latency_ms: number
  frontable: boolean
  last_checked: number
  created_at: number
}

interface ProxyConfig {
  provider: string
  front_domain: string
  hidden_domain: string
  port: number
  tls: boolean
  sni: string
  host_header: string
  xray_outbound: any
  singbox_outbound: any
}

export function DomainFrontingPage() {
  const [tab, setTab] = useState<'scan' | 'config' | 'domains'>('scan')
  const [providers, setProviders] = useState<KnownProvider[]>([])
  const [scanDomains, setScanDomains] = useState<CDNDomain[]>([])
  const [scanTarget, setScanTarget] = useState('')
  const [scanResult, setScanResult] = useState<ScanResult | null>(null)
  const [scanning, setScanning] = useState(false)
  const [scanAllLoading, setScanAllLoading] = useState(false)
  const [frontDomain, setFrontDomain] = useState('')
  const [hiddenDomain, setHiddenDomain] = useState('')
  const [providerName, setProviderName] = useState('cloudflare')
  const [proxyConfig, setProxyConfig] = useState<ProxyConfig | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    apiClient.get('/api/v1/domain-fronting/providers').then(r => setProviders(r.data.data || [])).catch(() => {})
    apiClient.get('/api/v1/domain-fronting/domains').then(r => setScanDomains(r.data.data || [])).catch(() => {})
  }, [])

  const handleScan = useCallback(async () => {
    if (!scanTarget) return
    setScanning(true)
    setScanResult(null)
    try {
      const res = await apiClient.get('/api/v1/domain-fronting/scan', { params: { domain: scanTarget } })
      setScanResult(res.data.data)
      // Refresh domain list
      const d = await apiClient.get('/api/v1/domain-fronting/domains')
      setScanDomains(d.data.data || [])
    } catch { setScanResult({ domain: scanTarget, provider: '', frontable: false, reachable: false, latency_ms: 0, tls_version: '', server_name: '', error: 'Scan failed' }) }
    finally { setScanning(false) }
  }, [scanTarget])

  const handleScanAll = useCallback(async () => {
    setScanAllLoading(true)
    try {
      const res = await apiClient.post('/api/v1/domain-fronting/scan-all')
      const d = await apiClient.get('/api/v1/domain-fronting/domains')
      setScanDomains(d.data.data || [])
    } catch {}
    finally { setScanAllLoading(false) }
  }, [])

  const handleGenerateConfig = useCallback(async () => {
    if (!frontDomain) return
    try {
      const res = await apiClient.get('/api/v1/domain-fronting/generate-config', {
        params: { front_domain: frontDomain, hidden_domain: hiddenDomain || 'your-server.com', provider: providerName }
      })
      setProxyConfig(res.data.data)
      setTab('config')
    } catch {}
  }, [frontDomain, hiddenDomain, providerName])

  const copyConfig = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-6 page-enter">
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-orange-500 to-red-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-white">Domain Fronting + CDN Proxy</h1>
            <p className="text-sm text-[#6868a0] mt-0.5">Discover frontable CDN domains & generate proxy configs</p>
          </div>
        </div>
      </div>

      <div className="flex gap-1 p-1 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] w-fit">
        <button onClick={() => setTab('scan')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'scan' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Scan Domains</button>
        <button onClick={() => setTab('config')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'config' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Proxy Config</button>
        <button onClick={() => setTab('domains')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'domains' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Scanned Domains</button>
      </div>

      {tab === 'scan' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">📡 Scan Domain</h3>
            <p className="text-xs text-[#6868a0] mb-4">Check if a CDN domain supports domain fronting by analyzing its TLS properties.</p>
            <div className="flex gap-3">
              <input value={scanTarget} onChange={e => setScanTarget(e.target.value)} placeholder="e.g., cloudflare.com" className="flex-1 px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" onKeyDown={e => e.key === 'Enter' && handleScan()} />
              <button onClick={handleScan} disabled={scanning || !scanTarget} className="px-5 py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 transition text-sm font-medium">{scanning ? 'Scanning...' : 'Scan'}</button>
            </div>
            <div className="mt-4 flex gap-2">
              <button onClick={handleScanAll} disabled={scanAllLoading} className="px-4 py-2 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-[#9898b8] rounded-lg hover:text-white transition text-xs font-medium">
                {scanAllLoading ? 'Scanning all...' : '🔄 Scan All Known Domains'}
              </button>
            </div>
          </div>
          {scanResult && (
            <div className="glass-card p-5">
              <h3 className="text-white font-semibold mb-4">📊 Result for {scanResult.domain}</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                  <span className="text-[#9898b8] text-sm">Frontable</span>
                  <span className={`text-sm font-bold ${scanResult.frontable ? 'text-emerald-400' : 'text-red-400'}`}>{scanResult.frontable ? '✅ YES' : '❌ NO'}</span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                  <span className="text-[#9898b8] text-sm">Reachable</span>
                  <span className={`text-sm ${scanResult.reachable ? 'text-emerald-400' : 'text-red-400'}`}>{scanResult.reachable ? '✓' : '✗'}</span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                  <span className="text-[#9898b8] text-sm">Provider</span>
                  <span className="text-white text-sm">{scanResult.provider || 'Unknown'}</span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                  <span className="text-[#9898b8] text-sm">Latency</span>
                  <span className="text-white text-sm">{scanResult.latency_ms}ms</span>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                  <span className="text-[#9898b8] text-sm">TLS</span>
                  <span className="text-white text-sm">{scanResult.tls_version}</span>
                </div>
                {scanResult.error && <div className="p-3 rounded-lg bg-red-500/10 text-red-400 text-xs">{scanResult.error}</div>}
              </div>
            </div>
          )}
        </div>
      )}

      {tab === 'config' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">⚙️ Generate Proxy Config</h3>
            <p className="text-xs text-[#6868a0] mb-4">Create a CDN-fronted xray/sing-box proxy configuration.</p>
            <div className="space-y-3">
              <div>
                <label className="block text-xs text-[#9898b8] mb-1">Front Domain (CDN)</label>
                <input value={frontDomain} onChange={e => setFrontDomain(e.target.value)} placeholder="cdn.cloudflare.com" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              </div>
              <div>
                <label className="block text-xs text-[#9898b8] mb-1">Hidden Domain (Your Server)</label>
                <input value={hiddenDomain} onChange={e => setHiddenDomain(e.target.value)} placeholder="your-server.com" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              </div>
              <div>
                <label className="block text-xs text-[#9898b8] mb-1">CDN Provider</label>
                <select value={providerName} onChange={e => setProviderName(e.target.value)} className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none">
                  <option value="cloudflare">Cloudflare</option>
                  <option value="fastly">Fastly</option>
                  <option value="akamai">Akamai</option>
                  <option value="cloudfront">CloudFront</option>
                </select>
              </div>
              <button onClick={handleGenerateConfig} disabled={!frontDomain} className="w-full px-5 py-2.5 bg-gradient-to-r from-purple-600 to-orange-600 text-white rounded-lg hover:from-purple-500 hover:to-orange-500 disabled:opacity-50 transition text-sm font-medium">Generate Config</button>
            </div>
          </div>
          {proxyConfig && (
            <div className="glass-card p-5">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-white font-semibold">📋 Generated Config</h3>
                <button onClick={() => copyConfig(JSON.stringify(proxyConfig, null, 2))} className="text-xs text-purple-400 hover:text-purple-300">{copied ? '✅ Copied!' : '📋 Copy'}</button>
              </div>
              <div className="space-y-2 mb-4">
                <div className="flex items-center justify-between p-2 rounded bg-[rgba(255,255,255,0.02)] text-xs"><span className="text-[#9898b8]">Provider</span><span className="text-white">{proxyConfig.provider}</span></div>
                <div className="flex items-center justify-between p-2 rounded bg-[rgba(255,255,255,0.02)] text-xs"><span className="text-[#9898b8]">Front Domain</span><span className="text-white">{proxyConfig.front_domain}</span></div>
                <div className="flex items-center justify-between p-2 rounded bg-[rgba(255,255,255,0.02)] text-xs"><span className="text-[#9898b8]">SNI</span><span className="text-white">{proxyConfig.sni}</span></div>
                <div className="flex items-center justify-between p-2 rounded bg-[rgba(255,255,255,0.02)] text-xs"><span className="text-[#9898b8]">Port</span><span className="text-white">{proxyConfig.port}</span></div>
              </div>
              <h4 className="text-xs text-[#9898b8] font-medium mb-2 uppercase">Xray Outbound</h4>
              <pre className="p-3 rounded-lg bg-[#08080f] text-emerald-300 text-[10px] font-mono overflow-x-auto max-h-40 border border-[rgba(255,255,255,0.06)]">{JSON.stringify(proxyConfig.xray_outbound, null, 2)}</pre>
              <h4 className="text-xs text-[#9898b8] font-medium mb-2 mt-3 uppercase">Sing-box Outbound</h4>
              <pre className="p-3 rounded-lg bg-[#08080f] text-emerald-300 text-[10px] font-mono overflow-x-auto max-h-40 border border-[rgba(255,255,255,0.06)]">{JSON.stringify(proxyConfig.singbox_outbound, null, 2)}</pre>
            </div>
          )}
        </div>
      )}

      {tab === 'domains' && (
        <div className="glass-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="table-modern">
              <thead>
                <tr>
                  <th>Domain</th>
                  <th>Provider</th>
                  <th>Status</th>
                  <th>Frontable</th>
                  <th>Latency</th>
                  <th>TLS</th>
                  <th>Checked</th>
                </tr>
              </thead>
              <tbody>
                {scanDomains.map(d => (
                  <tr key={d.id}>
                    <td className="text-white text-sm font-medium">{d.domain}</td>
                    <td><span className="px-2 py-0.5 rounded text-[10px] font-medium bg-purple-500/10 text-purple-300">{d.cdn_provider}</span></td>
                    <td><span className={`px-2 py-0.5 rounded text-[10px] font-medium ${d.status === 'active' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>{d.status}</span></td>
                    <td><span className={`text-xs ${d.frontable ? 'text-emerald-400' : 'text-[#585878]'}`}>{d.frontable ? '✅' : '—'}</span></td>
                    <td className="text-xs text-[#9898b8]">{d.latency_ms}ms</td>
                    <td className="text-xs text-[#9898b8]">{d.tls_version || '—'}</td>
                    <td className="text-xs text-[#585878]">{d.last_checked ? new Date(d.last_checked).toLocaleString() : '—'}</td>
                  </tr>
                ))}
                {scanDomains.length === 0 && <tr><td colSpan={7} className="text-center py-12 text-[#585878]">No domains scanned yet. Use the Scan tab to discover CDN domains.</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
