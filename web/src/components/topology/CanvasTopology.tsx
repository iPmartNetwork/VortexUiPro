import React, { useEffect, useRef, useState, useCallback } from 'react'

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

interface SimNode extends TopoNode {
  x: number; y: number
  vx: number; vy: number
  radius: number
  color: string
  glowIntensity: number
}

interface TopoEdge {
  source: number; target: number
  latency: number; active: boolean
}

interface Particle {
  x: number; y: number
  progress: number
  speed: number
  edge: TopoEdge
  sourceNode: SimNode
  targetNode: SimNode
}

interface Viewport {
  x: number; y: number
  scale: number
}

// ─── Colors ──────────────────────────────────────────────────────────

const COLORS = {
  leader: '#7C3AED',
  leaderGlow: 'rgba(124,58,237,0.4)',
  online: '#3B82F6',
  onlineGlow: 'rgba(59,130,246,0.3)',
  candidate: '#F59E0B',
  candidateGlow: 'rgba(245,158,11,0.3)',
  offline: '#6B7280',
  offlineGlow: 'rgba(107,114,128,0.2)',
  edge: 'rgba(139,92,246,0.15)',
  edgeActive: 'rgba(139,92,246,0.4)',
  edgeLatency: '#10B981',
  particle: '#A78BFA',
  bg: '#0a0a1a',
  grid: 'rgba(139,92,246,0.04)',
}

// ─── Canvas Topology ─────────────────────────────────────────────────

interface CanvasTopologyProps {
  width: number
  height: number
  onNodeSelect?: (node: TopoNode | null) => void
  wsData?: any
}

export function CanvasTopology({ width, height, onNodeSelect, wsData }: CanvasTopologyProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const animRef = useRef<number>(0)
  const mouseRef = useRef({ x: 0, y: 0, down: false, dragNode: null as SimNode | null, dragOffset: { x: 0, y: 0 } })
  const viewportRef = useRef<Viewport>({ x: 0, y: 0, scale: 1 })
  const nodesRef = useRef<SimNode[]>([])
  const edgesRef = useRef<TopoEdge[]>([])
  const particlesRef = useRef<Particle[]>([])
  const timeRef = useRef(0)
  const wsRef = useRef<WebSocket | null>(null)

  // ─── Initialize / Update from props ───────────────────────────────
  const initNodes = useCallback((peers: TopoNode[]) => {
    if (!peers || peers.length === 0) return
    const W = width; const H = height
    const simNodes: SimNode[] = peers.map((n, i) => ({
      ...n,
      x: W / 2 + (Math.random() - 0.5) * W * 0.6,
      y: H / 2 + (Math.random() - 0.5) * H * 0.6,
      vx: 0, vy: 0,
      radius: getNodeRadius(n),
      color: getNodeColor(n),
      glowIntensity: n.is_leader ? 1 : 0.5,
    }))
    nodesRef.current = simNodes

    // Create edges between online nodes
    const online = simNodes.filter(n => n.online)
    const newEdges: TopoEdge[] = []
    for (let i = 0; i < online.length; i++) {
      for (let j = i + 1; j < online.length; j++) {
        newEdges.push({
          source: online[i].id,
          target: online[j].id,
          latency: Math.random() * 30 + 1,
          active: true,
        })
      }
    }
    edgesRef.current = newEdges

    // Create particles
    createParticles(newEdges, simNodes)
  }, [width, height])

  // ─── Load initial data from wsData prop ────────────────────────────
  const initRef = useRef(0)
  useEffect(() => {
    if (wsData && Array.isArray(wsData) && wsData.length > 0 && initRef.current === 0) {
      initNodes(wsData)
      initRef.current = 1
    }
  }, [wsData])

  // ─── WebSocket for live updates ────────────────────────────────────
  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`)
    wsRef.current = ws

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'topology' || msg.type === 'cluster') {
          const topo = msg.payload || msg
          if (topo.peers && Array.isArray(topo.peers)) {
            const currentNodes = nodesRef.current
            const updated = currentNodes.map(n => {
              const update = topo.peers.find((p: any) => p.id === n.id)
              if (update) {
                return { ...n, ...update, x: n.x, y: n.y, vx: n.vx, vy: n.vy,
                  radius: getNodeRadius(update), color: getNodeColor(update) }
              }
              return n
            })
            // Add new nodes
            topo.peers.forEach((p: any) => {
              if (!updated.find(n => n.id === p.id)) {
                updated.push({
                  ...p,
                  x: Math.random() * width, y: Math.random() * height,
                  vx: 0, vy: 0,
                  radius: getNodeRadius(p), color: getNodeColor(p),
                  glowIntensity: p.is_leader ? 1 : 0.5,
                })
              }
            })
            nodesRef.current = updated

            // Rebuild edges
            const online = updated.filter(n => n.online)
            const newEdges: TopoEdge[] = []
            for (let i = 0; i < online.length; i++) {
              for (let j = i + 1; j < online.length; j++) {
                newEdges.push({
                  source: online[i].id, target: online[j].id,
                  latency: Math.random() * 30 + 1, active: true,
                })
              }
            }
            edgesRef.current = newEdges
            createParticles(newEdges, updated)
          }
        }
      } catch {}
    }
    return () => ws.close()
  }, [])

  // ─── Create Particles ─────────────────────────────────────────────
  function createParticles(edges: TopoEdge[], nodes: SimNode[]) {
    const particles: Particle[] = []
    edges.forEach(edge => {
      const src = nodes.find(n => n.id === edge.source)
      const tgt = nodes.find(n => n.id === edge.target)
      if (!src || !tgt) return
      for (let i = 0; i < 2; i++) {
        particles.push({
          x: src.x, y: src.y,
          progress: Math.random(),
          speed: 0.002 + Math.random() * 0.003,
          edge: { ...edge },
          sourceNode: src, targetNode: tgt,
        })
      }
    })
    particlesRef.current = particles
  }

  // ─── Physics Simulation ────────────────────────────────────────────
  function simulate(dt: number) {
    const nodes = nodesRef.current
    const edges = edgesRef.current
    const particles = particlesRef.current
    const W = width; const H = height
    const REPULSION = 50000; const ATTRACTION = 0.003
    const DAMPING = 0.92; const CENTER = 0.008
    const MAX_SPEED = 8

    // Repulsion
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const dx = nodes[j].x - nodes[i].x
        const dy = nodes[j].y - nodes[i].y
        const dist = Math.sqrt(dx * dx + dy * dy) || 1
        const force = REPULSION / (dist * dist)
        nodes[i].vx -= (dx / dist) * force * dt
        nodes[i].vy -= (dy / dist) * force * dt
        nodes[j].vx += (dx / dist) * force * dt
        nodes[j].vy += (dy / dist) * force * dt
      }
    }

    // Attraction along edges
    for (const edge of edges) {
      const src = nodes.find(n => n.id === edge.source)
      const tgt = nodes.find(n => n.id === edge.target)
      if (!src || !tgt) continue
      const dx = tgt.x - src.x
      const dy = tgt.y - src.y
      const dist = Math.sqrt(dx * dx + dy * dy) || 1
      const idealDist = 180
      const force = (dist - idealDist) * ATTRACTION
      src.vx += (dx / dist) * force * dt
      src.vy += (dy / dist) * force * dt
      tgt.vx -= (dx / dist) * force * dt
      tgt.vy -= (dy / dist) * force * dt
    }

    // Center gravity + damping + speed limit + bounds
    for (const n of nodes) {
      n.vx += (W / 2 - n.x) * CENTER * dt
      n.vy += (H / 2 - n.y) * CENTER * dt
      n.vx *= DAMPING
      n.vy *= DAMPING
      const speed = Math.sqrt(n.vx * n.vx + n.vy * n.vy)
      if (speed > MAX_SPEED) {
        n.vx = (n.vx / speed) * MAX_SPEED
        n.vy = (n.vy / speed) * MAX_SPEED
      }
      n.x += n.vx
      n.y += n.vy
      n.x = Math.max(40, Math.min(W - 40, n.x))
      n.y = Math.max(40, Math.min(H - 40, n.y))
    }

    // Update particles
    for (const p of particles) {
      p.progress += p.speed * dt * 60
      if (p.progress > 1) p.progress -= 1
      const src = nodes.find(n => n.id === p.edge.source) || p.sourceNode
      const tgt = nodes.find(n => n.id === p.edge.target) || p.targetNode
      p.x = src.x + (tgt.x - src.x) * p.progress
      p.y = src.y + (tgt.y - src.y) * p.progress
    }
  }

  // ─── Rendering ─────────────────────────────────────────────────────
  function render() {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const W = canvas.width
    const H = canvas.height
    const vp = viewportRef.current
    const nodes = nodesRef.current
    const edges = edgesRef.current
    const particles = particlesRef.current
    const mouse = mouseRef.current
    const time = timeRef.current

    // Clear
    ctx.fillStyle = COLORS.bg
    ctx.fillRect(0, 0, W, H)

    ctx.save()
    ctx.translate(vp.x, vp.y)
    ctx.scale(vp.scale, vp.scale)

    // Grid
    ctx.strokeStyle = COLORS.grid
    ctx.lineWidth = 1
    const gridSize = 60
    for (let x = -W; x < W * 2; x += gridSize) {
      ctx.beginPath(); ctx.moveTo(x, -H); ctx.lineTo(x, H * 2); ctx.stroke()
    }
    for (let y = -H; y < H * 2; y += gridSize) {
      ctx.beginPath(); ctx.moveTo(-W, y); ctx.lineTo(W * 2, y); ctx.stroke()
    }

    // Edges
    for (const edge of edges) {
      const src = nodes.find(n => n.id === edge.source)
      const tgt = nodes.find(n => n.id === edge.target)
      if (!src || !tgt) continue

      ctx.beginPath()
      ctx.moveTo(src.x, src.y)
      ctx.lineTo(tgt.x, tgt.y)
      ctx.strokeStyle = edge.active ? COLORS.edgeActive : COLORS.edge
      ctx.lineWidth = edge.active ? 1.5 : 0.5
      ctx.stroke()

      // Edge latency label
      if (edge.active) {
        const mx = (src.x + tgt.x) / 2
        const my = (src.y + tgt.y) / 2
        ctx.fillStyle = COLORS.edgeLatency
        ctx.font = '9px monospace'
        ctx.textAlign = 'center'
        ctx.fillText(`${edge.latency.toFixed(0)}ms`, mx, my - 8)
      }
    }

    // Particles (animated traffic)
    for (const p of particles) {
      ctx.beginPath()
      ctx.arc(p.x, p.y, 2.5, 0, Math.PI * 2)
      ctx.fillStyle = COLORS.particle
      ctx.shadowColor = COLORS.particle
      ctx.shadowBlur = 8
      ctx.fill()
      ctx.shadowBlur = 0
    }

    // Nodes
    for (const node of nodes) {
      const { x, y, radius, color, online, is_leader } = node

      // Glow
      const glowSize = radius + (is_leader ? 12 : 6) + Math.sin(time * 2) * 3
      const gradient = ctx.createRadialGradient(x, y, 0, x, y, glowSize)
      gradient.addColorStop(0, color + '40')
      gradient.addColorStop(1, 'transparent')
      ctx.fillStyle = gradient
      ctx.beginPath()
      ctx.arc(x, y, glowSize, 0, Math.PI * 2)
      ctx.fill()

      // Main circle
      ctx.beginPath()
      ctx.arc(x, y, radius, 0, Math.PI * 2)
      const nodeGrad = ctx.createRadialGradient(x - radius * 0.3, y - radius * 0.3, 0, x, y, radius)
      nodeGrad.addColorStop(0, '#fff')
      nodeGrad.addColorStop(0.3, color)
      nodeGrad.addColorStop(1, darkenColor(color, 0.4))
      ctx.fillStyle = nodeGrad
      ctx.fill()

      // Border
      ctx.strokeStyle = is_leader ? '#5B21B6' : 'rgba(255,255,255,0.1)'
      ctx.lineWidth = is_leader ? 3 : 1.5
      ctx.stroke()

      // Leader crown
      if (is_leader) {
        ctx.fillStyle = '#FBBF24'
        ctx.font = '14px sans-serif'
        ctx.textAlign = 'center'
        ctx.fillText('👑', x, y - radius - 14)
      }

      // Status dot
      const dotY = y - radius + 5
      ctx.beginPath()
      ctx.arc(x, dotY, 3.5, 0, Math.PI * 2)
      ctx.fillStyle = online ? '#10B981' : '#EF4444'
      ctx.fill()
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 1.5
      ctx.stroke()

      // CPU ring
      if (online) {
        ctx.beginPath()
        ctx.arc(x, y, radius + 4, -Math.PI / 2, -Math.PI / 2 + Math.PI * 2 * node.cpu_load)
        ctx.strokeStyle = node.cpu_load > 0.7 ? '#EF4444' : node.cpu_load > 0.4 ? '#F59E0B' : '#10B981'
        ctx.lineWidth = 2
        ctx.stroke()
      }

      // Name label
      ctx.fillStyle = 'rgba(255,255,255,0.85)'
      ctx.font = '11px system-ui, sans-serif'
      ctx.textAlign = 'center'
      ctx.fillText(node.name.length > 14 ? node.name.slice(0, 12) + '..' : node.name, x, y + radius + 16)

      // Role badge
      ctx.fillStyle = is_leader ? '#7C3AED' : node.role === 'candidate' ? '#F59E0B' : '#3B82F6'
      ctx.font = '8px system-ui, sans-serif'
      ctx.fillText(node.role.toUpperCase(), x, y + radius + 28)
    }

    // Drag indicator
    if (mouse.dragNode) {
      ctx.beginPath()
      ctx.arc(mouse.dragNode.x, mouse.dragNode.y, mouse.dragNode.radius + 10, 0, Math.PI * 2)
      ctx.strokeStyle = 'rgba(139,92,246,0.3)'
      ctx.lineWidth = 2
      ctx.setLineDash([4, 4])
      ctx.stroke()
      ctx.setLineDash([])
    }

    ctx.restore()

    // Stats overlay
    ctx.fillStyle = 'rgba(255,255,255,0.6)'
    ctx.font = '12px monospace'
    ctx.textAlign = 'left'
    const onlineCount = nodes.filter(n => n.online).length
    ctx.fillText(`${onlineCount}/${nodes.length} nodes · ${edges.length} connections · ${particles.length} flows`, 14, 22)
  }

  // ─── Animation Loop ────────────────────────────────────────────────
  function loop() {
    timeRef.current += 0.016
    simulate(0.016)
    render()
    animRef.current = requestAnimationFrame(loop)
  }

  useEffect(() => {
    animRef.current = requestAnimationFrame(loop)
    return () => cancelAnimationFrame(animRef.current)
  }, [])

  // ─── Attach passive wheel listener directly to canvas ref ──────────
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const wheelHandler = (e: WheelEvent) => {
      e.preventDefault()
      const vp = viewportRef.current
      const scaleFactor = e.deltaY > 0 ? 0.92 : 1.08
      const newScale = Math.max(0.2, Math.min(3, vp.scale * scaleFactor))
      const rect = canvas.getBoundingClientRect()
      const mx = e.clientX - rect.left
      const my = e.clientY - rect.top
      vp.x = mx - (mx - vp.x) * (newScale / vp.scale)
      vp.y = my - (my - vp.y) * (newScale / vp.scale)
      vp.scale = newScale
    }

    canvas.addEventListener('wheel', wheelHandler, { passive: false })
    return () => canvas.removeEventListener('wheel', wheelHandler)
  }, [])

  // ─── Mouse Event Handlers ─────────────────────────────────────────
  const getCanvasPos = useCallback((e: React.MouseEvent) => {
    const rect = canvasRef.current?.getBoundingClientRect()
    if (!rect) return { x: 0, y: 0 }
    const vp = viewportRef.current
    return {
      x: (e.clientX - rect.left - vp.x) / vp.scale,
      y: (e.clientY - rect.top - vp.y) / vp.scale,
    }
  }, [])

  const onMouseDown = (e: React.MouseEvent) => {
    const pos = getCanvasPos(e)
    const nodes = nodesRef.current
    // Check if clicked on a node
    for (let i = nodes.length - 1; i >= 0; i--) {
      const n = nodes[i]
      const dx = pos.x - n.x; const dy = pos.y - n.y
      if (dx * dx + dy * dy < (n.radius + 5) * (n.radius + 5)) {
        mouseRef.current.dragNode = n
        mouseRef.current.dragOffset = { x: pos.x - n.x, y: pos.y - n.y }
        onNodeSelect?.(n)
        return
      }
    }
    mouseRef.current.down = true
    mouseRef.current.x = e.clientX
    mouseRef.current.y = e.clientY
    onNodeSelect?.(null)
  }

  const onMouseMove = (e: React.MouseEvent) => {
    const mouse = mouseRef.current
    if (mouse.dragNode) {
      const pos = getCanvasPos(e)
      mouse.dragNode.x = pos.x - mouse.dragOffset.x
      mouse.dragNode.y = pos.y - mouse.dragOffset.y
    } else if (mouse.down) {
      const vp = viewportRef.current
      vp.x += e.clientX - mouse.x
      vp.y += e.clientY - mouse.y
      mouse.x = e.clientX
      mouse.y = e.clientY
    }
  }

  const onMouseUp = () => {
    mouseRef.current.down = false
    mouseRef.current.dragNode = null
  }

  const onWheel = (e: React.WheelEvent) => {
    e.preventDefault()
    const vp = viewportRef.current
    const scaleFactor = e.deltaY > 0 ? 0.92 : 1.08
    const newScale = Math.max(0.2, Math.min(3, vp.scale * scaleFactor))
    // Zoom towards cursor
    const rect = canvasRef.current?.getBoundingClientRect()
    if (rect) {
      const mx = e.clientX - rect.left
      const my = e.clientY - rect.top
      vp.x = mx - (mx - vp.x) * (newScale / vp.scale)
      vp.y = my - (my - vp.y) * (newScale / vp.scale)
    }
    vp.scale = newScale
  }

  return (
    <canvas
      ref={canvasRef}
      width={width}
      height={height}
      className="cursor-grab active:cursor-grabbing rounded-2xl"
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={onMouseUp}
      onMouseLeave={onMouseUp}
    />
  )
}

// ─── Helper Functions ────────────────────────────────────────────────

function getNodeColor(node: TopoNode): string {
  if (!node.online) return COLORS.offline
  if (node.is_leader) return COLORS.leader
  if (node.role === 'candidate') return COLORS.candidate
  return COLORS.online
}

function getNodeRadius(node: TopoNode): number {
  let r = node.is_leader ? 26 : 20
  if (node.user_count > 0) r += Math.min(node.user_count * 0.5, 12)
  return r
}

function darkenColor(hex: string, amount: number): string {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return `rgb(${r * (1 - amount)}, ${g * (1 - amount)}, ${b * (1 - amount)})`
}

// ─── Default Export ──────────────────────────────────────────────────

export function useTopologyData() {
  const [nodes, setNodes] = useState<TopoNode[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    try {
      const res = await fetch('/api/v1/cluster/topology')
      const data = await res.json()
      setNodes(data.peers || [])
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load(); const t = setInterval(load, 15000); return () => clearInterval(t) }, [load])

  return { nodes, loading, error, refresh: load }
}
