import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface Outbound {
  id: number
  tag: string
  protocol: string
  node_id: number
  settings: string
  stream_settings: string
  remark: string
  enable: boolean
  hidden: boolean
  created_at: number
  updated_at: number
}

export function OutboundsPage() {
  const [outbounds, setOutbounds] = useState<Outbound[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Outbound | null>(null)
  const [form, setForm] = useState({
    tag: '', protocol: 'freedom', node_id: 0, settings: '',
    stream_settings: '', remark: '', enable: true, hidden: false,
  })

  useEffect(() => { fetchOutbounds() }, [])

  const fetchOutbounds = async () => {
    try {
      setLoading(true)
      const { data } = await apiClient.get('/api/v1/outbounds')
      setOutbounds(data.outbounds || [])
    } catch {} finally { setLoading(false) }
  }

  const openCreate = () => {
    setEditing(null)
    setForm({ tag: '', protocol: 'freedom', node_id: 0, settings: '', stream_settings: '', remark: '', enable: true, hidden: false })
    setShowForm(true)
  }

  const openEdit = (ob: Outbound) => {
    setEditing(ob)
    setForm({
      tag: ob.tag,
      protocol: ob.protocol,
      node_id: ob.node_id,
      settings: ob.settings || '',
      stream_settings: ob.stream_settings || '',
      remark: ob.remark || '',
      enable: ob.enable,
      hidden: ob.hidden,
    })
    setShowForm(true)
  }

  const save = async () => {
    try {
      if (editing) {
        await apiClient.put(`/api/v1/outbounds/${editing.id}`, form)
      } else {
        await apiClient.post('/api/v1/outbounds', form)
      }
      setShowForm(false)
      fetchOutbounds()
    } catch (err) { console.error(err) }
  }

  const deleteOb = async (id: number) => {
    if (!confirm('Delete this outbound?')) return
    try {
      await apiClient.delete(`/api/v1/outbounds/${id}`)
      fetchOutbounds()
    } catch (err) { console.error(err) }
  }

  const toggleVisibility = async (ob: Outbound) => {
    try {
      await apiClient.put(`/api/v1/outbounds/${ob.id}/visibility`, { hidden: !ob.hidden })
      fetchOutbounds()
    } catch (err) { console.error(err) }
  }

  const protocolColors: Record<string, string> = {
    freedom: 'green', blackhole: 'red', dns: 'blue',
    vmess: 'purple', vless: 'purple', trojan: 'cyan',
    shadowsocks: 'yellow', socks: 'orange', http: 'blue',
    wireguard: 'emerald', direct: 'green',
  }

  const getBadge = (proto: string) => {
    const color = protocolColors[proto.toLowerCase()] || 'gray'
    return `badge badge-${color}`
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Outbound Management</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Proxy egress and routing outbound configurations</p>
            </div>
          </div>
          <button onClick={openCreate}
            className="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 transition text-sm font-medium">
            + New Outbound
          </button>
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
                  <th>Tag</th>
                  <th>Protocol</th>
                  <th>Node ID</th>
                  <th>Remark</th>
                  <th>Status</th>
                  <th>Visibility</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {outbounds.map(ob => (
                  <tr key={ob.id}>
                    <td><span className="font-mono text-sm text-[var(--text-primary)]">{ob.tag}</span></td>
                    <td><span className={getBadge(ob.protocol)}>{ob.protocol}</span></td>
                    <td className="text-[var(--text-secondary)]">{ob.node_id || '-'}</td>
                    <td className="text-[var(--text-secondary)] text-sm">{ob.remark || '-'}</td>
                    <td>
                      <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-medium ${
                        ob.enable ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'
                      }`}>
                        <span className={`w-1.5 h-1.5 rounded-full ${ob.enable ? 'bg-green-500' : 'bg-red-500'}`} />
                        {ob.enable ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                    <td>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                        ob.hidden ? 'bg-yellow-500/10 text-yellow-400' : 'bg-blue-500/10 text-blue-400'
                      }`}>
                        {ob.hidden ? 'Hidden' : 'Visible'}
                      </span>
                    </td>
                    <td className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button onClick={() => openEdit(ob)}
                          className="px-3 py-1 rounded-md bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 transition text-[11px] font-medium">
                          Edit
                        </button>
                        <button onClick={() => toggleVisibility(ob)}
                          className={`px-3 py-1 rounded-md transition text-[11px] font-medium ${
                            ob.hidden
                              ? 'bg-blue-500/10 text-blue-400 hover:bg-blue-500/20'
                              : 'bg-yellow-500/10 text-yellow-400 hover:bg-yellow-500/20'
                          }`}>
                          {ob.hidden ? 'Show' : 'Hide'}
                        </button>
                        <button onClick={() => deleteOb(ob.id)}
                          className="px-3 py-1 rounded-md bg-red-500/10 text-red-400 hover:bg-red-500/20 transition text-[11px] font-medium">
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {outbounds.length === 0 && (
                  <tr>
                    <td colSpan={7} className="text-center py-12 text-[var(--text-muted)] text-sm">
                      No outbounds configured. Create one to get started.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Create/Edit Modal */}
      {showForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowForm(false)}>
          <div className="glass-card p-6 w-full max-w-lg mx-4" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-4">
              {editing ? 'Edit Outbound' : 'New Outbound'}
            </h2>
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs text-[var(--text-secondary)] mb-1">Tag *</label>
                  <input value={form.tag} onChange={e => setForm({...form, tag: e.target.value})}
                    placeholder="outbound-tag"
                    className="w-full px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
                </div>
                <div>
                  <label className="block text-xs text-[var(--text-secondary)] mb-1">Protocol *</label>
                  <select value={form.protocol} onChange={e => setForm({...form, protocol: e.target.value})}
                    className="w-full px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] focus:border-purple-500 focus:outline-none text-sm">
                    <option value="freedom">Freedom (Direct)</option>
                    <option value="blackhole">Blackhole</option>
                    <option value="dns">DNS</option>
                    <option value="vmess">VMess</option>
                    <option value="vless">VLESS</option>
                    <option value="trojan">Trojan</option>
                    <option value="shadowsocks">Shadowsocks</option>
                    <option value="socks">SOCKS</option>
                    <option value="http">HTTP</option>
                    <option value="wireguard">WireGuard</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1">Remark</label>
                <input value={form.remark} onChange={e => setForm({...form, remark: e.target.value})}
                  placeholder="Optional description"
                  className="w-full px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
              </div>
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1">Node ID</label>
                <input type="number" value={form.node_id} onChange={e => setForm({...form, node_id: parseInt(e.target.value) || 0})}
                  placeholder="0 for local"
                  className="w-full px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm" />
              </div>
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1">Settings (JSON)</label>
                <textarea value={form.settings} onChange={e => setForm({...form, settings: e.target.value})}
                  rows={3} placeholder='{"domainStrategy": "AsIs"}'
                  className="w-full px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm font-mono" />
              </div>
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1">Stream Settings (JSON)</label>
                <textarea value={form.stream_settings} onChange={e => setForm({...form, stream_settings: e.target.value})}
                  rows={3} placeholder='{"security": "tls", "tlsSettings": {}}'
                  className="w-full px-3 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm font-mono" />
              </div>
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" checked={form.enable} onChange={e => setForm({...form, enable: e.target.checked})}
                    className="w-4 h-4 rounded border-[var(--border-light)] text-purple-600 focus:ring-purple-500 bg-[var(--bg-elevated)]" />
                  <span className="text-sm text-[var(--text-secondary)]">Enabled</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" checked={form.hidden} onChange={e => setForm({...form, hidden: e.target.checked})}
                    className="w-4 h-4 rounded border-[var(--border-light)] text-purple-600 focus:ring-purple-500 bg-[var(--bg-elevated)]" />
                  <span className="text-sm text-[var(--text-secondary)]">Hidden (from subscriptions)</span>
                </label>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-6">
              <button onClick={() => setShowForm(false)}
                className="px-4 py-2 bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] border border-[var(--border-light)] rounded-lg transition text-sm font-medium">
                Cancel
              </button>
              <button onClick={save} disabled={!form.tag || !form.protocol}
                className="px-5 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium">
                {editing ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
