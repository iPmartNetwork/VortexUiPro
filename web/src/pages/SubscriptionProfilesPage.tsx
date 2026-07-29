import { useState, useEffect } from 'react'
import { apiClient, apiGet, apiPost, apiDelete } from '../api/client'

interface SubProfile {
  id: number
  inbound_id: number
  dest: string
  port: number
  remark?: string
  enabled: boolean
  network?: string
  security?: string
  sni?: string
  fingerprint?: string
}

interface SubHost {
  id: number
  remark: string
  domain: string
  enable: boolean
}

export function SubscriptionProfilesPage() {
  const [profiles, setProfiles] = useState<SubProfile[]>([])
  const [hosts, setHosts] = useState<SubHost[]>([])
  const [formats, setFormats] = useState<{ name: string; description: string }[]>([])
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'profiles' | 'hosts' | 'info'>('profiles')
  const [showProfileModal, setShowProfileModal] = useState(false)
  const [showHostModal, setShowHostModal] = useState(false)
  const [profileForm, setProfileForm] = useState({ inbound_id: 0, dest: '', port: 443, remark: '', enabled: true, network: '', security: '', sni: '', fingerprint: '' })
  const [hostForm, setHostForm] = useState({ remark: '', domain: '' })

  const fetchAll = async () => {
    try {
      const [p, h, f] = await Promise.all([
        apiGet('/api/v1/sub-profiles'),
        apiGet('/api/v1/sub-hosts'),
        apiGet('/api/v1/sub-formats'),
      ])
      setProfiles(p.data?.profiles || p.data || [])
      setHosts(h.data?.hosts || h.data || [])
      setFormats(f.data?.formats || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  useEffect(() => { fetchAll() }, [])

  const createProfile = async () => {
    try {
      await apiPost('/api/v1/sub-profiles', profileForm)
      setShowProfileModal(false)
      setProfileForm({ inbound_id: 0, dest: '', port: 443, remark: '', enabled: true, network: '', security: '', sni: '', fingerprint: '' })
      fetchAll()
    } catch (err) { console.error(err) }
  }

  const deleteProfile = async (id: number) => {
    if (!confirm('Delete this profile?')) return
    await apiDelete(`/api/v1/sub-profiles/${id}`)
    fetchAll()
  }

  const createHost = async () => {
    try {
      await apiPost('/api/v1/sub-hosts', hostForm)
      setShowHostModal(false)
      setHostForm({ remark: '', domain: '' })
      fetchAll()
    } catch (err) { console.error(err) }
  }

  const deleteHost = async (id: number) => {
    if (!confirm('Delete this host?')) return
    await apiDelete(`/api/v1/sub-hosts/${id}`)
    fetchAll()
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
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Subscription Profiles</h1>
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">Multi-profile endpoints & custom hosts</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 glass-card p-1.5 w-fit">
        {(['profiles', 'hosts', 'info'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)} className={`px-4 py-2 rounded-lg text-xs font-medium transition-all ${tab === t ? 'bg-purple-500/15 text-purple-300 border border-purple-500/15' : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]'}`}>
            {t === 'profiles' ? `Profiles (${profiles.length})` : t === 'hosts' ? `Hosts (${hosts.length})` : 'Formats'}
          </button>
        ))}
      </div>

      {/* Profiles Tab */}
      {tab === 'profiles' && (
        <div className="glass-card p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-bold text-[var(--text-primary)]">Subscription Profiles</h2>
            <button onClick={() => setShowProfileModal(true)} className="btn-primary text-xs">+ Add Profile</button>
          </div>
          {profiles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--text-muted)]">
              <svg className="w-12 h-12 mb-3 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" />
              </svg>
              <p className="text-sm">No profiles configured</p>
            </div>
          ) : (
            <div className="space-y-2">
              {profiles.map(p => (
                <div key={p.id} className="flex items-center gap-3 p-3 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                  <div className={`w-2 h-2 rounded-full ${p.enabled ? 'bg-emerald-400' : 'bg-[var(--text-muted)]'}`} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-[var(--text-primary)]">{p.remark || p.dest}</span>
                      <span className="text-[10px] text-[var(--text-muted)]">inbound #{p.inbound_id}</span>
                    </div>
                    <p className="text-xs text-[var(--text-secondary)]">{p.dest}:{p.port} {p.network && `• ${p.network}`} {p.security && `• ${p.security}`}</p>
                  </div>
                  <button onClick={() => deleteProfile(p.id)} className="w-7 h-7 rounded-lg hover:bg-red-500/10 flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all">
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Hosts Tab */}
      {tab === 'hosts' && (
        <div className="glass-card p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-bold text-[var(--text-primary)]">Custom Subscription Hosts</h2>
            <button onClick={() => setShowHostModal(true)} className="btn-primary text-xs">+ Add Host</button>
          </div>
          {hosts.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--text-muted)]">
              <svg className="w-12 h-12 mb-3 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064" />
              </svg>
              <p className="text-sm">No custom hosts</p>
            </div>
          ) : (
            <div className="space-y-2">
              {hosts.map(h => (
                <div key={h.id} className="flex items-center gap-3 p-3 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                  <div className={`w-2 h-2 rounded-full ${h.enable ? 'bg-emerald-400' : 'bg-[var(--text-muted)]'}`} />
                  <div className="flex-1 min-w-0">
                    <span className="text-sm font-medium text-[var(--text-primary)]">{h.remark}</span>
                    <p className="text-xs text-[var(--text-secondary)]">{h.domain}</p>
                  </div>
                  <button onClick={() => deleteHost(h.id)} className="w-7 h-7 rounded-lg hover:bg-red-500/10 flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all">
                    <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Info Tab */}
      {tab === 'info' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Supported Formats</h2>
            <div className="space-y-2">
              {formats.map(f => (
                <div key={f.name} className="flex items-center gap-3 p-3 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                  <div className="w-7 h-7 rounded-lg bg-purple-500/10 flex items-center justify-center text-purple-300 text-xs font-bold">{f.name.slice(0, 3)}</div>
                  <div>
                    <span className="text-sm font-medium text-[var(--text-primary)]">{f.name}</span>
                    <p className="text-xs text-[var(--text-secondary)]">{f.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="glass-card p-5">
            <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Remark Variables</h2>
            <p className="text-xs text-[var(--text-muted)] mb-4">Use these variables in subscription remark templates</p>
            <div className="space-y-1.5 max-h-[400px] overflow-y-auto custom-scrollbar">
              {[ 
                { name: '{client_email}', desc: 'Client email address' },
                { name: '{client_id}', desc: 'Client UUID / ID' },
                { name: '{inbound_remark}', desc: 'Inbound remark name' },
                { name: '{inbound_tag}', desc: 'Inbound tag' },
                { name: '{inbound_protocol}', desc: 'Inbound protocol' },
                { name: '{inbound_port}', desc: 'Inbound port' },
                { name: '{user_data_limit}', desc: 'User data limit' },
                { name: '{user_expiry}', desc: 'User expiry date' },
                { name: '{user_status}', desc: 'Account status' },
                { name: '{subscription_url}', desc: 'Full subscription URL' },
              ].map(v => (
                <div key={v.name} className="flex items-center gap-2 p-2 rounded-lg bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                  <code className="text-[11px] text-purple-300 font-mono bg-purple-500/10 px-1.5 py-0.5 rounded">{v.name}</code>
                  <span className="text-xs text-[var(--text-muted)]">{v.desc}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Create Profile Modal */}
      {showProfileModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowProfileModal(false)}>
          <div className="glass-panel w-full max-w-lg mx-4 p-6" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">New Subscription Profile</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Inbound ID *</label>
                <input type="number" value={profileForm.inbound_id} onChange={e => setProfileForm(f => ({ ...f, inbound_id: +e.target.value }))} className="input-modern text-sm" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Port *</label>
                <input type="number" value={profileForm.port} onChange={e => setProfileForm(f => ({ ...f, port: +e.target.value }))} className="input-modern text-sm" />
              </div>
              <div className="col-span-2">
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Dest *</label>
                <input type="text" value={profileForm.dest} onChange={e => setProfileForm(f => ({ ...f, dest: e.target.value }))} className="input-modern text-sm" placeholder="e.g. 1.2.3.4 or example.com" />
              </div>
              <div className="col-span-2">
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Remark</label>
                <input type="text" value={profileForm.remark} onChange={e => setProfileForm(f => ({ ...f, remark: e.target.value }))} className="input-modern text-sm" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Network</label>
                <select value={profileForm.network} onChange={e => setProfileForm(f => ({ ...f, network: e.target.value }))} className="select-modern text-sm">
                  <option value="">Default</option>
                  <option value="tcp">TCP</option>
                  <option value="ws">WebSocket</option>
                  <option value="grpc">gRPC</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Security</label>
                <select value={profileForm.security} onChange={e => setProfileForm(f => ({ ...f, security: e.target.value }))} className="select-modern text-sm">
                  <option value="">None</option>
                  <option value="tls">TLS</option>
                  <option value="reality">REALITY</option>
                </select>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowProfileModal(false)} className="btn-ghost text-sm">Cancel</button>
              <button onClick={createProfile} className="btn-primary text-sm">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Create Host Modal */}
      {showHostModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowHostModal(false)}>
          <div className="glass-panel w-full max-w-lg mx-4 p-6" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">New Subscription Host</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Remark *</label>
                <input type="text" value={hostForm.remark} onChange={e => setHostForm(f => ({ ...f, remark: e.target.value }))} className="input-modern text-sm" placeholder="e.g. US Server" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Domain *</label>
                <input type="text" value={hostForm.domain} onChange={e => setHostForm(f => ({ ...f, domain: e.target.value }))} className="input-modern text-sm" placeholder="e.g. sub.example.com" />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowHostModal(false)} className="btn-ghost text-sm">Cancel</button>
              <button onClick={createHost} className="btn-primary text-sm">Create</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
