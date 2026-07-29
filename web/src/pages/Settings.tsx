import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import { useI18nStore } from '../hooks/useI18n'

// ─── Toggle Switch ───────────────────────────────────────────────────
function ToggleSwitch({ checked, onChange, label }: { checked: boolean; onChange: (v: boolean) => void; label?: string }) {
  return (
    <label className="relative inline-flex items-center cursor-pointer">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} className="sr-only peer" />
      <div className="w-11 h-6 rounded-full bg-[var(--border-light)] peer-focus:outline-none peer-checked:bg-[var(--accent-purple)] after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full"></div>
    </label>
  )
}

// ─── Settings Card ───────────────────────────────────────────────────
function SettingsCard({ title, icon, children }: { title: string; icon: string; children: React.ReactNode }) {
  return (
    <div className="glass-card p-6 page-enter">
      <div className="flex items-center gap-3 mb-5">
        <div className="w-8 h-8 rounded-lg bg-[rgba(139,92,246,0.1)] border border-[rgba(139,92,246,0.1)] flex items-center justify-center shrink-0">
          <span className="text-sm">{icon}</span>
        </div>
        <h3 className="text-base font-bold text-[var(--text-primary)]">{title}</h3>
      </div>
      <div className="space-y-4">
        {children}
      </div>
    </div>
  )
}

// ─── Settings Row ────────────────────────────────────────────────────
function SettingsRow({ label, desc, children }: { label: string; desc?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-[var(--border-light)] last:border-0">
      <div className="flex-1 min-w-0 pr-4">
        <p className="text-sm font-medium text-[var(--text-primary)]">{label}</p>
        {desc && <p className="text-xs text-[var(--text-muted)] mt-0.5">{desc}</p>}
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
}

// ─── Select Field ────────────────────────────────────────────────────
function SelectField({ value, onChange, options, label }: { value: string; onChange: (v: string) => void; options: { value: string; label: string }[]; label: string }) {
  return (
    <div>
      <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">{label}</label>
      <select value={value} onChange={(e) => onChange(e.target.value)} className="select-modern w-full text-sm">
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  )
}

// ─── Settings Page ───────────────────────────────────────────────────
export function SettingsPage() {
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    apiClient.get('/api/v1/settings')
      .then(r => setSettings(r.data.settings || {}))
      .catch(() => {})
  }, [])

  const updateSetting = (key: string, value: string) => {
    setSettings(prev => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await apiClient.put('/api/v1/settings', { settings })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (err) {
      console.error('Failed to save settings:', err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6 page-enter">
      {/* ── Header ──────────────────────────────────────────────── */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Settings</h1>
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">Panel configuration & preferences</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* General Settings */}
        <SettingsCard title="General" icon="⚙️">
          <SelectField
            label="Panel Language"
            value={settings['panel_language'] || 'en'}
            onChange={(v) => {
              updateSetting('panel_language', v)
              useI18nStore.getState().setLocale(v as 'en' | 'fa')
            }}
            options={[{ value: 'en', label: 'English' }, { value: 'fa', label: 'Persian (Farsi)' }]}
          />
          <SelectField
            label="Default Core"
            value={settings['default_core'] || 'xray'}
            onChange={(v) => updateSetting('default_core', v)}
            options={[{ value: 'xray', label: 'Xray-core' }, { value: 'singbox', label: 'Sing-box' }]}
          />
          <SelectField
            label="Log Level"
            value={settings['log_level'] || 'warn'}
            onChange={(v) => updateSetting('log_level', v)}
            options={[
              { value: 'debug', label: 'Debug' },
              { value: 'info', label: 'Info' },
              { value: 'warn', label: 'Warning' },
              { value: 'error', label: 'Error' },
            ]}
          />
        </SettingsCard>

        {/* Subscription Settings */}
        <SettingsCard title="Subscription" icon="🔗">
          <div>
            <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Subscription Path</label>
            <input
              type="text"
              value={settings['sub_path'] || '/sub'}
              onChange={(e) => updateSetting('sub_path', e.target.value)}
              className="input-modern text-sm"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Subscription Port</label>
            <input
              type="number"
              value={settings['sub_port'] || '2096'}
              onChange={(e) => updateSetting('sub_port', e.target.value)}
              className="input-modern text-sm"
            />
          </div>
          <SettingsRow label="Enable Subscription" desc="Allow clients to subscribe">
            <ToggleSwitch
              checked={settings['sub_enable'] !== 'false'}
              onChange={(v) => updateSetting('sub_enable', v ? 'true' : 'false')}
            />
          </SettingsRow>
        </SettingsCard>

        {/* Security Settings */}
        <SettingsCard title="Security" icon="🔒">
          <SettingsRow label="Two-Factor Auth (TOTP)" desc="Extra layer of login security">
            <span className={`badge ${settings['totp_enabled'] === 'true' ? 'badge-success' : 'badge-warning'}`}>
              {settings['totp_enabled'] === 'true' ? 'Enabled' : 'Disabled'}
            </span>
          </SettingsRow>
          <SettingsRow label="Auto-restart on Crash" desc="Automatically restart core on failure">
            <ToggleSwitch
              checked={settings['auto_restart_core'] !== 'false'}
              onChange={(v) => updateSetting('auto_restart_core', v ? 'true' : 'false')}
            />
          </SettingsRow>
        </SettingsCard>

        {/* Tunnel Monitor */}
        <SettingsCard title="Tunnel Monitor" icon="📡">
          <SettingsRow label="Enable Tunnel Health Monitor" desc="Monitor tunnel connectivity and restart core on failure">
            <ToggleSwitch
              checked={settings['tunnel_monitor_enabled'] === 'true'}
              onChange={(v) => updateSetting('tunnel_monitor_enabled', v ? 'true' : 'false')}
            />
          </SettingsRow>
          <div>
            <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Tunnel Monitor URL</label>
            <input
              type="text"
              value={settings['tunnel_monitor_url'] || ''}
              onChange={(e) => updateSetting('tunnel_monitor_url', e.target.value)}
              placeholder="https://example.com/check"
              className="input-modern text-sm"
            />
          </div>
          <p className="text-xs text-[var(--text-muted)] italic">
            Automatically monitor tunnel connectivity and restart the core on failure.
          </p>
        </SettingsCard>
      </div>

      {/* Save Button */}
      <div className="flex items-center justify-end gap-3 animate-fade-in">
        {saved && (
          <span className="text-[var(--success)] text-sm font-medium flex items-center gap-1.5">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
            Saved!
          </span>
        )}
        <button onClick={handleSave} disabled={saving} className="btn-primary text-sm">
          {saving ? (
            <><span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" /> Saving...</>
          ) : (
            <><svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" /></svg> Save Settings</>
          )}
        </button>
      </div>
    </div>
  )
}
