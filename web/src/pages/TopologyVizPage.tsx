import React, { useState, useCallback, useEffect, useRef } from 'react'
import { CanvasTopology, useTopologyData } from '../components/topology/CanvasTopology'

// ─── Types ───────────────────────────────────────────────────────────

interface TopoNode {
  id: number
  name: string
  address: string
  role: string
  status: string
  cpu_load: number
  memory_used: number
  region: string
  priority: number
  user_count: number
  latency_ms: number
  last_heartbeat: number
  online: boolean
  is_leader: boolean
}

// ─── Icons ───────────────────────────────────────────────────────────

const Icons = {
  ZoomIn: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
    </svg>
  ),
  ZoomOut: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
    </svg>
  ),
  Refresh: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
    </svg>
  ),
  Fullscreen: () => (
    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5v-4m0 4h-4m4 0l-5-5" />
    </svg>
  ),
}

// ─── Main Component ──────────────────────────────────────────────────

export default function TopologyVizPage() {
  const { nodes, loading, error, refresh } = useTopologyData()
  const [selectedNode, setSelectedNode] = useState<TopoNode | null>(null)
  const [dimensions, setDimensions] = useState({ width: 1200, height: 700 })
  const containerRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)

  // ─── Resize ─────────────────────────────────────────────────────
  useEffect(() => {
    function handleResize() {
      if (containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect()
        setDimensions({ width: rect.width, height: Math.max(500, window.innerHeight - 280) })
      }
    }
    handleResize()
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // ─── Format helpers ──────────────────────────────────────────────
  const formatTime = (ts: number) => ts ? new Date(ts).toLocaleString() : '-'
  const formatLoad = (val: number) => val ? (val * 100).toFixed(1) + '%' : '-'

  // ─── Node type badge color ──────────────────────────────────────
  const getRoleColor = (role: string) => {
    switch (role) {
      case 'leader': return 'bg-purple-500/20 text-purple-400 border-purple-500/30'
      case 'follower': return 'bg-blue-500/20 text-blue-400 border-blue-500/30'
      case 'candidate': return 'bg-amber-500/20 text-amber-400 border-amber-500/30'
      default: return 'bg-gray-500/20 text-gray-400 border-gray-500/30'
    }
  }

  // ─── Selected Node Detail Panel ─────────────────────────────────
  const NodeDetailPanel = ({ node }: { node: TopoNode }) => (
    <div className="w-80 bg-[#1a1a2e]/95 backdrop-blur-xl border border-[#2a2a4e] rounded-2xl shadow-2xl overflow-hidden">
      {/* Header */}
      <div className="p-5 border-b border-[#2a2a4e]">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-xl flex items-center justify-center text-lg font-bold ${
              !node.online ? 'bg-gray-500/20 text-gray-400' :
              node.is_leader ? 'bg-purple-500/20 text-purple-400' :
              'bg-blue-500/20 text-blue-400'
            }`}>
              {node.name[0].toUpperCase()}
            </div>
            <div>
              <h3 className="text-white font-bold">{node.name}</h3>
              <span className={`text-xs px-2 py-0.5 rounded-full border ${getRoleColor(node.role)}`}>
                {node.role}
              </span>
            </div>
          </div>
          <button onClick={() => setSelectedNode(null)} className="text-gray-500 hover:text-white transition-colors">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      {/* Details */}
      <div className="p-5 space-y-3">
        {[
          { label: 'Address', value: node.address, mono: true },
          { label: 'Status', value: node.online ? 'Online' : 'Offline', color: node.online ? 'text-emerald-400' : 'text-red-400' },
          { label: 'Region', value: node.region || '—' },
          { label: 'Priority', value: String(node.priority) },
          { label: 'CPU Load', value: formatLoad(node.cpu_load) },
          { label: 'Memory', value: formatLoad(node.memory_used) },
          { label: 'Users', value: String(node.user_count || 0) },
          { label: 'Latency', value: node.latency_ms ? `${node.latency_ms}ms` : '—' },
          { label: 'Last Heartbeat', value: formatTime(node.last_heartbeat), small: true },
        ].map((row, i) => (
          <div key={i} className="flex items-center justify-between py-1.5 border-b border-[#2a2a4e]/50 last:border-0">
            <span className="text-xs text-gray-400">{row.label}</span>
            <span className={`text-sm font-medium ${row.color || 'text-gray-200'} ${row.mono ? 'font-mono text-xs' : ''} ${row.small ? 'text-xs' : ''}`}>
              {row.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  )

  // ─── Loading State ──────────────────────────────────────────────
  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[500px]">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-purple-500/30 border-t-purple-500 rounded-full animate-spin" />
          <p className="text-gray-400 text-sm">Loading topology...</p>
        </div>
      </div>
    )
  }

  // ─── Stats Summary ──────────────────────────────────────────────
  const onlineCount = nodes.filter(n => n.online).length
  const leaderNode = nodes.find(n => n.is_leader)

  return (
    <div className="space-y-4">
      {/* ─── Header ───────────────────────────────────────────── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-3">
            <svg className="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
            Network Topology
          </h1>
          <p className="text-sm text-gray-400 mt-1">Real-time cluster topology with live traffic visualization</p>
        </div>
        <button onClick={refresh} className="flex items-center gap-2 px-4 py-2 rounded-xl bg-purple-500/10 hover:bg-purple-500/20 text-purple-400 border border-purple-500/20 transition-all text-sm">
          <Icons.Refresh /> Refresh
        </button>
      </div>

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* ─── Stats Bar ────────────────────────────────────────── */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
        {[
          { label: 'Total Nodes', value: nodes.length, color: 'text-blue-400' },
          { label: 'Online', value: onlineCount, color: 'text-emerald-400' },
          { label: 'Leader', value: leaderNode?.name || '—', color: 'text-purple-400' },
          { label: 'Connections', value: nodes.filter(n => n.online).length > 1 ? 
            (nodes.filter(n => n.online).length * (nodes.filter(n => n.online).length - 1)) / 2 : 0, color: 'text-amber-400' },
          { label: 'Regions', value: [...new Set(nodes.map(n => n.region).filter(Boolean))].length, color: 'text-cyan-400' },
        ].map((stat, i) => (
          <div key={i} className="p-4 rounded-xl bg-[#1a1a2e] border border-[#2a2a4e] shadow-sm">
            <p className="text-xs text-gray-400 mb-1">{stat.label}</p>
            <p className={`text-lg font-bold ${stat.color}`}>
              {typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}
            </p>
          </div>
        ))}
      </div>

      {/* ─── Main Canvas ──────────────────────────────────────── */}
      <div className="relative" ref={containerRef}>
        {/* Canvas */}
        <div className="rounded-2xl overflow-hidden border border-[#2a2a4e] shadow-xl bg-[#0a0a1a]">
          {dimensions.width > 0 && (
            <CanvasTopology
              width={dimensions.width}
              height={dimensions.height}
              onNodeSelect={(node) => setSelectedNode(node as TopoNode | null)}
              wsData={nodes}
            />
          )}
        </div>

        {/* Legend Overlay */}
        <div className="absolute top-4 left-4 bg-[#1a1a2e]/90 backdrop-blur-md border border-[#2a2a4e] rounded-xl p-3 text-xs space-y-1.5">
          <p className="text-gray-300 font-medium mb-1.5">Legend</p>
          {[
            { color: 'bg-purple-500', label: 'Leader Node' },
            { color: 'bg-blue-500', label: 'Online Node' },
            { color: 'bg-amber-500', label: 'Candidate' },
            { color: 'bg-gray-500', label: 'Offline' },
          ].map((item, i) => (
            <div key={i} className="flex items-center gap-2">
              <div className={`w-2.5 h-2.5 rounded-full ${item.color}`} />
              <span className="text-gray-400">{item.label}</span>
            </div>
          ))}
          <div className="border-t border-[#2a2a4e] my-1.5 pt-1.5">
            <div className="flex items-center gap-2">
              <div className="w-4 h-0.5 bg-purple-500/50" />
              <span className="text-gray-400">Connection</span>
            </div>
            <div className="flex items-center gap-2 mt-1">
              <div className="w-2.5 h-2.5 rounded-full bg-purple-400 animate-pulse" />
              <span className="text-gray-400">Traffic Flow</span>
            </div>
          </div>
        </div>

        {/* Controls — Fullscreen Only */}
        <div className="absolute bottom-4 right-4">
          <button onClick={() => {
            if (!document.fullscreenElement) {
              document.querySelector('.rounded-2xl')?.requestFullscreen()
            } else {
              document.exitFullscreen()
            }
          }}
            className="w-9 h-9 rounded-xl bg-[#1a1a2e]/90 backdrop-blur-md border border-[#2a2a4e] flex items-center justify-center text-gray-400 hover:text-white hover:border-purple-500/30 transition-all"
            title="Fullscreen">
            <Icons.Fullscreen />
          </button>
        </div>

        {/* Network Info Overlay */}
        {nodes.filter(n => n.online).length > 0 && (
          <div className="absolute bottom-4 left-4 bg-[#1a1a2e]/80 backdrop-blur-md border border-[#2a2a4e] rounded-xl px-3 py-2 text-xs text-gray-400">
            <span className="text-emerald-400">●</span> Live · Auto-refreshing every 15s
          </div>
        )}

        {/* Selected Node Panel */}
        {selectedNode && (
          <div className="absolute top-4 right-4 z-10">
            <NodeDetailPanel node={selectedNode} />
          </div>
        )}

        {/* Empty State */}
        {nodes.length === 0 && !loading && (
          <div className="flex flex-col items-center justify-center py-24 text-gray-500">
            <svg className="w-16 h-16 mb-4 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
            </svg>
            <p className="text-lg font-medium mb-1">No cluster nodes found</p>
            <p className="text-sm">Enable cluster mode and connect peers to see the topology</p>
          </div>
        )}
      </div>
    </div>
  )
}
