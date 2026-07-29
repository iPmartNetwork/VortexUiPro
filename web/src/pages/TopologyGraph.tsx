import React, { useEffect, useRef, useState, useCallback } from 'react';

interface TopoNode {
  id: number;
  name: string;
  address: string;
  role: string;
  status: string;
  cpu_load: number;
  memory_used: number;
  region: string;
  priority: number;
  user_count: number;
  last_heartbeat: number;
  online: boolean;
  is_leader: boolean;
}

interface TopoEdge {
  source: number;
  target: number;
  latency: number;
}

interface TopologyData {
  peers: TopoNode[];
  total: number;
  online: number;
}

// Force-directed physics simulation
interface SimNode extends TopoNode {
  x: number;
  y: number;
  vx: number;
  vy: number;
  fx?: number;
  fy?: number;
}

export function TopologyGraph() {
  const svgRef = useRef<SVGSVGElement>(null);
  const [nodes, setNodes] = useState<SimNode[]>([]);
  const [edges, setEdges] = useState<TopoEdge[]>([]);
  const [selectedNode, setSelectedNode] = useState<SimNode | null>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 500 });
  const animRef = useRef<number>(0);
  const wsRef = useRef<WebSocket | null>(null);

  // Load initial topology
  useEffect(() => {
    fetch('/api/v1/cluster/topology')
      .then(r => r.json())
      .then((data: TopologyData) => {
        const simNodes: SimNode[] = (data.peers || []).map((n, i) => ({
          ...n,
          x: 100 + Math.random() * (dimensions.width - 200),
          y: 100 + Math.random() * (dimensions.height - 200),
          vx: 0,
          vy: 0,
        }));
        setNodes(simNodes);

        // Create edges between all online pairs
        const onlineIds = simNodes.filter(n => n.online).map(n => n.id);
        const newEdges: TopoEdge[] = [];
        for (let i = 0; i < onlineIds.length; i++) {
          for (let j = i + 1; j < onlineIds.length; j++) {
            newEdges.push({
              source: onlineIds[i],
              target: onlineIds[j],
              latency: Math.floor(Math.random() * 50) + 1,
            });
          }
        }
        setEdges(newEdges);
      })
      .catch(() => {});

    // WebSocket for live updates
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws`);
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'topology' || msg.type === 'cluster') {
          const topo = msg.payload;
          if (topo.peers) {
            setNodes(prev => {
              const updated = [...prev];
              topo.peers.forEach((p: TopoNode) => {
                const idx = updated.findIndex(n => n.id === p.id);
                if (idx >= 0) {
                  updated[idx] = { ...updated[idx], ...p, x: updated[idx].x, y: updated[idx].y };
                } else {
                  updated.push({ ...p, x: Math.random() * 600, y: Math.random() * 400, vx: 0, vy: 0 });
                }
              });
              return updated;
            });
          }
        }
      } catch {}
    };
    wsRef.current = ws;

    return () => {
      ws.close();
      cancelAnimationFrame(animRef.current);
    };
  }, []);

  // Physics simulation
  useEffect(() => {
    if (nodes.length === 0) return;

    const W = dimensions.width;
    const H = dimensions.height;
    const REPULSION = 8000;
    const ATTRACTION = 0.005;
    const DAMPING = 0.9;
    const CENTER = 0.01;

    function simulate() {
      setNodes(prev => {
        const copy = prev.map(n => ({ ...n }));

        // Repulsion between all nodes
        for (let i = 0; i < copy.length; i++) {
          for (let j = i + 1; j < copy.length; j++) {
            const dx = copy[j].x - copy[i].x;
            const dy = copy[j].y - copy[i].y;
            const dist = Math.sqrt(dx * dx + dy * dy) || 1;
            const force = REPULSION / (dist * dist);
            const fx = (dx / dist) * force;
            const fy = (dy / dist) * force;
            copy[i].vx -= fx;
            copy[i].vy -= fy;
            copy[j].vx += fx;
            copy[j].vy += fy;
          }
        }

        // Attraction along edges
        for (const edge of edges) {
          const source = copy.find(n => n.id === edge.source);
          const target = copy.find(n => n.id === edge.target);
          if (!source || !target) continue;
          const dx = target.x - source.x;
          const dy = target.y - source.y;
          const dist = Math.sqrt(dx * dx + dy * dy) || 1;
          const force = (dist - 150) * ATTRACTION;
          source.vx += (dx / dist) * force;
          source.vy += (dy / dist) * force;
          target.vx -= (dx / dist) * force;
          target.vy -= (dy / dist) * force;
        }

        // Center gravity + damping + bounds
        for (const n of copy) {
          n.vx += (W / 2 - n.x) * CENTER;
          n.vy += (H / 2 - n.y) * CENTER;
          n.vx *= DAMPING;
          n.vy *= DAMPING;
          n.x += n.vx;
          n.y += n.vy;
          n.x = Math.max(30, Math.min(W - 30, n.x));
          n.y = Math.max(30, Math.min(H - 30, n.y));
        }

        return copy;
      });

      animRef.current = requestAnimationFrame(simulate);
    }

    animRef.current = requestAnimationFrame(simulate);
    return () => cancelAnimationFrame(animRef.current);
  }, [nodes.length, edges, dimensions]);

  // Resize handler
  const containerRef = useCallback((node: HTMLDivElement | null) => {
    if (node) {
      const rect = node.getBoundingClientRect();
      setDimensions({ width: rect.width, height: Math.max(400, rect.height) });
    }
  }, []);

  const getNodeColor = (node: TopoNode) => {
    if (!node.online) return '#9CA3AF';
    if (node.is_leader) return '#7C3AED';
    if (node.role === 'candidate') return '#F59E0B';
    return '#3B82F6';
  };

  const getNodeRadius = (node: TopoNode) => {
    let r = 20;
    if (node.is_leader) r = 28;
    if (node.user_count > 0) r += Math.min(node.user_count, 20);
    return r;
  };

  const getEdgeColor = (edge: TopoEdge) => {
    return edge.latency < 10 ? '#10B981' : edge.latency < 30 ? '#F59E0B' : '#EF4444';
  };

  return (
    <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] overflow-hidden">
      <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--border-light)]">
        <div className="flex items-center gap-3">
          <h3 className="font-semibold text-[var(--text-primary)]">Live Topology Map</h3>
          <div className="flex items-center gap-2 text-xs text-[var(--text-muted)]">
            <span className={`w-2 h-2 rounded-full ${nodes.some(n => n.is_leader) ? 'bg-purple-500' : 'bg-gray-300'}`} />
            <span>Leader</span>
            <span className="w-2 h-2 rounded-full bg-blue-500 ml-2" />
            <span>Online</span>
            <span className="w-2 h-2 rounded-full bg-gray-400 ml-2" />
            <span>Offline</span>
          </div>
        </div>
        <div className="text-xs text-[var(--text-muted)]">
          {nodes.filter(n => n.online).length}/{nodes.length} online
        </div>
      </div>

      <div ref={containerRef} className="relative" style={{ height: 500 }}>
        <svg ref={svgRef} width={dimensions.width} height={500} className="bg-gradient-to-br from-[var(--bg-surface)] to-[var(--bg-elevated)]">
          {/* Edges */}
          {edges.map(edge => {
            const source = nodes.find(n => n.id === edge.source);
            const target = nodes.find(n => n.id === edge.target);
            if (!source || !target) return null;
            return (
              <g key={`e-${edge.source}-${edge.target}`}>
                <line
                  x1={source.x} y1={source.y}
                  x2={target.x} y2={target.y}
                  stroke={getEdgeColor(edge)}
                  strokeWidth={1.5}
                  strokeOpacity={0.4}
                  className="transition-all duration-300"
                />
                <line
                  x1={source.x} y1={source.y}
                  x2={target.x} y2={target.y}
                  stroke={getEdgeColor(edge)}
                  strokeWidth={1}
                  strokeOpacity={0.15}
                  strokeDasharray="4 4"
                />
              </g>
            );
          })}

          {/* Nodes */}
          {nodes.map(node => (
            <g key={node.id}
              onClick={() => setSelectedNode(node)}
              className="cursor-pointer transition-transform duration-200 hover:scale-110"
              style={{ transform: `translate(${node.x}px, ${node.y}px)` }}
            >
              {/* Glow effect for leader */}
              {node.is_leader && (
                <circle r={getNodeRadius(node) + 8}
                  fill="none"
                  stroke="#7C3AED"
                  strokeWidth={2}
                  strokeOpacity={0.3}
                  className="animate-pulse"
                />
              )}
              {/* Main circle */}
              <circle r={getNodeRadius(node)}
                fill={getNodeColor(node)}
                stroke={node.is_leader ? '#5B21B6' : '#fff'}
                strokeWidth={3}
                className="drop-shadow-md"
              />
              {/* CPU load ring */}
              {node.online && (
                <circle r={getNodeRadius(node) + 4}
                  fill="none"
                  stroke="#10B981"
                  strokeWidth={2}
                  strokeOpacity={0.5}
                  strokeDasharray={`${node.cpu_load * 3} 100`}
                  transform="rotate(-90)"
                />
              )}
              {/* Node label */}
              <text textAnchor="middle" dy={getNodeRadius(node) + 14}
                className="text-xs fill-[var(--text-secondary)] font-medium"
                style={{ fontSize: '11px' }}
              >
                {node.name.length > 12 ? node.name.slice(0, 10) + '..' : node.name}
              </text>
              {/* Status dot */}
              <circle r={4} cy={-getNodeRadius(node) + 6}
                fill={node.online ? '#10B981' : '#EF4444'}
                stroke="#fff"
                strokeWidth={2}
              />
            </g>
          ))}

          {/* Empty state */}
          {nodes.length === 0 && (
            <text x={dimensions.width / 2} y={250}
              textAnchor="middle" className="fill-[var(--text-muted)] text-sm">
              No cluster nodes available
            </text>
          )}
        </svg>

        {/* Node detail panel */}
        {selectedNode && (
          <div className="absolute top-4 right-4 w-64 bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] shadow-xl p-4">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <div className={`w-3 h-3 rounded-full ${getNodeColor(selectedNode)}`} />
                <span className="font-semibold text-[var(--text-primary)]">{selectedNode.name}</span>
              </div>
              <button onClick={() => setSelectedNode(null)}
                className="text-[var(--text-muted)] hover:text-[var(--text-secondary)]">&times;</button>
            </div>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">Address</span>
                <span className="font-mono text-[var(--text-primary)]">{selectedNode.address}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">Role</span>
                <span className="font-medium capitalize">{selectedNode.role}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">Status</span>
                <span className={selectedNode.online ? 'text-green-600' : 'text-red-600'}>
                  {selectedNode.online ? 'Online' : 'Offline'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">CPU</span>
                <span className="text-[var(--text-primary)]">{(selectedNode.cpu_load * 100).toFixed(1)}%</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">Memory</span>
                <span className="text-[var(--text-primary)]">{(selectedNode.memory_used * 100).toFixed(1)}%</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">Region</span>
                <span className="text-[var(--text-primary)]">{selectedNode.region}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-[var(--text-secondary)]">Priority</span>
                <span className="text-[var(--text-primary)]">{selectedNode.priority}</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
