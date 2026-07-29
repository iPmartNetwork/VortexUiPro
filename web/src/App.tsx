import { Suspense, lazy, useEffect } from 'react'
import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from './hooks/useAuth'
import { PanelLayout } from './components/layout/PanelLayout'

// ─── Static imports (critical path — loaded immediately) ────────────
import { LoginPage } from './pages/Login'

// ═══ Lazy-loaded pages (code-split into separate chunks) ═══════════
const DashboardPage = lazy(() => import('./pages/Dashboard').then(m => ({ default: m.DashboardPage })))
const InboundsPage = lazy(() => import('./pages/Inbounds').then(m => ({ default: m.InboundsPage })))
const UsersPage = lazy(() => import('./pages/Users').then(m => ({ default: m.UsersPage })))
const NodesPage = lazy(() => import('./pages/Nodes').then(m => ({ default: m.NodesPage })))
const SettingsPage = lazy(() => import('./pages/Settings').then(m => ({ default: m.SettingsPage })))
const SubscriptionPage = lazy(() => import('./pages/Subscription').then(m => ({ default: m.SubscriptionPage })))
const PlansPage = lazy(() => import('./pages/Plans').then(m => ({ default: m.PlansPage })))
const AntiCensorshipPage = lazy(() => import('./pages/AntiCensorship').then(m => ({ default: m.AntiCensorshipPage })))
const UserPortalPage = lazy(() => import('./pages/UserPortal').then(m => ({ default: m.UserPortalPage })))
const AdminRolesPage = lazy(() => import('./pages/AdminRoles').then(m => ({ default: m.AdminRolesPage })))
const AdminsPage = lazy(() => import('./pages/Admins').then(m => ({ default: m.AdminsPage })))
const ApiTokensPage = lazy(() => import('./pages/ApiTokens').then(m => ({ default: m.ApiTokensPage })))
const OnlineMonitoringPage = lazy(() => import('./pages/OnlineMonitoring').then(m => ({ default: m.OnlineMonitoringPage })))
const TrafficManagementPage = lazy(() => import('./pages/TrafficManagement').then(m => ({ default: m.TrafficManagementPage })))
const ResellerManagementPage = lazy(() => import('./pages/ResellerManagement').then(m => ({ default: m.ResellerManagementPage })))
const TelegramSettingsPage = lazy(() => import('./pages/TelegramSettings').then(m => ({ default: m.TelegramSettingsPage })))
const BackupManagerPage = lazy(() => import('./pages/BackupManager').then(m => ({ default: m.BackupManagerPage })))
const MetricsDashboardPage = lazy(() => import('./pages/MetricsDashboard').then(m => ({ default: m.MetricsDashboardPage })))
const ClusterManagementPage = lazy(() => import('./pages/ClusterManagement').then(m => ({ default: m.ClusterManagementPage })))
const OutboundsPage = lazy(() => import('./pages/OutboundsPage').then(m => ({ default: m.OutboundsPage })))
const RoutingPage = lazy(() => import('./pages/RoutingPage').then(m => ({ default: m.RoutingPage })))
const SubscriptionProfilesPage = lazy(() => import('./pages/SubscriptionProfilesPage').then(m => ({ default: m.SubscriptionProfilesPage })))
const SecurityPage = lazy(() => import('./pages/SecurityPage').then(m => ({ default: m.SecurityPage })))
const ClientGroupsPage = lazy(() => import('./pages/ClientGroupsPage').then(m => ({ default: m.ClientGroupsPage })))
const FederationPage = lazy(() => import('./pages/FederationPage').then(m => ({ default: m.FederationPage })))
const NotFoundPage = lazy(() => import('./pages/NotFound').then(m => ({ default: m.NotFoundPage })))
const TerminalPage = lazy(() => import('./pages/TerminalPage').then(m => ({ default: m.TerminalPage })))
const LogStreamPage = lazy(() => import('./pages/LogStreamPage').then(m => ({ default: m.LogStreamPage })))
const WarpPage = lazy(() => import('./pages/WarpPage').then(m => ({ default: m.WarpPage })))
const TlsTricksPage = lazy(() => import('./pages/TlsTricksPage').then(m => ({ default: m.TlsTricksPage })))
const PluginPage = lazy(() => import('./pages/PluginPage').then(m => ({ default: m.PluginPage })))
const WebRTCPage = lazy(() => import('./pages/WebRTCPage'))
const TopologyVizPage = lazy(() => import('./pages/TopologyVizPage'))
const DomainFrontingPage = lazy(() => import('./pages/DomainFrontingPage').then(m => ({ default: m.DomainFrontingPage })))
const SmartDNSPage = lazy(() => import('./pages/SmartDNSPage').then(m => ({ default: m.SmartDNSPage })))
const DockerPage = lazy(() => import('./pages/DockerPage').then(m => ({ default: m.DockerPage })))
const XrayAPIPage = lazy(() => import('./pages/XrayAPIPage').then(m => ({ default: m.XrayAPIPage })))
const HealthPage = lazy(() => import('./pages/HealthPage').then(m => ({ default: m.HealthPage })))

// ─── Loading Fallback ──────────────────────────────────────────────
function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div className="relative">
        <div className="w-10 h-10 border-2 border-violet-500/30 border-t-violet-500 rounded-full animate-spin" />
        <div className="w-6 h-6 border-2 border-cyan-500/30 border-b-cyan-500 rounded-full animate-spin absolute inset-0 m-auto"
          style={{ animationDirection: 'reverse', animationDuration: '0.8s' }} />
      </div>
    </div>
  )
}

// ─── Protected Route ───────────────────────────────────────────────
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

// ─── Preload common pages after initial render ────────────────────
function usePreloadPages() {
  useEffect(() => {
    const timer = setTimeout(() => {
      // Dynamically import common pages in background (no render blocking)
      import('./pages/Dashboard')
      import('./pages/Inbounds')
      import('./pages/Users')
      import('./pages/Settings')
    }, 1000) // Wait 1s after initial render
    return () => clearTimeout(timer)
  }, [])
}

// ─── App Component ─────────────────────────────────────────────────
export function App() {
  const location = useLocation()
  usePreloadPages()

  return (
    <Suspense fallback={<PageLoader />}>
      <Routes location={location} key={location.pathname}>
        <Route path="/login" element={
          <div className="animate-fade-in">
            <Suspense fallback={<PageLoader />}>
              <LoginPage />
            </Suspense>
          </div>
        } />

        {/* Protected routes with panel layout */}
        <Route path="/" element={
          <ProtectedRoute>
            <div className="animate-fade-in">
              <PanelLayout />
            </div>
          </ProtectedRoute>
        }>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Suspense fallback={<PageLoader />}><DashboardPage /></Suspense>} />
          <Route path="inbounds" element={<Suspense fallback={<PageLoader />}><InboundsPage /></Suspense>} />
          <Route path="users" element={<Suspense fallback={<PageLoader />}><UsersPage /></Suspense>} />
          <Route path="nodes" element={<Suspense fallback={<PageLoader />}><NodesPage /></Suspense>} />
          <Route path="subscription" element={<Suspense fallback={<PageLoader />}><SubscriptionPage /></Suspense>} />
          <Route path="plans" element={<Suspense fallback={<PageLoader />}><PlansPage /></Suspense>} />
          <Route path="anticensor" element={<Suspense fallback={<PageLoader />}><AntiCensorshipPage /></Suspense>} />
          <Route path="portal" element={<Suspense fallback={<PageLoader />}><UserPortalPage /></Suspense>} />
          <Route path="admin-roles" element={<Suspense fallback={<PageLoader />}><AdminRolesPage /></Suspense>} />
          <Route path="admins" element={<Suspense fallback={<PageLoader />}><AdminsPage /></Suspense>} />
          <Route path="api-tokens" element={<Suspense fallback={<PageLoader />}><ApiTokensPage /></Suspense>} />
          <Route path="online" element={<Suspense fallback={<PageLoader />}><OnlineMonitoringPage /></Suspense>} />
          <Route path="traffic" element={<Suspense fallback={<PageLoader />}><TrafficManagementPage /></Suspense>} />
          <Route path="resellers" element={<Suspense fallback={<PageLoader />}><ResellerManagementPage /></Suspense>} />
          <Route path="telegram" element={<Suspense fallback={<PageLoader />}><TelegramSettingsPage /></Suspense>} />
          <Route path="backups" element={<Suspense fallback={<PageLoader />}><BackupManagerPage /></Suspense>} />
          <Route path="metrics" element={<Suspense fallback={<PageLoader />}><MetricsDashboardPage /></Suspense>} />
          <Route path="outbounds" element={<Suspense fallback={<PageLoader />}><OutboundsPage /></Suspense>} />
          <Route path="routing" element={<Suspense fallback={<PageLoader />}><RoutingPage /></Suspense>} />
          <Route path="sub-profiles" element={<Suspense fallback={<PageLoader />}><SubscriptionProfilesPage /></Suspense>} />
          <Route path="security" element={<Suspense fallback={<PageLoader />}><SecurityPage /></Suspense>} />
          <Route path="client-groups" element={<Suspense fallback={<PageLoader />}><ClientGroupsPage /></Suspense>} />
          <Route path="federation" element={<Suspense fallback={<PageLoader />}><FederationPage /></Suspense>} />
          <Route path="cluster" element={<Suspense fallback={<PageLoader />}><ClusterManagementPage /></Suspense>} />
          <Route path="terminal" element={<Suspense fallback={<PageLoader />}><TerminalPage /></Suspense>} />
          <Route path="logs" element={<Suspense fallback={<PageLoader />}><LogStreamPage /></Suspense>} />
          <Route path="warp" element={<Suspense fallback={<PageLoader />}><WarpPage /></Suspense>} />
          <Route path="tls-tricks" element={<Suspense fallback={<PageLoader />}><TlsTricksPage /></Suspense>} />
          <Route path="plugins" element={<Suspense fallback={<PageLoader />}><PluginPage /></Suspense>} />
          <Route path="webrtc" element={<Suspense fallback={<PageLoader />}><WebRTCPage /></Suspense>} />
          <Route path="topology" element={<Suspense fallback={<PageLoader />}><TopologyVizPage /></Suspense>} />
          <Route path="domain-fronting" element={<Suspense fallback={<PageLoader />}><DomainFrontingPage /></Suspense>} />
          <Route path="xray-api" element={<Suspense fallback={<PageLoader />}><XrayAPIPage /></Suspense>} />
          <Route path="smart-dns" element={<Suspense fallback={<PageLoader />}><SmartDNSPage /></Suspense>} />
          <Route path="docker" element={<Suspense fallback={<PageLoader />}><DockerPage /></Suspense>} />
          <Route path="health" element={<Suspense fallback={<PageLoader />}><HealthPage /></Suspense>} />
          <Route path="settings" element={<Suspense fallback={<PageLoader />}><SettingsPage /></Suspense>} />
        </Route>

        <Route path="*" element={<Suspense fallback={<PageLoader />}><NotFoundPage /></Suspense>} />
      </Routes>
    </Suspense>
  )
}
