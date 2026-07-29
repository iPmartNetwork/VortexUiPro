import { useState, useEffect } from 'react'
import { apiClient, apiGet, apiPost, apiPut, apiDelete } from '../api/client'

interface FederationProvider {
  id: number; name: string; api_url: string; api_key: string
  sync_users: boolean; sync_plans: boolean; sync_traffic: boolean
  status: string; last_sync_at: number
}

export function FederationPage() {
  const [providers, setProviders] = useState<FederationProvider[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', api_url: '', api_key: '', sync_users: true, sync_plans: true, sync_traffic: false })
  const [syncing, setSyncing] = useState(false)

  useEffect(() => { fetchProviders() }, [])

  const fetchProviders = async () => {
    try {
      const r = await apiGet('/api/v1/federation/providers')
      setProviders(r.data.providers || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  const createProvider = async () => {
    try {
      await apiPost('/api/v1/federation/providers', form)
      setShowCreate(false)
      setForm({ name: '', api_url: '', api_key: '', sync_users: true, sync_plans: true, sync_traffic: false })
      fetchProviders()
    } catch { }
  }

  const deleteProvider = async (id: number) => {
    if (!confirm('Remove this federation connection?')) return
    await apiDelete(`/api/v1/federation/providers/${id}`)
    fetchProviders()
  }

  const triggerSync = async (id?: number) => {
    setSyncing(true)
    try {
      if (id) await apiPost(`/api/v1/federation/sync/${id}`)
      else await apiPost('/api/v1/federation/sync')
      setTimeout(() => { setSyncing(false); fetchProviders() }, 3000)
    } catch { setSyncing(false) }
  }

  const formatTime = (ts: number) => {
    if (!ts) return 'Never'
    return new Date(ts).toLocaleString()
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
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center shadow-lg">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Federation</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Connect multiple VortexUiPro panels for cross-instance sync</p>
            </div>
          </div>
          <div className="flex gap-2">
            <button onClick={() => triggerSync()} disabled={syncing} className="btn-ghost text-xs">
              {syncing ? 'Syncing...' : 'Sync All'}
            </button>
            <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">+ Add Provider</button>
          </div>
        </div>
      </div>

      {/* Providers */}
      <div className="glass-card p-5">
        {providers.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-[var(--text-muted)]">
            <svg className="w-16 h-16 mb-4 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
            <p className="text-sm font-medium mb-1">No federation providers</p>
            <p className="text-xs">Connect remote VortexUiPro panels to sync users & plans</p>
          </div>
        ) : (
          <div className="space-y-3">
            {providers.map(p => (
              <div key={p.id} className="flex items-center gap-4 p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)] hover:border-[rgba(139,92,246,0.12)] transition-all">
                <div className={`w-3 h-3 rounded-full shrink-0 ${p.status === 'online' ? 'bg-emerald-400 shadow-[0_0_8px_rgba(16,185,129,0.4)]' : p.status === 'error' ? 'bg-red-400' : 'bg-[var(--text-muted)]'}`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold text-[var(--text-primary)]">{p.name}</h3>
                    <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${p.status === 'online' ? 'bg-emerald-500/10 text-emerald-300' : 'bg-[rgba(255,255,255,0.03)] text-[var(--text-muted)]'}`}>{p.status}</span>
                  </div>
                  <p className="text-xs text-[var(--text-secondary)] mt-0.5">{p.api_url}</p>
                  <div className="flex gap-3 mt-1.5">
                    <span className={`text-[10px] ${p.sync_users ? 'text-emerald-300' : 'text-[var(--text-muted)]'}`}>Users {p.sync_users ? '✓' : '✗'}</span>
                    <span className={`text-[10px] ${p.sync_plans ? 'text-emerald-300' : 'text-[var(--text-muted)]'}`}>Plans {p.sync_plans ? '✓' : '✗'}</span>
                    <span className={`text-[10px] ${p.sync_traffic ? 'text-emerald-300' : 'text-[var(--text-muted)]'}`}>Traffic {p.sync_traffic ? '✓' : '✗'}</span>
                  </div>
                  <p className="text-[10px] text-[var(--text-muted)] mt-1">Last sync: {formatTime(p.last_sync_at)}</p>
                </div>
                <div className="flex gap-1.5">
                  <button onClick={() => triggerSync(p.id)} className="w-7 h-7 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(6,182,212,0.1)] flex items-center justify-center text-[var(--text-muted)] hover:text-cyan-300 transition-all" title="Sync now">
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                  </button>
                  <button onClick={() => deleteProvider(p.id)} className="w-7 h-7 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-red-500/10 flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all" title="Remove provider">
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowCreate(false)}>
          <div className="glass-panel w-full max-w-lg mx-4 p-6" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">Add Federation Provider</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Provider Name *</label>
                <input value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className="input-modern text-sm w-full" placeholder="e.g. Secondary Panel" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">API URL *</label>
                <input value={form.api_url} onChange={e => setForm(f => ({ ...f, api_url: e.target.value }))} className="input-modern text-sm w-full" placeholder="https://panel2.example.com" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">API Key</label>
                <input value={form.api_key} onChange={e => setForm(f => ({ ...f, api_key: e.target.value }))} className="input-modern text-sm w-full" placeholder="Shared secret key" />
              </div>
              <div className="space-y-2 pt-2 border-t border-[var(--border-light)]">
                <p className="text-xs font-medium text-[var(--text-secondary)]">Sync Settings</p>
                {([
                  { key: 'sync_users' as const, label: 'Sync Users' },
                  { key: 'sync_plans' as const, label: 'Sync Plans' },
                  { key: 'sync_traffic' as const, label: 'Sync Traffic' },
                ]).map(s => (
                  <label key={s.key} className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" checked={form[s.key]} onChange={e => setForm(f => ({ ...f, [s.key]: e.target.checked }))} className="rounded border-gray-600" />
                    <span className="text-sm text-[var(--text-secondary)]">{s.label}</span>
                  </label>
                ))}
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowCreate(false)} className="btn-ghost text-sm">Cancel</button>
              <button onClick={createProvider} className="btn-primary text-sm">Add Provider</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
