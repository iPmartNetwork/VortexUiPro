import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '../api/client'

interface DNSConfig {
  id: number
  name: string
  upstream: string
  port: number
  doh_enabled: boolean
  dot_enabled: boolean
  rate_limit: number
  cache_ttl: number
  enabled: boolean
  created_at: number
}

interface DNSRule {
  id: number
  name: string
  domain_match: string
  action: string
  target_upstream: string
  enabled: boolean
  priority: number
  created_at: number
}

type TabKey = 'config' | 'rules' | 'resolve'

export function SmartDNSPage() {
  const [tab, setTab] = useState<TabKey>('config')
  const [configs, setConfigs] = useState<DNSConfig[]>([])
  const [rules, setRules] = useState<DNSRule[]>([])
  const [loading, setLoading] = useState(true)

  // Config form
  const [configForm, setConfigForm] = useState({ name: '', upstream: '', port: 853, doh_enabled: true, dot_enabled: false, rate_limit: 100, cache_ttl: 300 })

  // Rule form
  const [ruleForm, setRuleForm] = useState({ name: '', domain_match: '', action: 'block', target_upstream: '', priority: 0 })

  // Resolve
  const [resolveDomain, setResolveDomain] = useState('')
  const [resolveResult, setResolveResult] = useState<any>(null)

  useEffect(() => {
    Promise.all([
      apiClient.get('/api/v1/dns/configs').then(r => setConfigs(r.data.data || [])),
      apiClient.get('/api/v1/dns/rules').then(r => setRules(r.data.data || [])),
    ]).finally(() => setLoading(false))
  }, [])

  const saveConfig = useCallback(async () => {
    try {
      await apiClient.post('/api/v1/dns/configs', configForm)
      const r = await apiClient.get('/api/v1/dns/configs')
      setConfigs(r.data.data || [])
      setConfigForm({ name: '', upstream: '', port: 853, doh_enabled: true, dot_enabled: false, rate_limit: 100, cache_ttl: 300 })
    } catch {}
  }, [configForm])

  const deleteConfig = useCallback(async (id: number) => {
    try {
      await apiClient.delete(`/api/v1/dns/configs/${id}`)
      setConfigs(prev => prev.filter(c => c.id !== id))
    } catch {}
  }, [])

  const saveRule = useCallback(async () => {
    try {
      await apiClient.post('/api/v1/dns/rules', ruleForm)
      const r = await apiClient.get('/api/v1/dns/rules')
      setRules(r.data.data || [])
      setRuleForm({ name: '', domain_match: '', action: 'block', target_upstream: '', priority: 0 })
    } catch {}
  }, [ruleForm])

  const deleteRule = useCallback(async (id: number) => {
    try {
      await apiClient.delete(`/api/v1/dns/rules/${id}`)
      setRules(prev => prev.filter(r => r.id !== id))
    } catch {}
  }, [])

  const handleResolve = useCallback(async () => {
    if (!resolveDomain) return
    try {
      const res = await apiClient.get('/api/v1/dns/resolve', { params: { domain: resolveDomain } })
      setResolveResult(res.data.data)
    } catch { setResolveResult({ error: 'DNS resolution failed' }) }
  }, [resolveDomain])

  const loadAdBlock = useCallback(async () => {
    try {
      await apiClient.post('/api/v1/dns/load-ad-block')
      alert('Ad-block list loaded successfully!')
    } catch { alert('Failed to load ad-block list') }
  }, [])

  const clearCache = useCallback(async () => {
    try {
      await apiClient.post('/api/v1/dns/clear-cache')
      alert('DNS cache cleared!')
    } catch {}
  }, [])

  if (loading) return <div className="flex items-center justify-center min-h-[400px] text-[#6868a0]">Loading DNS system...</div>

  return (
    <div className="space-y-6 page-enter">
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m0 0c1.657 0 3 4.03 3 9s-1.343 9-3 9z" /></svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-white">Smart DNS System</h1>
            <p className="text-sm text-[#6868a0] mt-0.5">DoH/DoT resolver, ad-blocking, DNS routing rules</p>
          </div>
        </div>
      </div>

      <div className="flex gap-1 p-1 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] w-fit flex-wrap">
        <button onClick={() => setTab('config')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'config' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>DNS Configs</button>
        <button onClick={() => setTab('rules')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'rules' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Routing Rules</button>
        <button onClick={() => setTab('resolve')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'resolve' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>DNS Resolver</button>
      </div>

      {tab === 'config' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">➕ Add DNS Upstream</h3>
            <div className="space-y-3">
              <input value={configForm.name} onChange={e => setConfigForm(p => ({ ...p, name: e.target.value }))} placeholder="Config name" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <input value={configForm.upstream} onChange={e => setConfigForm(p => ({ ...p, upstream: e.target.value }))} placeholder="Upstream DNS (e.g., 8.8.8.8)" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <div className="grid grid-cols-2 gap-2">
                <input value={configForm.port} onChange={e => setConfigForm(p => ({ ...p, port: Number(e.target.value) }))} type="number" placeholder="Port" className="px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
                <input value={configForm.cache_ttl} onChange={e => setConfigForm(p => ({ ...p, cache_ttl: Number(e.target.value) }))} type="number" placeholder="Cache TTL (s)" className="px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              </div>
              <div className="flex gap-4">
                <label className="flex items-center gap-2 text-xs text-[#9898b8]"><input type="checkbox" checked={configForm.doh_enabled} onChange={e => setConfigForm(p => ({ ...p, doh_enabled: e.target.checked }))} className="accent-purple-500" /> DoH</label>
                <label className="flex items-center gap-2 text-xs text-[#9898b8]"><input type="checkbox" checked={configForm.dot_enabled} onChange={e => setConfigForm(p => ({ ...p, dot_enabled: e.target.checked }))} className="accent-purple-500" /> DoT</label>
              </div>
              <button onClick={saveConfig} className="w-full px-5 py-2.5 bg-gradient-to-r from-blue-600 to-cyan-600 text-white rounded-lg hover:from-blue-500 hover:to-cyan-500 transition text-sm font-medium">Save DNS Config</button>
            </div>
          </div>
          <div className="glass-card p-5">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-white font-semibold">📋 DNS Upstreams</h3>
              <div className="flex gap-2">
                <button onClick={clearCache} className="px-3 py-1.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-[#9898b8] rounded-lg hover:text-white text-xs">🗑 Clear Cache</button>
                <button onClick={loadAdBlock} className="px-3 py-1.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-[#9898b8] rounded-lg hover:text-white text-xs">🛡 Load Ad-Block</button>
              </div>
            </div>
            <div className="space-y-2 max-h-80 overflow-y-auto">
              {configs.map(c => (
                <div key={c.id} className="p-3 rounded-lg bg-[rgba(255,255,255,0.02)] flex items-center justify-between">
                  <div>
                    <p className="text-sm text-white font-medium">{c.name}</p>
                    <p className="text-xs text-[#6868a0]">{c.upstream}:{c.port} · {c.doh_enabled ? 'DoH' : ''}{c.dot_enabled ? ' DoT' : ''} · TTL {c.cache_ttl}s</p>
                  </div>
                  <button onClick={() => deleteConfig(c.id)} className="text-xs text-red-400 hover:text-red-300 px-2 py-1 rounded hover:bg-red-500/10">Delete</button>
                </div>
              ))}
              {configs.length === 0 && <p className="text-center py-8 text-[#585878] text-sm">No DNS configs yet</p>}
            </div>
          </div>
        </div>
      )}

      {tab === 'rules' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">➕ Add DNS Rule</h3>
            <div className="space-y-3">
              <input value={ruleForm.name} onChange={e => setRuleForm(p => ({ ...p, name: e.target.value }))} placeholder="Rule name" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <input value={ruleForm.domain_match} onChange={e => setRuleForm(p => ({ ...p, domain_match: e.target.value }))} placeholder="Domain match (glob: *.ads.com)" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <select value={ruleForm.action} onChange={e => setRuleForm(p => ({ ...p, action: e.target.value }))} className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none">
                <option value="block">Block</option>
                <option value="redirect">Redirect</option>
                <option value="custom">Custom Upstream</option>
              </select>
              {ruleForm.action === 'custom' && (
                <input value={ruleForm.target_upstream} onChange={e => setRuleForm(p => ({ ...p, target_upstream: e.target.value }))} placeholder="Target upstream (8.8.8.8)" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              )}
              <button onClick={saveRule} className="w-full px-5 py-2.5 bg-gradient-to-r from-cyan-600 to-teal-600 text-white rounded-lg hover:from-cyan-500 hover:to-teal-500 transition text-sm font-medium">Add Rule</button>
            </div>
          </div>
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">📋 DNS Routing Rules</h3>
            <div className="space-y-2 max-h-80 overflow-y-auto">
              {rules.map(r => (
                <div key={r.id} className="p-3 rounded-lg bg-[rgba(255,255,255,0.02)] flex items-center justify-between">
                  <div>
                    <p className="text-sm text-white font-medium">{r.name}</p>
                    <p className="text-xs text-[#6868a0]">{r.domain_match} → <span className={r.action === 'block' ? 'text-red-400' : 'text-emerald-400'}>{r.action}</span></p>
                  </div>
                  <button onClick={() => deleteRule(r.id)} className="text-xs text-red-400 hover:text-red-300 px-2 py-1 rounded hover:bg-red-500/10">Delete</button>
                </div>
              ))}
              {rules.length === 0 && <p className="text-center py-8 text-[#585878] text-sm">No DNS rules yet</p>}
            </div>
          </div>
        </div>
      )}

      {tab === 'resolve' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">🔍 DNS Lookup</h3>
            <p className="text-xs text-[#6868a0] mb-4">Resolve a domain through the Smart DNS system (DoH/DoT).</p>
            <div className="flex gap-3">
              <input value={resolveDomain} onChange={e => setResolveDomain(e.target.value)} placeholder="e.g., google.com" className="flex-1 px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" onKeyDown={e => e.key === 'Enter' && handleResolve()} />
              <button onClick={handleResolve} disabled={!resolveDomain} className="px-5 py-2.5 bg-cyan-600 text-white rounded-lg hover:bg-cyan-500 disabled:opacity-50 transition text-sm font-medium">Resolve</button>
            </div>
          </div>
          {resolveResult && (
            <div className="glass-card p-5">
              <h3 className="text-white font-semibold mb-4">📊 Result for {resolveDomain}</h3>
              {resolveResult.error ? (
                <div className="p-3 rounded-lg bg-red-500/10 text-red-400 text-sm">{resolveResult.error}</div>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                    <span className="text-[#9898b8] text-sm">IP</span>
                    <span className="text-white text-sm font-mono">{resolveResult.ip || resolveResult.answers?.[0] || 'N/A'}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                    <span className="text-[#9898b8] text-sm">Upstream</span>
                    <span className="text-white text-sm">{resolveResult.upstream || 'N/A'}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                    <span className="text-[#9898b8] text-sm">Protocol</span>
                    <span className="text-white text-sm">{resolveResult.protocol || resolveResult.type || 'A'}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-[rgba(255,255,255,0.02)]">
                    <span className="text-[#9898b8] text-sm">TTL</span>
                    <span className="text-white text-sm">{resolveResult.ttl || 'N/A'}</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
