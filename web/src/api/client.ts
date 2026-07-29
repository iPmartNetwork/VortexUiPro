import axios from 'axios'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
})

// Request interceptor to add auth token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor for 401 handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('username')
      localStorage.removeItem('role')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API endpoints
export const api = {
  // Auth
  login: (username: string, password: string) =>
    apiClient.post('/api/v1/login', { username, password }),

  getMe: () => apiClient.get('/api/v1/me'),

  // Users
  getUsers: () => apiClient.get('/api/v1/admin/users'),
  getUser: (id: number) => apiClient.get(`/api/v1/admin/users/${id}`),
  createUser: (data: { username: string; email?: string; data_limit?: number }) =>
    apiClient.post('/api/v1/admin/users', data),
  deleteUser: (id: number) => apiClient.delete(`/api/v1/admin/users/${id}`),
  
  getClients: (userId: number) =>
    apiClient.get(`/api/v1/admin/users/${userId}/clients`),
  addClient: (userId: number, data: { inbound_id: number; email: string }) =>
    apiClient.post(`/api/v1/admin/users/${userId}/clients`, data),

  // Inbounds
  getInbounds: () => apiClient.get('/api/v1/inbounds'),
  getInbound: (id: number) => apiClient.get(`/api/v1/inbounds/${id}`),
  createInbound: (data: any) => apiClient.post('/api/v1/inbounds', data),
  updateInbound: (id: number, data: any) => apiClient.put(`/api/v1/inbounds/${id}`, data),
  deleteInbound: (id: number) => apiClient.delete(`/api/v1/inbounds/${id}`),

  // Admin Roles (RBAC)
  getRoles: () => apiClient.get('/api/v1/roles'),
  getRole: (id: number) => apiClient.get(`/api/v1/roles/${id}`),
  createRole: (data: any) => apiClient.post('/api/v1/roles', data),
  updateRole: (id: number, data: any) => apiClient.put(`/api/v1/roles/${id}`, data),
  duplicateRole: (id: number) => apiClient.post(`/api/v1/roles/${id}/duplicate`),
  deleteRole: (id: number) => apiClient.delete(`/api/v1/roles/${id}`),

  // Admins Management
  getAdmins: () => apiClient.get('/api/v1/admins'),
  getAdmin: (id: number) => apiClient.get(`/api/v1/admins/${id}`),
  createAdmin: (data: any) => apiClient.post('/api/v1/admins', data),
  updateAdmin: (id: number, data: any) => apiClient.put(`/api/v1/admins/${id}`, data),
  deleteAdmin: (id: number) => apiClient.delete(`/api/v1/admins/${id}`),
  setAdminStatus: (id: number, enabled: boolean) => apiClient.put(`/api/v1/admins/${id}/status`, { enabled }),

  // API Tokens
  getApiTokens: () => apiClient.get('/api/v1/api-tokens'),
  createApiToken: (data: any) => apiClient.post('/api/v1/api-tokens', data),
  deleteApiToken: (id: number) => apiClient.delete(`/api/v1/api-tokens/${id}`),
  setApiTokenStatus: (id: number, enabled: boolean) => apiClient.put(`/api/v1/api-tokens/${id}/status`, { enabled }),

  // System
  getSystemStatus: () => apiClient.get('/api/v1/system/status'),
  getPerformance: () => apiClient.get('/api/v1/system/performance'),
  getCoreStatus: () => apiClient.get('/api/v1/system/core-status'),

  // Subscription
  getSubscriptionConfig: (clientId: string, format?: string) =>
    apiClient.get(`/sub/${clientId}`, { params: { format } }),
  getSubscriptionInfo: (clientId: string) =>
    apiClient.get(`/sub/${clientId}/info`),
}

// Generic CRUD helpers
export const apiGet = <T = any>(url: string, config?: any) =>
  apiClient.get<T>(url, config)

export const apiPost = <T = any>(url: string, data?: any, config?: any) =>
  apiClient.post<T>(url, data, config)

export const apiPut = <T = any>(url: string, data?: any, config?: any) =>
  apiClient.put<T>(url, data, config)

export const apiDelete = <T = any>(url: string, config?: any) =>
  apiClient.delete<T>(url, config)

export default api
