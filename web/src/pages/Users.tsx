import { useState, useEffect } from 'react'
import api from '../api/client'
import { formatBytes } from '../utils/format'

interface User {
  id: number
  username: string
  email: string
  status: string
  traffic_up: number
  traffic_down: number
  data_limit: number
  expiry_time: number
}

// ─── Status Badge ────────────────────────────────────────────────────
function StatusBadge({ status }: { status: string }) {
  const map: Record<string, { cls: string; label: string }> = {
    active: { cls: 'badge-success', label: 'Active' },
    limited: { cls: 'badge-warning', label: 'Limited' },
    expired: { cls: 'badge-danger', label: 'Expired' },
    disabled: { cls: 'badge-danger', label: 'Disabled' },
  }
  const s = map[status.toLowerCase()] || { cls: 'badge', label: status }
  return <span className={`badge ${s.cls}`}>{s.label}</span>
}

// ─── Traffic Bar ─────────────────────────────────────────────────────
function TrafficBar({ used, limit }: { used: number; limit: number }) {
  const pct = limit > 0 ? Math.min((used / limit) * 100, 100) : 0
  return (
    <div className="flex items-center gap-2">
      <div className="progress-bar flex-1 max-w-[100px]">
        <div className="progress-bar-fill" style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs text-[var(--text-muted)] font-mono whitespace-nowrap">
        {formatBytes(used)} / {limit > 0 ? formatBytes(limit) : '∞'}
      </span>
    </div>
  )
}

// ─── Users Page ──────────────────────────────────────────────────────
export function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [newUser, setNewUser] = useState({ username: '', email: '', data_limit: 0 })
  const [creating, setCreating] = useState(false)

  const fetchUsers = async () => {
    try {
      const res = await api.getUsers()
      setUsers(res.data.users || [])
    } catch (err) {
      console.error('Failed to fetch users:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchUsers() }, [])

  const handleCreate = async () => {
    if (!newUser.username.trim()) return
    setCreating(true)
    try {
      await api.createUser({
        ...newUser,
        data_limit: newUser.data_limit > 0 ? newUser.data_limit * 1073741824 : 0,
      })
      setShowCreate(false)
      setNewUser({ username: '', email: '', data_limit: 0 })
      fetchUsers()
    } catch (err) {
      console.error('Failed to create user:', err)
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this user?')) return
    try {
      await api.deleteUser(id)
      fetchUsers()
    } catch (err) {
      console.error('Failed to delete user:', err)
    }
  }

  return (
    <div className="space-y-6 page-enter">
      {/* ── Header ──────────────────────────────────────────────── */}
      <div className="glass-panel p-5">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-xl font-bold text-[var(--text-primary)]">Users</h1>
                <span className="badge badge-purple">{users.length} total</span>
              </div>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage panel users & clients</p>
            </div>
          </div>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="btn-primary text-sm"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            New User
          </button>
        </div>
      </div>

      {/* ── Create User Form ────────────────────────────────────── */}
      {showCreate && (
        <div className="glass-card p-6 animate-scale-in">
          <div className="flex items-center gap-3 mb-5">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-400 to-cyan-400 flex items-center justify-center">
              <svg className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
              </svg>
            </div>
            <h3 className="text-base font-bold text-[var(--text-primary)]">Create New User</h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Username *</label>
              <input
                type="text"
                placeholder="Enter username"
                value={newUser.username}
                onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                className="input-modern text-sm"
                autoFocus
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Email</label>
              <input
                type="email"
                placeholder="user@example.com"
                value={newUser.email}
                onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                className="input-modern text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Data Limit (GB)</label>
              <input
                type="number"
                placeholder="0 = Unlimited"
                value={newUser.data_limit || ''}
                onChange={(e) => setNewUser({ ...newUser, data_limit: Number(e.target.value) })}
                className="input-modern text-sm"
                min={0}
              />
            </div>
          </div>
          <div className="flex items-center gap-2.5 mt-5">
            <button onClick={handleCreate} disabled={creating || !newUser.username.trim()} className="btn-primary text-sm">
              {creating ? (
                <><span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" /> Creating...</>
              ) : (
                <><svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg> Create User</>
              )}
            </button>
            <button onClick={() => setShowCreate(false)} className="btn-ghost text-sm">Cancel</button>
          </div>
        </div>
      )}

      {/* ── Users Table ─────────────────────────────────────────── */}
      <div className="glass-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="table-modern">
            <thead>
              <tr>
                <th>ID</th>
                <th>Username</th>
                <th>Email</th>
                <th>Status</th>
                <th>Traffic Usage</th>
                <th>Data Limit</th>
                <th className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={7} className="p-12 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <div className="loading-spinner loading-spinner-lg" />
                      <p className="text-sm text-[var(--text-muted)]">Loading users...</p>
                    </div>
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={7}>
                    <div className="empty-state">
                      <div className="empty-state-icon">
                        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                        </svg>
                      </div>
                      <p className="empty-state-title">No users found</p>
                      <p className="empty-state-text">Create your first user to get started.</p>
                    </div>
                  </td>
                </tr>
              ) : (
                users.map((user) => (
                  <tr key={user.id}>
                    <td className="font-mono text-xs text-[var(--text-muted)]">#{user.id}</td>
                    <td>
                      <div className="flex items-center gap-2.5">
                        <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-[10px] font-bold">
                          {user.username[0].toUpperCase()}
                        </div>
                        <span className="font-medium text-sm text-[var(--text-primary)]">{user.username}</span>
                      </div>
                    </td>
                    <td className="text-sm text-[var(--text-secondary)]">{user.email || <span className="text-[var(--text-muted)]">--</span>}</td>
                    <td><StatusBadge status={user.status} /></td>
                    <td>
                      <TrafficBar used={user.traffic_up + user.traffic_down} limit={user.data_limit} />
                    </td>
                    <td className="text-sm text-[var(--text-secondary)]">
                      {user.data_limit > 0 ? formatBytes(user.data_limit) : <span className="text-[var(--text-muted)]">Unlimited</span>}
                    </td>
                    <td className="text-right">
                      <button
                        onClick={() => handleDelete(user.id)}
                        className="p-1.5 rounded-lg bg-[rgba(239,68,68,0.06)] hover:bg-[rgba(239,68,68,0.12)] text-[var(--text-muted)] hover:text-[var(--danger)] transition-all"
                        title="Delete user"
                      >
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                          <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
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
