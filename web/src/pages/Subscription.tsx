import { useState } from 'react'
import api from '../api/client'

export function SubscriptionPage() {
  const [clientId, setClientId] = useState('')
  const [config, setConfig] = useState('')
  const [format, setFormat] = useState('xray')
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)

  const fetchConfig = async () => {
    if (!clientId) return
    setLoading(true)
    try {
      const res = await api.getSubscriptionConfig(clientId, format)
      setConfig(res.data)
    } catch (err) {
      setConfig('Error: Invalid subscription or client not found')
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = () => {
    navigator.clipboard.writeText(config)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-6 fade-in">
      <div>
        <h1 className="text-xl font-bold text-[var(--text-primary)]">Subscription</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Generate subscription configurations for clients</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Controls */}
        <div className="glass-card p-6 space-y-4">
          <h3 className="text-base font-bold text-[var(--text-primary)]">Configuration</h3>
          
          <div>
            <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Client ID</label>
            <input
              type="text"
              value={clientId}
              onChange={(e) => setClientId(e.target.value)}
              placeholder="Enter client UUID"
              className="input-modern text-sm"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Format</label>
            <select
              value={format}
              onChange={(e) => setFormat(e.target.value)}
              className="select-modern w-full text-sm"
            >
              <option value="xray">Xray JSON</option>
              <option value="clash">Clash / Mihomo</option>
              <option value="singbox">Sing-box</option>
            </select>
          </div>

          <button
            onClick={fetchConfig}
            disabled={loading || !clientId}
            className="btn-primary w-full justify-center text-sm"
          >
            {loading ? 'Generating...' : 'Generate Config'}
          </button>

          <div className="border-t border-[var(--border-light)] pt-4">
            <h4 className="text-sm font-medium text-[var(--text-primary)] mb-2">Quick Links</h4>
            <div className="space-y-2 text-sm">
              <a href="#" className="block text-[var(--accent-light)] hover:text-[var(--accent-purple)] transition text-xs">Subscription Docs</a>
              <a href="#" className="block text-[var(--accent-light)] hover:text-[var(--accent-purple)] transition text-xs">API Reference</a>
            </div>
          </div>
        </div>

        {/* Config Output */}
        <div className="lg:col-span-2">
          <div className="glass-card">
            <div className="flex items-center justify-between p-4 border-b border-[var(--border-light)]">
              <h3 className="text-base font-bold text-[var(--text-primary)]">Generated Config</h3>
              {config && (
                <button
                  onClick={copyToClipboard}
                  className="btn-secondary text-xs py-1.5 px-3 gap-1.5"
                >
                  {copied ? (
                    <>✅ Copied!</>
                  ) : (
                    <>
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                      Copy
                    </>
                  )}
                </button>
              )}
            </div>
            <pre className="p-4 text-sm text-[var(--text-secondary)] font-mono overflow-x-auto max-h-96 overflow-y-auto">
              {config || <span className="text-[var(--text-muted)]">Enter a client ID and generate a config...</span>}
            </pre>
          </div>
        </div>
      </div>
    </div>
  )
}
