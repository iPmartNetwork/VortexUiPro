import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface TLSTrickConfig {
  id: string
  name: string
  type: string
  enabled: boolean
  description: string
  fragment_packets: string
  fragment_length: string
  fragment_sleep: string
  padding_type: string
  padding_size: string
  hello_fingerprint: string
  hello_alpn: string
}

interface TLSProfile {
  id: string
  name: string
  enabled: boolean
  description: string
  tricks: TLSTrickConfig[]
  created_at: number
}

const TRICK_ICONS: Record<string, string> = {
  fragment: '✂️',
  padding: '🧩',
  mixed_case: '🔀',
  tls_hello: '👋',
  tls_over_tls: '🔒',
  random_sni: '🎲',
}

export function TlsTricksPage() {
  const [profiles, setProfiles] = useState<TLSProfile[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedProfile, setSelectedProfile] = useState<TLSProfile | null>(null)

  useEffect(() => { fetchProfiles() }, [])

  const fetchProfiles = async () => {
    setLoading(true)
    try {
      const res = await apiClient.get('/api/v1/tls-tricks/profiles')
      setProfiles(res.data?.profiles || [])
    } catch { }
    setLoading(false)
  }

  const toggleProfile = async (id: string, enabled: boolean) => {
    try {
      await apiClient.put(`/api/v1/tls-tricks/profiles/${id}`, { enabled })
      fetchProfiles()
    } catch { }
  }

  const deleteProfile = async (id: string) => {
    if (!confirm('Delete this profile?')) return
    try {
      await apiClient.delete(`/api/v1/tls-tricks/profiles/${id}`)
      fetchProfiles()
    } catch { }
  }

  const generateConfig = async (profileId: string) => {
    try {
      const res = await apiClient.get(`/api/v1/tls-tricks/generate?profile_id=${profileId}`)
      alert(JSON.stringify(res.data, null, 2))
    } catch { }
  }

  if (loading) {
    return <div className="text-center py-20 text-[var(--text-muted)]">Loading...</div>
  }

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shadow-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">TLS Tricks Suite</h1>
            <p className="text-sm text-[var(--text-secondary)]">Anti-DPI profiles & traffic obfuscation</p>
          </div>
        </div>
      </div>

      {/* Profiles Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {profiles.map(profile => (
          <div key={profile.id} className={`glass-card p-5 transition-all duration-300 ${profile.enabled ? 'border-l-2 border-l-emerald-500' : 'opacity-60 hover:opacity-80'}`}>
            <div className="flex items-start justify-between mb-3">
              <div>
                <h3 className="text-sm font-bold text-[var(--text-primary)]">{profile.name}</h3>
                <p className="text-[10px] text-[var(--text-muted)] mt-0.5">{profile.id}</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" checked={profile.enabled} onChange={e => toggleProfile(profile.id, e.target.checked)} className="sr-only peer" />
                <div className="w-9 h-5 rounded-full bg-[var(--border-light)] peer-checked:bg-emerald-500 after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-4"></div>
              </label>
            </div>

            {profile.description && (
              <p className="text-xs text-[var(--text-secondary)] mb-3">{profile.description}</p>
            )}

            {/* Tricks */}
            <div className="space-y-1.5 mb-3">
              {profile.tricks.map(trick => (
                <div key={trick.id} className="flex items-center gap-2 p-1.5 rounded-lg bg-[rgba(255,255,255,0.02)]">
                  <span className="text-sm">{TRICK_ICONS[trick.type] || '🔧'}</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium text-[var(--text-primary)]">{trick.name}</p>
                    <p className="text-[10px] text-[var(--text-muted)] truncate">{trick.type}</p>
                  </div>
                  <div className={`w-1.5 h-1.5 rounded-full ${trick.enabled ? 'bg-emerald-400' : 'bg-gray-500'}`} />
                </div>
              ))}
            </div>

            {/* Actions */}
            <div className="flex gap-2 pt-2 border-t border-[rgba(255,255,255,0.06)]">
              <button onClick={() => generateConfig(profile.id)} className="flex-1 py-1.5 rounded-lg text-[10px] font-medium bg-[rgba(139,92,246,0.1)] text-purple-300 hover:bg-[rgba(139,92,246,0.2)] transition">Generate Config</button>
              <button onClick={() => deleteProfile(profile.id)} className="py-1.5 px-2 rounded-lg text-[10px] font-medium bg-red-500/10 text-red-400 hover:bg-red-500/20 transition">Delete</button>
            </div>
          </div>
        ))}

        {profiles.length === 0 && (
          <div className="col-span-full text-center py-12 text-[var(--text-muted)]">No TLS trick profiles configured</div>
        )}
      </div>
    </div>
  )
}
