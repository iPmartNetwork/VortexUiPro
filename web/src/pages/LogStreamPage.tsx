import { useState, useEffect, useRef } from 'react'

interface LogEntry {
  timestamp: number
  level: string
  source: string
  message: string
}

const LEVEL_COLORS: Record<string, string> = {
  error: 'text-red-400',
  warn: 'text-yellow-400',
  info: 'text-blue-300',
  debug: 'text-gray-400',
}

export function LogStreamPage() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [filter, setFilter] = useState('')
  const [levelFilter, setLevelFilter] = useState('info')
  const [sourceFilter, setSourceFilter] = useState('')
  const [connected, setConnected] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)
  const logRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    connectWebSocket()
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [levelFilter, sourceFilter, filter])

  const connectWebSocket = () => {
    if (wsRef.current) {
      wsRef.current.close()
    }

    const params = new URLSearchParams()
    params.set('level', levelFilter)
    if (sourceFilter) params.set('source', sourceFilter)
    if (filter) params.set('filter', filter)

    const ws = new WebSocket(
      `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/logs/ws?${params}`
    )

    ws.onopen = () => setConnected(true)
    ws.onclose = () => setConnected(false)
    ws.onmessage = (e) => {
      try {
        const entry = JSON.parse(e.data) as LogEntry
        setLogs(prev => {
          const next = [...prev, entry]
          return next.length > 1000 ? next.slice(-1000) : next
        })
      } catch { }
    }

    wsRef.current = ws
  }

  useEffect(() => {
    if (autoScroll && logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight
    }
  }, [logs, autoScroll])

  const filteredLogs = logs.filter(entry => {
    if (filter && !entry.message.toLowerCase().includes(filter.toLowerCase())) return false
    return true
  })

  const clearLogs = () => setLogs([])

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Live Logs</h1>
            <p className="text-sm text-[var(--text-secondary)]">Real-time log streaming</p>
          </div>
          <div className="ml-auto flex items-center gap-3">
            <div className={`w-2 h-2 rounded-full ${connected ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`} />
            <span className="text-xs text-[var(--text-muted)]">{connected ? 'Connected' : 'Disconnected'}</span>
            <span className="text-xs text-[var(--text-muted)]">{filteredLogs.length} entries</span>
            <button onClick={clearLogs} className="text-xs text-[var(--text-muted)] hover:text-[var(--text-primary)] transition">Clear</button>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="glass-card p-4">
        <div className="flex flex-wrap items-center gap-4">
          <div className="flex items-center gap-2">
            <label className="text-xs text-[var(--text-muted)]">Level:</label>
            <select value={levelFilter} onChange={e => setLevelFilter(e.target.value)} className="select-modern text-xs py-1">
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warning</option>
              <option value="error">Error</option>
            </select>
          </div>
          <div className="flex items-center gap-2">
            <label className="text-xs text-[var(--text-muted)]">Source:</label>
            <select value={sourceFilter} onChange={e => setSourceFilter(e.target.value)} className="select-modern text-xs py-1">
              <option value="">All</option>
              <option value="core">Core</option>
              <option value="panel">Panel</option>
              <option value="agent">Agent</option>
            </select>
          </div>
          <div className="flex items-center gap-2 flex-1 max-w-xs">
            <svg className="w-4 h-4 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="text"
              value={filter}
              onChange={e => setFilter(e.target.value)}
              placeholder="Filter logs..."
              className="flex-1 bg-transparent border-b border-[rgba(255,255,255,0.06)] text-sm text-[var(--text-primary)] placeholder-[#585878] outline-none focus:border-purple-500/30 transition"
            />
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" checked={autoScroll} onChange={e => setAutoScroll(e.target.checked)} className="rounded" />
            <span className="text-xs text-[var(--text-muted)]">Auto-scroll</span>
          </label>
        </div>
      </div>

      {/* Log Viewer */}
      <div ref={logRef} className="glass-card p-4 h-[600px] overflow-y-auto font-mono text-xs leading-relaxed">
        {filteredLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <p className="text-[#585878] mb-2">Waiting for logs...</p>
              <p className="text-[10px] text-[#3a3a5a]">Logs will appear in real-time</p>
            </div>
          </div>
        ) : (
          <div className="space-y-0.5">
            {filteredLogs.map((entry, i) => (
              <div key={i} className="flex gap-3 hover:bg-[rgba(255,255,255,0.02)] px-1 rounded">
                <span className="text-[#3a3a5a] shrink-0 w-16 text-right">{new Date(entry.timestamp).toLocaleTimeString()}</span>
                <span className={`uppercase text-[10px] font-bold w-12 shrink-0 ${LEVEL_COLORS[entry.level] || 'text-gray-400'}`}>{entry.level}</span>
                <span className="text-[#585878] shrink-0 w-12">[{entry.source}]</span>
                <span className="text-[var(--text-secondary)] break-all">{entry.message}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
