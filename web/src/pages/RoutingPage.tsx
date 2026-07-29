import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface RoutingRule {
  id: number
  inbound_tags?: string
  outbound_tag: string
  domain?: string
  ip?: string
  port?: string
  network?: string
  protocol?: string
  source_ip?: string
  balancer_tag?: string
  rule_type?: string
  enable: boolean
  created_at: number
}

export function RoutingPage() {
  const [rules, setRules] = useState<RoutingRule[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editRule, setEditRule] = useState<RoutingRule | null>(null)
  const [generated, setGenerated] = useState('')
  const [form, setForm] = useState({
    outbound_tag: '',
    domain: '',
    ip: '',
    port: '',
    network: '',
    protocol: '',
    source_ip: '',
    inbound_tags: '',
    enable: true,
  })

  const fetchRules = async () => {
    try {
      const r = await apiClient.get('/api/v1/routing/rules')
      setRules(r.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  const fetchGenerated = async () => {
    try {
      const r = await apiClient.get('/api/v1/routing/generate')
      setGenerated(typeof r.data === 'string' ? r.data : JSON.stringify(r.data, null, 2))
    } catch { /* ignore */ }
  }

  useEffect(() => { fetchRules(); fetchGenerated() }, [])

  const openCreate = () => {
    setEditRule(null)
    setForm({ outbound_tag: '', domain: '', ip: '', port: '', network: '', protocol: '', source_ip: '', inbound_tags: '', enable: true })
    setShowModal(true)
  }

  const openEdit = (rule: RoutingRule) => {
    setEditRule(rule)
    setForm({
      outbound_tag: rule.outbound_tag,
      domain: rule.domain || '',
      ip: rule.ip || '',
      port: rule.port || '',
      network: rule.network || '',
      protocol: rule.protocol || '',
      source_ip: rule.source_ip || '',
      inbound_tags: rule.inbound_tags || '',
      enable: rule.enable,
    })
    setShowModal(true)
  }

  const save = async () => {
    try {
      if (editRule) {
        await apiClient.put(`/api/v1/routing/rules/${editRule.id}`, form)
      } else {
        await apiClient.post('/api/v1/routing/rules', form)
      }
      setShowModal(false)
      fetchRules()
    } catch (err) { console.error(err) }
  }

  const toggle = async (id: number, enable: boolean) => {
    await apiClient.put(`/api/v1/routing/rules/${id}/toggle`, { enable })
    fetchRules()
  }

  const remove = async (id: number) => {
    if (!confirm('Delete this routing rule?')) return
    await apiClient.delete(`/api/v1/routing/rules/${id}`)
    fetchRules()
  }

  const protocolColors: Record<string, string> = {
    vmess: 'from-purple-500/20 to-purple-600/10 text-purple-300 border-purple-500/20',
    vless: 'from-cyan-500/20 to-cyan-600/10 text-cyan-300 border-cyan-500/20',
    trojan: 'from-emerald-500/20 to-emerald-600/10 text-emerald-300 border-emerald-500/20',
    shadowsocks: 'from-amber-500/20 to-amber-600/10 text-amber-300 border-amber-500/20',
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
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">Routing Rules</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage advanced traffic routing</p>
            </div>
          </div>
          <button onClick={openCreate} className="btn-primary text-sm">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add Rule
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Rules Table */}
        <div className="lg:col-span-2 glass-card p-5">
          <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Rules ({rules.filter(r => r.enable).length}/{rules.length} enabled)</h2>
          {rules.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--text-muted)]">
              <svg className="w-12 h-12 mb-3 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
              <p className="text-sm">No routing rules configured</p>
            </div>
          ) : (
            <div className="space-y-2 max-h-[500px] overflow-y-auto custom-scrollbar">
              {rules.map(rule => (
                <div key={rule.id} className="group flex items-center gap-3 p-3 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)] hover:border-[rgba(139,92,246,0.12)] transition-all">
                  <button onClick={() => toggle(rule.id, !rule.enable)} className={`w-8 h-8 rounded-lg flex items-center justify-center transition-all ${rule.enable ? 'bg-emerald-500/10 text-emerald-400' : 'bg-[rgba(255,255,255,0.03)] text-[var(--text-muted)]'}`}>
                    {rule.enable ? (
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    )}
                  </button>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-[var(--text-primary)]">{rule.outbound_tag}</span>
                      {rule.rule_type && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-purple-500/10 text-purple-300 border border-purple-500/10">
                          {rule.rule_type}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2 mt-1 flex-wrap">
                      {rule.domain && <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-300">domain</span>}
                      {rule.ip && <span className="text-[10px] px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-300">ip</span>}
                      {rule.port && <span className="text-[10px] px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-300">port:{rule.port}</span>}
                      {rule.network && <span className="text-[10px] px-1.5 py-0.5 rounded bg-pink-500/10 text-pink-300">{rule.network}</span>}
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-all">
                    <button onClick={() => openEdit(rule)} className="w-7 h-7 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(139,92,246,0.1)] flex items-center justify-center text-[var(--text-muted)] hover:text-purple-300 transition-all">
                      <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
                        <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                      </svg>
                    </button>
                    <button onClick={() => remove(rule.id)} className="w-7 h-7 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-red-500/10 flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all">
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

        {/* Generated Config Preview */}
        <div className="glass-card p-5 flex flex-col">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-bold text-[var(--text-primary)]">Generated Xray Config</h2>
            <button onClick={fetchGenerated} className="text-xs text-purple-400 hover:text-purple-300 transition">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          </div>
          <pre className="flex-1 text-[11px] text-[var(--text-secondary)] font-mono bg-[rgba(0,0,0,0.2)] rounded-xl p-4 overflow-auto max-h-[500px] custom-scrollbar whitespace-pre-wrap">
            {generated || 'Generate a config to preview...'}
          </pre>
        </div>
      </div>

      {/* Create/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowModal(false)}>
          <div className="glass-panel w-full max-w-2xl mx-4 p-6" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">
              {editRule ? 'Edit Rule' : 'New Routing Rule'}
            </h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2">
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Outbound Tag *</label>
                <input type="text" value={form.outbound_tag} onChange={e => setForm(f => ({ ...f, outbound_tag: e.target.value }))} className="input-modern text-sm" placeholder="e.g. direct, block, proxy-out" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Domain (JSON array)</label>
                <input type="text" value={form.domain} onChange={e => setForm(f => ({ ...f, domain: e.target.value }))} className="input-modern text-sm" placeholder='["example.com"]' />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">IP (JSON array)</label>
                <input type="text" value={form.ip} onChange={e => setForm(f => ({ ...f, ip: e.target.value }))} className="input-modern text-sm" placeholder='["10.0.0.0/8"]' />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Port</label>
                <input type="text" value={form.port} onChange={e => setForm(f => ({ ...f, port: e.target.value }))} className="input-modern text-sm" placeholder="e.g. 443, 80-443" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Network</label>
                <input type="text" value={form.network} onChange={e => setForm(f => ({ ...f, network: e.target.value }))} className="input-modern text-sm" placeholder="tcp, udp, ws, grpc" />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Protocol</label>
                <input type="text" value={form.protocol} onChange={e => setForm(f => ({ ...f, protocol: e.target.value }))} className="input-modern text-sm" placeholder='["vless","trojan"]' />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Source IP</label>
                <input type="text" value={form.source_ip} onChange={e => setForm(f => ({ ...f, source_ip: e.target.value }))} className="input-modern text-sm" placeholder='["1.2.3.4"]' />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Inbound Tags</label>
                <input type="text" value={form.inbound_tags} onChange={e => setForm(f => ({ ...f, inbound_tags: e.target.value }))} className="input-modern text-sm" placeholder='["inbound-tag"]' />
              </div>
            </div>
            <div className="flex items-center justify-between mt-6">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={form.enable} onChange={e => setForm(f => ({ ...f, enable: e.target.checked }))} className="rounded border-gray-600" />
                <span className="text-sm text-[var(--text-secondary)]">Enabled</span>
              </label>
              <div className="flex gap-2">
                <button onClick={() => setShowModal(false)} className="btn-ghost text-sm">Cancel</button>
                <button onClick={save} className="btn-primary text-sm">
                  {editRule ? 'Update' : 'Create'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
