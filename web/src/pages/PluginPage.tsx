import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface PluginInfo {
  id: string
  name: string
  version: string
  description: string
  author: string
  hooks: string[]
  enabled: boolean
  status: string
  loaded_at: number
  error: string
}

export function PluginPage() {
  const [plugins, setPlugins] = useState<PluginInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [showLoadModal, setShowLoadModal] = useState(false)
  const [loadForm, setLoadForm] = useState({ path: '', id: '', name: '', version: '1.0.0' })

  useEffect(() => { fetchPlugins() }, [])

  const fetchPlugins = async () => {
    setLoading(true)
    try {
      const res = await apiClient.get('/api/v1/plugins')
      setPlugins(res.data?.plugins || [])
    } catch { }
    setLoading(false)
  }

  const handleLoad = async () => {
    try {
      await apiClient.post('/api/v1/plugins/load', loadForm)
      setShowLoadModal(false)
      setLoadForm({ path: '', id: '', name: '', version: '1.0.0' })
      fetchPlugins()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to load plugin')
    }
  }

  const togglePlugin = async (id: string, enabled: boolean) => {
    try {
      await apiClient.put(`/api/v1/plugins/${id}`, { enabled })
      fetchPlugins()
    } catch { }
  }

  const unloadPlugin = async (id: string) => {
    if (!confirm('Unload this plugin?')) return
    try {
      await apiClient.delete(`/api/v1/plugins/${id}`)
      fetchPlugins()
    } catch { }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'loaded': return 'bg-emerald-400'
      case 'error': return 'bg-red-400'
      case 'disabled': return 'bg-gray-500'
      default: return 'bg-gray-500'
    }
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500 to-rose-600 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Plugins</h1>
            <p className="text-sm text-[var(--text-secondary)]">Extend functionality via plugins</p>
          </div>
          <div className="ml-auto">
            <button onClick={() => setShowLoadModal(true)} className="btn-primary text-sm">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Load Plugin
            </button>
          </div>
        </div>
      </div>

      {/* Plugin List */}
      <div className="glass-card p-6">
        {loading ? (
          <div className="text-center py-12 text-[var(--text-muted)]">Loading plugins...</div>
        ) : plugins.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-[var(--text-muted)] mb-2">No plugins loaded</p>
            <p className="text-xs text-[#585878]">Click "Load Plugin" to load a .so plugin file</p>
          </div>
        ) : (
          <div className="space-y-3">
            {plugins.map(plugin => (
              <div key={plugin.id} className="flex items-center justify-between p-4 rounded-xl bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] hover:border-[rgba(139,92,246,0.1)] transition-all">
                <div className="flex items-center gap-4">
                  <div className={`w-2.5 h-2.5 rounded-full ${getStatusColor(plugin.status)}`} />
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="text-sm font-semibold text-[var(--text-primary)]">{plugin.name}</h3>
                      <span className="text-[10px] text-[#585878]">v{plugin.version}</span>
                      {plugin.author && <span className="text-[10px] text-[#585878]">by {plugin.author}</span>}
                    </div>
                    <p className="text-xs text-[var(--text-secondary)] mt-0.5">{plugin.description || 'No description'}</p>
                    <div className="flex gap-1.5 mt-1.5">
                      {plugin.hooks.map(hook => (
                        <span key={hook} className="px-1.5 py-0.5 rounded text-[9px] font-mono bg-[rgba(139,92,246,0.08)] text-purple-300">{hook}</span>
                      ))}
                      <span className="text-[10px] text-[#585878] ml-1">{plugin.status}</span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" checked={plugin.enabled} onChange={e => togglePlugin(plugin.id, e.target.checked)} className="sr-only peer" />
                    <div className="w-9 h-5 rounded-full bg-[var(--border-light)] peer-checked:bg-emerald-500 after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-4"></div>
                  </label>
                  <button onClick={() => unloadPlugin(plugin.id)} className="text-xs text-red-400 hover:text-red-300 transition">Unload</button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Load Modal */}
      {showLoadModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowLoadModal(false)}>
          <div className="w-full max-w-md p-6 rounded-2xl bg-[#14142a] border border-[rgba(139,92,246,0.1)] shadow-2xl" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-4">Load Plugin</h2>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Plugin Path (.so file)</label>
                <input type="text" value={loadForm.path} onChange={e => setLoadForm({ ...loadForm, path: e.target.value })} className="input-modern w-full text-sm" placeholder="/etc/vortex/plugins/myplugin.so" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Plugin ID</label>
                  <input type="text" value={loadForm.id} onChange={e => setLoadForm({ ...loadForm, id: e.target.value })} className="input-modern w-full text-sm" placeholder="my-plugin" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Name</label>
                  <input type="text" value={loadForm.name} onChange={e => setLoadForm({ ...loadForm, name: e.target.value })} className="input-modern w-full text-sm" placeholder="My Plugin" />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1">Version</label>
                <input type="text" value={loadForm.version} onChange={e => setLoadForm({ ...loadForm, version: e.target.value })} className="input-modern w-full text-sm" placeholder="1.0.0" />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button onClick={() => setShowLoadModal(false)} className="btn-secondary text-sm">Cancel</button>
              <button onClick={handleLoad} className="btn-primary text-sm">Load</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
