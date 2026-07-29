import { useState, useEffect } from 'react'
import { apiGet, apiPost } from '../api/client'

interface UserWithTraffic {
  id: number
  username: string
  email: string
  traffic_up: number
  traffic_down: number
  data_limit: number
  status: string
  expiry_time: number
}

export function TrafficManagementPage() {
  const [users, setUsers] = useState<UserWithTraffic[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [message, setMessage] = useState('')

  useEffect(() => { fetchUsers() }, [])

  const fetchUsers = async () => {
    try {
      setLoading(true)
      const { data } = await apiGet('/api/v1/admin/users')
      setUsers(data.users || [])
    } catch {} finally { setLoading(false) }
  }

  const resetTraffic = async (userId: number) => {
    if (!confirm('Reset traffic for this user? This will re-enable all clients.')) return
    try {
      const { data } = await apiPost(`/api/v1/traffic/reset/${userId}`)
      setMessage(`✅ ${data.message}`)
      fetchUsers()
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`)
    }
  }

  const syncTraffic = async (email: string) => {
    try {
      const { data } = await apiGet(`/api/v1/traffic/sync?email=${encodeURIComponent(email)}`)
      setMessage(data.synced ? `✅ Synced: ↑${(data.up/1e9).toFixed(2)}G ↓${(data.down/1e9).toFixed(2)}G` : `ℹ️ ${data.message}`)
      fetchUsers()
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Sync failed'}`)
    }
  }

  const filtered = users.filter(u =>
    u.username.toLowerCase().includes(search.toLowerCase()) ||
    (u.email || '').toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Traffic Management</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Reset and sync user traffic</p>
            </div>
          </div>
          <button onClick={fetchUsers} className="px-4 py-2 rounded-lg bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-surface)] border border-[var(--border-light)] transition text-sm font-medium">
            ↻ Refresh
          </button>
        </div>
      </div>

      {/* Message */}
      {message && (
        <div className={`px-4 py-3 rounded-lg text-sm flex items-center justify-between ${
          message.includes('❌') ? 'bg-red-500/5 border border-red-500/20 text-red-400' : 'bg-green-500/5 border border-green-500/20 text-green-400'
        }`}>
          <span>{message}</span>
          <button onClick={() => setMessage('')} className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition">&times;</button>
        </div>
      )}

      {/* Search */}
      <input
        placeholder="Search by username or email..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full px-4 py-3 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-xl text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm transition-all"
      />

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
                  <th>User</th>
                  <th>Traffic ↑</th>
                  <th>Traffic ↓</th>
                  <th>Limit</th>
                  <th>Usage</th>
                  <th>Status</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((u) => {
                  const usage = u.traffic_up + u.traffic_down
                  const pct = u.data_limit > 0 ? (usage / u.data_limit * 100).toFixed(1) : '∞'
                  return (
                    <tr key={u.id}>
                      <td>
                        <div className="text-[var(--text-primary)] font-medium">{u.username}</div>
                        <div className="text-xs text-[var(--text-muted)]">{u.email}</div>
                      </td>
                      <td className="text-blue-400">↑ {(u.traffic_up / 1e9).toFixed(2)} GB</td>
                      <td className="text-yellow-400">↓ {(u.traffic_down / 1e9).toFixed(2)} GB</td>
                      <td className="text-[var(--text-secondary)]">{u.data_limit > 0 ? `${(u.data_limit / 1e9).toFixed(1)} GB` : '∞'}</td>
                      <td className={`font-medium ${
                        parseFloat(pct) > 90 ? 'text-red-400' : parseFloat(pct) > 70 ? 'text-yellow-400' : 'text-green-400'
                      }`}>{pct}%</td>
                      <td>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                          u.status === 'active' ? 'bg-green-500/10 text-green-400' :
                          u.status === 'limited' ? 'bg-yellow-500/10 text-yellow-400' :
                          'bg-red-500/10 text-red-400'
                        }`}>{u.status}</span>
                      </td>
                      <td className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button onClick={() => syncTraffic(u.email)}
                            className="px-3 py-1 rounded-md bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 transition text-[11px] font-medium">
                            Sync
                          </button>
                          <button onClick={() => resetTraffic(u.id)}
                            className="px-3 py-1 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition text-[11px] font-medium">
                            Reset
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
                {filtered.length === 0 && (
                  <tr>
                    <td colSpan={7} className="text-center py-12 text-[var(--text-muted)] text-sm">No users found</td>
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
