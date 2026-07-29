import { useState, useEffect, useRef } from 'react'
import { apiClient } from '../api/client'

interface TerminalSession {
  id: string
  node_id: number
  node_name: string
  created_at: string
}

export function TerminalPage() {
  const [sessions, setSessions] = useState<TerminalSession[]>([])
  const [activeSession, setActiveSession] = useState<string | null>(null)
  const [connected, setConnected] = useState(false)
  const [output, setOutput] = useState<string[]>([])
  const [input, setInput] = useState('')
  const [nodes, setNodes] = useState<{ id: number; name: string; address: string }[]>([])
  const [selectedNode, setSelectedNode] = useState<number | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const outputRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    apiClient.get('/api/v1/nodes').then(r => {
      setNodes(r.data?.nodes || [])
    }).catch(() => {})

    apiClient.get('/api/v1/terminal/sessions').then(r => {
      setSessions(r.data?.sessions || [])
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [output])

  const connect = () => {
    if (!selectedNode) return
    const ws = new WebSocket(`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/terminal/ws?node_id=${selectedNode}`)
    
    ws.onopen = () => {
      setConnected(true)
      setOutput(prev => [...prev, '\x1b[32m[Connected]\x1b[0m'])
    }

    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        if (msg.type === 'data') {
          setOutput(prev => [...prev, msg.data])
        } else if (msg.type === 'error') {
          setOutput(prev => [...prev, `\x1b[31m[Error] ${msg.data}\x1b[0m`])
        } else if (msg.type === 'connected') {
          setActiveSession(msg.data)
          setOutput(prev => [...prev, `\x1b[33m[Session: ${msg.data}]\x1b[0m`])
        } else if (msg.type === 'close') {
          setConnected(false)
          setOutput(prev => [...prev, `\x1b[31m[Disconnected: ${msg.data}]\x1b[0m`])
        }
      } catch { }
    }

    ws.onclose = () => {
      setConnected(false)
      setOutput(prev => [...prev, '\x1b[31m[Connection closed]\x1b[0m'])
    }

    ws.onerror = () => {
      setOutput(prev => [...prev, '\x1b[31m[WebSocket error]\x1b[0m'])
    }

    wsRef.current = ws
  }

  const disconnect = () => {
    if (wsRef.current) {
      wsRef.current.send(JSON.stringify({ type: 'close' }))
      wsRef.current.close()
      wsRef.current = null
    }
    setConnected(false)
    setActiveSession(null)
  }

  const sendInput = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && input && wsRef.current) {
      wsRef.current.send(JSON.stringify({ type: 'input', data: input + '\n' }))
      setInput('')
    }
  }

  const closeSession = async (id: string) => {
    try {
      await apiClient.delete(`/api/v1/terminal/sessions/${id}`)
      setSessions(prev => prev.filter(s => s.id !== id))
    } catch { }
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-cyan-500 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Web Terminal</h1>
            <p className="text-sm text-[var(--text-secondary)]">SSH console for nodes</p>
          </div>
          <div className="ml-auto flex items-center gap-3">
            <select
              value={selectedNode ?? ''}
              onChange={e => setSelectedNode(e.target.value ? Number(e.target.value) : null)}
              className="select-modern text-sm"
              disabled={connected}
            >
              <option value="">Select a node...</option>
              {nodes.map(n => (
                <option key={n.id} value={n.id}>{n.name} ({n.address})</option>
              ))}
            </select>
            {!connected ? (
              <button onClick={connect} disabled={!selectedNode} className="btn-primary text-sm">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                Connect
              </button>
            ) : (
              <button onClick={disconnect} className="btn-danger text-sm">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
                Disconnect
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Terminal Output */}
        <div className="lg:col-span-3">
          <div className="glass-card p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-[var(--text-primary)]">
                Terminal {connected && activeSession && <span className="text-emerald-400 text-xs ml-2">● Connected</span>}
              </h3>
              <div className="flex gap-2">
                <button onClick={() => setOutput([])} className="text-xs text-[var(--text-muted)] hover:text-[var(--text-primary)] transition">Clear</button>
              </div>
            </div>
            <div ref={outputRef} className="bg-black/80 rounded-lg p-4 h-[500px] overflow-y-auto font-mono text-sm leading-relaxed" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
              {output.length === 0 ? (
                <span className="text-[#585878]">Select a node and click Connect to start a terminal session...</span>
              ) : (
                output.map((line, i) => (
                  <div key={i} className="text-green-400/90" dangerouslySetInnerHTML={{
                    __html: line
                      .replace(/\x1b\[32m/g, '<span class="text-green-400">')
                      .replace(/\x1b\[31m/g, '<span class="text-red-400">')
                      .replace(/\x1b\[33m/g, '<span class="text-yellow-400">')
                      .replace(/\x1b\[0m/g, '</span>')
                  }} />
                ))
              )}
            </div>
            {connected && (
              <div className="flex items-center gap-2 mt-3">
                <span className="text-[var(--text-muted)] text-sm">$</span>
                <input
                  type="text"
                  value={input}
                  onChange={e => setInput(e.target.value)}
                  onKeyDown={sendInput}
                  className="flex-1 bg-transparent border-none outline-none text-sm text-green-400 font-mono placeholder-[#585878]"
                  placeholder="Type a command and press Enter..."
                  autoFocus
                />
              </div>
            )}
          </div>
        </div>

        {/* Active Sessions */}
        <div>
          <div className="glass-card p-4">
            <h3 className="text-sm font-semibold text-[var(--text-primary)] mb-3">Active Sessions</h3>
            {sessions.length === 0 ? (
              <p className="text-xs text-[var(--text-muted)]">No active sessions</p>
            ) : (
              <div className="space-y-2">
                {sessions.map(s => (
                  <div key={s.id} className="flex items-center justify-between p-2 rounded-lg bg-[rgba(255,255,255,0.02)]">
                    <div>
                      <p className="text-xs font-medium text-[var(--text-primary)]">{s.node_name}</p>
                      <p className="text-[10px] text-[var(--text-muted)]">{new Date(s.created_at).toLocaleTimeString()}</p>
                    </div>
                    <button onClick={() => closeSession(s.id)} className="text-red-400 hover:text-red-300 text-xs">Close</button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
