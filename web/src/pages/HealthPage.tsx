import { useState, useEffect, useCallback } from 'react'
import { useI18n } from '../hooks/useI18n'

// ─── Types ───────────────────────────────────────────────────────────
interface HealthConfig {
  id: number
  name: string
  target: string
  probe_type: string
  interval: number
  timeout: number
  threshold: number
  expected_code: number
  enabled: boolean
  created_at: number
  updated_at: number
}

interface HealthStatus {
  config_id: number
  name: string
  target: string
  probe_type: string
  status: string
  success_count: number
  failure_count: number
  consecutive_failures: number
  last_latency_ms: number
  avg_latency_ms: number
  uptime_pct: number
  last_check_at: number
  last_success_at: number
  last_error: string
  enabled: boolean
}

interface RecoveryRule {
  id: number
  name: string
  match_label: string
  action_type: string
  action_params: string
  cooldown: number
  max_retries: number
  enabled: boolean
  created_at: number
}

interface RecoveryAction {
  id: number
  rule_id: number
  check_id: number
  action_type: string
  target: string
  status: string
  result: string
  latency_ms: number
  created_at: number
}

interface HealthResult {
  id: number
  config_id: number
  target: string
  probe_type: string
  success: boolean
  latency_ms: number
  error: string
  created_at: number
}

// ─── API client ──────────────────────────────────────────────────────
const api = {
  async get<T>(path: string): Promise<T> {
    const res = await fetch(path, { credentials: 'include' })
    const json = await res.json()
    if (!res.ok || !json.success) throw new Error(json.error || 'API error')
    return json.data as T
  },
  async post<T>(path: string, body?: unknown): Promise<T> {
    const res = await fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: body ? JSON.stringify(body) : undefined,
    })
    const json = await res.json()
    if (!res.ok || !json.success) throw new Error(json.error || 'API error')
    return json.data as T
  },
  async put<T>(path: string, body: unknown): Promise<T> {
    const res = await fetch(path, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    })
    const json = await res.json()
    if (!res.ok || !json.success) throw new Error(json.error || 'API error')
    return json.data as T
  },
  async del(path: string): Promise<void> {
    const res = await fetch(path, { method: 'DELETE', credentials: 'include' })
    const json = await res.json()
    if (!res.ok || !json.success) throw new Error(json.error || 'API error')
  },
}

// ─── Helpers ─────────────────────────────────────────────────────────
function statusColor(status: string): string {
  switch (status) {
    case 'healthy': return 'bg-emerald-500'
    case 'warning': return 'bg-amber-500'
    case 'critical': return 'bg-red-500'
    case 'down': return 'bg-rose-600'
    default: return 'bg-slate-500'
  }
}

function statusBg(status: string): string {
  switch (status) {
    case 'healthy': return 'bg-emerald-500/10 border-emerald-500/20'
    case 'warning': return 'bg-amber-500/10 border-amber-500/20'
    case 'critical': return 'bg-red-500/10 border-red-500/20'
    case 'down': return 'bg-rose-500/10 border-rose-500/20'
    default: return 'bg-slate-500/10 border-slate-500/20'
  }
}

function probeColor(type: string): string {
  switch (type) {
    case 'tcp': return 'from-cyan-500 to-blue-600'
    case 'http': return 'from-emerald-500 to-teal-600'
    case 'ping': return 'from-violet-500 to-purple-600'
    case 'grpc': return 'from-rose-500 to-pink-600'
    default: return 'from-slate-500 to-slate-600'
  }
}

function formatTime(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return d.toLocaleTimeString() + ' ' + d.toLocaleDateString()
}

function formatLatency(ms: number): string {
  if (ms < 1) return '<1ms'
  if (ms < 1000) return ms.toFixed(1) + 'ms'
  return (ms / 1000).toFixed(2) + 's'
}

// ─── Add Config Modal ────────────────────────────────────────────────
function AddConfigModal({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const { t } = useI18n()
  const [form, setForm] = useState<Partial<HealthConfig>>({
    name: '', target: '', probe_type: 'tcp', interval: 30, timeout: 5, threshold: 3, expected_code: 200, enabled: true,
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSave = async () => {
    if (!form.name || !form.target) { setError('Name and target are required'); return }
    setSaving(true)
    try {
      await api.post('/api/v1/health/configs', form)
      onSaved()
      onClose()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4 bg-[#12121a] rounded-2xl border border-[rgba(139,92,246,0.12)] shadow-2xl">
        <div className="p-5 border-b border-[rgba(255,255,255,0.06)]">
          <h2 className="text-lg font-bold text-white">Add Health Check</h2>
        </div>
        <div className="p-5 space-y-4">
          {error && (
            <div className="px-4 py-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">{error}</div>
          )}
          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Name</label>
              <input value={form.name || ''} onChange={e => setForm({...form, name: e.target.value})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" placeholder="e.g., Xray API Health" />
            </div>
            <div className="col-span-2">
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Target</label>
              <input value={form.target || ''} onChange={e => setForm({...form, target: e.target.value})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" placeholder="host:port or http://..." />
            </div>
            <div>
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Probe Type</label>
              <select value={form.probe_type} onChange={e => setForm({...form, probe_type: e.target.value})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none">
                <option value="tcp">TCP</option>
                <option value="http">HTTP</option>
                <option value="ping">Ping</option>
                <option value="grpc">gRPC</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Interval (s)</label>
              <input type="number" value={form.interval} onChange={e => setForm({...form, interval: parseInt(e.target.value) || 30})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" />
            </div>
            <div>
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Timeout (s)</label>
              <input type="number" value={form.timeout} onChange={e => setForm({...form, timeout: parseInt(e.target.value) || 5})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" />
            </div>
            <div>
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Threshold</label>
              <input type="number" value={form.threshold} onChange={e => setForm({...form, threshold: parseInt(e.target.value) || 3})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" placeholder="Consecutive failures" />
            </div>
            {form.probe_type === 'http' && (
              <div>
                <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Expected Code</label>
                <input type="number" value={form.expected_code} onChange={e => setForm({...form, expected_code: parseInt(e.target.value) || 200})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              </div>
            )}
          </div>
        </div>
        <div className="p-5 border-t border-[rgba(255,255,255,0.06)] flex justify-end gap-3">
          <button onClick={onClose} className="px-4 py-2 rounded-lg text-sm text-[#9898b8] hover:text-white bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] transition-all">Cancel</button>
          <button onClick={handleSave} disabled={saving} className="px-4 py-2 rounded-lg text-sm font-semibold text-white bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-500 hover:to-purple-600 transition-all disabled:opacity-50">
            {saving ? 'Saving...' : 'Save Check'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ─── Add Recovery Rule Modal ─────────────────────────────────────────
function AddRuleModal({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const { t } = useI18n()
  const [form, setForm] = useState<Partial<RecoveryRule>>({
    name: '', match_label: '', action_type: 'restart_core', cooldown: 300, max_retries: 3, enabled: true,
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSave = async () => {
    if (!form.name) { setError('Name is required'); return }
    setSaving(true)
    try {
      await api.post('/api/v1/health/recovery-rules', form)
      onSaved()
      onClose()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4 bg-[#12121a] rounded-2xl border border-[rgba(139,92,246,0.12)] shadow-2xl">
        <div className="p-5 border-b border-[rgba(255,255,255,0.06)]">
          <h2 className="text-lg font-bold text-white">Add Recovery Rule</h2>
        </div>
        <div className="p-5 space-y-4">
          {error && (
            <div className="px-4 py-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">{error}</div>
          )}
          <div>
            <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Name</label>
            <input value={form.name || ''} onChange={e => setForm({...form, name: e.target.value})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" placeholder="e.g., Auto-restart Xray" />
          </div>
          <div>
            <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Match Label</label>
            <input value={form.match_label || ''} onChange={e => setForm({...form, match_label: e.target.value})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" placeholder="Matches check name/target (optional)" />
          </div>
          <div>
            <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Action Type</label>
            <select value={form.action_type} onChange={e => setForm({...form, action_type: e.target.value})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none">
              <option value="restart_core">Restart Core</option>
              <option value="restart_node">Restart Node</option>
              <option value="reboot">Reboot Server</option>
              <option value="webhook">Webhook</option>
              <option value="script">Script</option>
            </select>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Cooldown (s)</label>
              <input type="number" value={form.cooldown} onChange={e => setForm({...form, cooldown: parseInt(e.target.value) || 300})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" />
            </div>
            <div>
              <label className="block text-xs font-medium text-[#9898b8] mb-1.5">Max Retries</label>
              <input type="number" value={form.max_retries} onChange={e => setForm({...form, max_retries: parseInt(e.target.value) || 3})} className="w-full px-3 py-2 rounded-lg bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-white text-sm focus:border-purple-500/40 focus:outline-none" />
            </div>
          </div>
        </div>
        <div className="p-5 border-t border-[rgba(255,255,255,0.06)] flex justify-end gap-3">
          <button onClick={onClose} className="px-4 py-2 rounded-lg text-sm text-[#9898b8] hover:text-white bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] transition-all">Cancel</button>
          <button onClick={handleSave} disabled={saving} className="px-4 py-2 rounded-lg text-sm font-semibold text-white bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-500 hover:to-purple-600 transition-all disabled:opacity-50">
            {saving ? 'Saving...' : 'Save Rule'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ─── History Modal ────────────────────────────────────────────────────
function HistoryModal({ configId, configName, onClose }: { configId: number; configName: string; onClose: () => void }) {
  const [results, setResults] = useState<HealthResult[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<HealthResult[]>(`/api/v1/health/configs/${configId}/history?limit=100`)
      .then(setResults)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [configId])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="w-full max-w-2xl mx-4 bg-[#12121a] rounded-2xl border border-[rgba(139,92,246,0.12)] shadow-2xl max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="p-5 border-b border-[rgba(255,255,255,0.06)] flex items-center justify-between">
          <h2 className="text-lg font-bold text-white">History: {configName}</h2>
          <button onClick={onClose} className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] flex items-center justify-center text-[#9898b8] hover:text-white transition-all">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M18 6L6 18M6 6l12 12" /></svg>
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-5">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-6 h-6 border-2 border-purple-500 border-t-transparent rounded-full animate-spin" />
            </div>
          ) : results.length === 0 ? (
            <p className="text-center py-12 text-[#6868a0]">No history available</p>
          ) : (
            <div className="space-y-2">
              {results.map(r => (
                <div key={r.id} className={`flex items-center gap-3 px-4 py-3 rounded-xl border ${r.success ? 'bg-emerald-500/5 border-emerald-500/10' : 'bg-red-500/5 border-red-500/10'}`}>
                  <div className={`w-2 h-2 rounded-full ${r.success ? 'bg-emerald-500' : 'bg-red-500'}`} />
                  <div className="flex-1">
                    <span className="text-sm text-white">{r.success ? 'Success' : 'Failed'}</span>
                    {r.error && <span className="text-xs text-red-400 ml-2">({r.error})</span>}
                  </div>
                  <span className="text-xs text-[#6868a0]">{formatLatency(r.latency_ms)}</span>
                  <span className="text-xs text-[#585878]">{formatTime(r.created_at)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// ─── Main HealthPage ──────────────────────────────────────────────────
export function HealthPage() {
  const { t } = useI18n()
  const [configs, setConfigs] = useState<HealthConfig[]>([])
  const [statuses, setStatuses] = useState<HealthStatus[]>([])
  const [rules, setRules] = useState<RecoveryRule[]>([])
  const [recoveryHistory, setRecoveryHistory] = useState<RecoveryAction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [tab, setTab] = useState<'checks' | 'recovery' | 'history'>('checks')
  const [showAddConfig, setShowAddConfig] = useState(false)
  const [showAddRule, setShowAddRule] = useState(false)
  const [historyModal, setHistoryModal] = useState<{ id: number; name: string } | null>(null)
  const [runningCheck, setRunningCheck] = useState<number | null>(null)

  const loadData = useCallback(async () => {
    try {
      const [cfg, st, rls, rh] = await Promise.all([
        api.get<HealthConfig[]>('/api/v1/health/configs'),
        api.get<HealthStatus[]>('/api/v1/health/statuses'),
        api.get<RecoveryRule[]>('/api/v1/health/recovery-rules'),
        api.get<RecoveryAction[]>('/api/v1/health/recovery-history?limit=20'),
      ])
      setConfigs(cfg)
      setStatuses(st)
      setRules(rls)
      setRecoveryHistory(rh)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadData() }, [loadData])

  // Auto-refresh every 10s
  useEffect(() => {
    const interval = setInterval(loadData, 10000)
    return () => clearInterval(interval)
  }, [loadData])

  const handleDeleteConfig = async (id: number) => {
    if (!confirm('Delete this health check configuration?')) return
    try {
      await api.del(`/api/v1/health/configs/${id}`)
      loadData()
    } catch (e: any) {
      alert(e.message)
    }
  }

  const handleDeleteRule = async (id: number) => {
    if (!confirm('Delete this recovery rule?')) return
    try {
      await api.del(`/api/v1/health/recovery-rules/${id}`)
      loadData()
    } catch (e: any) {
      alert(e.message)
    }
  }

  const handleManualCheck = async (configId: number) => {
    setRunningCheck(configId)
    try {
      await api.post('/api/v1/health/manual-check', { config_id: configId })
      loadData()
    } catch (e: any) {
      alert(e.message)
    } finally {
      setRunningCheck(null)
    }
  }

  // Stats derived from statuses
  const total = statuses.length
  const healthy = statuses.filter(s => s.status === 'healthy').length
  const warning = statuses.filter(s => s.status === 'warning').length
  const critical = statuses.filter(s => s.status === 'critical' || s.status === 'down').length
  const avgLatency = statuses.reduce((sum, s) => sum + s.avg_latency_ms, 0) / (total || 1)
  const avgUptime = statuses.reduce((sum, s) => sum + s.uptime_pct, 0) / (total || 1)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="w-8 h-8 border-2 border-purple-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* ─── Header ────────────────────────────────────────────── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Health Check</h1>
          <p className="text-sm text-[#6868a0] mt-1">Smart health monitoring & auto-recovery system</p>
        </div>
        <div className="flex gap-2">
          {tab === 'checks' && (
            <button onClick={() => setShowAddConfig(true)} className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-500 hover:to-purple-600 text-white text-sm font-semibold transition-all shadow-lg shadow-purple-500/10">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 4v16m8-8H4" /></svg>
              Add Check
            </button>
          )}
          {tab === 'recovery' && (
            <button onClick={() => setShowAddRule(true)} className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-500 hover:to-purple-600 text-white text-sm font-semibold transition-all shadow-lg shadow-purple-500/10">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 4v16m8-8H4" /></svg>
              Add Rule
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="px-4 py-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">{error}</div>
      )}

      {/* ─── Status Overview Cards ──────────────────────────────── */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
        <div className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)]">
          <p className="text-xs text-[#6868a0] font-medium mb-1">Total Checks</p>
          <p className="text-2xl font-bold text-white">{total}</p>
        </div>
        <div className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)]">
          <p className="text-xs text-[#6868a0] font-medium mb-1">Healthy</p>
          <p className="text-2xl font-bold text-emerald-400">{healthy}</p>
        </div>
        <div className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)]">
          <p className="text-xs text-[#6868a0] font-medium mb-1">Warning</p>
          <p className="text-2xl font-bold text-amber-400">{warning}</p>
        </div>
        <div className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)]">
          <p className="text-xs text-[#6868a0] font-medium mb-1">Critical</p>
          <p className="text-2xl font-bold text-red-400">{critical}</p>
        </div>
        <div className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)]">
          <p className="text-xs text-[#6868a0] font-medium mb-1">Avg Uptime</p>
          <p className="text-2xl font-bold text-white">{avgUptime.toFixed(1)}%</p>
        </div>
      </div>

      {/* ─── Tabs ────────────────────────────────────────────────── */}
      <div className="flex gap-1 p-1 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] w-fit">
        <button onClick={() => setTab('checks')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'checks' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
          Health Checks
        </button>
        <button onClick={() => setTab('recovery')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'recovery' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
          Recovery Rules
        </button>
        <button onClick={() => setTab('history')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'history' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>
          Recovery History
        </button>
      </div>

      {/* ─── Tab: Health Checks ─────────────────────────────────--- */}
      {tab === 'checks' && (
        <div className="space-y-3">
          {configs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-2xl bg-[rgba(255,255,255,0.03)] flex items-center justify-center mb-4">
                <svg className="w-8 h-8 text-[#585878]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
              </div>
              <p className="text-[#6868a0] font-medium">No health checks configured</p>
              <p className="text-sm text-[#585878] mt-1">Add your first health check to start monitoring</p>
            </div>
          ) : (
            configs.map(cfg => {
              const status = statuses.find(s => s.config_id === cfg.id)
              return (
                <div key={cfg.id} className={`p-4 rounded-xl border transition-all hover:bg-[rgba(255,255,255,0.01)] ${status ? statusBg(status.status) : 'bg-[rgba(255,255,255,0.02)] border-[rgba(255,255,255,0.06)]'}`}>
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-3">
                      <div className={`w-8 h-8 rounded-lg bg-gradient-to-br ${probeColor(cfg.probe_type)} flex items-center justify-center text-white text-[10px] font-bold uppercase`}>
                        {cfg.probe_type}
                      </div>
                      <div>
                        <h3 className="text-sm font-semibold text-white">{cfg.name}</h3>
                        <p className="text-xs text-[#6868a0]">{cfg.target} · every {cfg.interval}s</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {status && (
                        <span className={`px-2 py-1 rounded-md text-[10px] font-bold uppercase ${status.status === 'healthy' ? 'bg-emerald-500/10 text-emerald-400' : status.status === 'warning' ? 'bg-amber-500/10 text-amber-400' : 'bg-red-500/10 text-red-400'}`}>
                          {status.status}
                        </span>
                      )}
                      <button onClick={() => handleManualCheck(cfg.id)} disabled={runningCheck === cfg.id} className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] flex items-center justify-center text-[#6868a0] hover:text-white transition-all" title="Run manual check">
                        {runningCheck === cfg.id ? (
                          <div className="w-3.5 h-3.5 border-2 border-purple-500 border-t-transparent rounded-full animate-spin" />
                        ) : (
                          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                        )}
                      </button>
                      <button onClick={() => setHistoryModal({ id: cfg.id, name: cfg.name })} className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] flex items-center justify-center text-[#6868a0] hover:text-white transition-all" title="View history">
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" /></svg>
                      </button>
                      <button onClick={() => handleDeleteConfig(cfg.id)} className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-red-500/10 flex items-center justify-center text-[#6868a0] hover:text-red-400 transition-all" title="Delete">
                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                      </button>
                    </div>
                  </div>
                  {status && (
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                      <div>
                        <p className="text-[10px] text-[#585878] font-medium">Latency</p>
                        <p className="text-xs font-semibold text-white">{formatLatency(status.last_latency_ms)}</p>
                      </div>
                      <div>
                        <p className="text-[10px] text-[#585878] font-medium">Avg Latency</p>
                        <p className="text-xs font-semibold text-white">{formatLatency(status.avg_latency_ms)}</p>
                      </div>
                      <div>
                        <p className="text-[10px] text-[#585878] font-medium">Uptime</p>
                        <p className="text-xs font-semibold text-white">{status.uptime_pct.toFixed(1)}%</p>
                      </div>
                      <div>
                        <p className="text-[10px] text-[#585878] font-medium">Failures</p>
                        <p className="text-xs font-semibold text-white">{status.consecutive_failures}/{cfg.threshold}</p>
                      </div>
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>
      )}

      {/* ─── Tab: Recovery Rules ───────────────────────────────── */}
      {tab === 'recovery' && (
        <div className="space-y-3">
          {rules.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-2xl bg-[rgba(255,255,255,0.03)] flex items-center justify-center mb-4">
                <svg className="w-8 h-8 text-[#585878]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
              </div>
              <p className="text-[#6868a0] font-medium">No recovery rules configured</p>
              <p className="text-sm text-[#585878] mt-1">Auto-recovery rules will respond when health checks fail</p>
            </div>
          ) : (
            rules.map(rule => (
              <div key={rule.id} className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] transition-all hover:bg-[rgba(255,255,255,0.01)]">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className={`w-8 h-8 rounded-lg ${rule.enabled ? 'bg-gradient-to-br from-rose-500 to-pink-600' : 'bg-[rgba(255,255,255,0.06)]'} flex items-center justify-center text-white text-[10px] font-bold uppercase`}>
                      <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold text-white">{rule.name}</h3>
                      <p className="text-xs text-[#6868a0]">{rule.action_type}{rule.match_label ? ` · matches: ${rule.match_label}` : ''}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`px-2 py-1 rounded-md text-[10px] font-bold ${rule.enabled ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-500/10 text-slate-400'}`}>
                      {rule.enabled ? 'Active' : 'Disabled'}
                    </span>
                    <button onClick={() => handleDeleteRule(rule.id)} className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.03)] hover:bg-red-500/10 flex items-center justify-center text-[#6868a0] hover:text-red-400 transition-all" title="Delete rule">
                      <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                    </button>
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <p className="text-[10px] text-[#585878] font-medium">Cooldown</p>
                    <p className="text-xs font-semibold text-white">{rule.cooldown}s</p>
                  </div>
                  <div>
                    <p className="text-[10px] text-[#585878] font-medium">Max Retries</p>
                    <p className="text-xs font-semibold text-white">{rule.max_retries}</p>
                  </div>
                  <div>
                    <p className="text-[10px] text-[#585878] font-medium">Created</p>
                    <p className="text-xs font-semibold text-white">{formatTime(rule.created_at)}</p>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* ─── Tab: Recovery History ──────────────────────────────── */}
      {tab === 'history' && (
        <div className="space-y-3">
          {recoveryHistory.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-2xl bg-[rgba(255,255,255,0.03)] flex items-center justify-center mb-4">
                <svg className="w-8 h-8 text-[#585878]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              </div>
              <p className="text-[#6868a0] font-medium">No recovery actions yet</p>
              <p className="text-sm text-[#585878] mt-1">Recovery actions will appear here when health checks fail</p>
            </div>
          ) : (
            recoveryHistory.map(action => (
              <div key={action.id} className={`p-4 rounded-xl border transition-all ${action.status === 'success' ? 'bg-emerald-500/5 border-emerald-500/10' : action.status === 'failed' ? 'bg-red-500/5 border-red-500/10' : 'bg-amber-500/5 border-amber-500/10'}`}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`w-2 h-2 rounded-full ${action.status === 'success' ? 'bg-emerald-500' : action.status === 'failed' ? 'bg-red-500' : 'bg-amber-500'}`} />
                    <div>
                      <span className="text-sm font-medium text-white">{action.action_type}</span>
                      <span className="text-xs text-[#6868a0] ml-2">→ {action.target}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={`px-2 py-1 rounded-md text-[10px] font-bold uppercase ${action.status === 'success' ? 'bg-emerald-500/10 text-emerald-400' : action.status === 'failed' ? 'bg-red-500/10 text-red-400' : 'bg-amber-500/10 text-amber-400'}`}>
                      {action.status}
                    </span>
                    <span className="text-xs text-[#585878]">{formatLatency(action.latency_ms)}</span>
                    <span className="text-xs text-[#585878]">{formatTime(action.created_at)}</span>
                  </div>
                </div>
                {action.result && (
                  <p className="text-xs text-[#6868a0] mt-2 ml-5">{action.result}</p>
                )}
              </div>
            ))
          )}
        </div>
      )}

      {/* ─── Modals ──────────────────────────────────────────────── */}
      {showAddConfig && <AddConfigModal onClose={() => setShowAddConfig(false)} onSaved={loadData} />}
      {showAddRule && <AddRuleModal onClose={() => setShowAddRule(false)} onSaved={loadData} />}
      {historyModal && <HistoryModal configId={historyModal.id} configName={historyModal.name} onClose={() => setHistoryModal(null)} />}
    </div>
  )
}

export default HealthPage
