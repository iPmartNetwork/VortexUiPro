import { useState, useEffect, useRef, useCallback } from 'react'
import { apiClient } from '../api/client'
import { formatBytes } from '../utils/format'
import {
  LineChart, Line, BarChart, Bar, AreaChart, Area,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell,
} from 'recharts'

// ─── Types ───────────────────────────────────────────────────────────
interface DashboardStats {
  total_users: number; active_users: number; expired_users: number
  total_inbounds: number; total_nodes: number; online_nodes: number
  total_tickets: number; open_tickets: number
  traffic_up: number; traffic_down: number
  revenue_total: number; transactions: number
  revenue_today: number; users_today: number
  memory_used_mb: number; uptime_seconds: number
}

interface TrafficPoint { date: string; up: number; down: number; total: number }
interface UserGrowthPoint { date: string; count: number; new: number }
interface RevenuePoint { date: string; amount: number; count: number }
interface TicketStats { total: number; open: number; answered: number; closed: number }

// ─── Custom Chart Tooltip ────────────────────────────────────────────
function ChartTooltip({ active, payload, label, formatter }: any) {
  if (!active || !payload?.length) return null
  return (
    <div className="bg-[rgba(13,13,26,0.9)] backdrop-blur-xl border border-[rgba(139,92,246,0.15)] rounded-xl px-4 py-3 shadow-[0_8px_32px_rgba(0,0,0,0.4)]">
      <p className="text-[#9898b8] text-xs mb-2">{label}</p>
      {payload.map((entry: any, i: number) => (
        <p key={i} className="text-sm font-semibold flex items-center gap-2" style={{ color: entry.color }}>
          <span className="w-2 h-2 rounded-full" style={{ background: entry.color }} />
          {entry.name}: {formatter ? formatter(entry.value) : entry.value}
        </p>
      ))}
    </div>
  )
}

// ─── Stat Card ───────────────────────────────────────────────────────
function StatCard({ label, value, icon, gradient, trend, delay = 0 }: {
  label: string; value: string | number; icon: string; gradient: string; trend?: { up: boolean; text: string }; delay?: number
}) {
  return (
    <div
      className="stat-card group cursor-default"
      style={{
        animation: `fadeInUp 0.5s ease-out ${delay}s both`,
      }}
    >
      <div className="glass-card p-5 h-full relative overflow-hidden">
        {/* Gradient accent line */}
        <div className="absolute top-0 left-4 right-4 h-[2px] rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-500"
          style={{ background: gradient }}
        />
        <div className="flex items-start justify-between relative z-10">
          <div className="flex-1 min-w-0">
            <p className="text-[#6868a0] text-[11px] font-semibold uppercase tracking-[0.08em] mb-1">{label}</p>
            <p className="text-[26px] font-bold text-white tracking-tight group-hover:bg-gradient-to-r group-hover:from-white group-hover:to-[#c0c0f0] group-hover:bg-clip-text group-hover:text-transparent transition-all duration-300">
              {value}
            </p>
            {trend && (
              <p className={`text-xs mt-1.5 flex items-center gap-1.5 ${trend.up ? 'text-emerald-400' : 'text-red-400'}`}>
                <span className={`inline-block w-0 h-0 border-x-[4px] border-x-transparent ${trend.up ? 'border-b-[6px] border-b-emerald-400' : 'border-t-[6px] border-t-red-400'}`} />
                <span>{trend.text}</span>
              </p>
            )}
          </div>
          <div className="stat-icon shrink-0" style={{ background: gradient }}>
            <div className="absolute inset-0 rounded-[inherit]" style={{ background: 'rgba(255,255,255,0.1)' }} />
            <div className="absolute inset-0 rounded-[inherit]" style={{ background: gradient, opacity: 0.3, filter: 'blur(4px)' }} />
            <span className="relative z-10">{icon}</span>
          </div>
        </div>
      </div>
    </div>
  )
}

// ─── Mini Metric ─────────────────────────────────────────────────────
function MiniMetric({ label, value, color }: { label: string; value: string | number; color: string }) {
  return (
    <div className="bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.03)] rounded-xl p-3 text-center hover:bg-[rgba(139,92,246,0.03)] hover:border-[rgba(139,92,246,0.08)] transition-all duration-300">
      <p className="text-xl font-bold" style={{ color }}>{value}</p>
      <p className="text-[11px] text-[#6868a0] font-medium mt-0.5">{label}</p>
    </div>
  )
}

// ─── Quick Action ────────────────────────────────────────────────────
function QuickAction({ label, desc, href, gradient }: { label: string; desc: string; href: string; gradient: string }) {
  return (
    <a
      href={href}
      className="group relative p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.03)] hover:border-[rgba(139,92,246,0.12)] transition-all duration-300 overflow-hidden"
    >
      <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 rounded-[inherit]"
        style={{ background: `linear-gradient(135deg, ${colorFromGradient(gradient)}10, transparent)` }}
      />
      <p className="text-sm font-semibold text-white relative z-10 group-hover:text-purple-300 transition-all duration-300">
        {label}
      </p>
      <p className="text-[11px] text-[#6868a0] mt-1 relative z-10">{desc}</p>
    </a>
  )
}

function colorFromGradient(g: string): string {
  if (g.includes('purple')) return '#8b5cf6'
  if (g.includes('cyan')) return '#06b6d4'
  if (g.includes('emerald')) return '#10b981'
  if (g.includes('pink')) return '#ec4899'
  if (g.includes('amber')) return '#f59e0b'
  return '#8b5cf6'
}

// ─── Chart Colors ────────────────────────────────────────────────────
const COLORS = {
  purple: '#8b5cf6', purpleLight: 'rgba(139,92,246,0.2)',
  cyan: '#06b6d4', cyanLight: 'rgba(6,182,212,0.2)',
  emerald: '#10b981', emeraldLight: 'rgba(16,185,129,0.2)',
  pink: '#ec4899',
  amber: '#f59e0b',
  grid: 'rgba(255,255,255,0.03)',
  text: '#6868a0',
}

const PIE_COLORS = ['#8b5cf6', '#10b981', '#f59e0b', '#ef4444', '#06b6d4']

// ─── Main Dashboard ──────────────────────────────────────────────────
export function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [trafficData, setTrafficData] = useState<TrafficPoint[]>([])
  const [growthData, setGrowthData] = useState<UserGrowthPoint[]>([])
  const [revenueData, setRevenueData] = useState<RevenuePoint[]>([])
  const [ticketStats, setTicketStats] = useState<TicketStats | null>(null)
  const [wsConnected, setWsConnected] = useState(false)
  const [loading, setLoading] = useState(true)
  const [trafficUp, setTrafficUp] = useState(0)
  const [trafficDown, setTrafficDown] = useState(0)
  const [onlineCount, setOnlineCount] = useState(0)
  const [notifications, setNotifications] = useState<{ text: string; time: string }[]>([])
  const [trafficDays, setTrafficDays] = useState(7)
  const [currentTime, setCurrentTime] = useState(new Date())
  const wsRef = useRef<WebSocket | null>(null)

  // ── Clock ──────────────────────────────────────────────────────────
  useEffect(() => {
    const i = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(i)
  }, [])

  // ── Fetch All Analytics ────────────────────────────────────────────
  const fetchAll = useCallback(async () => {
    try {
      const [s, t, g, r, tk] = await Promise.all([
        apiClient.get<DashboardStats>('/api/v1/analytics/stats'),
        apiClient.get<{ traffic: TrafficPoint[] }>(`/api/v1/analytics/traffic?days=${trafficDays}`),
        apiClient.get<{ growth: UserGrowthPoint[] }>('/api/v1/analytics/user-growth?days=30'),
        apiClient.get<{ revenue: RevenuePoint[] }>('/api/v1/analytics/revenue?days=30'),
        apiClient.get<TicketStats>('/api/v1/tickets/stats'),
      ])
      setStats(s.data)
      setTrafficData(t.data.traffic || [])
      setGrowthData(g.data.growth || [])
      setRevenueData(r.data.revenue || [])
      setTicketStats(tk.data)
    } catch (err) {
      console.error('fetch analytics error:', err)
    } finally {
      setLoading(false)
    }
  }, [trafficDays])

  useEffect(() => {
    fetchAll()
    const i = setInterval(fetchAll, 30000)
    return () => clearInterval(i)
  }, [fetchAll])

  // ── WebSocket ───────────────────────────────────────────────────────
  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`
    let isCancelled = false
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null

    const connect = () => {
      if (isCancelled) return
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws
      ws.onopen = () => {
        if (isCancelled) { ws.close(); return }
        setWsConnected(true)
        setNotifications(prev => [...prev.slice(-6), { text: '🟢 Real-time connected', time: 'just now' }])
      }
      ws.onmessage = (event) => {
        if (isCancelled) return
        try {
          const msg = JSON.parse(event.data)
          if (msg.type === 'traffic') {
            setTrafficUp(prev => prev + (msg.payload.up || 0))
            setTrafficDown(prev => prev + (msg.payload.down || 0))
            if (msg.payload.online) setOnlineCount(msg.payload.online)
          }
          if (msg.type === 'notification') {
            setNotifications(prev => [...prev.slice(-6), { text: `🔔 ${msg.payload.message || msg.payload.title}`, time: 'now' }])
          }
        } catch { /* ignore parse errors */ }
      }
      ws.onclose = () => {
        setWsConnected(false)
        if (!isCancelled) reconnectTimer = setTimeout(connect, 3000)
      }
      ws.onerror = () => ws.close()
    }
    connect()
    return () => {
      isCancelled = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      wsRef.current?.close()
    }
  }, [])

  // ── Derived Data ───────────────────────────────────────────────────
  const ticketPieData = ticketStats ? [
    { name: 'Open', value: ticketStats.open },
    { name: 'Answered', value: ticketStats.answered },
    { name: 'Closed', value: ticketStats.closed },
  ].filter(d => d.value > 0) : []

  const formatTime = (date: Date) =>
    date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })

  const formatUptime = (seconds: number) => {
    const d = Math.floor(seconds / 86400)
    const h = Math.floor((seconds % 86400) / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    return d > 0 ? `${d}d ${h}h ${m}m` : `${h}h ${m}m`
  }

  // ── Loading ────────────────────────────────────────────────────────
  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-[60vh] gap-4">
        <div className="loading-spinner loading-spinner-lg" />
        <p className="text-[#6868a0] text-sm font-medium">Loading dashboard...</p>
      </div>
    )
  }

  // ── Render ─────────────────────────────────────────────────────────
  return (
    <div className="space-y-6">
      {/* ═══ HEADER ═══════════════════════════════════════════════ */}
      <div
        className="glass-panel p-5"
        style={{ animation: 'fadeInUp 0.4s ease-out both' }}
      >
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-xl font-bold text-white">Dashboard</h1>
                <span className="px-2 py-0.5 rounded-full bg-[rgba(139,92,246,0.1)] border border-[rgba(139,92,246,0.15)] text-[11px] font-semibold text-purple-300">
                  v0.0.1
                </span>
              </div>
              <p className="text-[#9898b8] text-sm mt-0.5">System overview & real-time analytics</p>
            </div>
          </div>

          <div className="flex items-center gap-3 flex-wrap">
            {/* Time */}
            <div className="text-[11px] text-[#6868a0] font-mono bg-[rgba(255,255,255,0.02)] px-3 py-1.5 rounded-lg border border-[rgba(255,255,255,0.03)]">
              {formatTime(currentTime)}
            </div>

            {/* Days selector */}
            <select
              value={trafficDays}
              onChange={e => setTrafficDays(Number(e.target.value))}
              className="text-[11px] bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] rounded-lg px-3 py-1.5 text-[#c0c0d8] focus:border-purple-500/30 focus:outline-none transition-colors cursor-pointer"
            >
              <option value={7}>7 days</option>
              <option value={14}>14 days</option>
              <option value={30}>30 days</option>
              <option value={90}>90 days</option>
            </select>

            {/* Connection status */}
            <div className={`flex items-center gap-1.5 text-[11px] font-medium px-2.5 py-1.5 rounded-lg border transition-colors ${
              wsConnected
                ? 'bg-emerald-500/8 border-emerald-500/15 text-emerald-400'
                : 'bg-red-500/8 border-red-500/15 text-red-400'
            }`}>
              <span className={`w-1.5 h-1.5 rounded-full ${wsConnected ? 'bg-emerald-400 shadow-[0_0_6px_rgba(16,185,129,0.5)] animate-pulse' : 'bg-red-400'}`} />
              {wsConnected ? 'LIVE' : 'OFFLINE'}
            </div>
          </div>
        </div>
      </div>

      {/* ═══ STATS GRID ════════════════════════════════════════════ */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          label="Active Users"
          value={stats?.active_users || 0}
          icon="👥"
          gradient="linear-gradient(135deg, #8b5cf6, #a78bfa)"
          trend={{ up: (stats?.users_today || 0) > 0, text: `${stats?.users_today || 0} new today` }}
          delay={0.05}
        />
        <StatCard
          label="Total Inbounds"
          value={stats?.total_inbounds || 0}
          icon="📡"
          gradient="linear-gradient(135deg, #06b6d4, #22d3ee)"
          delay={0.1}
        />
        <StatCard
          label="Online Nodes"
          value={`${stats?.online_nodes || 0}/${stats?.total_nodes || 0}`}
          icon="🌐"
          gradient="linear-gradient(135deg, #10b981, #34d399)"
          delay={0.15}
        />
        <StatCard
          label="Cores Status"
          value="Active"
          icon="⚡"
          gradient="linear-gradient(135deg, #f59e0b, #fbbf24)"
          trend={{ up: true, text: 'System healthy' }}
          delay={0.2}
        />
        <StatCard
          label="Traffic Upload"
          value={formatBytes(trafficUp || stats?.traffic_up || 0)}
          icon="📤"
          gradient="linear-gradient(135deg, #8b5cf6, #ec4899)"
          delay={0.25}
        />
        <StatCard
          label="Traffic Download"
          value={formatBytes(trafficDown || stats?.traffic_down || 0)}
          icon="📥"
          gradient="linear-gradient(135deg, #06b6d4, #10b981)"
          delay={0.3}
        />
        <StatCard
          label="Total Revenue"
          value={stats ? `$${(stats.revenue_total / 100).toFixed(2)}` : '$0'}
          icon="💰"
          gradient="linear-gradient(135deg, #10b981, #6ee7b7)"
          trend={{ up: (stats?.revenue_today || 0) > 0, text: `$${((stats?.revenue_today || 0) / 100).toFixed(2)} today` }}
          delay={0.35}
        />
        <StatCard
          label="Open Tickets"
          value={stats?.open_tickets || 0}
          icon="🎫"
          gradient="linear-gradient(135deg, #ec4899, #f472b6)"
          delay={0.4}
        />
      </div>

      {/* ═══ CHART ROW 1: Traffic + Revenue ════════════════════════ */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* ── Traffic Chart ─────────────────────────────────────── */}
        <div
          className="glass-card p-5"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.2s both' }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-[rgba(139,92,246,0.1)] border border-[rgba(139,92,246,0.1)] flex items-center justify-center">
                <svg className="w-4 h-4 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h2 className="text-[15px] font-bold text-white">Traffic History</h2>
            </div>
            <div className="flex items-center gap-3 text-[11px]">
              <span className="flex items-center gap-1.5 text-purple-400"><span className="w-2 h-2 rounded-full bg-purple-500" /> Up</span>
              <span className="flex items-center gap-1.5 text-cyan-400"><span className="w-2 h-2 rounded-full bg-cyan-500" /> Down</span>
            </div>
          </div>
          <div className="h-[280px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trafficData} margin={{ top: 5, right: 5, bottom: 5, left: 0 }}>
                <defs>
                  <linearGradient id="cUp" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.25}/>
                    <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="cDown" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#06b6d4" stopOpacity={0.25}/>
                    <stop offset="95%" stopColor="#06b6d4" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke={COLORS.grid} />
                <XAxis dataKey="date" tick={{ fill: COLORS.text, fontSize: 10 }} axisLine={false} tickLine={false} />
                <YAxis tickFormatter={(v: number) => formatBytes(v)} tick={{ fill: COLORS.text, fontSize: 10 }} axisLine={false} tickLine={false} width={60} />
                <Tooltip content={<ChartTooltip formatter={formatBytes} />} />
                <Area type="monotone" dataKey="up" stroke={COLORS.purple} fill="url(#cUp)" name="Upload" strokeWidth={2} dot={false} />
                <Area type="monotone" dataKey="down" stroke={COLORS.cyan} fill="url(#cDown)" name="Download" strokeWidth={2} dot={false} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* ── Revenue Chart ─────────────────────────────────────── */}
        <div
          className="glass-card p-5"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.25s both' }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-[rgba(16,185,129,0.1)] border border-[rgba(16,185,129,0.1)] flex items-center justify-center">
                <svg className="w-4 h-4 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h2 className="text-[15px] font-bold text-white">Revenue (30 days)</h2>
            </div>
          </div>
          <div className="h-[280px]">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={revenueData} margin={{ top: 5, right: 5, bottom: 5, left: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={COLORS.grid} />
                <XAxis dataKey="date" tick={{ fill: COLORS.text, fontSize: 10 }} axisLine={false} tickLine={false} />
                <YAxis tickFormatter={(v: number) => `$${(v / 100).toFixed(0)}`} tick={{ fill: COLORS.text, fontSize: 10 }} axisLine={false} tickLine={false} width={50} />
                <Tooltip content={<ChartTooltip formatter={(v: number) => `$${(v / 100).toFixed(2)}`} />} />
                <Bar dataKey="amount" fill="url(#revGrad)" radius={[6, 6, 0, 0]} name="Revenue">
                  {revenueData.map((_, i) => (
                    <Cell key={i} fill={i % 2 === 0 ? '#10b981' : '#34d399'} />
                  ))}
                </Bar>
                <defs>
                  <linearGradient id="revGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#10b981" />
                    <stop offset="100%" stopColor="#34d399" />
                  </linearGradient>
                </defs>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* ═══ CHART ROW 2: Growth + Tickets ════════════════════════ */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* ── User Growth ───────────────────────────────────────── */}
        <div
          className="lg:col-span-2 glass-card p-5"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.3s both' }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-[rgba(6,182,212,0.1)] border border-[rgba(6,182,212,0.1)] flex items-center justify-center">
                <svg className="w-4 h-4 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                </svg>
              </div>
              <h2 className="text-[15px] font-bold text-white">User Growth (30 days)</h2>
            </div>
            <div className="flex items-center gap-3 text-[11px]">
              <span className="flex items-center gap-1.5 text-purple-400"><span className="w-2 h-2 rounded-full bg-purple-500" /> Total</span>
              <span className="flex items-center gap-1.5 text-emerald-400"><span className="w-2 h-2 rounded-full bg-emerald-500" /> New</span>
            </div>
          </div>
          <div className="h-[260px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={growthData} margin={{ top: 5, right: 5, bottom: 5, left: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={COLORS.grid} />
                <XAxis dataKey="date" tick={{ fill: COLORS.text, fontSize: 10 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: COLORS.text, fontSize: 10 }} axisLine={false} tickLine={false} width={40} />
                <Tooltip content={<ChartTooltip />} />
                <Line type="monotone" dataKey="count" stroke={COLORS.purple} strokeWidth={2} dot={false} name="Total Users" />
                <Line type="monotone" dataKey="new" stroke={COLORS.emerald} strokeWidth={2} dot={false} name="New Users" />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* ── Tickets ───────────────────────────────────────────── */}
        <div
          className="glass-card p-5 flex flex-col"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.35s both' }}
        >
          <div className="flex items-center gap-3 mb-4">
            <div className="w-8 h-8 rounded-lg bg-[rgba(236,72,153,0.1)] border border-[rgba(236,72,153,0.1)] flex items-center justify-center">
              <svg className="w-4 h-4 text-pink-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
              </svg>
            </div>
            <h2 className="text-[15px] font-bold text-white">Tickets</h2>
          </div>

          <div className="flex-1 flex items-center justify-center">
            {ticketPieData.length > 0 ? (
              <ResponsiveContainer width="100%" height={200}>
                <PieChart>
                  <Pie data={ticketPieData} cx="50%" cy="50%" innerRadius={50} outerRadius={80} paddingAngle={4} dataKey="value">
                    {ticketPieData.map((_, i) => <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />)}
                  </Pie>
                  <Tooltip content={<ChartTooltip />} />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex flex-col items-center gap-2 text-[#6868a0]">
                <svg className="w-10 h-10 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
                </svg>
                <span className="text-sm">No tickets</span>
              </div>
            )}
          </div>

          <div className="grid grid-cols-3 gap-2 mt-2">
            <MiniMetric label="Open" value={ticketStats?.open || 0} color="#f59e0b" />
            <MiniMetric label="Answered" value={ticketStats?.answered || 0} color="#8b5cf6" />
            <MiniMetric label="Closed" value={ticketStats?.closed || 0} color="#10b981" />
          </div>
        </div>
      </div>

      {/* ═══ BOTTOM ROW: Live Feed + Quick Actions + System ═══════ */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* ── Live Feed ─────────────────────────────────────────── */}
        <div
          className="glass-card p-5"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.4s both' }}
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-[rgba(139,92,246,0.1)] border border-[rgba(139,92,246,0.1)] flex items-center justify-center">
                <span className="text-sm">🔔</span>
              </div>
              <h2 className="text-[15px] font-bold text-white">Live Feed</h2>
            </div>
            <span className={`w-1.5 h-1.5 rounded-full ${wsConnected ? 'bg-emerald-400 shadow-[0_0_6px_rgba(16,185,129,0.5)]' : 'bg-[#6868a0]'}`} />
          </div>
          <div className="space-y-2 max-h-[200px] overflow-y-auto custom-scrollbar pr-1">
            {notifications.length === 0 ? (
              <div className="flex flex-col items-center gap-2 py-8 text-[#6868a0]">
                <svg className="w-8 h-8 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <span className="text-sm">No events yet</span>
              </div>
            ) : (
              notifications.map((n, i) => (
                <div key={i} className="flex items-start gap-2.5 text-sm text-[#c0c0d8] py-1.5 border-b border-[rgba(255,255,255,0.03)] last:border-0 group">
                  <span className="w-1.5 h-1.5 rounded-full bg-purple-500/50 mt-2 shrink-0 group-hover:bg-purple-400 transition-colors" />
                  <div className="flex-1 min-w-0">
                    <p className="truncate">{n.text}</p>
                    <p className="text-[10px] text-[#585878] mt-0.5">{n.time}</p>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* ── Quick Actions ─────────────────────────────────────── */}
        <div
          className="glass-card p-5"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.45s both' }}
        >
          <div className="flex items-center gap-3 mb-4">
            <div className="w-8 h-8 rounded-lg bg-[rgba(245,158,11,0.1)] border border-[rgba(245,158,11,0.1)] flex items-center justify-center">
              <svg className="w-4 h-4 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <h2 className="text-[15px] font-bold text-white">Quick Actions</h2>
          </div>
          <div className="grid grid-cols-2 gap-2.5">
            <QuickAction
              label="📡 Inbounds"
              desc={`${stats?.total_inbounds || 0} configured`}
              href="/inbounds"
              gradient="purple-cyan"
            />
            <QuickAction
              label="👥 Users"
              desc={`${stats?.total_users || 0} registered`}
              href="/users"
              gradient="cyan-emerald"
            />
            <QuickAction
              label="🌐 Nodes"
              desc={`${stats?.online_nodes || 0} online`}
              href="/nodes"
              gradient="emerald"
            />
            <QuickAction
              label="⚙️ Settings"
              desc="Configure panel"
              href="/settings"
              gradient="amber"
            />
          </div>
        </div>

        {/* ── System Info ───────────────────────────────────────── */}
        <div
          className="glass-card p-5"
          style={{ animation: 'fadeInUp 0.5s ease-out 0.5s both' }}
        >
          <div className="flex items-center gap-3 mb-4">
            <div className="w-8 h-8 rounded-lg bg-[rgba(6,182,212,0.1)] border border-[rgba(6,182,212,0.1)] flex items-center justify-center">
              <svg className="w-4 h-4 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
            </div>
            <h2 className="text-[15px] font-bold text-white">System</h2>
          </div>

          {stats && (
            <div className="space-y-3">
              {/* Uptime */}
              <div className="flex items-center justify-between py-2 px-3 rounded-lg bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.03)]">
                <span className="text-xs text-[#6868a0] font-medium">Uptime</span>
                <span className="text-sm font-semibold text-white">{formatUptime(stats.uptime_seconds)}</span>
              </div>
              {/* Memory */}
              <div className="flex items-center justify-between py-2 px-3 rounded-lg bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.03)]">
                <span className="text-xs text-[#6868a0] font-medium">Memory</span>
                <span className="text-sm font-semibold text-cyan-400">{stats.memory_used_mb.toFixed(0)} MB</span>
              </div>
              {/* Transactions */}
              <div className="flex items-center justify-between py-2 px-3 rounded-lg bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.03)]">
                <span className="text-xs text-[#6868a0] font-medium">Transactions</span>
                <span className="text-sm font-semibold text-purple-400">{stats.transactions}</span>
              </div>
              {/* Total Users */}
              <div className="flex items-center justify-between py-2 px-3 rounded-lg bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.03)]">
                <span className="text-xs text-[#6868a0] font-medium">Total Users</span>
                <span className="text-sm font-semibold text-emerald-400">{stats.total_users}</span>
              </div>
              {/* Active Users bar */}
              <div className="pt-2">
                <div className="flex items-center justify-between text-xs mb-2">
                  <span className="text-[#6868a0]">Active Rate</span>
                  <span className="text-white font-semibold">
                    {stats.total_users > 0 ? ((stats.active_users / stats.total_users) * 100).toFixed(0) : 0}%
                  </span>
                </div>
                <div className="progress-bar">
                  <div className="progress-bar-fill" style={{ width: `${stats.total_users > 0 ? (stats.active_users / stats.total_users) * 100 : 0}%` }} />
                </div>
              </div>
            </div>
          )}

          {!stats && (
            <div className="flex items-center justify-center h-32 text-[#6868a0] text-sm">
              No system data available
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
