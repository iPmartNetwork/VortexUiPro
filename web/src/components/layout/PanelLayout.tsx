import { useState, useEffect } from 'react'
import { Outlet, NavLink, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../../hooks/useAuth'
import { useThemeStore } from '../../hooks/useTheme'
import { useI18nStore, useI18n, localeList } from '../../hooks/useI18n'

// ─── Navigation Items ────────────────────────────────────────────────
interface NavItem {
  label: string
  path: string
  icon: React.ReactNode
  badge?: { text: string; color: string }
}

const navItems: NavItem[] = [
  {
    label: 'Dashboard',
    path: '/dashboard',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" rx="1.5" />
        <rect x="14" y="3" width="7" height="7" rx="1.5" />
        <rect x="3" y="14" width="7" height="7" rx="1.5" />
        <rect x="14" y="14" width="7" height="7" rx="1.5" />
      </svg>
    ),
  },
  {
    label: 'Inbounds',
    path: '/inbounds',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
      </svg>
    ),
  },
  {
    label: 'Users',
    path: '/users',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
    ),
  },
  {
    label: 'Outbounds',
    path: '/outbounds',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
      </svg>
    ),
  },
  {
    label: 'Routing',
    path: '/routing',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
      </svg>
    ),
  },
  {
    label: 'Sub Profiles',
    path: '/sub-profiles',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
      </svg>
    ),
  },
  {
    label: 'Nodes',
    path: '/nodes',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
    ),
  },
  {
    label: 'Subscription',
    path: '/subscription',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
      </svg>
    ),
  },  { label: 'Domain Fronting', path: '/domain-fronting',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  { label: 'Xray API', path: '/xray-api',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
  },
  { label: 'Smart DNS', path: '/smart-dns',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M12 6v6l4 2" />
      </svg>
    ),
  },
  {
    label: 'Docker',
    path: '/docker',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
    ),
  },
  {
    label: 'Anti-Censorship',
    path: '/anticensor',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
  },
  {
    label: 'Portal',
    path: '/portal',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
      </svg>
    ),
  },
  {
    label: 'Roles',
    path: '/admin-roles',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
  },
  {
    label: 'Admins',
    path: '/admins',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
    ),
  },
  {
    label: 'API Tokens',
    path: '/api-tokens',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
      </svg>
    ),
  },
  {
    label: 'Online',
    path: '/online',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
    ),
  },
  {
    label: 'Traffic',
    path: '/traffic',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
      </svg>
    ),
  },
  {
    label: 'Health',
    path: '/health',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
  },
  {
    label: 'Resellers',
    path: '/resellers',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
      </svg>
    ),
  },
  {
    label: 'Telegram',
    path: '/telegram',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>
    ),
  },
  {
    label: 'Backup',
    path: '/backups',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M7 10l5 5 5-5M12 15V3" />
      </svg>
    ),
  },
  {
    label: 'Federation',
    path: '/federation',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
    ),
  },
  {
    label: 'Cluster',
    path: '/cluster',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
    ),
  },
  {
    label: 'Metrics',
    path: '/metrics',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
    ),
  },
  {
    label: 'Plans',
    path: '/plans',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    label: 'Security',
    path: '/security',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>
    ),
  },
  {
    label: 'Terminal',
    path: '/terminal',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
    ),
  },
  {
    label: 'Live Logs',
    path: '/logs',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
  },
  {
    label: 'WARP+',
    path: '/warp',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    label: 'TLS Tricks',
    path: '/tls-tricks',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
  },
  {
    label: 'Plugins',
    path: '/plugins',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
      </svg>
    ),
  },
  {
    label: 'WebRTC',
    path: '/webrtc',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    label: 'Topology',
    path: '/topology',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
    ),
  },
  {
    label: 'Client Groups',
    path: '/client-groups',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
      </svg>
    ),
  },
  {
    label: 'Settings',
    path: '/settings',
    icon: (
      <svg className="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
        <path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
    ),
  },
]

// ─── Grouped sections ────────────────────────────────────────────────
const navGroups = [
  { label: 'Main', items: ['/dashboard', '/inbounds', '/users', '/outbounds', '/nodes', '/routing'] },
  { label: 'Services', items: ['/subscription', '/sub-profiles', '/anticensor', '/portal', '/client-groups'] },
  { label: 'Admin', items: ['/admin-roles', '/admins', '/api-tokens', '/security'] },
  { label: 'Monitoring', items: ['/online', '/traffic', '/health'] },
  { label: 'Management', items: ['/resellers', '/telegram', '/backups'] },
  { label: 'Advanced', items: ['/terminal', '/logs', '/warp', '/tls-tricks', '/plugins', '/webrtc', '/topology', '/federation', '/cluster', '/metrics', '/plans', '/settings'] },
]

// ─── Sidebar Component ───────────────────────────────────────────────
function Sidebar({ collapsed, setCollapsed }: { collapsed: boolean; setCollapsed: (v: boolean) => void }) {
  const { username, logout } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()
  const { t } = useI18n()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  // Get current section label
  const currentSection = navGroups.find(g =>
    g.items.some(p => location.pathname === p || location.pathname.startsWith(p + '/'))
  )

  return (
    <aside
      className={`${
        collapsed ? 'w-[72px]' : 'w-[260px]'
      } h-screen flex flex-col relative z-20 transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)]`}
    >
      {/* Glass background */}
      <div className="absolute inset-0 bg-[rgba(8,8,15,0.85)] backdrop-blur-[30px] border-r border-[rgba(139,92,246,0.08)]" />
      <div className="absolute inset-0 bg-gradient-to-b from-[rgba(139,92,246,0.03)] to-transparent pointer-events-none" />

      <div className="relative flex flex-col h-full">
        {/* ── Logo ──────────────────────────────────────────────── */}
        <div className="h-[68px] flex items-center justify-center shrink-0 border-b border-[rgba(139,92,246,0.06)] relative">
          <div className={`flex items-center gap-3 ${collapsed ? 'justify-center w-full' : 'px-5 w-full'}`}>
            <div className="w-9 h-9 rounded-[10px] bg-gradient-to-br from-purple-500 via-purple-600 to-cyan-500 flex items-center justify-center shadow-[0_4px_16px_rgba(139,92,246,0.25)] shrink-0 relative overflow-hidden group cursor-pointer transition-transform duration-200 hover:scale-105">
              <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent" />
              <svg className="w-4.5 h-4.5 text-white relative z-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            {!collapsed && (
              <>
                <div className="flex-1 min-w-0">
                  <div className="text-[15px] font-bold text-white tracking-tight">
                    Vortex<span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-cyan-400">UiPro</span>
                  </div>
                  <p className="text-[10px] text-[#585878] font-medium mt-[-2px]">v0.0.1</p>
                </div>
                {/* Collapse button */}
                <button
                  onClick={() => setCollapsed(true)}
                  className="w-6 h-6 rounded-md bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] flex items-center justify-center text-[#585878] hover:text-white transition-all shrink-0"
                  title="Collapse sidebar"
                >
                  <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
                    <path d="M15 18l-6-6 6-6" />
                  </svg>
                </button>
              </>
            )}
            {collapsed && (
              <button
                onClick={() => setCollapsed(false)}
                className="absolute -right-3 top-1/2 -translate-y-1/2 w-6 h-6 rounded-full bg-[#1c1c30] border border-[rgba(139,92,246,0.12)] flex items-center justify-center text-[#6868a0] hover:text-white hover:border-purple-500/30 transition-all shadow-lg"
                title="Expand sidebar"
              >
                <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
                  <path d="M9 18l6-6-6-6" />
                </svg>
              </button>
            )}
          </div>
        </div>

        {/* ── Section label ───────────────────────────────────────── */}
        {!collapsed && currentSection && (
          <div className="px-5 pt-4 pb-1">
            <span className="text-[10px] font-semibold uppercase tracking-[0.12em] text-[#585878]">
              {currentSection.label}
            </span>
          </div>
        )}

        {/* ── Navigation ──────────────────────────────────────────── */}
        <nav className="flex-1 overflow-y-auto overflow-x-hidden px-3 py-3 space-y-0.5 scrollbar-thin">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path || location.pathname.startsWith(item.path + '/')
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={`sidebar-link group ${isActive ? 'sidebar-link-active' : ''}`}
                title={collapsed ? item.label : undefined}
              >
                <span className="shrink-0">{item.icon}</span>
                {!collapsed && (
                  <>                      <span className="flex-1 truncate">{resolveNavLabel(item.path, t) || item.label}</span>
                    {item.badge && (
                      <span className={`px-1.5 py-0.5 rounded text-[9px] font-bold uppercase ${item.badge.color}`}>
                        {item.badge.text}
                      </span>
                    )}
                    {isActive && (
                      <span className="w-1.5 h-1.5 rounded-full bg-purple-500 shadow-[0_0_6px_rgba(139,92,246,0.5)]" />
                    )}
                  </>
                )}
                {collapsed && isActive && (
                  <span className="absolute right-1.5 top-1/2 -translate-y-1/2 w-1.5 h-1.5 rounded-full bg-purple-500 shadow-[0_0_6px_rgba(139,92,246,0.5)]" />
                )}
              </NavLink>
            )
          })}
        </nav>

        {/* ── Theme Toggle ──────────────────────────────────────── */}
        {!collapsed && (
          <div className="px-3 py-1">
            <ThemeToggle />
          </div>
        )}
        {collapsed && (
          <div className="flex justify-center py-1">
            <ThemeToggleCollapsed />
          </div>
        )}

        {/* ── Locale Switcher ────────────────────────────────────── */}
        {!collapsed && (
          <div className="px-3 pb-1">
            <LocaleSwitcher />
          </div>
        )}

        {/* ── User Section ────────────────────────────────────────── */}
        <div className="shrink-0 p-3 border-t border-[rgba(139,92,246,0.06)] relative">
          <div className={`flex items-center gap-3 ${collapsed ? 'justify-center' : ''}`}>
            <div className="w-9 h-9 rounded-[10px] bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-xs font-bold shrink-0 shadow-[0_2px_8px_rgba(139,92,246,0.2)]">
              {username?.[0]?.toUpperCase() || 'A'}
            </div>
            {!collapsed && (
              <>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-white truncate">{username || 'Admin'}</p>
                  <p className="text-[10px] text-[#585878] font-medium">Administrator</p>
                </div>
                <button
                  onClick={handleLogout}
                  className="w-8 h-8 rounded-lg bg-[rgba(255,255,255,0.02)] hover:bg-red-500/10 flex items-center justify-center text-[#6868a0] hover:text-red-400 transition-all shrink-0"
                  title="Logout"
                >
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                  </svg>
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </aside>
  )
}

// ─── Mobile Topbar ────────────────────────────────────────────────────
// ─── Theme Toggle (expanded) ─────────────────────────────────────────
function ThemeToggle() {
  const { isDark, theme, toggle, setTheme, reset } = useThemeStore()
  return (
    <div className="space-y-1.5">
      <button
        onClick={toggle}
        className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl bg-[rgba(255,255,255,0.02)] hover:bg-[rgba(139,92,246,0.06)] border border-[rgba(139,92,246,0.06)] hover:border-[rgba(139,92,246,0.15)] transition-all duration-300 group"
        title={isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
      >
        <div className={`w-7 h-7 rounded-lg flex items-center justify-center transition-all duration-300 ${isDark ? 'bg-amber-500/10 text-amber-400' : 'bg-indigo-500/10 text-indigo-400'}`}>
          {isDark ? (
            <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <circle cx="12" cy="12" r="5" />
              <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
            </svg>
          ) : (
            <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
            </svg>
          )}
        </div>
        <div className="flex-1 text-left">
          <p className="text-xs font-medium text-[var(--text-secondary)]">{isDark ? 'Light Mode' : 'Dark Mode'}</p>
          <p className="text-[10px] text-[var(--text-muted)]">{theme === 'system' ? 'System' : 'Manual'}</p>
        </div>
        <div className={`w-4 h-4 rounded-full border-2 transition-all duration-300 flex items-center justify-center ${isDark ? 'border-amber-400' : 'border-indigo-400'}`}>
          <div className={`w-2 h-2 rounded-full transition-all duration-300 ${isDark ? 'bg-amber-400' : 'bg-indigo-400'}`} />
        </div>
      </button>
      <div className="flex gap-1">
        <button onClick={() => setTheme('dark')} className={`flex-1 py-1.5 rounded-lg text-[10px] font-medium transition-all ${theme === 'dark' ? 'bg-purple-500/20 text-purple-300 border border-purple-500/20' : 'bg-[rgba(255,255,255,0.02)] text-[var(--text-muted)] hover:text-[var(--text-secondary)] border border-transparent'}`}>Dark</button>
        <button onClick={() => setTheme('light')} className={`flex-1 py-1.5 rounded-lg text-[10px] font-medium transition-all ${theme === 'light' ? 'bg-purple-500/20 text-purple-300 border border-purple-500/20' : 'bg-[rgba(255,255,255,0.02)] text-[var(--text-muted)] hover:text-[var(--text-secondary)] border border-transparent'}`}>Light</button>
        <button onClick={reset} className={`flex-1 py-1.5 rounded-lg text-[10px] font-medium transition-all ${theme === 'system' ? 'bg-purple-500/20 text-purple-300 border border-purple-500/20' : 'bg-[rgba(255,255,255,0.02)] text-[var(--text-muted)] hover:text-[var(--text-secondary)] border border-transparent'}`} title="System preference">System</button>
      </div>
    </div>
  )
}

// ─── Theme Toggle (collapsed) ─────────────────────────────────────────
function ThemeToggleCollapsed() {
  const { isDark, toggle } = useThemeStore()
  return (
    <button
      onClick={toggle}
      className={`w-8 h-8 rounded-lg flex items-center justify-center transition-all duration-300 hover:scale-110 ${isDark ? 'bg-amber-500/10 text-amber-400 hover:bg-amber-500/20' : 'bg-indigo-500/10 text-indigo-400 hover:bg-indigo-500/20'}`}
      title={isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
    >
      {isDark ? (
        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <circle cx="12" cy="12" r="5" />
          <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
        </svg>
      ) : (
        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
        </svg>
      )}
    </button>
  )
}

// ─── Mobile Topbar ────────────────────────────────────────────────────
function MobileTopbar({ onToggle }: { onToggle: () => void }) {
  const { username } = useAuthStore()
  const { isDark, toggle } = useThemeStore()

  return (
    <div className="lg:hidden flex items-center justify-between h-14 px-4 bg-[rgba(8,8,15,0.9)] backdrop-blur-[20px] border-b border-[rgba(139,92,246,0.06)]">
      <button onClick={onToggle} className="w-9 h-9 rounded-lg bg-[rgba(255,255,255,0.03)] flex items-center justify-center text-[#9898b8] hover:text-white transition">
        <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <path d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
      <div className="flex items-center gap-2.5">
        <div className="text-sm font-semibold text-white">
          Vortex<span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-cyan-400">UiPro</span>
        </div>
        <button
          onClick={toggle}
          className={`w-7 h-7 rounded-md flex items-center justify-center transition ${isDark ? 'bg-amber-500/10 text-amber-400' : 'bg-indigo-500/10 text-indigo-400'}`}
        >
          {isDark ? (
            <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <circle cx="12" cy="12" r="5" />
              <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
            </svg>
          ) : (
            <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
            </svg>
          )}
        </button>
        <div className="w-7 h-7 rounded-md bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white text-[10px] font-bold">
          {username?.[0]?.toUpperCase() || 'A'}
        </div>
      </div>
    </div>
  )
}

// ─── i18n nav label helper ──────────────────────────────────────────
const navI18nKeys: Record<string, string> = {
  '/dashboard': 'nav.dashboard',
  '/inbounds': 'nav.inbounds',
  '/users': 'nav.users',
  '/outbounds': 'nav.outbounds',
  '/routing': 'nav.routing',
  '/sub-profiles': 'nav.subProfiles',
  '/nodes': 'nav.nodes',
  '/subscription': 'nav.subscription',
  '/anticensor': 'nav.anticensor',
  '/portal': 'nav.portal',
  '/admin-roles': 'nav.roles',
  '/admins': 'nav.admins',
  '/api-tokens': 'nav.apiTokens',
  '/online': 'nav.online',
  '/traffic': 'nav.traffic',
  '/health': 'nav.health',
  '/resellers': 'nav.resellers',
  '/telegram': 'nav.telegram',
  '/backups': 'nav.backup',
  '/federation': 'nav.federation',
  '/cluster': 'nav.cluster',
  '/metrics': 'nav.metrics',
  '/plans': 'nav.plans',
  '/settings': 'nav.settings',
  '/security': 'nav.security',
  '/client-groups': 'nav.clientGroups',
  '/terminal': 'nav.terminal',
  '/logs': 'nav.logs',
  '/warp': 'nav.warp',
  '/tls-tricks': 'nav.tlsTricks',
  '/plugins': 'nav.plugins',
  '/webrtc': 'nav.webrtc',
  '/topology': 'nav.topology',
  '/domain-fronting': 'nav.domainFronting',
  '/xray-api': 'nav.xrayAPI',
  '/smart-dns': 'nav.smartDNS',
  '/docker': 'nav.docker',
}

function resolveNavLabel(path: string, translator: (key: string) => string): string {
  const key = navI18nKeys[path]
  if (!key) return ''
  return translator(key)
}

// ─── Locale Switcher (multi-language) ───────────────────────────────
function LocaleSwitcher() {
  const { locale, setLocale } = useI18nStore()
  const [open, setOpen] = useState(false)
  const current = localeList.find(l => l.code === locale) || localeList[0]

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
        className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl bg-[rgba(255,255,255,0.02)] hover:bg-[rgba(139,92,246,0.06)] border border-[rgba(139,92,246,0.06)] hover:border-[rgba(139,92,246,0.15)] transition-all duration-300 group"
        title={`Language: ${current.name}`}
      >
        <div className="w-7 h-7 rounded-lg flex items-center justify-center bg-emerald-500/10 text-emerald-400">
          <span className="text-xs">{current.flag}</span>
        </div>
        <div className="flex-1 text-left">
          <p className="text-xs font-medium text-[var(--text-secondary)]">{current.nativeName}</p>
          <p className="text-[10px] text-[var(--text-muted)]">{current.dir === 'rtl' ? 'RTL' : 'LTR'}</p>
        </div>
        <svg className={`w-3 h-3 text-[#585878] transition-transform duration-200 ${open ? 'rotate-180' : ''}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>

      {open && (
        <div className="absolute bottom-full left-0 right-0 mb-1 rounded-xl bg-[#16162a] border border-[rgba(139,92,246,0.12)] shadow-2xl shadow-purple-500/5 overflow-hidden max-h-72 overflow-y-auto">
          {localeList.map((lang) => (
            <button
              key={lang.code}
              onClick={() => { setLocale(lang.code); setOpen(false) }}
              className={`w-full flex items-center gap-3 px-3 py-2.5 transition-all duration-150 hover:bg-[rgba(139,92,246,0.08)] ${locale === lang.code ? 'bg-[rgba(139,92,246,0.06)] text-white' : 'text-[#9898b8]'}`}
            >
              <span className="text-base">{lang.flag}</span>
              <div className="flex-1 text-left">
                <p className="text-xs font-medium">{lang.nativeName}</p>
                <p className="text-[9px] text-[#6868a0]">{lang.name} · {lang.dir.toUpperCase()}</p>
              </div>
              {locale === lang.code && (
                <svg className="w-3.5 h-3.5 text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
                  <path d="M5 13l4 4L19 7" />
                </svg>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── Panel Layout ─────────────────────────────────────────────────────
export function PanelLayout() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
    // Check localStorage for sidebar state
    const saved = localStorage.getItem('sidebarCollapsed')
    if (saved === 'true') setSidebarCollapsed(true)
  }, [])

  useEffect(() => {
    localStorage.setItem('sidebarCollapsed', String(sidebarCollapsed))
  }, [sidebarCollapsed])

  return (
    <div className="min-h-screen bg-[#08080f] flex">
      {/* ── Desktop Sidebar ────────────────────────────────────── */}
      <div className="hidden lg:flex">
        <Sidebar collapsed={sidebarCollapsed} setCollapsed={setSidebarCollapsed} />
      </div>

      {/* ── Mobile Sidebar Overlay ─────────────────────────────── */}
      {mobileSidebarOpen && (
        <div
          className="fixed inset-0 z-40 lg:hidden bg-black/50 backdrop-blur-sm"
          onClick={() => setMobileSidebarOpen(false)}
        >
          <div
            className="w-[260px] h-full"
            onClick={(e) => e.stopPropagation()}
            style={{ animation: 'slideInLeft 0.25s ease-out both' }}
          >
            <Sidebar collapsed={false} setCollapsed={() => {}} />
          </div>
        </div>
      )}

      {/* ── Main Content ──────────────────────────────────────────── */}
      <div className="flex-1 flex flex-col min-h-screen overflow-hidden">
        <MobileTopbar onToggle={() => setMobileSidebarOpen(true)} />

        <main
          className="flex-1 overflow-y-auto"
          style={{
            opacity: mounted ? 1 : 0,
            transition: 'opacity 0.3s ease',
          }}
        >
          {/* Animated background */}
          <div className="animated-bg fixed inset-0 pointer-events-none z-0" />

          <div className="relative z-10 p-4 sm:p-6 lg:p-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
