import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface WARPConfig {
  enabled: boolean
  license_key: string
  endpoint: string
  address_v4: string
  dns: string
  mtu: number
  outbound_tag: string
  auto_connect: boolean
  connected: boolean
}

interface WARPStatus {
  connected: boolean
  endpoint: string
  address_v4: string
  uptime_seconds: number
  error: string
}

export function WarpPage() {
  const [config, setConfig] = useState<WARPConfig | null>(null)
  const [status, setStatus] = useState<WARPStatus | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    setLoading(true)
    try {
      const [cfgRes, statusRes] = await Promise.all([
        apiClient.get('/api/v1/warp/config'),
        apiClient.get('/api/v1/warp/status'),
      ])
      setConfig(cfgRes.data)
      setStatus(statusRes.data)
    } catch { }
    setLoading(false)
  }

  const handleConnect = async () => {
    try {
      await apiClient.post('/api/v1/warp/connect')
      fetchData()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to connect')
    }
  }

  const handleDisconnect = async () => {
    try {
      await apiClient.post('/api/v1/warp/disconnect')
      fetchData()
    } catch { }
  }

  const handleSave = async () => {
    if (!config) return
    try {
      await apiClient.put('/api/v1/warp/config', config)
      alert('Configuration saved')
    } catch { }
  }

  if (loading) {
    return <div className="text-center py-20 text-[var(--text-muted)]">Loading...</div>
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-sky-500 to-blue-600 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">WARP+ Outbound</h1>
            <p className="text-sm text-[var(--text-secondary)]">Cloudflare WARP tunnel integration</p>
          </div>
          <div className="ml-auto">
            {status?.connected ? (
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
                <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                <span className="text-xs font-medium text-emerald-400">Connected</span>
                <span className="text-[10px] text-[var(--text-muted)]">{Math.floor((status?.uptime_seconds || 0) / 60)}m</span>
              </div>
            ) : (
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-red-500/10 border border-red-500/20">
                <div className="w-2 h-2 rounded-full bg-red-400" />
                <span className="text-xs font-medium text-red-400">Disconnected</span>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Config */}
        <div className="lg:col-span-2">
          <div className="glass-card p-6">
            <h3 className="text-base font-bold text-[var(--text-primary)] mb-4">Configuration</h3>
            {config && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Endpoint</label>
                    <input type="text" value={config.endpoint} onChange={e => setConfig({ ...config, endpoint: e.target.value })} className="input-modern text-sm w-full" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Outbound Tag</label>
                    <input type="text" value={config.outbound_tag} onChange={e => setConfig({ ...config, outbound_tag: e.target.value })} className="input-modern text-sm w-full" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Address (IPv4)</label>
                    <input type="text" value={config.address_v4} onChange={e => setConfig({ ...config, address_v4: e.target.value })} className="input-modern text-sm w-full" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">DNS</label>
                    <input type="text" value={config.dns} onChange={e => setConfig({ ...config, dns: e.target.value })} className="input-modern text-sm w-full" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">MTU</label>
                    <input type="number" value={config.mtu} onChange={e => setConfig({ ...config, mtu: Number(e.target.value) })} className="input-modern text-sm w-full" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">License Key</label>
                    <input type="text" value={config.license_key} onChange={e => setConfig({ ...config, license_key: e.target.value })} className="input-modern text-sm w-full" placeholder="WARP+ license key" />
                  </div>
                </div>
                <div className="flex items-center justify-between pt-3 border-t border-[rgba(255,255,255,0.06)]">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" checked={config.auto_connect} onChange={e => setConfig({ ...config, auto_connect: e.target.checked })} className="rounded" />
                    <span className="text-xs text-[var(--text-secondary)]">Auto-connect on startup</span>
                  </label>
                  <div className="flex gap-2">
                    {!status?.connected ? (
                      <button onClick={handleConnect} className="btn-primary text-sm">
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                        Connect
                      </button>
                    ) : (
                      <button onClick={handleDisconnect} className="btn-danger text-sm">
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                        Disconnect
                      </button>
                    )}
                    <button onClick={handleSave} className="btn-secondary text-sm">Save Config</button>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Status */}
        <div>
          <div className="glass-card p-6">
            <h3 className="text-base font-bold text-[var(--text-primary)] mb-4">Connection Status</h3>
            {status && (
              <div className="space-y-3">
                <div className="flex justify-between py-2 border-b border-[rgba(255,255,255,0.06)]">
                  <span className="text-xs text-[var(--text-muted)]">Status</span>
                  <span className={`text-xs font-medium ${status.connected ? 'text-emerald-400' : 'text-red-400'}`}>
                    {status.connected ? 'Connected' : 'Disconnected'}
                  </span>
                </div>
                <div className="flex justify-between py-2 border-b border-[rgba(255,255,255,0.06)]">
                  <span className="text-xs text-[var(--text-muted)]">Endpoint</span>
                  <span className="text-xs text-[var(--text-primary)]">{status.endpoint}</span>
                </div>
                <div className="flex justify-between py-2 border-b border-[rgba(255,255,255,0.06)]">
                  <span className="text-xs text-[var(--text-muted)]">IPv4</span>
                  <span className="text-xs text-[var(--text-primary)]">{status.address_v4}</span>
                </div>
                <div className="flex justify-between py-2 border-b border-[rgba(255,255,255,0.06)]">
                  <span className="text-xs text-[var(--text-muted)]">Uptime</span>
                  <span className="text-xs text-[var(--text-primary)]">{Math.floor(status.uptime_seconds / 60)}m {status.uptime_seconds % 60}s</span>
                </div>
                {status.error && (
                  <div className="p-2 rounded-lg bg-red-500/10 text-xs text-red-400">{status.error}</div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
