import { create } from 'zustand'

type Theme = 'dark' | 'light' | 'system'

interface ThemeState {
  theme: Theme
  resolved: 'dark' | 'light'
  isDark: boolean
  isLight: boolean
  toggle: () => void
  setTheme: (t: Theme) => void
  reset: () => void
}

const getInitialTheme = (): Theme => {
  const stored = localStorage.getItem('vortex-theme') as Theme | null
  if (stored === 'light' || stored === 'dark' || stored === 'system') return stored
  return 'dark'
}

const getSystemTheme = (): 'dark' | 'light' => {
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
  }
  return 'dark'
}

const resolveTheme = (theme: Theme): 'dark' | 'light' => {
  if (theme === 'system') return getSystemTheme()
  return theme
}

const applyTheme = (resolved: 'dark' | 'light') => {
  const root = document.documentElement
  if (resolved === 'dark') {
    root.classList.remove('light')
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
    root.classList.add('light')
  }
}

let mediaQuery: MediaQueryList | null = null
let mediaListener: (() => void) | null = null

const setupSystemListener = (setTheme: (t: Theme) => void) => {
  if (typeof window === 'undefined' || !window.matchMedia) return
  if (mediaQuery) {
    mediaQuery.removeEventListener('change', mediaListener!)
  }
  mediaQuery = window.matchMedia('(prefers-color-scheme: light)')
  mediaListener = () => {
    const stored = localStorage.getItem('vortex-theme') as Theme | null
    if (stored === 'system') {
      const resolved = getSystemTheme()
      applyTheme(resolved)
      // Update the store without triggering re-render loops
      const state = useThemeStore.getState()
      useThemeStore.setState({
        resolved,
        isDark: resolved === 'dark',
        isLight: resolved === 'light',
      })
    }
  }
  mediaQuery.addEventListener('change', mediaListener)
}

// Apply on init
const initial = getInitialTheme()
const initialResolved = resolveTheme(initial)
applyTheme(initialResolved)

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: initial,
  resolved: initialResolved,
  isDark: initialResolved === 'dark',
  isLight: initialResolved === 'light',

  toggle: () => {
    const current = get().resolved
    const next: Theme = current === 'dark' ? 'light' : 'dark'
    const resolved = resolveTheme(next)
    applyTheme(resolved)
    localStorage.setItem('vortex-theme', next)
    set({ theme: next, resolved, isDark: resolved === 'dark', isLight: resolved === 'light' })
  },

  setTheme: (t: Theme) => {
    const resolved = resolveTheme(t)
    applyTheme(resolved)
    localStorage.setItem('vortex-theme', t)
    set({ theme: t, resolved, isDark: resolved === 'dark', isLight: resolved === 'light' })
    if (t === 'system') {
      setupSystemListener(() => {})
    }
  },

  reset: () => {
    localStorage.removeItem('vortex-theme')
    const resolved = getSystemTheme()
    applyTheme(resolved)
    set({ theme: 'system', resolved, isDark: resolved === 'dark', isLight: resolved === 'light' })
    setupSystemListener(() => {})
  },
}))

// Initial setup for system listener
if (initial === 'system') {
  setupSystemListener((t: Theme) => {})
}
