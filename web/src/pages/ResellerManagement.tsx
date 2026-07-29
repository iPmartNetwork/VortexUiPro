import { useState, useEffect } from 'react'
import { apiGet } from '../api/client'

interface ResellerInfo {
  id: number
  username: string
  role: string
  traffic_limit: number
  user_limit: number
  client_count: number
  total_traffic: number
}

export function ResellerManagementPage() {
  const [resellers, setResellers] = useState<ResellerInfo[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => { fetchData() }, [])

  const fetchData = async () => {
    try {
      setLoading(true)
      const { data } = await apiGet('/api/v1/resellers/stats')
      setResellers(data.resellers || [])
    } catch {} finally { setLoading(false) }
  }

  const totalClients = resellers.reduce((s, r) => s + r.client_count, 0)
  const totalTraffic = resellers.reduce((s, r) => s + r.total_traffic, 0)

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Reseller Management</h1>
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">Monitor reseller accounts and resource usage</p>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-3">
        <div className="glass-card p-5">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wide mb-1">Resellers</p>
          <p className="text-2xl font-bold text-purple-400">{resellers.length}</p>
        </div>
        <div className="glass-card p-5">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wide mb-1">Total Clients</p>
          <p className="text-2xl font-bold text-green-400">{totalClients}</p>
        </div>
        <div className="glass-card p-5">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wide mb-1">Total Traffic</p>
          <p className="text-2xl font-bold text-yellow-400">{(totalTraffic / 1e9).toFixed(1)} GB</p>
        </div>
      </div>

      {/* Table */}
      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="w-8 h-8 border-2 border-purple-500/30 border-t-purple-500 rounded-full animate-spin" />
        </div>
      ) : (
        <div className="glass-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="table-modern">
              <thead>
                <tr>
                  <th>Username</th>
                  <th>Role</th>
                  <th>Clients</th>
                  <th>Traffic Used</th>
                  <th>Traffic Limit</th>
                  <th>User Limit</th>
                  <th className="text-right">Usage</th>
                </tr>
              </thead>
              <tbody>
                {resellers.map((r) => {
                  const usagePct = r.traffic_limit > 0 ? ((r.total_traffic / r.traffic_limit) * 100).toFixed(1) : '∞'
                  const userPct = r.user_limit > 0 ? ((r.client_count / r.user_limit) * 100).toFixed(1) : '∞'
                  return (
                    <tr key={r.id}>
                      <td className="text-[var(--text-primary)] font-medium">{r.username}</td>
                      <td><span className="badge badge-purple">{r.role}</span></td>
                      <td className="text-[var(--text-primary)]">{r.client_count}</td>
                      <td className="text-[var(--text-secondary)]">{(r.total_traffic / 1e9).toFixed(2)} GB</td>
                      <td className="text-[var(--text-secondary)]">{r.traffic_limit > 0 ? `${(r.traffic_limit / 1e9).toFixed(1)} GB` : '∞'}</td>
                      <td className="text-[var(--text-secondary)]">{r.user_limit > 0 ? r.user_limit : '∞'}</td>
                      <td className="text-right">
                        <div className={`text-xs ${
                          parseFloat(usagePct) > 90 ? 'text-red-400' : parseFloat(usagePct) > 70 ? 'text-yellow-400' : 'text-[var(--text-muted)]'
                        }`}>
                          Traffic: {usagePct}% | Users: {userPct}%
                        </div>
                      </td>
                    </tr>
                  )
                })}
                {resellers.length === 0 && (
                  <tr>
                    <td colSpan={7} className="text-center py-12 text-[var(--text-muted)] text-sm">No resellers found</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
