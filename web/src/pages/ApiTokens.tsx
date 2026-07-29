import { useState, useEffect } from 'react'
import { apiGet, apiPost, apiDelete, apiPut } from '../api/client'

interface TokenView {
  id: number; name: string; token?: string; kind: string; subjectUsername: string
  subjectRoleName: string; scopes: string[]; expiresAt: number; expired: boolean
  enabled: boolean; createdAt: number
}

export function ApiTokensPage() {
  const [tokens, setTokens] = useState<TokenView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [tokenName, setTokenName] = useState('')
  const [expiresDays, setExpiresDays] = useState(90)
  const [newTokenPlain, setNewTokenPlain] = useState('')

  useEffect(() => { fetchTokens() }, [])

  const fetchTokens = async () => {
    try { setLoading(true); const { data } = await apiGet('/api/v1/api-tokens'); setTokens(data.tokens || []) }
    catch (err: any) { setError(err?.response?.data?.error || 'Failed to load tokens') }
    finally { setLoading(false) }
  }

  const createToken = async () => {
    try {
      const expiresAt = expiresDays > 0 ? Math.floor(Date.now() / 1000) + expiresDays * 86400 : 0
      const { data } = await apiPost('/api/v1/api-tokens', { name: tokenName, kind: 'service', expiresAt })
      setNewTokenPlain(data.token || ''); setTokenName(''); fetchTokens()
    } catch (err: any) { setError(err?.response?.data?.error || 'Failed to create token') }
  }

  const toggleToken = async (id: number, enabled: boolean) => { try { await apiPut(`/api/v1/api-tokens/${id}/status`, { enabled }); fetchTokens() } catch (err: any) { setError(err?.response?.data?.error || 'Failed to update token') } }
  const deleteToken = async (id: number) => { if (!confirm('Delete this API token?')) return; try { await apiDelete(`/api/v1/api-tokens/${id}`); fetchTokens() } catch (err: any) { setError(err?.response?.data?.error || 'Failed to delete token') } }

  return (
    <div className="page-enter max-w-[1000px] mx-auto py-6 px-4 space-y-6">
      <div className="glass-panel p-5 flex items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" /></svg>
          </div>
          <div><h1 className="text-xl font-bold text-[var(--text-primary)]">API Tokens</h1><p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage service API tokens</p></div>
        </div>
        <button onClick={() => { setShowCreate(true); setNewTokenPlain('') }} className="btn-primary text-sm">+ New Token</button>
      </div>

      {error && <div className="glass-card p-3 flex items-center justify-between border-red-500/20"><span className="text-sm text-red-400">{error}</span><button onClick={() => setError('')} className="text-red-400 hover:text-red-300">×</button></div>}

      {loading ? <div className="flex justify-center py-12"><div className="loading-spinner loading-spinner-lg" /></div> : (
        <div className="glass-card overflow-hidden">
          <table className="table-modern">
            <thead><tr><th>Name</th><th>Kind</th><th>Scopes</th><th>Expires</th><th>Status</th><th className="text-right">Actions</th></tr></thead>
            <tbody>{tokens.map((token) => (
              <tr key={token.id}>
                <td className="text-sm font-medium text-[var(--text-primary)]">{token.name}</td>
                <td><span className={`badge ${token.kind === 'service' ? 'badge-cyan' : 'badge-purple'}`}>{token.kind}</span></td>
                <td className="text-xs text-[var(--text-muted)]">{token.scopes?.join(', ') || 'all'}</td>
                <td className={`text-xs ${token.expired ? 'text-[var(--danger)]' : 'text-[var(--text-muted)]'}`}>{token.expiresAt > 0 ? new Date(token.expiresAt * 1000).toLocaleDateString() : 'Never'}{token.expired && ' (expired)'}</td>
                <td><span className={`badge ${token.enabled ? 'badge-success' : 'badge-danger'}`}>{token.enabled ? 'Active' : 'Disabled'}</span></td>
                <td className="text-right"><div className="flex items-center justify-end gap-1.5">
                  <button onClick={() => toggleToken(token.id, !token.enabled)} className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-all ${token.enabled ? 'bg-amber-500/10 text-amber-400 hover:bg-amber-500/20' : 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20'}`}>{token.enabled ? 'Disable' : 'Enable'}</button>
                  <button onClick={() => deleteToken(token.id)} className="px-2.5 py-1 rounded-lg text-xs font-medium bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-all">Delete</button>
                </div></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 z-50" onClick={() => setShowCreate(false)}>
          <div className="glass-card p-7 w-full max-w-[450px]" onClick={e => e.stopPropagation()}>
            {newTokenPlain ? (
              <><h2 className="text-lg font-bold text-emerald-400 mb-3">✅ Token Created!</h2>
              <p className="text-xs text-amber-400 mb-4">Copy this token now. You won't see it again!</p>
              <div className="p-3 rounded-lg bg-[var(--bg-deep)] font-mono text-xs text-cyan-400 break-all mb-4 border border-[var(--border-light)]">{newTokenPlain}
                <button onClick={() => navigator.clipboard.writeText(newTokenPlain)} className="ml-2 px-2 py-0.5 rounded bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] text-[10px] border border-[var(--border-light)]">Copy</button></div>
              <button onClick={() => { setShowCreate(false); setNewTokenPlain('') }} className="btn-primary w-full text-sm justify-center">Done</button></>
            ) : (
              <><h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">Create API Token</h2>
              <div className="space-y-3">
                <input placeholder="Token Name" value={tokenName} onChange={(e) => setTokenName(e.target.value)} className="input-modern text-sm" />
                <input type="number" placeholder="Expires in (days, 0 = never)" value={expiresDays} onChange={(e) => setExpiresDays(Number(e.target.value))} className="input-modern text-sm" />
              </div>
              <div className="flex justify-end gap-2.5 mt-5">
                <button onClick={() => setShowCreate(false)} className="btn-ghost text-sm">Cancel</button>
                <button onClick={createToken} disabled={!tokenName} className="btn-primary text-sm disabled:opacity-50">Create</button>
              </div></>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
