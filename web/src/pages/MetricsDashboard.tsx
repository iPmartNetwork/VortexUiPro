import { useState, useRef, useEffect, useCallback } from 'react'
import { apiGet } from '../api/client'
import { useI18n } from '../hooks/useI18n'

// ─── Types ───────────────────────────────────────────────────────────

interface MetricsSnapshot {
  users_total: number
  inbounds_total: number
  clients_total: number
  nodes_total: number
  online_now: number
  traffic_up_gb: number
  traffic_down_gb: number

  cpu_percent: number
  cpu_threads: number
  load_avg_1: number
  load_avg_5: number
  load_avg_15: number

  memory_total_mb: number
  memory_used_mb: number
  memory_pct: number

  disk_total_mb: number
  disk_used_mb: number
  disk_free_mb: number
  disk_used_percent: number

  net_bytes_sent: number
  net_bytes_recv: number
  net_packets_sent: number
  net_packets_recv: number

  go_routines: number
  uptime_seconds: number
  uptime_human: string

  hostname: string
  os: string
  arch: string
  started_at: number
}

interface MetricPoint {
  t: number
  cpu: number
  mem: number
  disk: number
  net_sent: number
  net_recv: number
  online: number
  goroutines: number
}

interface WSMessage {
  type: string
  payload: any
  time: number
}

// ─── Mini Sparkline ─────────────────────────────────────────────────

function Sparkline({ data, color, height = 32, width: w = 120 }: {
  data: number[]
  color: string
  height?: number
  width?: number
}) {
  if (!data || data.length < 2) return <div className="h-8" />

  const max = Math.max(...data, 1)
  const min = Math.min(...data, 0)
  const range = max - min || 1
  const stepX = w / (data.length - 1)

  const pts = data.map((v, i) => {
    const x = i * stepX
    const y = height - ((v - min) / range) * height
    return `${x},${y}`
  }).join(' ')

  return (
    <svg viewBox={`0 0 ${w} ${height}`} className="w-full h-full" preserveAspectRatio="none">
      <polyline
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        points={pts}
        className="drop-shadow-[0_0_3px_var(--tw-shadow)]"
        style={{ filter: `drop-shadow(0 0 2px ${color}40)` }}
      />
    </svg>
  )
}

// ─── Gauge Widget ───────────────────────────────────────────────────

function GaugeWidget({ label, value, unit, percent, color, sparklineData, subtitle }: {
  label: string
  value: string | number
  unit?: string
  percent: number
  color: string
  sparklineData?: number[]
  subtitle?: string
}) {
  const clamped = Math.min(Math.max(percent, 0), 100)
  const dashArray = 2 * Math.PI * 36
  const dashOffset = dashArray - (clamped / 100) * dashArray

  const warning = clamped > 80 ? 'from-red-500 to-orange-500' :
    clamped > 60 ? 'from-yellow-500 to-orange-500' :
    `from-${color}-500 to-${color}-400`

  return (
    <div className="glass-card p-5 relative overflow-hidden group hover:border-[var(--border-hover)] transition-all duration-300">
      {/* Subtle gradient background */}
      <div className={`absolute inset-0 bg-gradient-to-br ${warning} opacity-[0.03] group-hover:opacity-[0.05] transition-opacity`} />

      <div className="relative z-10">
        {/* Gauge circle */}
        <div className="flex items-center gap-4">
          <div className="relative shrink-0">
            <svg width="88" height="88" viewBox="0 0 88 88" className="transform -rotate-90">
              <circle cx="44" cy="44" r="36" fill="none" stroke="currentColor" strokeWidth="5"
                className="text-[var(--border-light)] opacity-50" />
              <circle cx="44" cy="44" r="36" fill="none" strokeWidth="5" strokeLinecap="round"
                stroke={color}
                strokeDasharray={dashArray}
                strokeDashoffset={dashOffset}
                className="transition-all duration-700 ease-out"
                style={{ filter: `drop-shadow(0 0 4px ${color}60)` }}
              />
            </svg>
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-lg font-bold text-[var(--text-primary)]" style={{ color }}>
                {Math.round(clamped)}%
              </span>
            </div>
          </div>

          <div className="flex-1 min-w-0">
            <p className="text-xs text-[var(--text-muted)] uppercase tracking-wider font-medium">{label}</p>
            <p className="text-xl font-bold text-[var(--text-primary)] mt-0.5 font-mono">
              {value}
              {unit && <span className="text-sm text-[var(--text-secondary)] ml-1 font-normal">{unit}</span>}
            </p>
            {subtitle && (
              <p className="text-xs text-[var(--text-secondary)] mt-1">{subtitle}</p>
            )}
          </div>
        </div>

        {/* Sparkline */}
        {sparklineData && sparklineData.length > 1 && (
          <div className="mt-3 h-8 opacity-60 group-hover:opacity-100 transition-opacity">
            <Sparkline data={sparklineData} color={color} height={32} width={120} />
          </div>
        )}
      </div>
    </div>
  )
}

// ─── Stat Tile ──────────────────────────────────────────────────────

function StatTile({ label, value, icon, color, subtitle, trend }: {
  label: string
  value: string | number
  icon: string
  color: string
  subtitle?: string
  trend?: { value: number; up: boolean }
}) {
  return (
    <div className="glass-card p-4 relative overflow-hidden group hover:border-[var(--border-hover)] transition-all duration-300">
      <div className={`absolute top-0 right-0 w-24 h-24 bg-gradient-to-bl from-${color}-500/5 to-transparent rounded-bl-full group-hover:from-${color}-500/10 transition-all`} />
      <div className="relative z-10">
        <div className="flex items-center justify-between mb-2">
          <p className="text-xs text-[var(--text-muted)] uppercase tracking-wider font-medium">{label}</p>
          <span className="text-lg opacity-60 group-hover:opacity-100 transition-opacity">{icon}</span>
        </div>
        <p className="text-2xl font-bold text-[var(--text-primary)] font-mono">{value}</p>
        {subtitle && (
          <p className="text-xs text-[var(--text-secondary)] mt-1">{subtitle}</p>
        )}
        {trend && (
          <div className={`flex items-center gap-1 mt-1.5 text-xs ${trend.up ? 'text-green-400' : 'text-red-400'}`}>
            <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={trend.up ? 'M5 10l7-7m0 0l7 7m-7-7v18' : 'M19 14l-7 7m0 0l-7-7m7 7V3'} />
            </svg>
            <span>{trend.value}%</span>
          </div>
        )}
      </div>
    </div>
  )
}

// ─── Bar Chart (Mini) ───────────────────────────────────────────────

function MiniBarChart({ data, color, height = 40 }: { data: number[]; color: string; height?: number }) {
  if (!data || data.length === 0) return null
  const max = Math.max(...data, 1)
  return (
    <div className="flex items-end gap-[2px] h-full" style={{ height }}>
      {data.map((v, i) => (
        <div
          key={i}
          className="flex-1 rounded-t-sm transition-all duration-500"
          style={{
            height: `${(v / max) * 100}%`,
            backgroundColor: color,
            opacity: 0.4 + (v / max) * 0.6,
          }}
        />
      ))}
    </div>
  )
}

// ─── Network Chart ──────────────────────────────────────────────────

function NetworkChart({ sent, recv }: { sent: number[]; recv: number[] }) {
  const max = Math.max(...sent, ...recv, 1)

  return (
    <div className="glass-card p-5">
      <h3 className="text-sm font-bold text-[var(--text-primary)] mb-3 flex items-center gap-2">
        <svg className="w-4 h-4 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
        </svg>
        {t('metrics.network')}
      </h3>
      <div className="flex gap-6 items-start">
        {/* Upload */}
        <div className="flex-1">
          <p className="text-xs text-[var(--text-muted)] mb-2 flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-blue-400" />
            {t('metrics.upload')}
          </p>
          <div className="h-12 flex items-end gap-[2px]">
            {sent.map((v, i) => (
              <div key={i} className="flex-1 rounded-t-sm transition-all duration-500"
                style={{ height: `${(v / max) * 100}%`, backgroundColor: '#60a5fa', opacity: 0.4 + (v / max) * 0.6 }} />
            ))}
          </div>
        </div>
        {/* Download */}
        <div className="flex-1">
          <p className="text-xs text-[var(--text-muted)] mb-2 flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-emerald-400" />
            {t('metrics.download')}
          </p>
          <div className="h-12 flex items-end gap-[2px]">
            {recv.map((v, i) => (
              <div key={i} className="flex-1 rounded-t-sm transition-all duration-500"
                style={{ height: `${(v / max) * 100}%`, backgroundColor: '#34d399', opacity: 0.4 + (v / max) * 0.6 }} />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

// ─── Formatters ─────────────────────────────────────────────────────

function bytesToHuman(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

function durationHuman(sec: number): string {
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = Math.floor(sec % 60)
  const parts: string[] = []
  if (d > 0) parts.push(`${d}d`)
  if (h > 0) parts.push(`${h}h`)
  if (m > 0) parts.push(`${m}m`)
  parts.push(`${s}s`)
  return parts.join(' ')
}

// ─── Main Component ────────────────────────────────────────────────

export function MetricsDashboardPage() {
  const { t } = useI18n()
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null)
  const [loading, setLoading] = useState(true)
  const [history, setHistory] = useState<MetricPoint[]>([])
  const [wsStatus, setWsStatus] = useState<'connected' | 'disconnected' | 'connecting'>('connecting')
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>()

  // ─── WebSocket Connection ─────────────────────────────────────
  const connectWS = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${window.location.host}/ws`

    try {
      const ws = new WebSocket(url)
      wsRef.current = ws
      setWsStatus('connecting')

      ws.onopen = () => {
        setWsStatus('connected')
      }

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          if (msg.type === 'system' && msg.payload) {
            setMetrics(msg.payload as MetricsSnapshot)
            setLoading(false)
          }
        } catch { /* ignore parse errors */ }
      }

      ws.onclose = () => {
        setWsStatus('disconnected')
        reconnectTimer.current = setTimeout(connectWS, 5000)
      }

      ws.onerror = () => {
        ws.close()
      }
    } catch {
      setWsStatus('disconnected')
      reconnectTimer.current = setTimeout(connectWS, 5000)
    }
  }, [])

  // ─── Fetch HTTP (fallback + initial load + history) ──────────
  const fetchMetrics = useCallback(async () => {
    try {
      const { data } = await apiGet('/api/v1/metrics')
      setMetrics(data)
      setLoading(false)
    } catch { /* WS will handle */ }
  }, [])

  const fetchHistory = useCallback(async () => {
    try {
      const { data } = await apiGet('/api/v1/metrics/history')
      if (data?.history) setHistory(data.history as MetricPoint[])
    } catch { /* ignore */ }
  }, [])

  // ─── Init ────────────────────────────────────────────────────
  useEffect(() => {
    fetchMetrics()
    fetchHistory() // One-time initial load of server history
    connectWS()

    // HTTP polling fallback if WS fails (only metrics, not history — WS drives history)
    const pollTimer = setInterval(fetchMetrics, 30000)

    return () => {
      clearInterval(pollTimer)
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [])

  // ─── Update history from WS ──────────────────────────────────
  useEffect(() => {
    if (!metrics) return
    const pt: MetricPoint = {
      t: Date.now(),
      cpu: metrics.cpu_percent || 0,
      mem: metrics.memory_pct || 0,
      disk: metrics.disk_used_percent || 0,
      net_sent: metrics.net_bytes_sent || 0,
      net_recv: metrics.net_bytes_recv || 0,
      online: metrics.online_now || 0,
      goroutines: metrics.go_routines || 0,
    }
    setHistory(prev => {
      const next = [...prev, pt]
      return next.length > 120 ? next.slice(-120) : next // keep last 120 points (30 min)
    })
  }, [metrics])

  // ─── Extract chart data ──────────────────────────────────────
  const cpuData = history.map(p => p.cpu)
  const memData = history.map(p => p.mem)
  const diskData = history.map(p => p.disk)
  const onlineData = history.map(p => p.online)
  const goroutineData = history.map(p => p.goroutines)
  const netSentData = history.map(p => p.net_sent)
  const netRecvData = history.map(p => p.net_recv)

  const memoryGB = metrics ? (metrics.memory_total_mb / 1024).toFixed(1) : '0'
  const trafficTotal = metrics ? (metrics.traffic_up_gb + metrics.traffic_down_gb) : 0

  // ─── Render ──────────────────────────────────────────────────
  return (
    <div className="space-y-6 page-enter">
      {/* ─── Header ──────────────────────────────────────────── */}
      <div className="glass-panel p-5">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div className="flex items-center gap-4">
            <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-violet-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
              <svg className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold text-[var(--text-primary)]">{t('metrics.title')}</h1>
              <p className="text-sm text-[var(--text-secondary)] mt-0.5">{t('metrics.subtitle')}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {/* WebSocket status */}
            <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium
              ${wsStatus === 'connected' ? 'bg-green-500/10 text-green-400 border border-green-500/20' :
                wsStatus === 'connecting' ? 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/20' :
                'bg-red-500/10 text-red-400 border border-red-500/20'}`}>
              <span className={`w-1.5 h-1.5 rounded-full ${
                wsStatus === 'connected' ? 'bg-green-400 animate-pulse' :
                wsStatus === 'connecting' ? 'bg-yellow-400 animate-pulse' : 'bg-red-400'
              }`} />
              {wsStatus === 'connected' ? t('metrics.live') : wsStatus === 'connecting' ? t('metrics.connecting') : t('metrics.offline')}
            </div>
            <a href="/metrics" target="_blank" rel="noopener noreferrer"
              className="px-3 py-1.5 bg-orange-600/20 text-orange-400 rounded-lg hover:bg-orange-600/30 transition text-xs font-medium border border-orange-500/20">
              {t('metrics.prometheus')}
            </a>
            <button onClick={fetchMetrics}
              className="p-2 rounded-lg bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-surface)] border border-[var(--border-light)] transition">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-24">
          <div className="relative">
            <div className="w-10 h-10 border-2 border-violet-500/30 border-t-violet-500 rounded-full animate-spin" />
            <div className="w-6 h-6 border-2 border-cyan-500/30 border-b-cyan-500 rounded-full animate-spin absolute inset-0 m-auto" style={{ animationDirection: 'reverse', animationDuration: '0.8s' }} />
          </div>
        </div>
      ) : !metrics ? (
        <div className="glass-card p-16 text-center">
          <div className="w-16 h-16 rounded-full bg-[var(--bg-elevated)] flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <p className="text-[var(--text-muted)] text-sm">{t('metrics.noMetrics')}</p>
        </div>
      ) : (
        <>
          {/* ─── System Gauges Row ────────────────────────────── */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
            <GaugeWidget
              label={t('metrics.cpu')}
              value={`${metrics.cpu_percent.toFixed(1)}%`}
              percent={metrics.cpu_percent}
              color="#8b5cf6"
              sparklineData={cpuData}
              subtitle={`${metrics.cpu_threads} ${t('metrics.threads')} · Load: ${metrics.load_avg_1.toFixed(1)}`}
            />
            <GaugeWidget
              label={t('metrics.memory')}
              value={`${metrics.memory_used_mb.toFixed(0)} MB`}
              unit={`/ ${memoryGB} GB`}
              percent={metrics.memory_pct}
              color="#06b6d4"
              sparklineData={memData}
              subtitle={`${(metrics.memory_used_mb / 1024).toFixed(1)} GB ${t('metrics.used')}`}
            />
            <GaugeWidget
              label={t('metrics.disk')}
              value={`${metrics.disk_used_percent.toFixed(1)}%`}
              percent={metrics.disk_used_percent}
              color="#10b981"
              sparklineData={diskData}
              subtitle={`${(metrics.disk_used_mb / 1024).toFixed(1)} GB / ${(metrics.disk_total_mb / 1024).toFixed(1)} GB`}
            />
          </div>

          {/* ─── Stats Row ─────────────────────────────────────── */}
          <div className="grid gap-3 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6">
            <StatTile label={t('metrics.users')} value={metrics.users_total} color="purple" icon="👤" />
            <StatTile label={t('metrics.inbounds')} value={metrics.inbounds_total} color="blue" icon="📡" />
            <StatTile label={t('metrics.clients')} value={metrics.clients_total} color="green" icon="🔗" />
            <StatTile label={t('metrics.nodes')} value={metrics.nodes_total} color="yellow" icon="🖥️" />
            <StatTile label={t('metrics.onlineUsers')} value={metrics.online_now} color="orange" icon="🟢" subtitle={t('metrics.realtime')} />
            <StatTile label={t('metrics.uptime')} value={durationHuman(metrics.uptime_seconds)} color="cyan" icon="⏱️" />
          </div>

          {/* ─── Network & Traffic ──────────────────────────────── */}
          <div className="grid gap-4 grid-cols-1 lg:grid-cols-2">
            <NetworkChart sent={netSentData} recv={netRecvData} />

            <div className="glass-card p-5">
              <h3 className="text-sm font-bold text-[var(--text-primary)] mb-3 flex items-center gap-2">
                <svg className="w-4 h-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0l-4-4m4 4l-4 4M5 3l.964 4.714A6 6 0 0011.75 13h.5a6 6 0 005.786-4.714L19 3" />
                </svg>
                {t('metrics.traffic')}
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 rounded-lg bg-[var(--bg-elevated)] border border-[var(--border-light)]">                    <p className="text-xs text-[var(--text-muted)] mb-1 flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-blue-400" /> {t('metrics.trafficUp')}
                  </p>
                  <p className="text-lg font-bold text-blue-400 font-mono">{metrics.traffic_up_gb.toFixed(2)} <span className="text-xs text-[var(--text-secondary)]">GB</span></p>
                </div>
                <div className="p-3 rounded-lg bg-[var(--bg-elevated)] border border-[var(--border-light)]">
                  <p className="text-xs text-[var(--text-muted)] mb-1 flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-emerald-400" /> {t('metrics.trafficDown')}
                  </p>
                  <p className="text-lg font-bold text-emerald-400 font-mono">{metrics.traffic_down_gb.toFixed(2)} <span className="text-xs text-[var(--text-secondary)]">GB</span></p>
                </div>
                <div className="p-3 rounded-lg bg-[var(--bg-elevated)] border border-[var(--border-light)] col-span-2">
                  <p className="text-xs text-[var(--text-muted)] mb-1">{t('metrics.trafficTotal')}</p>
                  <p className="text-xl font-bold text-[var(--text-primary)] font-mono">{trafficTotal.toFixed(2)} <span className="text-sm text-[var(--text-secondary)] font-normal">GB</span></p>
                </div>
              </div>
            </div>
          </div>

          {/* ─── Load & Online Sparklines ──────────────────────── */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
            {/* CPU Load Averages */}
            <div className="glass-card p-4">
              <p className="text-xs text-[var(--text-muted)] uppercase tracking-wider mb-2">{t('metrics.loadAvg')}</p>
              <div className="flex gap-4">
                {[
                  { label: '1m', value: metrics.load_avg_1, color: '#f59e0b' },
                  { label: '5m', value: metrics.load_avg_5, color: '#22d3ee' },
                  { label: '15m', value: metrics.load_avg_15, color: '#a78bfa' },
                ].map(item => (
                  <div key={item.label} className="text-center">
                    <p className="text-xs text-[var(--text-muted)]">{item.label}</p>
                    <p className="text-lg font-bold font-mono text-[var(--text-primary)]" style={{ color: item.color }}>
                      {item.value.toFixed(2)}
                    </p>
                  </div>
                ))}
              </div>
            </div>

            {/* Online Users Sparkline */}
            <div className="glass-card p-4">
              <p className="text-xs text-[var(--text-muted)] uppercase tracking-wider mb-2">{t('metrics.onlineUsers')} ({t('metrics.history')})</p>
              <div className="h-16">
                <MiniBarChart data={onlineData} color="#f97316" height={60} />
              </div>
              <p className="text-right text-xs text-[var(--text-secondary)] mt-1">
                {t('metrics.current')}: <span className="text-orange-400 font-bold font-mono">{metrics.online_now}</span>
              </p>
            </div>

            {/* Go Routines Sparkline */}
            <div className="glass-card p-4">
              <p className="text-xs text-[var(--text-muted)] uppercase tracking-wider mb-2">{t('metrics.goRoutines')}</p>
              <div className="h-16">
                <MiniBarChart data={goroutineData} color="#8b5cf6" height={60} />
              </div>
              <p className="text-right text-xs text-[var(--text-secondary)] mt-1">
                {t('metrics.current')}: <span className="text-violet-400 font-bold font-mono">{metrics.go_routines}</span>
              </p>
            </div>

            {/* Memory bar */}
            <div className="glass-card p-4">
              <p className="text-xs text-[var(--text-muted)] uppercase tracking-wider mb-2">{t('metrics.memoryUsage')}</p>
              <div className="h-16 flex items-center justify-center">
                <div className="w-full bg-[var(--bg-elevated)] rounded-full h-5 overflow-hidden border border-[var(--border-light)]">
                  <div className="h-full rounded-full transition-all duration-1000 ease-out"
                    style={{
                      width: `${metrics.memory_pct}%`,
                      background: `linear-gradient(90deg, #06b6d4, ${metrics.memory_pct > 80 ? '#ef4444' : metrics.memory_pct > 60 ? '#f59e0b' : '#06b6d4'})`,
                    }}
                  />
                </div>
              </div>
              <p className="text-right text-xs text-[var(--text-secondary)] mt-1">
                {metrics.memory_pct.toFixed(1)}% · {metrics.memory_used_mb.toFixed(0)} MB / {memoryGB} GB
              </p>
            </div>
          </div>

          {/* ─── System Info ────────────────────────────────────── */}
          <div className="glass-card p-5">
            <h3 className="text-sm font-bold text-[var(--text-primary)] mb-3 flex items-center gap-2">
              <svg className="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              {t('metrics.sysInfo')}
            </h3>
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
              {[
                { label: t('metrics.hostname'), value: metrics.hostname || '-' },
                { label: t('metrics.os'), value: metrics.os || '-' },
                { label: t('metrics.architecture'), value: metrics.arch || '-' },
                { label: t('metrics.cpuThreads'), value: metrics.cpu_threads.toString() },
                { label: t('metrics.goRoutines'), value: metrics.go_routines.toLocaleString() },
                { label: t('metrics.uptime'), value: metrics.uptime_human },
              ].map(item => (
                <div key={item.label} className="p-3 rounded-lg bg-[var(--bg-elevated)] border border-[var(--border-light)]">
                  <p className="text-xs text-[var(--text-muted)] mb-0.5">{item.label}</p>
                  <p className="text-sm font-mono font-semibold text-[var(--text-primary)] truncate">{item.value}</p>
                </div>
              ))}
            </div>
          </div>

          {/* ─── Prometheus ──────────────────────────────────────── */}
          <details className="glass-card p-5 group">
            <summary className="cursor-pointer text-sm font-bold text-[var(--text-primary)] flex items-center gap-2 select-none">
              <svg className="w-4 h-4 text-orange-400 group-open:rotate-90 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
              {t('metrics.prometheus')} Scrape Endpoint
            </summary>
            <div className="mt-3 text-sm text-[var(--text-secondary)]">
              <p className="mb-2">
                Configure Prometheus to scrape{' '}
                <code className="px-1.5 py-0.5 rounded bg-[var(--bg-elevated)] text-orange-400 text-xs font-mono">http://your-server:8080/metrics</code>
              </p>
              <a href="/metrics" target="_blank" rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-orange-600/20 text-orange-400 rounded-lg hover:bg-orange-600/30 transition text-xs font-medium border border-orange-500/20">
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
                Open /metrics
              </a>
            </div>
          </details>
        </>
      )}
    </div>
  )
}
