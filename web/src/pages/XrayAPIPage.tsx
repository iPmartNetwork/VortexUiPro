import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '../api/client'

interface ProcessInfo {
  pid: number
  status: string
  uptime: string
  apiPort: number
  binary: string
  config: string
  online: number
  lastError?: string
}

interface TrafficStat {
  tag: string
  up: number
  down: number
  time: string
}

interface ClientTraffic {
  email: string
  up: number
  down: number
}

interface OnlineUser {
  email: string
  ips: Array<{ ip: string; lastSeen: number }>
}

interface BalancerInfo {
  tag: string
  override: string
  selected: string[]
}

interface TestRouteResult {
  matched: boolean
  outboundTag: string
  groupTags?: string[]
}

type TabKey = 'dashboard' | 'traffic' | 'users' | 'routing' | 'logs'

export function XrayAPIPage() {
  const [tab, setTab] = useState<TabKey>('dashboard')
  const [processInfo, setProcessInfo] = useState<ProcessInfo | null>(null)
  const [traffic, setTraffic] = useState<TrafficStat[]>([])
  const [clientTraffic, setClientTraffic] = useState<ClientTraffic[]>([])
  const [onlineUsers, setOnlineUsers] = useState<OnlineUser[]>([])
  const [logs, setLogs] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [configText, setConfigText] = useState('')
  const [configResult, setConfigResult] = useState<{ valid?: boolean; error?: string } | null>(null)

  // Route test
  const [routeTest, setRouteTest] = useState({ domain: '', port: 443, protocol: '', email: '' })
  const [routeResult, setRouteResult] = useState<TestRouteResult | null>(null)

  const loadDashboard = useCallback(async () => {
    try {
      const [pRes, tRes, uRes, lRes] = await Promise.all([
        apiClient.get('/api/v1/xray/process').catch(() => ({ data: { data: null } })),
        apiClient.get('/api/v1/xray/traffic').catch(() => ({ data: { data: {} } })),
        apiClient.get('/api/v1/xray/online-users').catch(() => ({ data: { data: [] } })),
        apiClient.get('/api/v1/xray/logs').catch(() => ({ data: { data: [] } })),
      ])
      if (pRes.data.data) setProcessInfo(pRes.data.data)
      if (tRes.data.data) {
        setTraffic(tRes.data.data.inbound_traffic || [])
        setClientTraffic(tRes.data.data.client_traffic || [])
      }
      if (uRes.data.data) setOnlineUsers(uRes.data.data)
      if (lRes.data.data) setLogs(lRes.data.data)
    } catch {}
    finally { setLoading(false) }
  }, [])

  useEffect(() => { loadDashboard() }, [loadDashboard])

  // Auto-refresh dashboard every 10s
  useEffect(() => {
    if (tab !== 'dashboard') return
    const interval = setInterval(loadDashboard, 10000)
    return () => clearInterval(interval)
  }, [tab, loadDashboard])

  const handleValidateConfig = useCallback(async () => {
    if (!configText) return
    try {
      const res = await apiClient.post('/api/v1/xray/validate', { config: configText })
      setConfigResult(res.data)
    } catch (err: any) {
      setConfigResult({ valid: false, error: err.response?.data?.error || 'Validation failed' })
    }
  }, [configText])

  const handleTestRoute = useCallback(async () => {
    if (!routeTest.domain) return
    try {
      const res = await apiClient.post('/api/v1/xray/test-route', routeTest)
      setRouteResult(res.data.data)
    } catch {}
  }, [routeTest])

  if (loading) return <div className="flex items-center justify-center min-h-[400px] text-[#6868a0]">Loading Xray API...</div>

  return (
    <div className="space-y-6 page-enter">
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-orange-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-white">Xray Core gRPC API</h1>
            <p className="text-sm text-[#6868a0] mt-0.5">Real-time core statistics, traffic monitoring, online users & routing test</p>
          </div>
        </div>
      </div>

      {/* Process Status Bar */}
      {processInfo && (
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3">
          {[
            { label: 'Status', value: processInfo.status, color: processInfo.status === 'running' ? 'text-emerald-400' : 'text-red-400' },
            { label: 'PID', value: String(processInfo.pid || '—'), color: 'text-white' },
            { label: 'Uptime', value: processInfo.uptime, color: 'text-white' },
            { label: 'API Port', value: String(processInfo.apiPort), color: 'text-white' },
            { label: 'Online Users', value: String(processInfo.online), color: 'text-cyan-400' },
            { label: 'Binary', value: processInfo.binary.split('/').pop() || 'xray', color: 'text-[#9898b8] text-xs' },
            { label: 'Last Error', value: processInfo.lastError ? '⚠ Error' : 'None', color: processInfo.lastError ? 'text-red-400' : 'text-emerald-400' },
          ].map((s, i) => (
            <div key={i} className="glass-card px-3 py-2.5 text-center">
              <p className="text-[9px] text-[#6868a0] uppercase tracking-wider">{s.label}</p>
              <p className={`text-sm font-bold mt-0.5 truncate ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 p-1 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] w-fit flex-wrap">
        {(['dashboard', 'traffic', 'users', 'routing', 'logs'] as TabKey[]).map(t => (
          <button key={t} onClick={() => setTab(t)} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all capitalize ${tab === t ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
            {t === 'dashboard' && '📊'} {t === 'traffic' && '📈'} {t === 'users' && '👤'} {t === 'routing' && '🔀'} {t === 'logs' && '📋'}
            {' '}{t}
          </button>
        ))}
      </div>

      {tab === 'dashboard' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Traffic Overview */}
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">📊 Traffic Stats</h3>
            <div className="space-y-2 max-h-80 overflow-y-auto">
              {traffic.map(t => (
                <div key={t.tag} className="p-3 rounded-lg bg-[rgba(255,255,255,0.02)] flex items-center justify-between">
                  <span className="text-sm text-white font-medium">{t.tag}</span>
                  <div className="text-right text-xs">
                    <p className="text-emerald-400">↑ {formatBytes(t.up)}</p>
                    <p className="text-blue-400">↓ {formatBytes(t.down)}</p>
                  </div>
                </div>
              ))}
              {traffic.length === 0 && <p className="text-center py-8 text-[#585878]">No traffic data. Make sure xray is running with stats enabled.</p>}
            </div>
          </div>

          {/* Recent Logs */}
          <div className="glass-card p-5">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-semibold">📋 Recent Logs</h3>
              <button onClick={loadDashboard} className="text-xs text-purple-400 hover:text-purple-300">🔄 Refresh</button>
            </div>
            <div className="space-y-1 max-h-80 overflow-y-auto font-mono text-[10px]">
              {logs.length > 0 ? logs.slice(-20).reverse().map((line, i) => (
                <div key={i} className="text-[#9898b8] hover:text-white transition truncate">
                  {line}
                </div>
              )) : <p className="text-center py-8 text-[#585878]">No logs available</p>}
            </div>
          </div>
        </div>
      )}

      {tab === 'traffic' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">📈 Inbound Traffic</h3>
            <div className="overflow-x-auto">
              <table className="table-modern">
                <thead>
                  <tr>
                    <th>Tag</th>
                    <th>Upload</th>
                    <th>Download</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {traffic.map(t => (
                    <tr key={t.tag}>
                      <td className="text-white text-sm font-medium">{t.tag}</td>
                      <td className="text-emerald-400 text-xs">{formatBytes(t.up)}</td>
                      <td className="text-blue-400 text-xs">{formatBytes(t.down)}</td>
                      <td className="text-[#9898b8] text-xs">{formatBytes(t.up + t.down)}</td>
                    </tr>
                  ))}
                  {traffic.length === 0 && <tr><td colSpan={4} className="text-center py-8 text-[#585878]">No traffic data</td></tr>}
                </tbody>
              </table>
            </div>
          </div>
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">👤 Client Traffic</h3>
            <div className="overflow-x-auto">
              <table className="table-modern">
                <thead>
                  <tr>
                    <th>Email</th>
                    <th>Upload</th>
                    <th>Download</th>
                  </tr>
                </thead>
                <tbody>
                  {clientTraffic.slice(0, 50).map(ct => (
                    <tr key={ct.email}>
                      <td className="text-white text-sm font-medium truncate max-w-[200px]">{ct.email}</td>
                      <td className="text-emerald-400 text-xs">{formatBytes(ct.up)}</td>
                      <td className="text-blue-400 text-xs">{formatBytes(ct.down)}</td>
                    </tr>
                  ))}
                  {clientTraffic.length === 0 && <tr><td colSpan={3} className="text-center py-8 text-[#585878]">No client traffic</td></tr>}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {tab === 'users' && (
        <div className="glass-card p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-white font-semibold">👤 Online Users ({onlineUsers.length})</h3>
            <button onClick={loadDashboard} className="text-xs text-purple-400 hover:text-purple-300">🔄 Refresh</button>
          </div>
          <div className="overflow-x-auto">
            <table className="table-modern">
              <thead>
                <tr>
                  <th>Email</th>
                  <th>IP Addresses</th>
                  <th>Active Connections</th>
                </tr>
              </thead>
              <tbody>
                {onlineUsers.map(u => (
                  <tr key={u.email}>
                    <td className="text-white text-sm font-medium">{u.email}</td>
                    <td>
                      <div className="flex flex-wrap gap-1">
                        {u.ips.map(ip => (
                          <span key={ip.ip} className="px-1.5 py-0.5 rounded bg-[rgba(255,255,255,0.03)] text-[10px] text-[#9898b8] font-mono">
                            {ip.ip}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="text-xs text-[#9898b8]">{u.ips.length}</td>
                  </tr>
                ))}
                {onlineUsers.length === 0 && <tr><td colSpan={3} className="text-center py-12 text-[#585878]">No users currently online</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === 'routing' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Route Test */}
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">🔀 Route Test</h3>
            <p className="text-xs text-[#6868a0] mb-4">Test which outbound the running xray core would select for a connection.</p>
            <div className="space-y-3">
              <div>
                <label className="block text-xs text-[#9898b8] mb-1">Domain</label>
                <input value={routeTest.domain} onChange={e => setRouteTest(p => ({ ...p, domain: e.target.value }))} placeholder="e.g., google.com" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="block text-xs text-[#9898b8] mb-1">Port</label>
                  <input value={routeTest.port} onChange={e => setRouteTest(p => ({ ...p, port: Number(e.target.value) }))} type="number" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
                </div>
                <div>
                  <label className="block text-xs text-[#9898b8] mb-1">Protocol</label>
                  <input value={routeTest.protocol} onChange={e => setRouteTest(p => ({ ...p, protocol: e.target.value }))} placeholder="e.g., tls, http" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
                </div>
              </div>
              <button onClick={handleTestRoute} disabled={!routeTest.domain} className="w-full px-5 py-2.5 bg-gradient-to-r from-purple-600 to-cyan-600 text-white rounded-lg hover:from-purple-500 hover:to-cyan-500 disabled:opacity-50 transition text-sm font-medium">Test Route</button>
            </div>
          </div>

          <div className="space-y-6">
            {/* Route Test Result */}
            {routeResult && (
              <div className="glass-card p-5">
                <h3 className="text-white font-semibold mb-4">📊 Route Result</h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                    <span className="text-[#9898b8] text-sm">Matched</span>
                    <span className={`text-sm font-bold ${routeResult.matched ? 'text-emerald-400' : 'text-yellow-400'}`}>{routeResult.matched ? '✅ Yes' : '⚠ No rule matched'}</span>
                  </div>
                  {routeResult.matched && (
                    <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                      <span className="text-[#9898b8] text-sm">Outbound Tag</span>
                      <span className="text-white text-sm font-mono">{routeResult.outboundTag}</span>
                    </div>
                  )}
                  {routeResult.groupTags && routeResult.groupTags.length > 0 && (
                    <div className="p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                      <p className="text-[#9898b8] text-xs mb-2">Group Tags</p>
                      <div className="flex flex-wrap gap-1">
                        {routeResult.groupTags.map(tag => (
                          <span key={tag} className="px-2 py-0.5 rounded text-[10px] font-medium bg-purple-500/10 text-purple-300">{tag}</span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Config Validation */}
            <div className="glass-card p-5">
              <h3 className="text-white font-semibold mb-3">✅ Validate Config</h3>
              <textarea value={configText} onChange={e => setConfigText(e.target.value)} rows={4} placeholder="Paste xray JSON config to validate..." className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-xs font-mono focus:border-purple-500/40 focus:outline-none" />
              <button onClick={handleValidateConfig} disabled={!configText} className="mt-3 w-full px-5 py-2.5 bg-gradient-to-r from-cyan-600 to-teal-600 text-white rounded-lg hover:from-cyan-500 hover:to-teal-500 disabled:opacity-50 transition text-sm font-medium">Validate</button>
              {configResult && (
                <div className={`mt-3 p-3 rounded-lg text-sm ${configResult.valid ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
                  {configResult.valid ? '✅ Config is valid!' : `❌ ${configResult.error}`}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {tab === 'logs' && (
        <div className="glass-card p-5">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-white font-semibold">📋 Core Logs</h3>
            <button onClick={loadDashboard} className="text-xs text-purple-400 hover:text-purple-300">🔄 Refresh</button>
          </div>
          <div className="bg-[#08080f] rounded-lg border border-[rgba(255,255,255,0.06)] p-4 max-h-[600px] overflow-y-auto">
            {logs.length > 0 ? logs.map((line, i) => (
              <div key={i} className="text-[11px] leading-5 font-mono text-[#9898b8] hover:text-white transition border-b border-[rgba(255,255,255,0.02)] last:border-0 py-1">
                {line}
              </div>
            )) : <p className="text-center py-12 text-[#585878]">No logs available. Xray may not be running.</p>}
          </div>
        </div>
      )}
    </div>
  )
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}
