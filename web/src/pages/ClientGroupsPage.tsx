import { useState, useEffect } from 'react'
import { apiClient, apiGet, apiPost, apiPut, apiDelete } from '../api/client'

interface ClientGroup {
  id: number; name: string; description?: string; member_count: number; created_at: number
}

interface GroupedClient {
  id: string; email: string; enable: boolean; groups: string[]; inbound_id: number; user_id: number
}

export function ClientGroupsPage() {
  const [groups, setGroups] = useState<ClientGroup[]>([])
  const [allClients, setAllClients] = useState<GroupedClient[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedGroup, setSelectedGroup] = useState<ClientGroup | null>(null)
  const [groupClients, setGroupClients] = useState<GroupedClient[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [showAddClients, setShowAddClients] = useState(false)
  const [form, setForm] = useState({ name: '', description: '' })
  const [selectedClientIDs, setSelectedClientIDs] = useState<string[]>([])

  const fetchAll = async () => {
    try {
      const [g, c] = await Promise.all([
        apiGet('/api/v1/client-groups'),
        apiGet('/api/v1/client-groups/with-clients'),
      ])
      setGroups(g.data.groups || [])
      setAllClients(c.data.clients || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  useEffect(() => { fetchAll() }, [])

  const selectGroup = async (g: ClientGroup) => {
    setSelectedGroup(g)
    try {
      const r = await apiGet(`/api/v1/client-groups/${g.id}`)
      setGroupClients(r.data.clients || [])
    } catch { setGroupClients([]) }
  }

  const createGroup = async () => {
    try {
      await apiPost('/api/v1/client-groups', form)
      setShowCreate(false); setForm({ name: '', description: '' }); fetchAll()
    } catch { }
  }

  const deleteGroup = async (id: number) => {
    if (!confirm('Delete this group?')) return
    await apiDelete(`/api/v1/client-groups/${id}`)
    if (selectedGroup?.id === id) setSelectedGroup(null)
    fetchAll()
  }

  const addClients = async () => {
    if (!selectedGroup || selectedClientIDs.length === 0) return
    try {
      await apiPost(`/api/v1/client-groups/${selectedGroup.id}/clients/bulk`, { client_ids: selectedClientIDs })
      setShowAddClients(false); setSelectedClientIDs([]); selectGroup(selectedGroup)
    } catch { }
  }

  const removeClient = async (clientID: string) => {
    if (!selectedGroup) return
    await apiDelete(`/api/v1/client-groups/${selectedGroup.id}/clients?client_id=${clientID}`)
    selectGroup(selectedGroup)
  }

  const toggleClientSelect = (id: string) => {
    setSelectedClientIDs(prev => prev.includes(id) ? prev.filter(c => c !== id) : [...prev, id])
  }

  if (loading) return (
    <div className="flex items-center justify-center h-[60vh]">
      <div className="loading-spinner loading-spinner-lg" />
    </div>
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-emerald-500 flex items-center justify-center shadow-lg">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Client Groups</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Organize clients into groups for bulk operations</p>
            </div>
          </div>
          <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">+ Create Group</button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Groups List */}
        <div className="glass-card p-5">
          <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Groups ({groups.length})</h2>
          {groups.length === 0 ? (
            <div className="text-center py-8 text-[var(--text-muted)] text-sm">No groups configured</div>
          ) : (
            <div className="space-y-2">
              {groups.map(g => (
                <div key={g.id}
                  onClick={() => selectGroup(g)}
                  className={`flex items-center gap-3 p-3 rounded-xl border cursor-pointer transition-all ${selectedGroup?.id === g.id ? 'bg-purple-500/10 border-purple-500/20' : 'bg-[rgba(255,255,255,0.02)] border-[var(--border-light)] hover:border-[rgba(139,92,246,0.12)]'}`}
                >
                  <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500/20 to-cyan-500/20 flex items-center justify-center text-xs font-bold text-purple-300">
                    {g.name[0]}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-[var(--text-primary)] truncate">{g.name}</p>
                    <p className="text-[10px] text-[var(--text-muted)]">{g.member_count} members</p>
                  </div>
                  <button onClick={(e) => { e.stopPropagation(); deleteGroup(g.id) }} className="w-6 h-6 hover:bg-red-500/10 rounded flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 opacity-0 group-hover:opacity-100 transition-all">
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Group Details */}
        <div className="lg:col-span-2 glass-card p-5">
          {!selectedGroup ? (
            <div className="flex flex-col items-center justify-center h-64 text-[var(--text-muted)]">
              <svg className="w-16 h-16 mb-4 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857" />
              </svg>
              <p className="text-sm">Select a group to manage members</p>
            </div>
          ) : (
            <>
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-sm font-bold text-[var(--text-primary)]">{selectedGroup.name}</h2>
                  {selectedGroup.description && <p className="text-xs text-[var(--text-muted)]">{selectedGroup.description}</p>}
                </div>
                <div className="flex gap-2">
                  <button onClick={() => setShowAddClients(true)} className="btn-primary text-xs">+ Add Clients</button>
                  <button onClick={() => deleteGroup(selectedGroup.id)} className="btn-ghost text-xs text-red-400">Delete</button>
                </div>
              </div>
              <p className="text-xs text-[var(--text-secondary)] mb-3">{groupClients.length} clients in this group</p>
              <div className="space-y-2 max-h-[400px] overflow-y-auto custom-scrollbar">
                {groupClients.length === 0 ? (
                  <div className="text-center py-8 text-[var(--text-muted)] text-sm">No clients in this group</div>
                ) : groupClients.map((c: any) => (
                  <div key={c.id || c.email} className="flex items-center justify-between p-3 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                    <div className="flex items-center gap-3">
                      <div className={`w-2 h-2 rounded-full ${c.enable ? 'bg-emerald-400' : 'bg-[var(--text-muted)]'}`} />
                      <div>
                        <p className="text-sm font-medium text-[var(--text-primary)]">{c.email}</p>
                        <p className="text-[10px] text-[var(--text-muted)]">ID: {c.id?.slice(0, 8)}...</p>
                      </div>
                    </div>
                    <button onClick={() => removeClient(c.id)} className="w-7 h-7 hover:bg-red-500/10 rounded flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all">
                      <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Create Group Modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowCreate(false)}>
          <div className="glass-panel w-full max-w-md mx-4 p-6" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">Create Group</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Name *</label>
                <input type="text" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="input-modern text-sm w-full" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Description</label>
                <textarea value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className="input-modern text-sm w-full" rows={3} />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowCreate(false)} className="btn-ghost text-sm">Cancel</button>
              <button onClick={createGroup} className="btn-primary text-sm">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Add Clients Modal */}
      {showAddClients && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowAddClients(false)}>
          <div className="glass-panel w-full max-w-2xl mx-4 p-6" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">Add Clients to {selectedGroup?.name}</h2>
            <div className="space-y-2 max-h-[400px] overflow-y-auto custom-scrollbar mb-4">
              {allClients.filter(c => !groupClients.find((gc: any) => gc.id === c.id)).map(c => (
                <label key={c.id} className="flex items-center gap-3 p-3 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)] cursor-pointer hover:border-[rgba(139,92,246,0.12)] transition-all">
                  <input type="checkbox" checked={selectedClientIDs.includes(c.id)} onChange={() => toggleClientSelect(c.id)} className="rounded border-gray-600" />
                  <div className="flex-1">
                    <p className="text-sm font-medium text-[var(--text-primary)]">{c.email}</p>
                    <p className="text-[10px] text-[var(--text-muted)]">Groups: {c.groups.join(', ') || 'None'}</p>
                  </div>
                </label>
              ))}
            </div>
            <div className="flex justify-between items-center">
              <span className="text-xs text-[var(--text-muted)]">{selectedClientIDs.length} selected</span>
              <div className="flex gap-2">
                <button onClick={() => setShowAddClients(false)} className="btn-ghost text-sm">Cancel</button>
                <button onClick={addClients} className="btn-primary text-sm" disabled={selectedClientIDs.length === 0}>Add Selected</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
