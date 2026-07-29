import React, { useEffect, useState, useCallback } from 'react'
import { api } from '../api/client'
import { useI18n } from '../hooks/useI18n'
// CSS animations used instead of framer-motion

// ─── Types ───────────────────────────────────────────────────────────

interface ICEServer {
  urls: string[]
  username?: string
  credential?: string
}

interface ICEConfig {
  stun_servers: string[]
  turn_servers: TURNServer[]
  ice_servers: ICEServer[]
}

interface TURNServer {
  id: number
  address: string
  username: string
  password: string
  realm: string
  protocol: string
  status: string
  region: string
  bandwidth: number
  enabled: boolean
  created_at: number
}

interface WebRTCPeer {
  id: string
  node_id: string
  name: string
  protocol: string
  mode: string
  status: string
  local_addr: string
  remote_addr: string
  latency: number
  bandwidth: number
  connected_at: number
  last_seen: number
  created_at: number
}

interface P2PMeshConfig {
  enabled: boolean
  mesh_name: string
  role: string
  listen_port: number
  max_peers: number
  auto_reconnect: boolean
  discovery: string
  encryption: boolean
  heartbeat_sec: number
  data_channel: boolean
}

interface PeerStats {
  total: number
  connected: number
  connecting: number
  disconnected: number
  turn_servers: number
  mesh_enabled: boolean
}

interface NATType {
  type: string
  public_ip: string
  public_port: number
  behind_nat: boolean
  description: string
}

// ─── Icon Components ─────────────────────────────────────────────────

const Icons = {
  Globe: () => (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  Server: () => (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
    </svg>
  ),
  Nodes: () => (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
    </svg>
  ),
  Shield: () => (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
    </svg>
  ),
  Plus: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
    </svg>
  ),
  Refresh: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
    </svg>
  ),
  Check: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  ),
  X: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
    </svg>
  ),
  Activity: () => (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
    </svg>
  ),
}

// ─── Modal Component ─────────────────────────────────────────────────

function Modal({ open, onClose, title, children }: { open: boolean; onClose: () => void; title: string; children: React.ReactNode }) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center backdrop-blur-sm bg-black/40" onClick={onClose}>
      <div
        className="bg-white dark:bg-[#1a1a2e] rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700/30 w-full max-w-lg mx-4 overflow-hidden animate-scale-in"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700/30">
          <h2 className="text-lg font-bold text-gray-900 dark:text-white">{title}</h2>
          <button onClick={onClose} className="p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700/50 text-gray-400">
            <Icons.X />
          </button>
        </div>
        <div className="px-6 py-4">{children}</div>
      </div>
    </div>
  )
}

// ─── Main Component ──────────────────────────────────────────────────

export default function WebRTCPage() {
  const { t, isRTL } = useI18n()

  // State
  const [iceConfig, setIceConfig] = useState<ICEConfig | null>(null)
  const [turnServers, setTurnServers] = useState<TURNServer[]>([])
  const [peers, setPeers] = useState<WebRTCPeer[]>([])
  const [peerStats, setPeerStats] = useState<PeerStats | null>(null)
  const [meshConfig, setMeshConfig] = useState<P2PMeshConfig | null>(null)
  const [natType, setNatType] = useState<NATType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showAddTURN, setShowAddTURN] = useState(false)
  const [testResult, setTestResult] = useState<{ address: string; reachable: boolean; latency_ms: number } | null>(null)

  // Form state
  const [turnForm, setTurnForm] = useState({
    address: '', username: '', password: '', realm: '',
    protocol: 'udp', region: '', bandwidth: 100
  })

  // ─── Load Data ─────────────────────────────────────────────────────
  const loadData = useCallback(async () => {
    try {
      setLoading(true)
      const [iceRes, turnRes, peersRes, statsRes, meshRes, natRes] = await Promise.all([
        api.get('/api/v1/webrtc/ice-config'),
        api.get('/api/v1/webrtc/turn-servers'),
        api.get('/api/v1/webrtc/peers'),
        api.get('/api/v1/webrtc/peers/stats'),
        api.get('/api/v1/webrtc/mesh-config'),
        api.get('/api/v1/webrtc/nat-type').catch(() => null),
      ])
      setIceConfig(iceRes.data)
      setTurnServers(turnRes.data || [])
      setPeers(peersRes.data || [])
      setPeerStats(statsRes.data)
      setMeshConfig(meshRes.data)
      setNatType(natRes?.data || null)
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadData(); const t = setInterval(loadData, 10000); return () => clearInterval(t) }, [loadData])

  // ─── Actions ──────────────────────────────────────────────────────

  const addTURNServer = async () => {
    try {
      await api.post('/api/v1/webrtc/turn-servers', turnForm)
      setShowAddTURN(false)
      setTurnForm({ address: '', username: '', password: '', realm: '', protocol: 'udp', region: '', bandwidth: 100 })
      loadData()
    } catch (err: any) { setError(err.message) }
  }

  const deleteTURN = async (id: number) => {
    try {
      await api.delete(`/api/v1/webrtc/turn-servers/${id}`)
      loadData()
    } catch (err: any) { setError(err.message) }
  }

  const testTURN = async (address: string) => {
    try {
      const res = await api.post('/api/v1/webrtc/turn-servers/test', { address })
      setTestResult({ address, ...res.data })
      setTimeout(() => setTestResult(null), 5000)
    } catch (err: any) { setError(err.message) }
  }

  const disconnectPeer = async (id: string) => {
    try {
      await api.delete(`/api/v1/webrtc/peers/${id}`)
      loadData()
    } catch (err: any) { setError(err.message) }
  }

  const updateMeshConfig = async (cfg: Partial<P2PMeshConfig>) => {
    try {
      await api.put('/api/v1/webrtc/mesh-config', { ...meshConfig, ...cfg })
      loadData()
    } catch (err: any) { setError(err.message) }
  }

  // ─── Status Badge ─────────────────────────────────────────────────
  const StatusBadge = ({ status }: { status: string }) => {
    const colors: Record<string, string> = {
      connected: 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/30',
      connecting: 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border-amber-500/30',
      disconnected: 'bg-gray-500/20 text-gray-600 dark:text-gray-400 border-gray-500/30',
      failed: 'bg-red-500/20 text-red-600 dark:text-red-400 border-red-500/30',
      online: 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/30',
      offline: 'bg-gray-500/20 text-gray-600 dark:text-gray-400 border-gray-500/30',
    }
    return (
      <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium border ${colors[status] || colors.disconnected}`}>
        {status}
      </span>
    )
  }

  // ─── Loading ──────────────────────────────────────────────────────
  if (loading && !iceConfig) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin" />
          <p className="text-gray-400 text-sm">{t('loading') || 'Loading WebRTC configuration...'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6" dir={isRTL ? 'rtl' : 'ltr'}>
      {/* ─── Header ───────────────────────────────────────────── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <Icons.Globe /> WebRTC & P2P Mesh
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Manage direct connections, STUN/TURN servers, and peer-to-peer mesh network
          </p>
        </div>
        <button onClick={loadData} className="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-500 border border-indigo-500/20 transition-all">
          <Icons.Refresh /> {t('refresh') || 'Refresh'}
        </button>
      </div>

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-600 dark:text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* ─── Stats Cards ─────────────────────────────────────── */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {[
          { label: 'Connected Peers', value: peerStats?.connected || 0, icon: <Icons.Nodes />, color: 'emerald' },
          { label: 'TURN Servers', value: peerStats?.turn_servers || 0, icon: <Icons.Server />, color: 'blue' },
          { label: 'Mesh Mode', value: meshConfig?.role || 'disabled', icon: <Icons.Shield />, color: 'purple' },
          { label: 'NAT Type', value: natType?.type || 'unknown', icon: <Icons.Activity />, color: 'amber' },
        ].map((stat, i) => (
          <div
            style={{ animationDelay: `${i * 50}ms` }}
            className="p-5 rounded-2xl bg-white dark:bg-[#1a1a2e] border border-gray-200 dark:border-gray-700/30 shadow-sm hover:shadow-md transition-shadow animate-fade-in-up"
          >
            <div className="flex items-center gap-3 mb-3">
              <div className={`p-2 rounded-lg ${stat.color === 'emerald' ? 'bg-emerald-500/10 text-emerald-500' : stat.color === 'blue' ? 'bg-blue-500/10 text-blue-500' : stat.color === 'purple' ? 'bg-purple-500/10 text-purple-500' : 'bg-amber-500/10 text-amber-500'}`}>{stat.icon}</div>
            </div>
            <p className={`text-2xl font-bold text-gray-900 dark:text-white`}>
              {typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}
            </p>            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{stat.label}</p>
          </div>
        ))}
        </div>

      {/* ─── ICE Configuration ───────────────────────────────── */}
      <div className="p-6 rounded-2xl bg-white dark:bg-[#1a1a2e] border border-gray-200 dark:border-gray-700/30 shadow-sm">
        <h2 className="text-lg font-bold text-gray-900 dark:text-white mb-4">ICE Configuration</h2>
        {iceConfig && (
          <div className="space-y-3">
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-2">STUN Servers</h3>
              <div className="flex flex-wrap gap-2">
                {iceConfig.stun_servers.map((s, i) => (
                  <span key={i} className="px-3 py-1.5 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400 text-xs font-mono border border-blue-500/20">
                    {s}
                  </span>
                ))}
              </div>
            </div>
            {iceConfig.turn_servers.length > 0 && (
              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-2">TURN Servers</h3>
                <div className="flex flex-wrap gap-2">
                  {iceConfig.turn_servers.map((t, i) => (
                    <span key={i} className="px-3 py-1.5 rounded-lg bg-purple-500/10 text-purple-600 dark:text-purple-400 text-xs font-mono border border-purple-500/20">
                      {t.address}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* ─── TURN Servers ────────────────────────────────────── */}
      <div className="p-6 rounded-2xl bg-white dark:bg-[#1a1a2e] border border-gray-200 dark:border-gray-700/30 shadow-sm">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-bold text-gray-900 dark:text-white">TURN Servers</h2>
          <button onClick={() => setShowAddTURN(true)} className="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white transition-all text-sm font-medium">
            <Icons.Plus /> Add TURN Server
          </button>
        </div>

        {turnServers.length === 0 ? (
          <p className="text-gray-400 text-center py-8">No TURN servers configured. Add one to enable relay connections.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 dark:border-gray-700/30">
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Address</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Protocol</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Region</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Bandwidth</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Status</th>
                  <th className="text-right py-3 px-3 text-gray-500 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {turnServers.map((server) => (
                  <tr key={server.id} className="border-b border-gray-100 dark:border-gray-800/50 hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors">
                    <td className="py-3 px-3 text-gray-900 dark:text-white font-mono text-xs">{server.address}</td>
                    <td className="py-3 px-3"><span className="uppercase text-xs font-mono text-gray-500">{server.protocol}</span></td>
                    <td className="py-3 px-3 text-gray-500">{server.region || '—'}</td>
                    <td className="py-3 px-3 text-gray-500">{server.bandwidth} Mbps</td>
                    <td className="py-3 px-3"><StatusBadge status={server.status} /></td>
                    <td className="py-3 px-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button onClick={() => testTURN(server.address)} className="p-1.5 rounded-lg hover:bg-blue-500/10 text-blue-500 transition-colors" title="Test connection">
                          <Icons.Activity />
                        </button>
                        <button onClick={() => deleteTURN(server.id)} className="p-1.5 rounded-lg hover:bg-red-500/10 text-red-500 transition-colors" title="Delete">
                          <Icons.X />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {testResult && (
          <div className={`mt-4 p-3 rounded-xl text-sm animate-fade-in-up ${testResult.reachable ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-red-500/10 text-red-600 dark:text-red-400'}`}>
            {testResult.reachable
              ? `✅ ${testResult.address} reachable (${testResult.latency_ms.toFixed(1)}ms)`
              : `❌ ${testResult.address} not reachable`}
          </div>
        )}
      </div>

      {/* ─── P2P Mesh Configuration ──────────────────────────── */}
      {meshConfig && (
        <div className="p-6 rounded-2xl bg-white dark:bg-[#1a1a2e] border border-gray-200 dark:border-gray-700/30 shadow-sm">
          <h2 className="text-lg font-bold text-gray-900 dark:text-white mb-4">P2P Mesh Network</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { label: 'Enabled', value: meshConfig.enabled ? 'Yes' : 'No', key: 'enabled' as const },
              { label: 'Mesh Name', value: meshConfig.mesh_name, key: 'mesh_name' as const },
              { label: 'Role', value: meshConfig.role, key: 'role' as const },
              { label: 'Max Peers', value: String(meshConfig.max_peers), key: 'max_peers' as const },
              { label: 'Discovery', value: meshConfig.discovery, key: 'discovery' as const },
              { label: 'Encryption', value: meshConfig.encryption ? 'Enabled' : 'Disabled', key: 'encryption' as const },
              { label: 'Heartbeat', value: `${meshConfig.heartbeat_sec}s`, key: 'heartbeat_sec' as const },
              { label: 'Auto Reconnect', value: meshConfig.auto_reconnect ? 'Yes' : 'No', key: 'auto_reconnect' as const },
            ].map((item) => (
              <div key={item.key} className="p-3 rounded-xl bg-gray-50 dark:bg-gray-800/50">
                <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">{item.label}</p>
                <div className="flex items-center gap-2">
                  {item.key === 'enabled' && (
                    <button
                      onClick={() => updateMeshConfig({ enabled: !meshConfig.enabled })}
                      className={`relative inline-flex h-5 w-9 rounded-full transition-colors ${meshConfig.enabled ? 'bg-indigo-500' : 'bg-gray-300 dark:bg-gray-600'}`}
                    >
                      <span className={`inline-block h-4 w-4 rounded-full bg-white shadow-sm transform transition-transform mt-0.5 ${meshConfig.enabled ? 'translate-x-4 ml-0.5' : 'translate-x-0.5'}`} />
                    </button>
                  )}
                  <span className="text-sm font-semibold text-gray-900 dark:text-white">{item.value}</span>
                </div>
              </div>
            ))}
          </div>
          {/* Role selector */}
          <div className="mt-4 flex gap-2">
            {['direct', 'relay', 'hybrid'].map((role) => (
              <button
                key={role}
                onClick={() => updateMeshConfig({ role })}
                className={`px-4 py-2 rounded-xl text-sm font-medium capitalize transition-all ${
                  meshConfig.role === role
                    ? 'bg-indigo-500 text-white shadow-lg shadow-indigo-500/20'
                    : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700'
                }`}
              >
                {role}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* ─── Peers Table ─────────────────────────────────────── */}
      <div className="p-6 rounded-2xl bg-white dark:bg-[#1a1a2e] border border-gray-200 dark:border-gray-700/30 shadow-sm">
        <h2 className="text-lg font-bold text-gray-900 dark:text-white mb-4">Connected Peers ({peers.length})</h2>
        {peers.length === 0 ? (
          <p className="text-gray-400 text-center py-8">No peers connected. Enable P2P mesh or wait for incoming connections.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 dark:border-gray-700/30">
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Peer ID</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Name</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Mode</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Protocol</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Latency</th>
                  <th className="text-left py-3 px-3 text-gray-500 font-medium">Status</th>
                  <th className="text-right py-3 px-3 text-gray-500 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {peers.map((peer) => (
                  <tr key={peer.id} className="border-b border-gray-100 dark:border-gray-800/50 hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors">
                    <td className="py-3 px-3 font-mono text-xs text-gray-900 dark:text-white">{peer.id}</td>
                    <td className="py-3 px-3 text-gray-700 dark:text-gray-300">{peer.name}</td>
                    <td className="py-3 px-3"><span className="capitalize text-xs">{peer.mode}</span></td>
                    <td className="py-3 px-3"><span className="font-mono text-xs">{peer.protocol}</span></td>
                    <td className="py-3 px-3">
                      {peer.latency > 0 ? (
                        <span className={`font-mono text-xs ${peer.latency < 50 ? 'text-emerald-500' : peer.latency < 150 ? 'text-amber-500' : 'text-red-500'}`}>
                          {peer.latency.toFixed(1)}ms
                        </span>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                    <td className="py-3 px-3"><StatusBadge status={peer.status} /></td>
                    <td className="py-3 px-3 text-right">
                      <button onClick={() => disconnectPeer(peer.id)} className="p-1.5 rounded-lg hover:bg-red-500/10 text-red-500 transition-colors" title="Disconnect">
                        <Icons.X />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ─── Add TURN Modal ──────────────────────────────────── */}
      <Modal open={showAddTURN} onClose={() => setShowAddTURN(false)} title="Add TURN Server">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Address *</label>
            <input type="text" value={turnForm.address} onChange={(e) => setTurnForm({ ...turnForm, address: e.target.value })}
              className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none"
              placeholder="turn:turn.example.com:3478" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username</label>
              <input type="text" value={turnForm.username} onChange={(e) => setTurnForm({ ...turnForm, username: e.target.value })}
                className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Password</label>
              <input type="password" value={turnForm.password} onChange={(e) => setTurnForm({ ...turnForm, password: e.target.value })}
                className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Protocol</label>
              <select value={turnForm.protocol} onChange={(e) => setTurnForm({ ...turnForm, protocol: e.target.value })}
                className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none">
                <option value="udp">UDP</option>
                <option value="tcp">TCP</option>
                <option value="tls">TLS</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Bandwidth (Mbps)</label>
              <input type="number" value={turnForm.bandwidth} onChange={(e) => setTurnForm({ ...turnForm, bandwidth: Number(e.target.value) })}
                className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Realm</label>
              <input type="text" value={turnForm.realm} onChange={(e) => setTurnForm({ ...turnForm, realm: e.target.value })}
                className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Region</label>
              <input type="text" value={turnForm.region} onChange={(e) => setTurnForm({ ...turnForm, region: e.target.value })}
                className="w-full px-3 py-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500/50 outline-none" />
            </div>
          </div>
          <button onClick={addTURNServer} disabled={!turnForm.address}
            className="w-full py-3 rounded-xl bg-indigo-500 hover:bg-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium transition-all">
            Add TURN Server
          </button>
        </div>
      </Modal>
    </div>
  )
}
