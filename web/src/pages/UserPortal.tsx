import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import { formatBytes, formatDate } from '../utils/format'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

// ─── Types ───────────────────────────────────────────────────────────

interface PortalClient {
  id: string
  email: string
  enable: boolean
  inbound_id: number
  total_gb: number
  expiry_time: number
  up_mbps: number
  down_mbps: number
  traffic_up: number
  traffic_down: number
  data_limit: number
  usage_percent: number
}

interface PortalTraffic {
  traffic_up: number
  traffic_down: number
  usage_percent: number
  data_limit: number
  expiry_time: number
  status: string
}

interface PortalTicket {
  id: number
  subject: string
  message: string
  status: string
  created_at: number
}

// ─── Component ───────────────────────────────────────────────────────

export function UserPortalPage() {
  const [clients, setClients] = useState<PortalClient[]>([])
  const [traffic, setTraffic] = useState<PortalTraffic | null>(null)
  const [tickets, setTickets] = useState<PortalTicket[]>([])
  const [activeTab, setActiveTab] = useState<'overview' | 'configs' | 'tickets'>('overview')
  const [ticketSubject, setTicketSubject] = useState('')
  const [ticketMessage, setTicketMessage] = useState('')
  const [sendingTicket, setSendingTicket] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  useEffect(() => {
    fetchPortalData()
  }, [])

  const fetchPortalData = async () => {
    try {
      const [clientsRes, trafficRes, ticketsRes] = await Promise.all([
        apiClient.get('/api/v1/portal/clients'),
        apiClient.get('/api/v1/portal/traffic'),
        apiClient.get('/api/v1/portal/tickets'),
      ])
      setClients(clientsRes.data.clients || [])
      setTraffic(trafficRes.data)
      setTickets(ticketsRes.data.tickets || [])
    } catch (err) {
      console.error('Failed to fetch portal data:', err)
    }
  }

  const createTicket = async () => {
    if (!ticketSubject || !ticketMessage) return
    setSendingTicket(true)
    try {
      await apiClient.post('/api/v1/portal/tickets', { subject: ticketSubject, message: ticketMessage })
      setTicketSubject('')
      setTicketMessage('')
      // Refresh tickets
      const res = await apiClient.get('/api/v1/portal/tickets')
      setTickets(res.data.tickets || [])
    } catch (err) {
      console.error('Failed to create ticket:', err)
    } finally {
      setSendingTicket(false)
    }
  }

  const copyConfig = async (clientId: string) => {
    try {
      const res = await apiClient.get(`/sub/${clientId}`, { params: { format: 'json' } })
      await navigator.clipboard.writeText(JSON.stringify(res.data, null, 2))
      setCopiedId(clientId)
      setTimeout(() => setCopiedId(null), 3000)
    } catch {
      // Try plain format
      try {
        const res = await apiClient.get(`/sub/${clientId}`)
        await navigator.clipboard.writeText(res.data)
        setCopiedId(clientId)
        setTimeout(() => setCopiedId(null), 3000)
      } catch {}
    }
  }

  // Generate mock traffic history data for chart
  const trafficHistory = clients.length > 0 ? Array.from({ length: 7 }, (_, i) => {
    const date = new Date()
    date.setDate(date.getDate() - (6 - i))
    const usage = clients.reduce((sum, c) => sum + (c.usage_percent || 0), 0)
    return {
      date: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      usage: Math.max(0, usage * (0.5 + Math.random() * 0.5)),
      upload: Math.floor(Math.random() * 500),
      download: Math.floor(Math.random() * 2000),
    }
  }) : []

  const tabs = [
    { id: 'overview' as const, label: 'Overview', icon: '📊' },
    { id: 'configs' as const, label: 'My Configs', icon: '🔗' },
    { id: 'tickets' as const, label: 'Support Tickets', icon: '🎫' },
  ]

  return (
    <div className="space-y-6 fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">👤 User Portal</h1>
          <p className="text-[var(--text-secondary)] mt-1">Manage your subscriptions and support tickets</p>
        </div>
        <div className="flex items-center gap-2">
          {traffic && (
            <>
              <div className={`px-3 py-1.5 rounded-lg text-xs font-medium ${
                traffic.status === 'active' ? 'bg-green-500/10 text-green-400 border border-green-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'
              }`}>
                {traffic.status === 'active' ? '🟢 Active' : '🔴 Expired'}
              </div>
              <button onClick={fetchPortalData} className="p-2 rounded-lg bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:text-white hover:bg-[var(--bg-surface)] transition border border-[var(--border-light)]">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </button>
            </>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-2">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              activeTab === tab.id
                ? 'bg-purple-600 text-white shadow-lg shadow-purple-600/20'
                : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] hover:text-white border border-[var(--border-light)]'
            }`}
          >
            <span>{tab.icon}</span>
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      {/* ─── Overview Tab ──────────────────────────────────────────── */}
      {activeTab === 'overview' && (
        <>
          {/* Stats Cards */}
          {traffic && (
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="glass-card p-4">
                <p className="text-[var(--text-muted)] text-xs uppercase tracking-wide mb-1">Upload</p>
                <p className="text-white text-xl font-bold">{formatBytes(traffic.traffic_up || 0)}</p>
              </div>
              <div className="glass-card p-4">
                <p className="text-[var(--text-muted)] text-xs uppercase tracking-wide mb-1">Download</p>
                <p className="text-white text-xl font-bold">{formatBytes(traffic.traffic_down || 0)}</p>
              </div>
              <div className="glass-card p-4">
                <p className="text-[var(--text-muted)] text-xs uppercase tracking-wide mb-1">Data Limit</p>
                <p className="text-white text-xl font-bold">{formatBytes(traffic.data_limit || 0)}</p>
              </div>
              <div className="glass-card p-4">
                <p className="text-[var(--text-muted)] text-xs uppercase tracking-wide mb-1">Usage</p>
                <p className="text-white text-xl font-bold">{traffic.usage_percent?.toFixed(1) || '0'}%</p>
                <div className="mt-2 w-full bg-[var(--bg-surface)] rounded-full h-1.5">
                  <div className={`h-1.5 rounded-full transition-all duration-500 ${
                    (traffic.usage_percent || 0) > 80 ? 'bg-red-500' : (traffic.usage_percent || 0) > 50 ? 'bg-yellow-500' : 'bg-green-500'
                  }`} style={{ width: `${Math.min(traffic.usage_percent || 0, 100)}%` }} />
                </div>
              </div>
            </div>
          )}

          {/* Traffic Chart */}
          <div className="glass-card p-6">
            <h3 className="text-white font-semibold mb-4">📈 7-Day Traffic Overview</h3>
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={trafficHistory}>
                  <defs>
                    <linearGradient id="colorUpload" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#7c3aed" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#7c3aed" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="colorDownload" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                  <XAxis dataKey="date" stroke="#64748b" tick={{ fontSize: 11 }} />
                  <YAxis stroke="#64748b" tick={{ fontSize: 11 }} tickFormatter={v => formatBytes(v)} />
                  <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#fff' }} />
                  <Area type="monotone" dataKey="upload" stroke="#7c3aed" strokeWidth={2} fill="url(#colorUpload)" name="Upload" />
                  <Area type="monotone" dataKey="download" stroke="#3b82f6" strokeWidth={2} fill="url(#colorDownload)" name="Download" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Expiry info */}
          {traffic && traffic.expiry_time > 0 && (
            <div className="glass-card p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">⏰</span>
                  <div>
                    <p className="text-white text-sm font-medium">Subscription Expiry</p>
                    <p className="text-[var(--text-secondary)] text-xs">{formatDate(traffic.expiry_time)}</p>
                  </div>
                </div>
                {traffic.expiry_time > 0 && (
                  <span className={`px-3 py-1 rounded-lg text-xs font-medium ${
                    traffic.expiry_time > Date.now() ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'
                  }`}>
                    {traffic.expiry_time > Date.now() ? '✅ Active' : '❌ Expired'}
                  </span>
                )}
              </div>
            </div>
          )}
        </>
      )}

      {/* ─── Configs Tab ─────────────────────────────────────────────── */}
      {activeTab === 'configs' && (
        <div className="space-y-4">
          {clients.length === 0 ? (
            <div className="glass-card p-8 text-center">
              <p className="text-[var(--text-muted)] text-sm">No active configurations. Contact your provider to get started.</p>
            </div>
          ) : (
            clients.map(client => (
              <div key={client.id} className="glass-card p-5 hover:border-purple-500/30 transition-all">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`w-2 h-2 rounded-full ${client.enable ? 'bg-green-500' : 'bg-red-500'}`} />
                      <span className="text-white font-medium text-sm">{client.email}</span>
                      <span className="text-[var(--text-muted)] text-xs font-mono">#{client.id.slice(0, 8)}</span>
                    </div>
                    <div className="flex flex-wrap gap-3 mt-2">
                      <span className="text-xs text-[var(--text-secondary)]">⬆️ {formatBytes(client.traffic_up || 0)}</span>
                      <span className="text-xs text-[var(--text-secondary)]">⬇️ {formatBytes(client.traffic_down || 0)}</span>
                      <span className="text-xs text-[var(--text-secondary)]">📦 {formatBytes(client.data_limit || client.total_gb)}</span>
                      {client.expiry_time > 0 && (
                        <span className="text-xs text-[var(--text-secondary)]">⏰ {formatDate(client.expiry_time)}</span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => copyConfig(client.id)}
                      className="px-3 py-1.5 rounded-lg bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:text-purple-400 hover:bg-[var(--bg-surface)] transition border border-[var(--border-light)] text-xs font-medium"
                    >
                      {copiedId === client.id ? '✅ Copied!' : '📋 Copy'}
                    </button>
                    <a
                      href={`/sub/${client.id}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="px-3 py-1.5 rounded-lg bg-purple-600/20 text-purple-300 hover:bg-purple-600/30 transition text-xs font-medium"
                    >
                      🔗 Link
                    </a>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* ─── Tickets Tab ─────────────────────────────────────────────── */}
      {activeTab === 'tickets' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Create Ticket */}
          <div className="glass-card p-6">
            <h3 className="text-white font-semibold mb-4">🎫 New Support Ticket</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1.5">Subject</label>
                <input
                  type="text"
                  value={ticketSubject}
                  onChange={e => setTicketSubject(e.target.value)}
                  placeholder="Brief description of your issue"
                  className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-white placeholder-dark-400 focus:border-purple-500 focus:outline-none text-sm"
                />
              </div>
              <div>
                <label className="block text-xs text-[var(--text-secondary)] mb-1.5">Message</label>
                <textarea
                  value={ticketMessage}
                  onChange={e => setTicketMessage(e.target.value)}
                  placeholder="Describe your issue in detail..."
                  rows={4}
                  className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-white placeholder-dark-400 focus:border-purple-500 focus:outline-none text-sm resize-none"
                />
              </div>
              <button
                onClick={createTicket}
                disabled={sendingTicket || !ticketSubject || !ticketMessage}
                className="w-full px-4 py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 transition text-sm font-medium"
              >
                {sendingTicket ? 'Sending...' : 'Submit Ticket'}
              </button>
            </div>
          </div>

          {/* Tickets List */}
          <div className="glass-card p-6">
            <h3 className="text-white font-semibold mb-4">📋 Your Tickets</h3>
            {tickets.length === 0 ? (
              <p className="text-[var(--text-muted)] text-sm text-center py-8">No tickets yet.</p>
            ) : (
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {tickets.map(ticket => (
                  <div key={ticket.id} className="p-3 rounded-lg bg-[var(--bg-elevated)]/50 border border-[var(--border-light)] hover:border-dark-600 transition">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-white text-sm font-medium">{ticket.subject}</span>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${
                        ticket.status === 'open' ? 'bg-green-500/10 text-green-400' :
                        ticket.status === 'answered' ? 'bg-yellow-500/10 text-yellow-400' :
                        'bg-[var(--bg-surface)] text-[var(--text-secondary)]'
                      }`}>
                        {ticket.status}
                      </span>
                    </div>
                    <p className="text-[var(--text-muted)] text-xs truncate">{ticket.message}</p>
                    <p className="text-[var(--text-muted)] text-[10px] mt-1">{formatDate(ticket.created_at)}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
