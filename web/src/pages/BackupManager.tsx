import { useState, useEffect } from 'react'
import { apiGet, apiPost, apiDelete } from '../api/client'

interface BackupEntry {
  id: string
  name: string
  scope: string
  size: number
  admin_name: string
  encrypted: boolean
  encryption_key_id: number
  remote_storage_name: string
  telegram_sent: boolean
  created_at: number
  status: string
}

interface EncKey {
  id: number
  name: string
  active: boolean
  created_at: number
}

interface RemoteStorageCfg {
  id: number
  name: string
  type: string
  enabled: boolean
  s3_endpoint: string
  s3_region: string
  s3_bucket: string
  s3_prefix: string
  gdrive_folder_id: string
}

export function BackupManagerPage() {
  const [backups, setBackups] = useState<BackupEntry[]>([])
  const [encKeys, setEncKeys] = useState<EncKey[]>([])
  const [storageCfgs, setStorageCfgs] = useState<RemoteStorageCfg[]>([])
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [scope, setScope] = useState('full')
  const [adminId, setAdminId] = useState('')
  const [autoEnabled, setAutoEnabled] = useState(false)
  const [autoHours, setAutoHours] = useState(24)
  const [tab, setTab] = useState<'backups' | 'encryption' | 'storage'>('backups')
  const [showKeyModal, setShowKeyModal] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [showStorageModal, setShowStorageModal] = useState(false)
  const [storageForm, setStorageForm] = useState<Partial<RemoteStorageCfg>>({})
  const [syncingId, setSyncingId] = useState<string | null>(null)

  useEffect(() => { fetchData() }, [])

  const fetchData = async () => {
    try {
      setLoading(true)
      const [bkp, keys, storage] = await Promise.all([
        apiGet('/api/v1/backups'),
        apiGet('/api/v1/backups/encryption/keys').catch(() => ({ data: [] })),
        apiGet('/api/v1/backups/remote-storage').catch(() => ({ data: [] })),
      ])
      setBackups(bkp.data?.backups || [])
      setEncKeys(keys.data?.data || [])
      setStorageCfgs(storage.data?.data || [])
    } catch {} finally { setLoading(false) }
  }

  const createBackup = async () => {
    try {
      const payload: any = { name, scope }
      if (scope === 'reseller' && adminId) payload.admin_id = parseInt(adminId)
      const { data } = await apiPost('/api/v1/backups', payload)
      setMessage(`✅ Backup created: ${data.name || data.id}`)
      setShowCreate(false)
      fetchData()
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`)
    }
  }

  const downloadBackup = (id: string) => window.open(`/api/v1/backups/${id}/download`, '_blank')

  const deleteBackup = async (id: string) => {
    if (!confirm('Delete this backup?')) return
    try { await apiDelete(`/api/v1/backups/${id}`); setMessage('✅ Backup deleted'); fetchData() }
    catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const restoreBackup = async (id: string) => {
    if (!confirm('Restore from this backup? This will add data to the database.')) return
    try { await apiPost(`/api/v1/backups/${id}/restore`, { scope }); setMessage('✅ Restore initiated') }
    catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const syncToRemote = async (backupId: string, storageId: number) => {
    setSyncingId(backupId)
    try {
      await apiPost(`/api/v1/backups/${backupId}/sync`, { storage_id: storageId })
      setMessage('✅ Synced to remote storage')
      fetchData()
    } catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
    finally { setSyncingId(null) }
  }

  const sendToTelegram = async (backupId: string) => {
    try {
      await apiPost(`/api/v1/backups/${backupId}/telegram`, {})
      setMessage('✅ Telegram notification sent')
      fetchData()
    } catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const generateKey = async () => {
    try {
      const { data } = await apiPost('/api/v1/backups/encryption/keys', { name: keyName })
      setMessage(`✅ Key generated: ${data.data?.name || keyName}`)
      setShowKeyModal(false)
      setKeyName('')
      fetchData()
    } catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const deleteKey = async (id: number) => {
    if (!confirm('Delete this encryption key? Backups encrypted with it will not be decryptable.')) return
    try { await apiDelete(`/api/v1/backups/encryption/keys/${id}`); setMessage('✅ Key deleted'); fetchData() }
    catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const saveStorageConfig = async () => {
    try {
      await apiPost('/api/v1/backups/remote-storage', storageForm)
      setMessage('✅ Storage config saved')
      setShowStorageModal(false)
      setStorageForm({})
      fetchData()
    } catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const deleteStorage = async (id: number) => {
    if (!confirm('Delete this storage config?')) return
    try { await apiDelete(`/api/v1/backups/remote-storage/${id}`); setMessage('✅ Config deleted'); fetchData() }
    catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const saveAutoConfig = async () => {
    try {
      await apiPost('/api/v1/backups/auto-config', { enabled: autoEnabled, interval_hours: autoHours })
      setMessage(autoEnabled ? '✅ Auto-backup enabled' : '✅ Auto-backup disabled')
    } catch (err: any) { setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`) }
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024*1024) return `${(bytes/1024).toFixed(1)} KB`
    return `${(bytes/(1024*1024)).toFixed(1)} MB`
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M7 10l5 5 5-5M12 15V3" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Backup Manager</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Advanced backup with encryption, remote storage & Telegram</p>
            </div>
          </div>
          {tab === 'backups' && (
            <button onClick={() => setShowCreate(true)}
              className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 transition text-sm font-medium whitespace-nowrap">+ Create Backup</button>
          )}
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

      {/* Tabs */}
      <div className="flex gap-1 p-1 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] w-fit">
        <button onClick={() => setTab('backups')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'backups' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
          Backups
        </button>
        <button onClick={() => setTab('encryption')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'encryption' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
          Encryption Keys
        </button>
        <button onClick={() => setTab('storage')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'storage' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
          Remote Storage
        </button>
      </div>

      {/* Tab: Backups */}
      {tab === 'backups' && (
        <>
          {/* Auto-backup */}
          <div className="glass-card p-5">
            <h3 className="text-base font-bold text-[var(--text-primary)] mb-3">Auto-Backup</h3>
            <div className="flex flex-wrap items-center gap-3">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={autoEnabled} onChange={() => setAutoEnabled(!autoEnabled)}
                  className="w-4 h-4 rounded border-[var(--border-light)] text-purple-600 focus:ring-purple-500 bg-[var(--bg-elevated)]" />
                <span className="text-sm text-[var(--text-secondary)]">Enable</span>
              </label>
              <input type="number" value={autoHours} onChange={(e) => setAutoHours(Number(e.target.value))}
                className="w-20 px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] text-sm focus:border-purple-500 focus:outline-none" />
              <span className="text-sm text-[var(--text-muted)]">hours interval</span>
              <button onClick={saveAutoConfig}
                className="px-4 py-2 bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-[var(--border-light)] rounded-lg hover:bg-[var(--bg-elevated)] transition text-sm font-medium">Save</button>
            </div>
          </div>

          {/* Backup table */}
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
                      <th>Name</th>
                      <th>Scope</th>
                      <th>Size</th>
                      <th>Enc</th>
                      <th>Remote</th>
                      <th>Telegram</th>
                      <th>Status</th>
                      <th>Date</th>
                      <th className="text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {backups.map((b) => (
                      <tr key={b.id}>
                        <td className="text-[var(--text-primary)] font-medium">{b.name}</td>
                        <td>
                          <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                            b.scope === 'full' ? 'bg-green-500/10 text-green-400' :
                            b.scope === 'reseller' ? 'bg-yellow-500/10 text-yellow-400' :
                            'bg-blue-500/10 text-blue-400'
                          }`}>{b.scope}</span>
                        </td>
                        <td className="text-xs text-[var(--text-muted)]">{formatSize(b.size)}</td>
                        <td>
                          <span className={`text-xs ${b.encrypted ? 'text-emerald-400' : 'text-[#585878]'}`}>
                            {b.encrypted ? '🔒' : '—'}
                          </span>
                        </td>
                        <td className="text-xs text-[var(--text-muted)]">{b.remote_storage_name || '—'}</td>
                        <td>
                          <span className={`text-xs ${b.telegram_sent ? 'text-blue-400' : 'text-[#585878]'}`}>
                            {b.telegram_sent ? '📤' : '—'}
                          </span>
                        </td>
                        <td>
                          <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                            b.status === 'completed' ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'
                          }`}>{b.status}</span>
                        </td>
                        <td className="text-xs text-[var(--text-muted)]">{new Date(b.created_at).toLocaleString()}</td>
                        <td className="text-right">
                          <div className="flex items-center justify-end gap-1 flex-wrap">
                            <button onClick={() => downloadBackup(b.id)}
                              className="px-2 py-1 rounded-md bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 transition text-[10px] font-medium">DL</button>
                            <button onClick={() => restoreBackup(b.id)}
                              className="px-2 py-1 rounded-md bg-yellow-500/10 text-yellow-400 hover:bg-yellow-500/20 transition text-[10px] font-medium">Restore</button>
                            {storageCfgs.length > 0 && (
                              <button onClick={() => syncToRemote(b.id, storageCfgs[0].id)} disabled={syncingId === b.id}
                                className="px-2 py-1 rounded-md bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 transition text-[10px] font-medium disabled:opacity-50">
                                {syncingId === b.id ? '...' : '☁️'}
                              </button>
                            )}
                            <button onClick={() => sendToTelegram(b.id)}
                              className="px-2 py-1 rounded-md bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 transition text-[10px] font-medium">📤</button>
                            <button onClick={() => deleteBackup(b.id)}
                              className="px-2 py-1 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition text-[10px] font-medium">🗑</button>
                          </div>
                        </td>
                      </tr>
                    ))}
                    {backups.length === 0 && (
                      <tr><td colSpan={9} className="text-center py-12 text-[var(--text-muted)] text-sm">No backups yet</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}

      {/* Tab: Encryption Keys */}
      {tab === 'encryption' && (
        <div className="space-y-3">
          <div className="flex justify-end">
            <button onClick={() => setShowKeyModal(true)}
              className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 transition text-sm font-medium">+ Generate Key</button>
          </div>
          {encKeys.length === 0 ? (
            <div className="glass-card p-8 text-center text-[var(--text-muted)]">
              <p className="text-lg mb-2">🔐</p>
              <p>No encryption keys yet</p>
              <p className="text-sm mt-1">Generate a key to enable AES-256 backup encryption</p>
            </div>
          ) : (
            <div className="glass-card overflow-hidden">
              <table className="table-modern">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Active</th>
                    <th>Created</th>
                    <th className="text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {encKeys.map(k => (
                    <tr key={k.id}>
                      <td className="text-[var(--text-primary)] font-medium">{k.name}</td>
                      <td>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${k.active ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-500/10 text-slate-400'}`}>
                          {k.active ? 'Active' : 'Inactive'}
                        </span>
                      </td>
                      <td className="text-xs text-[var(--text-muted)]">{new Date(k.created_at).toLocaleString()}</td>
                      <td className="text-right">
                        <button onClick={() => deleteKey(k.id)}
                          className="px-3 py-1 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition text-[11px] font-medium">Delete</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Tab: Remote Storage */}
      {tab === 'storage' && (
        <div className="space-y-3">
          <div className="flex justify-end">
            <button onClick={() => { setStorageForm({ type: 's3', enabled: true }); setShowStorageModal(true) }}
              className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 transition text-sm font-medium">+ Add Storage</button>
          </div>
          {storageCfgs.length === 0 ? (
            <div className="glass-card p-8 text-center text-[var(--text-muted)]">
              <p className="text-lg mb-2">☁️</p>
              <p>No remote storage configured</p>
              <p className="text-sm mt-1">Add S3 or Google Drive for off-site backup sync</p>
            </div>
          ) : (
            storageCfgs.map(cfg => (
              <div key={cfg.id} className="glass-card p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className={`w-8 h-8 rounded-lg ${cfg.type === 's3' ? 'bg-orange-500/10 text-orange-400' : 'bg-blue-500/10 text-blue-400'} flex items-center justify-center font-bold text-xs`}>
                      {cfg.type === 's3' ? 'S3' : 'GD'}
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold text-white">{cfg.name}</h3>
                      <p className="text-xs text-[#6868a0]">{cfg.type === 's3' ? `${cfg.s3_bucket}` : `Google Drive`}</p>
                    </div>
                  </div>
                  <button onClick={() => deleteStorage(cfg.id)}
                    className="px-3 py-1.5 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition text-[11px] font-medium">Delete</button>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                  <div><span className="text-[#585878]">Endpoint:</span> <span className="text-white ml-1">{cfg.s3_endpoint || '—'}</span></div>
                  <div><span className="text-[#585878]">Region:</span> <span className="text-white ml-1">{cfg.s3_region || '—'}</span></div>
                  <div><span className="text-[#585878]">Prefix:</span> <span className="text-white ml-1">{cfg.s3_prefix || '—'}</span></div>
                  <div><span className="text-[#585878]">Status:</span> <span className={`ml-1 ${cfg.enabled ? 'text-emerald-400' : 'text-slate-400'}`}>{cfg.enabled ? 'Active' : 'Inactive'}</span></div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Create Backup Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
          <div className="glass-card p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-4">Create Backup</h2>
            <div className="space-y-3">
              <input placeholder="Backup Name" value={name} onChange={(e) => setName(e.target.value)}
                className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
              <select value={scope} onChange={(e) => setScope(e.target.value)}
                className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] focus:border-purple-500 focus:outline-none text-sm">
                <option value="full">Full System</option>
                <option value="system">System Only</option>
                <option value="users">Users & Data Only</option>
                <option value="reseller">Reseller Scope</option>
              </select>
              {scope === 'reseller' && (
                <input placeholder="Admin ID" value={adminId} onChange={(e) => setAdminId(e.target.value)}
                  className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
              )}
              {encKeys.length > 0 && (
                <p className="text-xs text-emerald-400">🔒 Backup will be auto-encrypted with AES-256</p>
              )}
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowCreate(false)}
                className="px-4 py-2 bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-[var(--border-light)] rounded-lg hover:bg-[var(--bg-elevated)] transition text-sm font-medium">Cancel</button>
              <button onClick={createBackup} disabled={!name}
                className="px-5 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Generate Key Modal */}
      {showKeyModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowKeyModal(false)}>
          <div className="glass-card p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-4">Generate AES-256 Key</h2>
            <p className="text-sm text-[var(--text-secondary)] mb-4">New backups will be automatically encrypted with this key. Old keys will be deactivated.</p>
            <input placeholder="Key Name (e.g., production-2026)" value={keyName} onChange={(e) => setKeyName(e.target.value)}
              className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowKeyModal(false)}
                className="px-4 py-2 bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-[var(--border-light)] rounded-lg hover:bg-[var(--bg-elevated)] transition text-sm font-medium">Cancel</button>
              <button onClick={generateKey} disabled={!keyName}
                className="px-5 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium">Generate</button>
            </div>
          </div>
        </div>
      )}

      {/* Storage Config Modal */}
      {showStorageModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowStorageModal(false)}>
          <div className="glass-card p-6 w-full max-w-lg mx-4" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-4">Add Remote Storage</h2>
            <div className="space-y-3">
              <input placeholder="Name" value={storageForm.name || ''} onChange={e => setStorageForm({...storageForm, name: e.target.value})}
                className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
              <select value={storageForm.type || 's3'} onChange={e => setStorageForm({...storageForm, type: e.target.value})}
                className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] focus:border-purple-500 focus:outline-none text-sm">
                <option value="s3">S3-Compatible</option>
                <option value="gdrive">Google Drive</option>
              </select>
              {storageForm.type === 's3' && (
                <>
                  <input placeholder="Endpoint (e.g., https://s3.amazonaws.com)" value={storageForm.s3_endpoint || ''} onChange={e => setStorageForm({...storageForm, s3_endpoint: e.target.value})}
                    className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
                  <div className="grid grid-cols-2 gap-3">
                    <input placeholder="Region" value={storageForm.s3_region || ''} onChange={e => setStorageForm({...storageForm, s3_region: e.target.value})}
                      className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
                    <input placeholder="Bucket" value={storageForm.s3_bucket || ''} onChange={e => setStorageForm({...storageForm, s3_bucket: e.target.value})}
                      className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
                  </div>
                  <input placeholder="Path Prefix (optional)" value={storageForm.s3_prefix || ''} onChange={e => setStorageForm({...storageForm, s3_prefix: e.target.value})}
                    className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
                </>
              )}
              {storageForm.type === 'gdrive' && (
                <input placeholder="Google Drive Folder ID" value={storageForm.gdrive_folder_id || ''} onChange={e => setStorageForm({...storageForm, gdrive_folder_id: e.target.value})}
                  className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
              )}
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowStorageModal(false)}
                className="px-4 py-2 bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-[var(--border-light)] rounded-lg hover:bg-[var(--bg-elevated)] transition text-sm font-medium">Cancel</button>
              <button onClick={saveStorageConfig} disabled={!storageForm.name}
                className="px-5 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium">Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
