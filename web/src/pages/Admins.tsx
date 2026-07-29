import { useState, useEffect } from 'react'
import { apiGet, apiPost, apiPut, apiDelete } from '../api/client'

interface AdminView {
  id: number
  username: string
  email: string
  role: string
  roleId: number
  roleName: string
  status: string
  created_at: number
}

export function AdminsPage() {
  const [admins, setAdmins] = useState<AdminView[]>([])
  const [roles, setRoles] = useState<{ id: number; name: string; slug: string }[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [selectedRole, setSelectedRole] = useState(0)

  useEffect(() => { fetchData() }, [])

  const fetchData = async () => {
    try { setLoading(true); const [a, r] = await Promise.all([apiGet('/api/v1/admins'), apiGet('/api/v1/roles')]); setAdmins(a.data.admins || []); setRoles(r.data.roles || []) }
    catch (err: any) { setError(err?.response?.data?.error || 'Failed to load data') }
    finally { setLoading(false) }
  }

  const createAdmin = async () => {
    try { await apiPost('/api/v1/admins', { username, password, email, roleId: selectedRole }); setShowCreate(false); setUsername(''); setPassword(''); setEmail(''); fetchData() }
    catch (err: any) { setError(err?.response?.data?.error || 'Failed to create admin') }
  }

  const toggleStatus = async (id: number, enabled: boolean) => { try { await apiPut(`/api/v1/admins/${id}/status`, { enabled }); fetchData() } catch (err: any) { setError(err?.response?.data?.error || 'Failed to update status') } }
  const deleteAdmin = async (id: number) => { if (!confirm('Delete this admin?')) return; try { await apiDelete(`/api/v1/admins/${id}`); fetchData() } catch (err: any) { setError(err?.response?.data?.error || 'Failed to delete admin') } }

  return (
    <div className="page-enter max-w-[1000px] mx-auto py-6 px-4 space-y-6">
      <div className="glass-panel p-5 flex items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
          </div>
          <div><h1 className="text-xl font-bold text-[var(--text-primary)]">Admins</h1><p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage panel administrators</p></div>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">+ New Admin</button>
      </div>

      {error && <div className="glass-card p-3 flex items-center justify-between border-red-500/20"><span className="text-sm text-red-400">{error}</span><button onClick={() => setError('')} className="text-red-400 hover:text-red-300">×</button></div>}

      {loading ? <div className="flex justify-center py-12"><div className="loading-spinner loading-spinner-lg" /></div> : (
        <div className="glass-card overflow-hidden">
          <table className="table-modern">
            <thead><tr><th>ID</th><th>Username</th><th>Role</th><th>Status</th><th>Created</th><th className="text-right">Actions</th></tr></thead>
            <tbody>{admins.map((admin) => (
              <tr key={admin.id}>
                <td className="font-mono text-xs text-[var(--text-muted)]">{admin.id}</td>
                <td><div className="text-sm font-medium text-[var(--text-primary)]">{admin.username}</div>{admin.email && <div className="text-xs text-[var(--text-muted)]">{admin.email}</div>}</td>
                <td><span className="badge badge-purple text-[10px]">{admin.roleName || admin.role}</span></td>
                <td><span className={`badge ${admin.status === 'active' ? 'badge-success' : 'badge-danger'}`}>{admin.status}</span></td>
                <td className="text-sm text-[var(--text-secondary)]">{new Date(admin.created_at).toLocaleDateString()}</td>
                <td className="text-right"><div className="flex items-center justify-end gap-1.5">
                  <button onClick={() => toggleStatus(admin.id, admin.status !== 'active')} className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-all ${admin.status === 'active' ? 'bg-amber-500/10 text-amber-400 hover:bg-amber-500/20' : 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20'}`}>{admin.status === 'active' ? 'Disable' : 'Enable'}</button>
                  <button onClick={() => deleteAdmin(admin.id)} className="px-2.5 py-1 rounded-lg text-xs font-medium bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-all">Delete</button>
                </div></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 z-50" onClick={() => setShowCreate(false)}>
          <div className="glass-card p-7 w-full max-w-[450px]" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">Create Admin</h2>
            <div className="space-y-3">
              <input placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} className="input-modern text-sm" />
              <input type="password" placeholder="Password (min 6 chars)" value={password} onChange={(e) => setPassword(e.target.value)} className="input-modern text-sm" />
              <input placeholder="Email (optional)" value={email} onChange={(e) => setEmail(e.target.value)} className="input-modern text-sm" />
              <select value={selectedRole} onChange={(e) => setSelectedRole(Number(e.target.value))} className="select-modern w-full text-sm">
                <option value={0}>Select Role</option>
                {roles.filter((r) => !r.slug.includes('owner')).map((r) => (<option key={r.id} value={r.id}>{r.name}</option>))}
              </select>
            </div>
            <div className="flex justify-end gap-2.5 mt-5">
              <button onClick={() => setShowCreate(false)} className="btn-ghost text-sm">Cancel</button>
              <button onClick={createAdmin} disabled={!username || !password} className="btn-primary text-sm disabled:opacity-50">{!username || !password ? 'Fill required fields' : 'Create'}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
