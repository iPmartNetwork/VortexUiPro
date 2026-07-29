import { useState, useEffect, useRef } from 'react'
import { apiGet } from '../api/client'

interface OnlineUser {
  client_id: string
  email: string
  inbound_id: number
  ip: string
  device: string
  connected_at: number
  last_active: number
  traffic_up: number
  traffic_down: number
}

interface ActivityRecord {
  client_id: string
  email: string
  action: string
  ip: string
  timestamp: number
}

export function OnlineMonitoringPage() {
  const [online, setOnline] = useState<OnlineUser[]>([])
  const [total, setTotal] = useState(0)
  const [activities, setActivities] = useState<ActivityRecord[]>([])
  const [byInbound, setByInbound] = useState<Record<string, number>>({})
  const [tab, setTab] = useState<'online' | 'activity'>('online')
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 15000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`)
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        if (msg.type === 'online_count') {
          setTotal(msg.payload?.online || 0)
        }
      } catch {}
    }
    wsRef.current = ws
    return () => ws.close()
  }, [])

  const fetchData = async () => {
    try {
      const [onlineRes, activityRes] = await Promise.all([
        apiGet('/api/v1/monitor/online'),
        apiGet('/api/v1/monitor/activity?limit=30'),
      ])
      setOnline(onlineRes.data.online || [])
      setTotal(onlineRes.data.total || 0)
      setByInbound(onlineRes.data.by_inbound || {})
      setActivities(activityRes.data.activities || [])
    } catch {}
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Online Monitoring</h1>
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">Real-time user activity tracking</p>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-3">
        <div className="glass-card p-5">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wide mb-1">Online Now</p>
          <p className="text-2xl font-bold text-green-400">{total}</p>
        </div>
        <div className="glass-card p-5">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wide mb-1">Total Users Today</p>
          <p className="text-2xl font-bold text-purple-400">{activities.length}</p>
        </div>
        <div className="glass-card p-5">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wide mb-1">Active Inbounds</p>
          <p className="text-2xl font-bold text-yellow-400">{Object.keys(byInbound).length}</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-2">
        <button
          onClick={() => setTab('online')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            tab === 'online'
              ? 'bg-purple-600 text-white shadow-lg shadow-purple-600/20'
              : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] hover:text-[var(--text-primary)] border border-[var(--border-light)]'
          }`}
        >
          Online Users ({total})
        </button>
        <button
          onClick={() => setTab('activity')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
            tab === 'activity'
              ? 'bg-purple-600 text-white shadow-lg shadow-purple-600/20'
              : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] hover:text-[var(--text-primary)] border border-[var(--border-light)]'
          }`}
        >
          Recent Activity
        </button>
      </div>

      {/* Online Users Table */}
      {tab === 'online' && (
        <div className="glass-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="table-modern">
              <thead>
                <tr>
                  <th>Email</th>
                  <th>IP</th>
                  <th>Device</th>
                  <th>Inbound</th>
                  <th>Connected</th>
                  <th>Traffic</th>
                </tr>
              </thead>
              <tbody>
                {online.map((u) => (
                  <tr key={u.email}>
                    <td>
                      <span className="flex items-center gap-2">
                        <span className="w-2 h-2 rounded-full bg-green-500" />
                        <span className="text-[var(--text-primary)]">{u.email}</span>
                      </span>
                    </td>
                    <td className="font-mono text-xs text-[var(--text-secondary)]">{u.ip || '-'}</td>
                    <td className="text-[var(--text-secondary)]">{u.device || '-'}</td>
                    <td><span className="badge badge-purple">#{u.inbound_id}</span></td>
                    <td className="text-xs text-[var(--text-muted)]">{new Date(u.connected_at).toLocaleTimeString()}</td>
                    <td className="text-xs text-[var(--text-secondary)]">
                      ↑ {(u.traffic_up / 1e9).toFixed(1)}G ↓ {(u.traffic_down / 1e9).toFixed(1)}G
                    </td>
                  </tr>
                ))}
                {online.length === 0 && (
                  <tr>
                    <td colSpan={6} className="text-center py-12 text-[var(--text-muted)] text-sm">No users online</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Activity Table */}
      {tab === 'activity' && (
        <div className="glass-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="table-modern">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Email</th>
                  <th>Action</th>
                  <th>IP</th>
                </tr>
              </thead>
              <tbody>
                {activities.map((a, i) => (
                  <tr key={i}>
                    <td className="text-xs text-[var(--text-muted)]">{new Date(a.timestamp).toLocaleTimeString()}</td>
                    <td className="text-[var(--text-primary)]">{a.email}</td>
                    <td>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                        a.action === 'connect' ? 'bg-green-500/10 text-green-400' :
                        a.action === 'disconnect' ? 'bg-red-500/10 text-red-400' :
                        'bg-blue-500/10 text-blue-400'
                      }`}>{a.action}</span>
                    </td>
                    <td className="font-mono text-xs text-[var(--text-secondary)]">{a.ip || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
