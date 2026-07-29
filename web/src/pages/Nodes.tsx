import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface Node {
  id: number
  name: string
  address: string
  port: number
  status: string
  core_type: string
  location: string
  country: string
  enable: boolean
  cpu_load?: number
  memory_used?: number
  last_heartbeat?: number
  remark?: string
}

// ─── Status Indicator ────────────────────────────────────────────────
function NodeStatus({ status }: { status: string }) {
  const colors: Record<string, string> = {
    online: 'text-[var(--success)]',
    error: 'text-[var(--danger)]',
    offline: 'text-[var(--text-muted)]',
    connecting: 'text-[var(--warning)]',
  }
  const dots: Record<string, string> = {
    online: 'online',
    error: 'danger',
    offline: 'offline',
    connecting: 'warning',
  }
  const cls = colors[status.toLowerCase()] || 'text-[var(--text-muted)]'
  const dot = dots[status.toLowerCase()] || 'offline'
  return (
    <span className={`inline-flex items-center gap-1.5 text-xs font-medium ${cls}`}>
      <span className={`notification-dot ${dot}`} />
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  )
}

// ─── Core Badge ──────────────────────────────────────────────────────
function CoreBadge({ core }: { core: string }) {
  const c = core.toLowerCase()
  const cls = c === 'xray' ? 'badge-purple' : c === 'sing-box' ? 'badge-cyan' : c === 'hysteria2' ? 'badge-warning' : 'badge'
  return <span className={`badge ${cls}`}>{core}</span>
}

// ─── Mini Stat Card ──────────────────────────────────────────────────
function MiniStatCard({ label, value, icon, color }: { label: string; value: string | number; icon: string; color: string }) {
  return (
    <div className="glass-card p-4 flex items-center gap-3 animate-scale-in">
      <div className="w-9 h-9 rounded-lg flex items-center justify-center text-sm shrink-0" style={{ background: `${color}0.1`, color }}>
        {icon}
      </div>
      <div>
        <p className="text-[11px] font-medium text-[var(--text-muted)] uppercase tracking-wider">{label}</p>
        <p className="text-lg font-bold text-[var(--text-primary)]">{value}</p>
      </div>
    </div>
  )
}

// ─── Nodes Page ──────────────────────────────────────────────────────
export function NodesPage() {
  const [nodes, setNodes] = useState<Node[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiClient.get('/api/v1/nodes')
      .then(r => setNodes(r.data.nodes || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const onlineCount = nodes.filter(n => n.status === 'online').length
  const errorCount = nodes.filter(n => n.status === 'error').length
  const avgCpu = nodes.filter(n => n.cpu_load != null).reduce((s, n) => s + (n.cpu_load || 0), 0) / Math.max(nodes.filter(n => n.cpu_load != null).length, 1)

  return (
    <div className="space-y-6 page-enter">
      {/* ── Header ──────────────────────────────────────────────── */}
      <div className="glass-panel p-5">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-400 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
              </svg>
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-xl font-bold text-[var(--text-primary)]">Nodes</h1>
                <span className="badge badge-purple">{nodes.length} total</span>
              </div>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage proxy server nodes</p>
            </div>
          </div>
          <button className="btn-primary text-sm">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add Node
          </button>
        </div>
      </div>

      {/* ── Stats Row ───────────────────────────────────────────── */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <MiniStatCard label="Total" value={nodes.length} icon="🖥️" color="#8b5cf6" />
        <MiniStatCard label="Online" value={onlineCount} icon="🟢" color="#10b981" />
        <MiniStatCard label="Errors" value={errorCount} icon="🔴" color="#ef4444" />
        <MiniStatCard label="Avg CPU" value={nodes.filter(n => n.cpu_load != null).length > 0 ? `${avgCpu.toFixed(1)}%` : '--'} icon="⚡" color="#f59e0b" />
      </div>

      {/* ── Nodes Table ─────────────────────────────────────────── */}
      <div className="glass-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="table-modern">
            <thead>
              <tr>
                <th>Name</th>
                <th>Address</th>
                <th>Core</th>
                <th>Status</th>
                <th>CPU</th>
                <th>Memory</th>
                <th>Location</th>
                <th className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={8} className="p-12 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <div className="loading-spinner loading-spinner-lg" />
                      <p className="text-sm text-[var(--text-muted)]">Loading nodes...</p>
                    </div>
                  </td>
                </tr>
              ) : nodes.length === 0 ? (
                <tr>
                  <td colSpan={8}>
                    <div className="empty-state">
                      <div className="empty-state-icon">
                        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                        </svg>
                      </div>
                      <p className="empty-state-title">No nodes configured</p>
                      <p className="empty-state-text">Add your first node to start proxying traffic.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                nodes.map((node) => (
                  <tr key={node.id}>
                    <td>
                      <div className="flex items-center gap-2.5">
                        <div className="w-2 h-2 rounded-full" style={{
                          background: node.status === 'online' ? '#10b981' : node.status === 'error' ? '#ef4444' : '#585878',
                          boxShadow: node.status === 'online' ? '0 0 6px rgba(16,185,129,0.5)' : 'none',
                        }} />
                        <span className="font-medium text-sm text-[var(--text-primary)]">{node.name}</span>
                      </div>
                    </td>
                    <td>
                      <span className="font-mono text-xs text-[var(--text-secondary)]">{node.address}:{node.port}</span>
                    </td>
                    <td><CoreBadge core={node.core_type} /></td>
                    <td><NodeStatus status={node.status} /></td>
                    <td>
                      <div className="flex items-center gap-2">
                        <div className="progress-bar w-16">
                          <div className="progress-bar-fill" style={{ width: `${Math.min(node.cpu_load || 0, 100)}%` }} />
                        </div>
                        <span className="text-xs font-mono text-[var(--text-muted)]">{node.cpu_load != null ? `${node.cpu_load.toFixed(0)}%` : '--'}</span>
                      </div>
                    </td>
                    <td className="text-sm text-[var(--text-secondary)]">
                      {node.memory_used != null ? `${(node.memory_used / 1024 / 1024).toFixed(0)} MB` : '--'}
                    </td>
                    <td className="text-sm text-[var(--text-secondary)]">{node.location || node.country || <span className="text-[var(--text-muted)]">--</span>}</td>
                    <td className="text-right">
                      <button className="p-1.5 rounded-lg bg-[rgba(139,92,246,0.06)] hover:bg-[rgba(139,92,246,0.12)] text-[var(--text-muted)] hover:text-[var(--accent-purple)] transition-all" title="Edit">
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                          <path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
