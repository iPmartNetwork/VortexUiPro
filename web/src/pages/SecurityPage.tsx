import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

export function SecurityPage() {
  const [tab, setTab] = useState<'geo' | 'password' | 'ip'>('geo')
  const [blockedCountries, setBlockedCountries] = useState<string[]>([])
  const [bannedIPs, setBannedIPs] = useState<string[]>([])
  const [whitelistedIPs, setWhitelistedIPs] = useState<string[]>([])
  const [newIP, setNewIP] = useState({ ip: '', reason: '', type: 'ban' as 'ban' | 'whitelist' })
  const [policy, setPolicy] = useState({ min_length: 8, require_upper: true, require_lower: true, require_number: true, require_special: false })
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  const countryList = [
    { code: 'IR', name: 'Iran' }, { code: 'CN', name: 'China' },
    { code: 'RU', name: 'Russia' }, { code: 'CU', name: 'Cuba' },
    { code: 'KP', name: 'North Korea' }, { code: 'SY', name: 'Syria' },
  ]

  useEffect(() => { fetchData() }, [])

  const fetchData = async () => {
    try {
      const [g, b, w, p] = await Promise.all([
        apiClient.get('/api/v1/settings/security/geo-block'),
        apiClient.get('/api/v1/settings/security/banned-ips'),
        apiClient.get('/api/v1/settings/security/whitelisted-ips'),
        apiClient.get('/api/v1/settings/security/password-policy'),
      ])
      setBlockedCountries(g.data.blocked_countries || [])
      setBannedIPs(b.data.banned_ips || [])
      setWhitelistedIPs(w.data.whitelisted_ips || [])
      if (p.data) setPolicy(p.data)
    } catch { /* ignore */ }
  }

  const toggleCountry = async (code: string) => {
    const updated = blockedCountries.includes(code)
      ? blockedCountries.filter(c => c !== code)
      : [...blockedCountries, code]
    setBlockedCountries(updated)
    await apiClient.put('/api/v1/settings/security/geo-block', { countries: updated })
  }

  const savePolicy = async () => {
    setSaving(true)
    try {
      await apiClient.put('/api/v1/settings/security/password-policy', policy)
      setSaved(true); setTimeout(() => setSaved(false), 2000)
    } catch { } finally { setSaving(false) }
  }

  const addIP = async () => {
    if (!newIP.ip) return
    try {
      if (newIP.type === 'ban') {
        await apiClient.post('/api/v1/settings/security/banned-ips', { ip: newIP.ip, reason: newIP.reason })
      } else {
        await apiClient.post('/api/v1/settings/security/whitelisted-ips', { ip: newIP.ip, reason: newIP.reason })
      }
      setNewIP({ ip: '', reason: '', type: 'ban' })
      fetchData()
    } catch { /* ignore */ }
  }

  const removeBanIP = async (ip: string) => {
    await apiClient.delete('/api/v1/settings/security/banned-ips?ip=' + encodeURIComponent(ip))
    fetchData()
  }

  const removeWhitelistIP = async (ip: string) => {
    await apiClient.delete('/api/v1/settings/security/whitelisted-ips?ip=' + encodeURIComponent(ip))
    fetchData()
  }

  return (
    <div className="space-y-6">
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-red-500 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Security Settings</h1>
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">Geo-blocking, password policy & IP management</p>
          </div>
        </div>
      </div>

      <div className="flex gap-1 glass-card p-1.5 w-fit">
        {([{ id: 'geo' as const, label: `Geo-Block (${blockedCountries.length})` },
           { id: 'password' as const, label: 'Password Policy' },
           { id: 'ip' as const, label: `IP Mgmt (${bannedIPs.length})` }]).map(t => (
          <button key={t.id} onClick={() => setTab(t.id)} className={`px-4 py-2 rounded-lg text-xs font-medium transition-all ${tab === t.id ? 'bg-purple-500/15 text-purple-300 border border-purple-500/15' : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]'}`}>
            {t.label}
          </button>
        ))}
      </div>

      {/* Geo-Block */}
      {tab === 'geo' && (
        <div className="glass-card p-5">
          <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Country Blocking</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            {countryList.map(c => (
              <button key={c.code} onClick={() => toggleCountry(c.code)} className={`flex items-center gap-3 p-3 rounded-xl border transition-all ${blockedCountries.includes(c.code) ? 'bg-red-500/10 border-red-500/20 text-red-300' : 'bg-[rgba(255,255,255,0.02)] border-[var(--border-light)] text-[var(--text-secondary)] hover:border-[rgba(139,92,246,0.12)]'}`}>
                <div className={`w-3 h-3 rounded border-2 flex items-center justify-center ${blockedCountries.includes(c.code) ? 'border-red-400 bg-red-400' : 'border-[var(--text-muted)]'}`}>
                  {blockedCountries.includes(c.code) && <svg className="w-2 h-2 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                </div>
                <span className="text-sm font-medium">{c.name} ({c.code})</span>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Password Policy */}
      {tab === 'password' && (
        <div className="glass-card p-5">
          <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Password Requirements</h2>
          <div className="space-y-4 max-w-md">
            <div>
              <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Minimum Length</label>
              <input type="number" value={policy.min_length} onChange={e => setPolicy(p => ({ ...p, min_length: +e.target.value }))} className="input-modern text-sm w-full" min={4} max={64} />
            </div>
            {(['require_upper', 'require_lower', 'require_number', 'require_special'] as const).map(f => (
              <label key={f} className="flex items-center justify-between py-2 border-b border-[var(--border-light)] cursor-pointer">
                <span className="text-sm text-[var(--text-primary)]">{f === 'require_upper' ? 'Require Uppercase (A-Z)' : f === 'require_lower' ? 'Require Lowercase (a-z)' : f === 'require_number' ? 'Require Number (0-9)' : 'Require Special (!@#$)'}</span>
                <input type="checkbox" checked={policy[f]} onChange={e => setPolicy(p => ({ ...p, [f]: e.target.checked }))} className="rounded border-gray-600" />
              </label>
            ))}
            <div className="flex justify-end pt-2">
              {saved && <span className="text-sm text-emerald-400 mr-3">Saved!</span>}
              <button onClick={savePolicy} disabled={saving} className="btn-primary text-sm">
                {saving ? 'Saving...' : 'Save Policy'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* IP Management */}
      {tab === 'ip' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-5">
            <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Banned IPs ({bannedIPs.length})</h2>
            <div className="flex gap-2 mb-4">
              <input type="text" value={newIP.ip} onChange={e => setNewIP(n => ({ ...n, ip: e.target.value }))} placeholder="IP address" className="input-modern text-sm flex-1" />
              <button onClick={addIP} className="btn-primary text-xs">Ban</button>
            </div>
            <div className="space-y-2 max-h-[300px] overflow-y-auto">
              {bannedIPs.length === 0 ? (
                <p className="text-sm text-[var(--text-muted)] text-center py-8">No banned IPs</p>
              ) : bannedIPs.map((entry, i) => {
                const [ip, ...reasonParts] = entry.split(':')
                const reason = reasonParts.join(':')
                return (
                  <div key={i} className="flex items-center justify-between p-2.5 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                    <div>
                      <span className="text-sm text-[var(--text-primary)] font-mono">{ip}</span>
                      {reason && <p className="text-[10px] text-[var(--text-muted)]">{reason}</p>}
                    </div>
                    <button onClick={() => removeBanIP(ip)} className="w-6 h-6 hover:bg-red-500/10 rounded flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all">
                      <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  </div>
                )
              })}
            </div>
          </div>
          <div className="glass-card p-5">
            <h2 className="text-sm font-bold text-[var(--text-primary)] mb-4">Whitelisted IPs ({whitelistedIPs.length})</h2>
            <div className="flex gap-2 mb-4">
              <input type="text" value={newIP.ip} onChange={e => setNewIP(n => ({ ...n, ip: e.target.value }))} placeholder="IP address" className="input-modern text-sm flex-1" />
              <button onClick={() => { newIP.type = 'whitelist'; addIP() }} className="btn-primary text-xs bg-emerald-500/20 text-emerald-300 border-emerald-500/20 hover:bg-emerald-500/30">Allow</button>
            </div>
            <div className="space-y-2 max-h-[300px] overflow-y-auto">
              {whitelistedIPs.length === 0 ? (
                <p className="text-sm text-[var(--text-muted)] text-center py-8">No whitelisted IPs</p>
              ) : whitelistedIPs.map((entry, i) => {
                const [ip, ...reasonParts] = entry.split(':')
                return (
                  <div key={i} className="flex items-center justify-between p-2.5 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[var(--border-light)]">
                    <span className="text-sm text-[var(--text-primary)] font-mono">{ip}</span>
                    <button onClick={() => removeWhitelistIP(ip)} className="w-6 h-6 hover:bg-red-500/10 rounded flex items-center justify-center text-[var(--text-muted)] hover:text-red-400 transition-all">
                      <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
