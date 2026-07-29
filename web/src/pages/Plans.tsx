import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import { formatBytes } from '../utils/format'

interface Plan {
  id: number
  name: string
  description: string
  price: number
  data_limit: number
  speed_limit: number
  device_limit: number
  duration: number
  protocol: string
  enabled: boolean
}

interface Order {
  id: number
  user_id: number
  plan_id: number
  amount: number
  status: string
  created_at: number
}

export function PlansPage() {
  const [plans, setPlans] = useState<Plan[]>([])
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreatePlan, setShowCreatePlan] = useState(false)
  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null)
  const [showPayment, setShowPayment] = useState(false)
  const [paymentMethod, setPaymentMethod] = useState<'zarinpal' | 'nowpayments'>('zarinpal')

  const [newPlan, setNewPlan] = useState({
    name: '', description: '', price: 0, data_limit: 0,
    speed_limit: 0, device_limit: 0, duration: 30, protocol: 'all',
  })

  useEffect(() => {
    Promise.all([
      apiClient.get('/api/v1/plans'),
      apiClient.get('/api/v1/orders'),
    ]).then(([p, o]) => {
      setPlans(p.data.plans || [])
      setOrders(o.data.orders || [])
    }).catch(() => {}).finally(() => setLoading(false))
  }, [])

  const handleCreatePlan = async () => {
    try {
      await apiClient.post('/api/v1/plans', newPlan)
      setShowCreatePlan(false)
      setNewPlan({ name: '', description: '', price: 0, data_limit: 0, speed_limit: 0, device_limit: 0, duration: 30, protocol: 'all' })
      const res = await apiClient.get('/api/v1/plans')
      setPlans(res.data.plans || [])
    } catch (err) { console.error('create plan error:', err) }
  }

  const handleDeletePlan = async (id: number) => {
    if (!confirm('Delete this plan?')) return
    try {
      await apiClient.delete(`/api/v1/plans/${id}`)
      setPlans(prev => prev.filter(p => p.id !== id))
    } catch (err) { console.error('delete plan error:', err) }
  }

  const handleBuyPlan = (plan: Plan) => {
    setSelectedPlan(plan)
    setShowPayment(true)
  }

  // Format price from cents
  const formatPrice = (cents: number) => `$${(cents / 100).toFixed(2)}`

  // Format duration
  const formatDuration = (d: number) => d >= 30 ? `${d / 30} month(s)` : `${d} day(s)`

  if (loading) {
    return <div className="flex items-center justify-center h-64"><div className="loading-spinner w-10 h-10" /></div>
  }

  return (
    <div className="space-y-6 fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--text-primary)]">Plans & Orders</h1>
          <p className="text-[var(--text-secondary)] mt-1">Manage service plans and view order history</p>
        </div>
        <button onClick={() => setShowCreatePlan(!showCreatePlan)}
          className="px-4 py-2 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 transition text-sm font-medium">
          + New Plan
        </button>
      </div>

      {/* Create Plan Form */}
      {showCreatePlan && (
        <div className="glass-card p-6 fade-in">
          <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-4">Create New Plan</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <input type="text" placeholder="Plan Name" value={newPlan.name}
              onChange={e => setNewPlan({...newPlan, name: e.target.value})}
              className="px-4 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none" />
            <input type="number" placeholder="Price (cents)" value={newPlan.price || ''}
              onChange={e => setNewPlan({...newPlan, price: Number(e.target.value)})}
              className="px-4 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none" />
            <input type="number" placeholder="Duration (days)" value={newPlan.duration}
              onChange={e => setNewPlan({...newPlan, duration: Number(e.target.value)})}
              className="px-4 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none" />
            <input type="number" placeholder="Data Limit (GB)" value={newPlan.data_limit ? newPlan.data_limit / 1073741824 : ''}
              onChange={e => setNewPlan({...newPlan, data_limit: Number(e.target.value) * 1073741824})}
              className="px-4 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none" />
            <input type="number" placeholder="Speed Limit (Mbps)" value={newPlan.speed_limit || ''}
              onChange={e => setNewPlan({...newPlan, speed_limit: Number(e.target.value)})}
              className="px-4 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none" />
            <input type="number" placeholder="Device Limit" value={newPlan.device_limit || ''}
              onChange={e => setNewPlan({...newPlan, device_limit: Number(e.target.value)})}
              className="px-4 py-2 bg-[var(--bg-elevated)] border border-[var(--border-light)] rounded-lg text-[var(--text-primary)] placeholder-[var(--text-muted)] focus:border-purple-500 focus:outline-none" />
          </div>
          <div className="mt-4 flex gap-2">
            <button onClick={handleCreatePlan} className="px-4 py-2 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 transition text-sm">Create</button>
            <button onClick={() => setShowCreatePlan(false)} className="px-4 py-2 bg-[var(--bg-surface)] text-[var(--text-secondary)] rounded-lg hover:bg-dark-600 transition text-sm">Cancel</button>
          </div>
        </div>
      )}

      {/* Plans Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {plans.length === 0 ? (
          <div className="lg:col-span-3 glass-card p-8 text-center">
            <div className="text-4xl mb-3">📦</div>
            <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-2">No Plans Yet</h3>
            <p className="text-[var(--text-secondary)] text-sm">Create your first service plan to start selling.</p>
          </div>
        ) : plans.map(plan => (
          <div key={plan.id} className={`glass-card p-6 hover:border-purple-500/30 transition-all group ${!plan.enabled ? 'opacity-50' : ''}`}>
            <div className="flex items-start justify-between mb-4">
              <div>
                <h3 className="text-lg font-bold text-[var(--text-primary)] group-hover:text-purple-300 transition-colors">{plan.name}</h3>
                {plan.description && <p className="text-[var(--text-muted)] text-sm mt-1">{plan.description}</p>}
              </div>
              {!plan.enabled && <span className="text-xs px-2 py-1 rounded bg-[var(--bg-surface)] text-[var(--text-muted)]">Disabled</span>}
            </div>
            <div className="text-3xl font-bold text-purple-400 mb-4">{formatPrice(plan.price)}</div>
            <div className="space-y-2 text-sm text-[var(--text-secondary)] mb-6">
              <div className="flex justify-between"><span>Data</span><span className="text-[var(--text-primary)]">{formatBytes(plan.data_limit)}</span></div>
              <div className="flex justify-between"><span>Speed</span><span className="text-[var(--text-primary)]">{plan.speed_limit > 0 ? `${plan.speed_limit} Mbps` : 'Unlimited'}</span></div>
              <div className="flex justify-between"><span>Devices</span><span className="text-[var(--text-primary)]">{plan.device_limit > 0 ? plan.device_limit : 'Unlimited'}</span></div>
              <div className="flex justify-between"><span>Duration</span><span className="text-[var(--text-primary)]">{formatDuration(plan.duration)}</span></div>
              <div className="flex justify-between"><span>Protocol</span><span className="text-[var(--text-primary)] capitalize">{plan.protocol}</span></div>
            </div>
            <div className="flex gap-2">
              <button onClick={() => handleBuyPlan(plan)}
                className="flex-1 py-2 bg-purple-600 text-[var(--text-primary)] rounded-lg hover:bg-purple-500 transition text-sm font-medium">
                Buy Now
              </button>
              <button onClick={() => handleDeletePlan(plan.id)}
                className="p-2 text-[var(--text-muted)] hover:text-red-400 hover:bg-[var(--bg-surface)] rounded-lg transition">
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Payment Modal */}
      {showPayment && selectedPlan && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 z-50" onClick={() => setShowPayment(false)}>
          <div className="glass-card p-6 max-w-md w-full fade-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-xl font-bold text-[var(--text-primary)] mb-4">Purchase {selectedPlan.name}</h3>
            <div className="text-3xl font-bold text-purple-400 mb-6 text-center">{formatPrice(selectedPlan.price)}</div>
            
            <div className="space-y-3 mb-6">
              <label className="flex items-center gap-3 p-3 rounded-lg bg-[var(--bg-elevated)]/50 cursor-pointer hover:bg-[var(--bg-elevated)] transition"
                onClick={() => setPaymentMethod('zarinpal')}>
                <input type="radio" checked={paymentMethod === 'zarinpal'} onChange={() => setPaymentMethod('zarinpal')} className="accent-purple-500" />
                <div>
                  <p className="text-[var(--text-primary)] text-sm font-medium">ZarinPal</p>
                  <p className="text-[var(--text-muted)] text-xs">Iranian payment gateway (IRR/IRT)</p>
                </div>
              </label>
              <label className="flex items-center gap-3 p-3 rounded-lg bg-[var(--bg-elevated)]/50 cursor-pointer hover:bg-[var(--bg-elevated)] transition"
                onClick={() => setPaymentMethod('nowpayments')}>
                <input type="radio" checked={paymentMethod === 'nowpayments'} onChange={() => setPaymentMethod('nowpayments')} className="accent-purple-500" />
                <div>
                  <p className="text-[var(--text-primary)] text-sm font-medium">NOWPayments</p>
                  <p className="text-[var(--text-muted)] text-xs">Cryptocurrency (BTC, USDT, TRX, etc.)</p>
                </div>
              </label>
            </div>

            <button className="w-full py-3 bg-gradient-to-r from-purple-600 to-indigo-600 text-[var(--text-primary)] rounded-lg font-medium hover:from-purple-500 hover:to-indigo-500 transition">
              Proceed to Payment
            </button>
            <button onClick={() => { setShowPayment(false); setSelectedPlan(null) }}
              className="w-full mt-2 py-2 text-[var(--text-muted)] hover:text-[var(--text-primary)] transition text-sm">
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Orders Table */}
      <div className="glass-card overflow-hidden">
        <div className="p-4 border-b border-[var(--border-light)]">
          <h3 className="text-lg font-semibold text-[var(--text-primary)]">Order History</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-[var(--border-light)]">
                <th className="text-left p-4 text-[var(--text-secondary)] text-sm font-medium">ID</th>
                <th className="text-left p-4 text-[var(--text-secondary)] text-sm font-medium">Amount</th>
                <th className="text-left p-4 text-[var(--text-secondary)] text-sm font-medium">Status</th>
                <th className="text-left p-4 text-[var(--text-secondary)] text-sm font-medium">Date</th>
              </tr>
            </thead>
            <tbody>
              {orders.length === 0 ? (
                <tr><td colSpan={4} className="p-8 text-center text-[var(--text-muted)]">No orders yet.</td></tr>
              ) : orders.map(order => (
                <tr key={order.id} className="border-b border-[var(--border-light)]/50 hover:bg-[var(--bg-elevated)]/30 transition">
                  <td className="p-4 text-[var(--text-primary)] text-sm font-mono">#{order.id}</td>
                  <td className="p-4 text-[var(--text-primary)] text-sm">{formatPrice(order.amount)}</td>
                  <td className="p-4">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      order.status === 'paid' ? 'bg-green-500/10 text-green-400' :
                      order.status === 'pending' ? 'bg-yellow-500/10 text-yellow-400' :
                      order.status === 'cancelled' ? 'bg-red-500/10 text-red-400' :
                      'bg-[var(--bg-surface)] text-[var(--text-secondary)]'
                    }`}>{order.status}</span>
                  </td>
                  <td className="p-4 text-[var(--text-secondary)] text-sm">{new Date(order.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
