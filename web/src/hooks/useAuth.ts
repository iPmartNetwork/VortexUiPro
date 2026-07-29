import { create } from 'zustand'
import { apiClient } from '../api/client'

interface AuthState {
  token: string | null
  refreshToken: string | null
  username: string | null
  role: string | null
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  refresh: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),
  username: localStorage.getItem('username'),
  role: localStorage.getItem('role'),
  isAuthenticated: !!localStorage.getItem('access_token'),

  login: async (username: string, password: string) => {
    const response = await apiClient.post('/api/v1/login', { username, password })
    const data = response.data

    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    localStorage.setItem('username', username)
    localStorage.setItem('role', data.role || 'admin')

    set({
      token: data.access_token,
      refreshToken: data.refresh_token,
      username,
      role: data.role || 'admin',
      isAuthenticated: true,
    })
  },

  logout: () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    set({
      token: null,
      refreshToken: null,
      username: null,
      role: null,
      isAuthenticated: false,
    })
  },

  refresh: async () => {
    const refreshToken = get().refreshToken
    if (!refreshToken) {
      get().logout()
      return
    }
    try {
      const response = await apiClient.post('/api/v1/auth/refresh', {
        refresh_token: refreshToken,
      })
      const data = response.data
      localStorage.setItem('access_token', data.access_token)
      set({ token: data.access_token, isAuthenticated: true })
    } catch {
      get().logout()
    }
  },
}))
