import { useState } from 'react'
import { apiPost } from '../api/client'

export function TelegramSettingsPage() {
  const [tab, setTab] = useState<'link' | 'notify' | 'test'>('link')
  const [email, setEmail] = useState('')
  const [chatId, setChatId] = useState('')
  const [testChatId, setTestChatId] = useState('')
  const [notifyEmail, setNotifyEmail] = useState('')
  const [message, setMessage] = useState('')

  const linkClient = async () => {
    try {
      await apiPost('/api/v1/telegram/client/link', { email, chat_id: chatId })
      setMessage(`✅ Telegram linked to ${email}`)
      setEmail(''); setChatId('')
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`)
    }
  }

  const removeLink = async () => {
    try {
      await apiPost('/api/v1/telegram/client/link', { email, chat_id: '', remove: true })
      setMessage(`✅ Telegram unlinked from ${email}`)
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`)
    }
  }

  const sendTest = async () => {
    try {
      await apiPost('/api/v1/telegram/test', { chat_id: testChatId })
      setMessage('✅ Test notification sent! Check your Telegram.')
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`)
    }
  }

  const sendUsage = async () => {
    try {
      await apiPost('/api/v1/telegram/notify', { email: notifyEmail })
      setMessage(`✅ Usage notification sent to ${notifyEmail}`)
    } catch (err: any) {
      setMessage(`❌ ${err?.response?.data?.error || 'Failed'}`)
    }
  }

  const tabs = [
    { id: 'link' as const, label: 'Link Client', icon: '🔗' },
    { id: 'notify' as const, label: 'Send Report', icon: '📊' },
    { id: 'test' as const, label: 'Test', icon: '🧪' },
  ]

  return (
    <div className="space-y-6 page-enter">
      {/* Header */}
      <div className="glass-panel p-5">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <div>
            <h1 className="text-xl font-bold text-[var(--text-primary)]">Telegram Bot</h1>
            <p className="text-sm text-[var(--text-secondary)] mt-0.5">Client notifications via Telegram</p>
          </div>
        </div>
      </div>

      {/* Message */}
      {message && (
        <div className={`px-4 py-3 rounded-lg text-sm flex items-center justify-between ${
          message.includes('❌') ? 'bg-red-500/5 border border-red-500/20 text-red-400' : 'bg-green-500/5 border border-green-500/20 text-green-400'
        }`}>
          <span>{message}</span>
          <button onClick={() => setMessage('')} className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition">&times;</button>
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-2">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              tab === t.id
                ? 'bg-purple-600 text-white shadow-lg shadow-purple-600/20'
                : 'bg-[var(--bg-elevated)] text-[var(--text-secondary)] hover:bg-[var(--bg-surface)] hover:text-[var(--text-primary)] border border-[var(--border-light)]'
            }`}
          >
            <span>{t.icon}</span>
            <span>{t.label}</span>
          </button>
        ))}
      </div>

      {/* Link Client */}
      {tab === 'link' && (
        <div className="glass-card p-6">
          <h3 className="text-base font-bold text-[var(--text-primary)] mb-4">Link Telegram to Client</h3>
          <div className="space-y-3">
            <input
              placeholder="Client Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
            />
            <input
              placeholder="Telegram Chat ID"
              value={chatId}
              onChange={(e) => setChatId(e.target.value)}
              className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
            />
          </div>
          <div className="flex gap-2 mt-4">
            <button
              onClick={linkClient}
              disabled={!email || !chatId}
              className="px-5 py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium"
            >
              Link
            </button>
            <button
              onClick={removeLink}
              disabled={!email}
              className="px-5 py-2.5 bg-red-600/20 text-red-400 rounded-lg hover:bg-red-600/30 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium"
            >
              Remove Link
            </button>
          </div>
        </div>
      )}

      {/* Notify */}
      {tab === 'notify' && (
        <div className="glass-card p-6">
          <h3 className="text-base font-bold text-[var(--text-primary)] mb-2">Send Usage Report</h3>
          <p className="text-sm text-[var(--text-secondary)] mb-4">
            Send a traffic usage report to a client via Telegram. The client must have a linked Telegram chat.
          </p>
          <input
            placeholder="Client Email"
            value={notifyEmail}
            onChange={(e) => setNotifyEmail(e.target.value)}
            className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
          />
          <button
            onClick={sendUsage}
            disabled={!notifyEmail}
            className="mt-3 px-5 py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium"
          >
            Send Usage Report
          </button>
        </div>
      )}

      {/* Test */}
      {tab === 'test' && (
        <div className="glass-card p-6">
          <h3 className="text-base font-bold text-[var(--text-primary)] mb-2">Test Notification</h3>
          <p className="text-sm text-[var(--text-secondary)] mb-4">
            Send a test message to verify Telegram bot is working.
          </p>
          <input
            placeholder="Telegram Chat ID"
            value={testChatId}
            onChange={(e) => setTestChatId(e.target.value)}
            className="w-full px-4 py-2.5 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none text-sm"
          />
          <button
            onClick={sendTest}
            disabled={!testChatId}
            className="mt-3 px-5 py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium"
          >
            Send Test
          </button>
        </div>
      )}
    </div>
  )
}
