import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '../api/client'

interface Container {
  id: string
  name: string
  image: string
  status: string
  state: string
  ports: string
  created: string
  uptime: string
}

interface DockerImage {
  id: string
  repository: string
  tag: string
  size: string
  created: string
}

interface DockerStatus {
  running: boolean
  version: string
  containers: number
  running_containers: number
  images: number
  used_memory: string
}

type TabKey = 'containers' | 'images' | 'status'

export function DockerPage() {
  const [tab, setTab] = useState<TabKey>('containers')
  const [containers, setContainers] = useState<Container[]>([])
  const [images, setImages] = useState<DockerImage[]>([])
  const [status, setStatus] = useState<DockerStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  // Create container form
  const [createForm, setCreateForm] = useState({ name: '', image: '', port_maps: '' })
  const [pullImage, setPullImage] = useState('')

  const loadData = useCallback(async () => {
    try {
      const [cRes, iRes, sRes] = await Promise.all([
        apiClient.get('/api/v1/docker/containers'),
        apiClient.get('/api/v1/docker/images'),
        apiClient.get('/api/v1/docker/status'),
      ])
      setContainers(cRes.data.data || [])
      setImages(iRes.data.data || [])
      setStatus(sRes.data.data)
    } catch {}
    finally { setLoading(false) }
  }, [])

  useEffect(() => { loadData() }, [loadData])

  const containerAction = useCallback(async (id: string, action: string) => {
    setActionLoading(`${action}-${id}`)
    try {
      await apiClient.post(`/api/v1/docker/containers/${id}/${action}`)
      await loadData()
    } catch {}
    finally { setActionLoading(null) }
  }, [loadData])

  const createContainer = useCallback(async () => {
    try {
      const portMaps = createForm.port_maps.split(',').map(p => p.trim()).filter(Boolean)
      await apiClient.post('/api/v1/docker/containers', {
        name: createForm.name,
        image: createForm.image,
        port_maps: portMaps,
      })
      setCreateForm({ name: '', image: '', port_maps: '' })
      await loadData()
    } catch {}
  }, [createForm, loadData])

  const handlePullImage = useCallback(async () => {
    if (!pullImage) return
    try {
      await apiClient.post('/api/v1/docker/images/pull', { image: pullImage })
      setPullImage('')
      await loadData()
    } catch {}
  }, [pullImage, loadData])

  const removeImage = useCallback(async (id: string) => {
    try {
      await apiClient.delete(`/api/v1/docker/images/${id}`)
      await loadData()
    } catch {}
  }, [loadData])

  const removeContainer = useCallback(async (id: string) => {
    try {
      await apiClient.delete(`/api/v1/docker/containers/${id}`)
      await loadData()
    } catch {}
  }, [loadData])

  const getContainerLogs = useCallback(async (id: string) => {
    try {
      const res = await apiClient.get(`/api/v1/docker/containers/${id}/logs`, { params: { tail: 50 } })
      const logs = res.data.data || 'No logs'
      alert(logs)
    } catch {}
  }, [])

  if (loading) return <div className="flex items-center justify-center min-h-[400px] text-[#6868a0]">Loading Docker...</div>

  return (
    <div className="space-y-6 page-enter">
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" /></svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-white">Docker Native Mode</h1>
            <p className="text-sm text-[#6868a0] mt-0.5">Container management for xray/sing-box core instances</p>
          </div>
        </div>
      </div>

      {status && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {[
            { label: 'Docker Engine', value: status.version || (status.running ? 'Running' : 'Stopped'), color: status.running ? 'text-emerald-400' : 'text-red-400' },
            { label: 'Containers', value: `${status.running_containers}/${status.containers} running`, color: 'text-white' },
            { label: 'Images', value: String(status.images), color: 'text-white' },
            { label: 'Memory', value: status.used_memory || 'N/A', color: 'text-white' },
          ].map((s, i) => (
            <div key={i} className="glass-card px-4 py-3 text-center">
              <p className="text-[10px] text-[#6868a0] uppercase tracking-wider">{s.label}</p>
              <p className={`text-lg font-bold mt-1 ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>
      )}

      <div className="flex gap-1 p-1 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)] w-fit flex-wrap">
        <button onClick={() => setTab('containers')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'containers' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Containers</button>
        <button onClick={() => setTab('images')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'images' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Images</button>
        <button onClick={() => setTab('status')} className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === 'status' ? 'bg-purple-500/20 text-purple-300' : 'text-[#6868a0] hover:text-white'}`}>Docker Status</button>
      </div>

      {tab === 'containers' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">➕ Create Container</h3>
            <div className="space-y-3">
              <input value={createForm.name} onChange={e => setCreateForm(p => ({ ...p, name: e.target.value }))} placeholder="Container name" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <input value={createForm.image} onChange={e => setCreateForm(p => ({ ...p, image: e.target.value }))} placeholder="Image (e.g., teddysun/xray:latest)" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <input value={createForm.port_maps} onChange={e => setCreateForm(p => ({ ...p, port_maps: e.target.value }))} placeholder="Port maps: 80:80,443:443" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" />
              <button onClick={createContainer} className="w-full px-5 py-2.5 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-lg hover:from-blue-500 hover:to-indigo-500 transition text-sm font-medium">Create & Start</button>
            </div>
          </div>
          <div className="lg:col-span-2 glass-card overflow-hidden">
            <div className="p-4 border-b border-[rgba(255,255,255,0.06)] flex items-center justify-between">
              <h3 className="text-white font-semibold">📋 Running Containers</h3>
              <button onClick={loadData} className="text-xs text-purple-400 hover:text-purple-300">🔄 Refresh</button>
            </div>
            <div className="overflow-x-auto">
              <table className="table-modern">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Image</th>
                    <th>Status</th>
                    <th>Ports</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {containers.map(c => (
                    <tr key={c.id}>
                      <td className="text-white text-sm font-medium">{c.name}</td>
                      <td className="text-xs text-[#9898b8]">{c.image}</td>
                      <td>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${c.state === 'running' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
                          {c.state || c.status}
                        </span>
                      </td>
                      <td className="text-xs text-[#9898b8]">{c.ports || '—'}</td>
                      <td>
                        <div className="flex gap-1">
                          {c.state !== 'running' && (
                            <button onClick={() => containerAction(c.id, 'start')} disabled={actionLoading === `start-${c.id}`} className="px-2 py-1 rounded bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 text-[10px] font-medium transition disabled:opacity-50">Start</button>
                          )}
                          {c.state === 'running' && (
                            <button onClick={() => containerAction(c.id, 'stop')} disabled={actionLoading === `stop-${c.id}`} className="px-2 py-1 rounded bg-red-500/10 text-red-400 hover:bg-red-500/20 text-[10px] font-medium transition disabled:opacity-50">Stop</button>
                          )}
                          {c.state === 'running' && (
                            <button onClick={() => containerAction(c.id, 'restart')} disabled={actionLoading === `restart-${c.id}`} className="px-2 py-1 rounded bg-yellow-500/10 text-yellow-400 hover:bg-yellow-500/20 text-[10px] font-medium transition disabled:opacity-50">Restart</button>
                          )}
                          <button onClick={() => getContainerLogs(c.id)} className="px-2 py-1 rounded bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 text-[10px] font-medium transition">Logs</button>
                          <button onClick={() => removeContainer(c.id)} className="px-2 py-1 rounded bg-red-500/10 text-red-400 hover:bg-red-500/20 text-[10px] font-medium transition">Delete</button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {containers.length === 0 && <tr><td colSpan={5} className="text-center py-12 text-[#585878]">No containers. Create one or check Docker engine.</td></tr>}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {tab === 'images' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="glass-card p-5">
            <h3 className="text-white font-semibold mb-3">⬇️ Pull Image</h3>
            <div className="space-y-3">
              <input value={pullImage} onChange={e => setPullImage(e.target.value)} placeholder="e.g., teddysun/xray:latest" className="w-full px-4 py-2.5 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] rounded-lg text-white text-sm focus:border-purple-500/40 focus:outline-none" onKeyDown={e => e.key === 'Enter' && handlePullImage()} />
              <button onClick={handlePullImage} disabled={!pullImage} className="w-full px-5 py-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-lg hover:from-indigo-500 hover:to-purple-500 disabled:opacity-50 transition text-sm font-medium">Pull Image</button>
              <button onClick={async () => { await apiClient.post('/api/v1/docker/images/prune'); await loadData() }} className="w-full px-3 py-2 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.08)] text-[#9898b8] rounded-lg hover:text-white text-xs">🧹 Prune Unused Images</button>
            </div>
          </div>
          <div className="lg:col-span-2 glass-card overflow-hidden">
            <div className="p-4 border-b border-[rgba(255,255,255,0.06)]">
              <h3 className="text-white font-semibold">📦 Available Images</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="table-modern">
                <thead>
                  <tr>
                    <th>Repository</th>
                    <th>Tag</th>
                    <th>Size</th>
                    <th>Created</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {images.map(img => (
                    <tr key={img.id}>
                      <td className="text-white text-sm font-medium">{img.repository}</td>
                      <td><span className="px-2 py-0.5 rounded text-[10px] font-medium bg-purple-500/10 text-purple-300">{img.tag}</span></td>
                      <td className="text-xs text-[#9898b8]">{img.size}</td>
                      <td className="text-xs text-[#585878]">{img.created}</td>
                      <td>
                        <button onClick={() => removeImage(img.id)} className="px-2 py-1 rounded bg-red-500/10 text-red-400 hover:bg-red-500/20 text-[10px] font-medium transition">Delete</button>
                      </td>
                    </tr>
                  ))}
                  {images.length === 0 && <tr><td colSpan={5} className="text-center py-12 text-[#585878]">No images. Pull one from Docker Hub.</td></tr>}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {tab === 'status' && (
        <div className="glass-card p-6">
          <h3 className="text-white font-semibold mb-6">🐳 Docker Engine Details</h3>
          {status ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {[
                { label: 'Engine Status', value: status.running ? '✅ Running' : '❌ Not running', color: status.running ? 'text-emerald-400' : 'text-red-400' },
                { label: 'Docker Version', value: status.version || 'N/A', color: 'text-white' },
                { label: 'Total Containers', value: String(status.containers), color: 'text-white' },
                { label: 'Running Containers', value: String(status.running_containers), color: 'text-emerald-400' },
                { label: 'Total Images', value: String(status.images), color: 'text-white' },
                { label: 'Memory Usage', value: status.used_memory || 'N/A', color: 'text-white' },
              ].map((s, i) => (
                <div key={i} className="p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.06)]">
                  <p className="text-xs text-[#6868a0] mb-1">{s.label}</p>
                  <p className={`text-lg font-bold ${s.color}`}>{s.value}</p>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-[#585878] py-8 text-center">Unable to connect to Docker engine. Make sure Docker is running.</p>
          )}
        </div>
      )}
    </div>
  )
}
