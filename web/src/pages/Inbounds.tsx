import { useState, useEffect } from 'react'
import api from '../api/client'
import { formatBytes } from '../utils/format'

interface Inbound {
  id: number
  tag: string
  protocol: string
  port: number
  remark: string
  enable: boolean
  status?: string
  up_mbps?: number
  down_mbps?: number
  traffic_up?: number
  traffic_down?: number
}

// ─── Protocol Badge ──────────────────────────────────────────────────
function ProtocolBadge({ protocol }: { protocol: string }) {
  const colors: Record<string, string> = {
    vmess: 'badge-cyan',
    vless: 'badge-purple',
    trojan: 'badge-success',
    shadowsocks: 'badge-warning',
    http: 'badge',
    socks: 'badge',
  }
  const cls = colors[protocol.toLowerCase()] || 'badge-purple'
  return <span className={`badge ${cls}`}>{protocol}</span>
}

// ─── Status Dot ──────────────────────────────────────────────────────
function StatusDot({ active }: { active: boolean }) {
  return (
    <span className={`inline-flex items-center gap-1.5 text-xs font-medium ${active ? 'text-[var(--success)]' : 'text-[var(--danger)]'}`}>
      <span className={`notification-dot ${active ? 'online' : 'danger'}`} />
      {active ? 'Active' : 'Inactive'}
    </span>
  )
}

// ─── Inbounds Page ───────────────────────────────────────────────────
export function InboundsPage() {
  const [inbounds, setInbounds] = useState<Inbound[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchInbounds = async () => {
      try {
        const res = await api.getInbounds()
        setInbounds(res.data.inbounds || [])
      } catch (err) {
        console.error('Failed to fetch inbounds:', err)
      } finally {
        setLoading(false)
      }
    }
    fetchInbounds()
  }, [])

  return (
    <div className="space-y-6 page-enter">
      {/* ── Header ──────────────────────────────────────────────── */}
      <div className="glass-panel p-5">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
              </svg>
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-xl font-bold text-[var(--text-primary)]">Inbounds</h1>
                <span className="badge badge-cyan">{inbounds.length} total</span>
              </div>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage proxy inbound connections</p>
            </div>
          </div>
          <button className="btn-primary text-sm">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            New Inbound
          </button>
        </div>
      </div>

      {/* ── Table ────────────────────────────────────────────────── */}
      <div className="glass-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="table-modern">
            <thead>
              <tr>
                <th>ID</th>
                <th>Tag</th>
                <th>Protocol</th>
                <th>Port</th>
                <th>Remark</th>
                <th>Traffic</th>
                <th>Status</th>
                <th className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={8} className="p-12 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <div className="loading-spinner loading-spinner-lg" />
                      <p className="text-sm text-[var(--text-muted)]">Loading inbounds...</p>
                    </div>
                  </td>
                </tr>
              ) : inbounds.length === 0 ? (
                <tr>
                  <td colSpan={8}>
                    <div className="empty-state">
                      <div className="empty-state-icon">
                        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                        </svg>
                      </div>
                      <p className="empty-state-title">No inbounds yet</p>
                      <p className="empty-state-text">Create your first inbound to start proxying traffic.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                inbounds.map((ib, idx) => (
                  <tr key={ib.id} className={`animate-fade-in stagger-${(idx % 8) + 1}`}>
                    <td className="font-mono text-xs text-[var(--text-muted)]">#{ib.id}</td>
                    <td>
                      <span className="font-mono text-sm font-semibold text-[var(--text-primary)]">{ib.tag}</span>
                    </td>
                    <td><ProtocolBadge protocol={ib.protocol} /></td>
                    <td>
                      <span className="font-mono text-sm text-[var(--text-primary)]">{ib.port}</span>
                    </td>
                    <td className="text-sm text-[var(--text-secondary)] max-w-[200px] truncate">{ib.remark || '--'}</td>
                    <td>
                      {ib.traffic_up != null ? (
                        <span className="text-xs text-[var(--text-muted)] font-mono">
                          ↓{formatBytes(ib.traffic_down || 0)} ↑{formatBytes(ib.traffic_up || 0)}
                        </span>
                      ) : (
                        <span className="text-xs text-[var(--text-muted)]">--</span>
                      )}
                    </td>
                    <td><StatusDot active={ib.enable} /></td>
                    <td className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button className="p-1.5 rounded-lg bg-[rgba(139,92,246,0.06)] hover:bg-[rgba(139,92,246,0.12)] text-[var(--text-muted)] hover:text-[var(--accent-purple)] transition-all" title="Edit">
                          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button className="p-1.5 rounded-lg bg-[rgba(239,68,68,0.06)] hover:bg-[rgba(239,68,68,0.12)] text-[var(--text-muted)] hover:text-[var(--danger)] transition-all" title="Delete">
                          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
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
