import React, { useState, useEffect } from 'react';
import { TopologyGraph } from './TopologyGraph';

interface ClusterNode {
  id: number;
  name: string;
  address: string;
  peer_port: number;
  role: string;
  status: string;
  priority: number;
  term: number;
  cpu_load: number;
  memory_used: number;
  last_heartbeat: number;
  last_synced_at: number;
  region: string;
  enabled: boolean;
  created_at: number;
  updated_at: number;
}

interface SyncEvent {
  id: number;
  type: string;
  source_id: number;
  entity_id: string;
  status: string;
  detail: string;
  created_at: number;
}

interface ClusterStatus {
  enabled: boolean;
  node_id: number;
  node_name: string;
  addr: string;
  region: string;
  peers: { id: number; name: string; address: string; priority: number; online: boolean }[];
  started: boolean;
  election: Record<string, any>;
  conflict_resolver: Record<string, any>;
}

type TabType = 'nodes' | 'topology' | 'events' | 'election';

export function ClusterManagementPage() {
  const [status, setStatus] = useState<ClusterStatus | null>(null);
  const [nodes, setNodes] = useState<ClusterNode[]>([]);
  const [events, setEvents] = useState<SyncEvent[]>([]);
  const [activeTab, setActiveTab] = useState<TabType>('nodes');
  const [showAddModal, setShowAddModal] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 10000);
    return () => clearInterval(interval);
  }, []);

  async function loadData() {
    try {
      const [statusRes, nodesRes, eventsRes] = await Promise.all([
        fetch('/api/v1/cluster/status').then(r => r.json()),
        fetch('/api/v1/cluster/nodes').then(r => r.json()),
        fetch('/api/v1/cluster/sync-events?limit=20').then(r => r.json()),
      ]);
      setStatus(statusRes);
      setNodes(Array.isArray(nodesRes) ? nodesRes : []);
      setEvents(Array.isArray(eventsRes) ? eventsRes : []);
    } catch (err: any) {
      setError(err.message || 'Failed to load cluster data');
    } finally {
      setLoading(false);
    }
  }

  function formatTime(ts: number) {
    if (!ts) return '-';
    return new Date(ts).toLocaleString();
  }

  function formatLoad(val: number) {
    return val ? val.toFixed(1) + '%' : '-';
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'online': return 'text-green-600 bg-green-50';
      case 'offline': return 'text-[var(--text-secondary)] bg-[var(--bg-surface)]';
      case 'syncing': return 'text-yellow-600 bg-yellow-50';
      default: return 'text-[var(--text-secondary)] bg-[var(--bg-surface)]';
    }
  }

  function getRoleBadge(role: string) {
    switch (role) {
      case 'leader': return 'bg-purple-100 text-purple-800';
      case 'follower': return 'bg-blue-100 text-blue-800';
      case 'candidate': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-[var(--text-secondary)]';
    }
  }

  async function forceElection() {
    try {
      const res = await fetch('/api/v1/cluster/election/force', { method: 'POST' });
      const data = await res.json();
      if (res.ok) {
        loadData();
      } else {
        setError(data.error || 'Failed to trigger election');
      }
    } catch (err: any) {
      setError(err.message);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-[var(--text-primary)]">Cluster Management</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            Multi-node mesh — Peer-to-peer gRPC protocol with leader election & data sync
          </p>
        </div>
        {status?.enabled && (
          <div className="flex items-center gap-2">
            <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${status?.started ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
              {status?.started ? '▶ Running' : '⏹ Stopped'}
            </span>
          </div>
        )}
      </div>

      {/* Error display */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
          {error}
          <button onClick={() => setError('')} className="float-right font-bold">&times;</button>
        </div>
      )}

      {/* Cluster disabled state */}
      {status && !status.enabled && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-8 text-center">
          <div className="text-4xl mb-3">🌐</div>
          <h3 className="text-lg font-semibold text-yellow-800 mb-2">Cluster Mode Disabled</h3>
          <p className="text-yellow-600 text-sm max-w-md mx-auto">
            Set <code className="bg-yellow-100 px-2 py-0.5 rounded">VORTEX_CLUSTER_ENABLED=true</code> and configure <code className="bg-yellow-100 px-2 py-0.5 rounded">VORTEX_CLUSTER_PEERS</code> environment variables to enable multi-node clustering with leader election and data synchronization.
          </p>
        </div>
      )}

      {status?.enabled && (
        <>
          {/* Quick Stats */}
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-4">
              <div className="text-xs text-[var(--text-secondary)] uppercase tracking-wide">This Node</div>
              <div className="text-lg font-bold text-[var(--text-primary)] mt-1">{status.node_name}</div>
              <div className="text-xs text-[var(--text-muted)]">{status.addr}</div>
            </div>
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-4">
              <div className="text-xs text-[var(--text-secondary)] uppercase tracking-wide">Leader</div>
              <div className="text-lg font-bold text-purple-900 mt-1">{status.election?.leader_name || '-'}</div>
              <div className="text-xs text-[var(--text-muted)]">Term #{status.election?.term || 0}</div>
            </div>
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-4">
              <div className="text-xs text-[var(--text-secondary)] uppercase tracking-wide">Peers</div>
              <div className="text-lg font-bold text-[var(--text-primary)] mt-1">
                {status.peers?.filter(p => p.online).length || 0}
                <span className="text-sm text-[var(--text-muted)] font-normal"> / {status.peers?.length || 0} online</span>
              </div>
            </div>
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-4">
              <div className="text-xs text-[var(--text-secondary)] uppercase tracking-wide">Role</div>
              <div className="mt-1">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleBadge(status.election?.role || 'follower')}`}>
                  {status.election?.role || 'follower'}
                </span>
              </div>
            </div>
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-4">
              <div className="text-xs text-[var(--text-secondary)] uppercase tracking-wide">Conflicts</div>
              <div className="text-lg font-bold text-[var(--text-primary)] mt-1">
                {status.conflict_resolver?.tracked_entities || 0}
              </div>
              <div className="text-xs text-[var(--text-muted)]">tracked entities</div>
            </div>
          </div>

          {/* Tabs */}
          <div className="border-b border-[var(--border-light)]">
            <nav className="flex gap-6">
              {[
                { id: 'nodes' as TabType, label: 'Nodes', icon: '🖥' },
                { id: 'topology' as TabType, label: 'Topology', icon: '🔗' },
                { id: 'election' as TabType, label: 'Election', icon: '🗳' },
                { id: 'events' as TabType, label: 'Sync Events', icon: '📋' },
              ].map(tab => (
                <button key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 py-3 px-1 border-b-2 text-sm font-medium transition-colors ${activeTab === tab.id ? 'border-purple-600 text-purple-600' : 'border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:border-[var(--border-light)]'}`}
                >
                  <span>{tab.icon}</span>
                  <span>{tab.label}</span>
                </button>
              ))}
            </nav>
          </div>

          {/* Tab Content: Nodes */}
          {activeTab === 'nodes' && (
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] overflow-hidden">
              <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--border-light)]">
                <h3 className="font-semibold text-[var(--text-primary)]">Cluster Nodes</h3>
                <button onClick={() => setShowAddModal(true)}
                  className="px-4 py-2 bg-purple-600 text-white text-sm font-medium rounded-lg hover:bg-purple-700 transition-colors">
                  + Add Node
                </button>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="bg-[var(--bg-surface)] text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">
                      <th className="px-6 py-3">Name</th>
                      <th className="px-6 py-3">Address</th>
                      <th className="px-6 py-3">Role</th>
                      <th className="px-6 py-3">Status</th>
                      <th className="px-6 py-3">Priority</th>
                      <th className="px-6 py-3">Region</th>
                      <th className="px-6 py-3">Term</th>
                      <th className="px-6 py-3">Last Heartbeat</th>
                      <th className="px-6 py-3"></th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[var(--border-light)]">
                    {nodes.map(node => (
                      <tr key={node.id} className="hover:bg-[var(--bg-surface)] transition-colors">
                        <td className="px-6 py-4">
                          <div className="font-medium text-[var(--text-primary)]">{node.name}</div>
                        </td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)] font-mono">{node.address}</td>
                        <td className="px-6 py-4">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleBadge(node.role)}`}>
                            {node.role}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(node.status)}`}>
                            {node.status}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)]">{node.priority}</td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)]">{node.region}</td>
                        <td className="px-6 py-4 text-sm font-mono text-[var(--text-secondary)]">{node.term}</td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)]">{formatTime(node.last_heartbeat)}</td>
                        <td className="px-6 py-4">
                          <button className="text-purple-600 hover:text-purple-800 text-sm font-medium">Edit</button>
                        </td>
                      </tr>
                    ))}
                    {nodes.length === 0 && (
                      <tr>
                        <td colSpan={9} className="px-6 py-8 text-center text-sm text-[var(--text-secondary)]">
                          No cluster nodes found. Add a node to get started.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Tab Content: Topology */}
          {activeTab === 'topology' && (
            <div className="space-y-6">
              <TopologyGraph />
              <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-6">
                <h3 className="font-semibold text-[var(--text-primary)] mb-4">Node Details</h3>
                {status && (
                  <div className="space-y-3">
                    <div className="grid grid-cols-2 gap-3">
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Node Name</div>
                        <div className="font-medium text-[var(--text-primary)]">{status.node_name}</div>
                      </div>
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Address</div>
                        <div className="font-medium text-[var(--text-primary)] font-mono">{status.addr}</div>
                      </div>
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Region</div>
                        <div className="font-medium text-[var(--text-primary)]">{status.region}</div>
                      </div>
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Node ID</div>
                        <div className="font-medium text-[var(--text-primary)]">{status.node_id}</div>
                      </div>
                    </div>
                    <div className="flex gap-2 mt-4">
                      <button onClick={forceElection}
                        className="px-4 py-2 bg-yellow-500 text-white text-sm font-medium rounded-lg hover:bg-yellow-600 transition-colors">
                        ⚡ Force Election
                      </button>
                      <button onClick={loadData}
                        className="px-4 py-2 bg-gray-100 text-[var(--text-primary)] text-sm font-medium rounded-lg hover:bg-gray-200 transition-colors">
                        🔄 Refresh
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Tab Content: Election */}
          {activeTab === 'election' && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-6">
                <h3 className="font-semibold text-[var(--text-primary)] mb-4">Leader Election Status</h3>
                {status?.election && (
                  <div className="space-y-4">
                    <div className="flex items-center gap-4 p-4 bg-purple-50 rounded-lg border border-purple-100">
                      <div className="text-3xl">👑</div>
                      <div>
                        <div className="text-sm text-purple-600 font-medium">Current Leader</div>
                        <div className="text-lg font-bold text-purple-900">{status.election.leader_name || '-'}</div>
                        <div className="text-xs text-purple-500">Node ID: {status.election.leader_id || '-'}</div>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Term</div>
                        <div className="font-bold text-lg text-[var(--text-primary)]">{status.election.term || 0}</div>
                      </div>
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Role</div>
                        <div className={`font-medium text-lg ${getRoleBadge(status.election.role || 'follower')}`}>
                          {status.election.role || 'follower'}
                        </div>
                      </div>
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Priority</div>
                        <div className="font-bold text-lg text-[var(--text-primary)]">{status.election.priority || 0}</div>
                      </div>
                      <div className="bg-[var(--bg-surface)] rounded-lg p-3">
                        <div className="text-xs text-[var(--text-secondary)]">Voted For</div>
                        <div className="font-medium text-lg text-[var(--text-primary)]">{status.election.voted_for || '-'}</div>
                      </div>
                    </div>
                  </div>
                )}
                {!status?.election && (
                  <div className="text-sm text-[var(--text-secondary)] text-center py-4">No election data available</div>
                )}
              </div>

              <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] p-6">
                <h3 className="font-semibold text-[var(--text-primary)] mb-4">Election Algorithm</h3>
                <div className="prose prose-sm max-w-none">
                  <p className="text-[var(--text-secondary)]">The cluster uses a <strong>Bully Algorithm</strong> with Raft-style terms for leader election:</p>
                  <ul className="space-y-2 text-sm text-[var(--text-secondary)] mt-3">
                    <li className="flex gap-2">
                      <span className="text-purple-500 font-bold">1.</span>
                      <span>When a follower node detects heartbeat timeout (<strong>15s</strong>), it becomes a <strong>candidate</strong>.</span>
                    </li>
                    <li className="flex gap-2">
                      <span className="text-purple-500 font-bold">2.</span>
                      <span>The candidate requests votes from peers with <strong>higher priority</strong>.</span>
                    </li>
                    <li className="flex gap-2">
                      <span className="text-purple-500 font-bold">3.</span>
                      <span>If the candidate gets a <strong>majority</strong>, it becomes the leader.</span>
                    </li>
                    <li className="flex gap-2">
                      <span className="text-purple-500 font-bold">4.</span>
                      <span>The leader broadcasts <strong>heartbeats</strong> every <strong>5s</strong> and manages data sync.</span>
                    </li>
                    <li className="flex gap-2">
                      <span className="text-purple-500 font-bold">5.</span>
                      <span><strong>Conflict resolution</strong> uses Last-Write-Wins (LWW) with timestamps.</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Tab Content: Sync Events */}
          {activeTab === 'events' && (
            <div className="bg-[var(--bg-elevated)] rounded-xl border border-[var(--border-light)] overflow-hidden">
              <div className="px-6 py-4 border-b border-[var(--border-light)]">
                <h3 className="font-semibold text-[var(--text-primary)]">Sync Events</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="bg-[var(--bg-surface)] text-left text-xs font-medium text-[var(--text-secondary)] uppercase tracking-wider">
                      <th className="px-6 py-3">ID</th>
                      <th className="px-6 py-3">Type</th>
                      <th className="px-6 py-3">Source</th>
                      <th className="px-6 py-3">Entity</th>
                      <th className="px-6 py-3">Status</th>
                      <th className="px-6 py-3">Detail</th>
                      <th className="px-6 py-3">Time</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[var(--border-light)]">
                    {events.map(evt => (
                      <tr key={evt.id} className="hover:bg-[var(--bg-surface)] transition-colors">
                        <td className="px-6 py-4 text-sm font-mono text-[var(--text-secondary)]">{evt.id}</td>
                        <td className="px-6 py-4">
                          <span className="px-2 py-0.5 text-xs rounded-full bg-purple-50 text-purple-700 font-medium">
                            {evt.type}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)]">Node #{evt.source_id}</td>
                        <td className="px-6 py-4 text-sm font-mono text-[var(--text-secondary)]">{evt.entity_id || '-'}</td>
                        <td className="px-6 py-4">
                          <span className={`px-2 py-0.5 text-xs rounded-full font-medium ${
                            evt.status === 'applied' ? 'bg-green-50 text-green-700' :
                            evt.status === 'broadcast' ? 'bg-blue-50 text-blue-700' :
                            evt.status === 'failed' ? 'bg-red-50 text-red-700' :
                            'bg-[var(--bg-surface)] text-[var(--text-primary)]'
                          }`}>
                            {evt.status}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)]">{evt.detail || '-'}</td>
                        <td className="px-6 py-4 text-sm text-[var(--text-secondary)]">{formatTime(evt.created_at)}</td>
                      </tr>
                    ))}
                    {events.length === 0 && (
                      <tr>
                        <td colSpan={7} className="px-6 py-8 text-center text-sm text-[var(--text-secondary)]">
                          No sync events yet.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}

      {/* Add Node Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowAddModal(false)}>
          <div className="bg-[var(--bg-elevated)] rounded-2xl p-6 w-full max-w-md shadow-2xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-bold text-[var(--text-primary)] mb-4">Add Cluster Node</h3>
            <form onSubmit={async (e) => {
              e.preventDefault();
              const form = e.target as HTMLFormElement;
              const data = Object.fromEntries(new FormData(form));
              try {
                const res = await fetch('/api/v1/cluster/nodes', {
                  method: 'POST',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify(data),
                });
                if (res.ok) {
                  setShowAddModal(false);
                  loadData();
                } else {
                  const err = await res.json();
                  setError(err.error || 'Failed to add node');
                }
              } catch (err: any) {
                setError(err.message);
              }
            }}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-[var(--text-primary)] mb-1">Node Name</label>
                  <input name="name" required
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none"
                    placeholder="node-2" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--text-primary)] mb-1">Address</label>
                  <input name="address" required
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none"
                    placeholder="192.168.1.2:1337" />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-primary)] mb-1">Priority</label>
                    <input name="priority" type="number" defaultValue={50}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-[var(--text-primary)] mb-1">Region</label>
                    <input name="region" defaultValue="default"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
                  </div>
                </div>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button type="button" onClick={() => setShowAddModal(false)}
                  className="px-4 py-2 text-sm font-medium text-[var(--text-primary)] hover:text-[var(--text-primary)] transition-colors">
                  Cancel
                </button>
                <button type="submit"
                  className="px-4 py-2 bg-purple-600 text-white text-sm font-medium rounded-lg hover:bg-purple-700 transition-colors">
                  Add Node
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
