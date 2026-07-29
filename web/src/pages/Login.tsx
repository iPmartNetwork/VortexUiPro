import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../hooks/useAuth'

// ─── Floating particles for background ──────────────────────────────
function ParticleField() {
  const particles = Array.from({ length: 20 }, (_, i) => ({
    id: i,
    x: Math.random() * 100,
    y: Math.random() * 100,
    size: Math.random() * 3 + 1,
    speed: Math.random() * 30 + 20,
    delay: Math.random() * 10,
    color: ['rgba(139,92,246,0.3)', 'rgba(6,182,212,0.2)', 'rgba(16,185,129,0.15)'][i % 3],
  }))

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {particles.map((p) => (
        <div
          key={p.id}
          className="absolute rounded-full"
          style={{
            left: `${p.x}%`,
            top: `${p.y}%`,
            width: `${p.size}px`,
            height: `${p.size}px`,
            background: p.color,
            boxShadow: `0 0 ${p.size * 4}px ${p.color}`,
            animation: `particleDrift ${p.speed}s linear ${p.delay}s infinite`,
          }}
        />
      ))}
      <style>{`
        @keyframes particleDrift {
          0% { transform: translateY(0) translateX(0); opacity: 0; }
          10% { opacity: 0.6; }
          90% { opacity: 0.6; }
          100% { transform: translateY(-100vh) translateX(${Math.random() * 100 - 50}px); opacity: 0; }
        }
      `}</style>
    </div>
  )
}

// ─── Animated gradient border ───────────────────────────────────────
function GradientBorder() {
  return (
    <div
      className="absolute inset-0 rounded-[inherit]"
      style={{
        padding: '1px',
        background: 'linear-gradient(135deg, rgba(139,92,246,0.3), rgba(6,182,212,0.2), rgba(139,92,246,0.1), rgba(6,182,212,0.3))',
        backgroundSize: '300% 300%',
        WebkitMask: 'linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)',
        WebkitMaskComposite: 'xor',
        maskComposite: 'exclude',
        animation: 'borderShift 4s ease-in-out infinite',
      }}
    >
      <style>{`
        @keyframes borderShift {
          0% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
          100% { background-position: 0% 50%; }
        }
      `}</style>
    </div>
  )
}

// ─── Floating Shapes ────────────────────────────────────────────────
function FloatingShapes() {
  const shapes = [
    { type: 'circle', size: 60, top: '15%', left: '5%', delay: '0s', color: 'rgba(139,92,246,0.06)' },
    { type: 'square', size: 40, top: '25%', right: '10%', delay: '2s', color: 'rgba(6,182,212,0.05)' },
    { type: 'circle', size: 80, bottom: '20%', left: '10%', delay: '4s', color: 'rgba(16,185,129,0.04)' },
    { type: 'square', size: 50, bottom: '30%', right: '5%', delay: '1s', color: 'rgba(236,72,153,0.05)' },
    { type: 'circle', size: 100, top: '50%', left: '50%', delay: '3s', color: 'rgba(139,92,246,0.03)' },
  ]

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {shapes.map((s, i) => (
        <div
          key={i}
          className="absolute"
          style={{
            width: s.size,
            height: s.size,
            top: s.top,
            left: s.left,
            right: s.right,
            bottom: s.bottom,
            background: s.color,
            borderRadius: s.type === 'circle' ? '50%' : '20%',
            border: '1px solid rgba(255,255,255,0.03)',
            animation: `floatShape ${6 + i * 2}s ease-in-out infinite`,
            animationDelay: s.delay,
          }}
        />
      ))}
      <style>{`
        @keyframes floatShape {
          0%, 100% { transform: translateY(0) rotate(0deg); }
          25% { transform: translateY(-15px) rotate(5deg); }
          50% { transform: translateY(5px) rotate(-3deg); }
          75% { transform: translateY(-8px) rotate(2deg); }
        }
      `}</style>
    </div>
  )
}

// ─── Login Page ──────────────────────────────────────────────────────
export function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [mounted, setMounted] = useState(false)
  const login = useAuthStore((s) => s.login)
  const navigate = useNavigate()

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(username, password)
      navigate('/dashboard')
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Invalid credentials. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative overflow-hidden bg-[#08080f]">
      {/* ── Animated Background Layers ──────────────────────────── */}
      <div className="absolute inset-0">
        <div className="absolute inset-0" style={{
          background: `
            radial-gradient(ellipse at 20% 50%, rgba(139, 92, 246, 0.08) 0%, transparent 60%),
            radial-gradient(ellipse at 80% 20%, rgba(6, 182, 212, 0.05) 0%, transparent 60%),
            radial-gradient(ellipse at 50% 80%, rgba(16, 185, 129, 0.03) 0%, transparent 60%)
          `,
        }} />
        <div className="absolute inset-0 opacity-[0.03]" style={{
          backgroundImage: 'linear-gradient(rgba(139, 92, 246, 0.5) 1px, transparent 1px), linear-gradient(90deg, rgba(139, 92, 246, 0.5) 1px, transparent 1px)',
          backgroundSize: '60px 60px',
        }} />
      </div>

      <FloatingShapes />
      <ParticleField />

      {/* ── Login Card ──────────────────────────────────────────── */}
      <div
        className="w-full max-w-md relative z-10"
        style={{
          opacity: mounted ? 1 : 0,
          transform: mounted ? 'translateY(0)' : 'translateY(20px)',
          transition: 'all 0.6s cubic-bezier(0.22, 1, 0.36, 1)',
        }}
      >
        <div className="relative">
          <GradientBorder />
          <div className="relative bg-[rgba(13,13,26,0.6)] backdrop-blur-[40px] saturate-[1.5] rounded-[20px] p-[40px] shadow-[0_0_60px_rgba(139,92,246,0.08),0_8px_32px_rgba(0,0,0,0.3)]">
            {/* ── Logo ──────────────────────────────────────────── */}
            <div className="text-center mb-8">
              <div className="w-[68px] h-[68px] mx-auto mb-5 rounded-[18px] bg-gradient-to-br from-purple-500 via-purple-600 to-cyan-500 flex items-center justify-center shadow-[0_8px_32px_rgba(139,92,246,0.3)] relative overflow-hidden group cursor-pointer transition-transform duration-300 hover:scale-105">
                <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent" />
                <svg className="w-8 h-8 text-white relative z-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                <div className="absolute inset-0 rounded-[18px] ring-1 ring-inset ring-white/10" />
              </div>

              <h1 className="text-[28px] font-bold text-white mb-1.5 tracking-tight">
                Vortex<span className="bg-gradient-to-r from-purple-400 to-cyan-400 bg-clip-text text-transparent">UiPro</span>
              </h1>
              <p className="text-[#9898b8] text-sm font-medium">Ultimate Proxy Management Panel</p>
            </div>

            {/* ── Form ──────────────────────────────────────────── */}
            <form onSubmit={handleSubmit} className="space-y-5">
              {/* Username */}
              <div>
                <label className="block text-sm font-medium text-[#c0c0d8] mb-2">Username</label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg className="w-4 h-4 text-[#585878] group-focus-within:text-purple-400 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </div>
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="w-full pl-11 pr-4 py-3 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.04)] rounded-xl text-white placeholder-[#585878] focus:border-purple-500/40 focus:outline-none focus:bg-[rgba(139,92,246,0.04)] focus:shadow-[0_0_0_3px_rgba(139,92,246,0.08)] transition-all duration-300 text-sm"
                    placeholder="Enter your username"
                    autoComplete="username"
                    autoFocus
                    required
                  />
                </div>
              </div>

              {/* Password */}
              <div>
                <label className="block text-sm font-medium text-[#c0c0d8] mb-2">Password</label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <svg className="w-4 h-4 text-[#585878] group-focus-within:text-purple-400 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                  </div>
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full pl-11 pr-11 py-3 bg-[rgba(255,255,255,0.03)] border border-[rgba(255,255,255,0.04)] rounded-xl text-white placeholder-[#585878] focus:border-purple-500/40 focus:outline-none focus:bg-[rgba(139,92,246,0.04)] focus:shadow-[0_0_0_3px_rgba(139,92,246,0.08)] transition-all duration-300 text-sm"
                    placeholder="Enter your password"
                    autoComplete="current-password"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute inset-y-0 right-0 pr-4 flex items-center text-[#585878] hover:text-[#9898b8] transition-colors"
                    tabIndex={-1}
                  >
                    {showPassword ? (
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* Error */}
              {error && (
                <div className="p-3.5 bg-red-500/8 border border-red-500/15 rounded-xl text-red-400 text-sm flex items-start gap-3 animate-fade-in">
                  <svg className="w-5 h-5 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{error}</span>
                </div>
              )}

              {/* Submit */}
              <button
                type="submit"
                disabled={loading}
                className="relative w-full py-3.5 rounded-xl font-semibold text-white text-sm overflow-hidden group transition-all duration-300 disabled:opacity-60 disabled:cursor-not-allowed"
                style={{
                  background: 'linear-gradient(135deg, #8b5cf6, #06b6d4)',
                  boxShadow: '0 4px 20px rgba(139, 92, 246, 0.3)',
                }}
                onMouseEnter={(e) => {
                  if (!loading) {
                    e.currentTarget.style.boxShadow = '0 8px 30px rgba(139, 92, 246, 0.4)'
                    e.currentTarget.style.transform = 'translateY(-1px)'
                  }
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.boxShadow = '0 4px 20px rgba(139, 92, 246, 0.3)'
                  e.currentTarget.style.transform = 'translateY(0)'
                }}
              >
                <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/10 to-white/0 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                {loading ? (
                  <span className="flex items-center justify-center gap-2.5">
                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    <span>Signing in...</span>
                  </span>
                ) : (
                  <span className="flex items-center justify-center gap-2.5">
                    <span>Sign In</span>
                    <svg className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                    </svg>
                  </span>
                )}
              </button>
            </form>

            {/* ── Footer ────────────────────────────────────────── */}
            <div className="mt-8 pt-6 border-t border-[rgba(139,92,246,0.08)]">
              <div className="flex items-center justify-center gap-2 text-xs text-[#585878]">
                <div className="w-1 h-1 rounded-full bg-purple-500/50" />
                <span>Default: admin / admin123</span>
                <div className="w-1 h-1 rounded-full bg-cyan-500/50" />
              </div>
              <p className="text-center mt-3 text-[11px] text-[#404060] font-mono">
                v0.0.1 · Built with precision
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
