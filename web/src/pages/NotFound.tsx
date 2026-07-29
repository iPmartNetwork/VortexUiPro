import { useNavigate } from 'react-router-dom'

export function NotFoundPage() {
  const navigate = useNavigate()
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6 page-enter">
      <div className="text-7xl mb-2">🌀</div>
      <h1 className="text-4xl font-bold text-[var(--text-primary)]">404</h1>
      <p className="text-[var(--text-secondary)] text-lg">Page not found</p>
      <p className="text-[var(--text-muted)] text-sm max-w-md text-center">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <button onClick={() => navigate('/dashboard')} className="btn-primary mt-2">
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" /></svg>
        Back to Dashboard
      </button>
    </div>
  )
}
