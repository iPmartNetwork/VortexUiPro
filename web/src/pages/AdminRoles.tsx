import { useState, useEffect } from 'react'
import { apiClient, apiGet, apiPost, apiPut, apiDelete } from '../api/client'

interface AdminRole {
  id: number
  name: string
  slug: string
  builtIn: boolean
  ownerRole: boolean
  permissions: Record<string, any>
  limits: Record<string, any>
  features: Record<string, any>
  access: Record<string, any>
  adminCount: number
  createdAt: number
  updatedAt: number
}

interface PermissionEditor {
  resource: string
  actions: { key: string; label: string }[]
}

const RESOURCES: PermissionEditor[] = [
  { resource: 'users', actions: [{ key: 'view', label: 'View Users' }, { key: 'create', label: 'Create Users' }, { key: 'update', label: 'Update Users' }, { key: 'delete', label: 'Delete Users' }, { key: 'resetUsage', label: 'Reset Usage' }, { key: 'setOwner', label: 'Change Owner' }] },
  { resource: 'inbounds', actions: [{ key: 'view', label: 'View Inbounds' }, { key: 'create', label: 'Create Inbounds' }, { key: 'update', label: 'Update Inbounds' }, { key: 'delete', label: 'Delete Inbounds' }] },
  { resource: 'admins', actions: [{ key: 'view', label: 'View Admins' }, { key: 'create', label: 'Create Admins' }, { key: 'update', label: 'Update Admins' }, { key: 'delete', label: 'Delete Admins' }] },
  { resource: 'roles', actions: [{ key: 'view', label: 'View Roles' }, { key: 'create', label: 'Create Roles' }, { key: 'update', label: 'Update Roles' }, { key: 'delete', label: 'Delete Roles' }] },
  { resource: 'nodes', actions: [{ key: 'view', label: 'View Nodes' }, { key: 'create', label: 'Create Nodes' }, { key: 'update', label: 'Update Nodes' }, { key: 'delete', label: 'Delete Nodes' }] },
  { resource: 'settings', actions: [{ key: 'view', label: 'View Settings' }, { key: 'update', label: 'Update Settings' }] },
  { resource: 'system', actions: [{ key: 'view', label: 'View System' }] },
]

export function AdminRolesPage() {
  const [roles, setRoles] = useState<AdminRole[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editingRole, setEditingRole] = useState<AdminRole | null>(null)
  const [showEditor, setShowEditor] = useState(false)
  const [roleName, setRoleName] = useState('')
  const [permissions, setPermissions] = useState<Record<string, any>>({})
  const [features, setFeatures] = useState<Record<string, boolean>>({ blockLimitedAdmins: false, disconnectUsersWhenLimited: true, disconnectUsersWhenDisabled: true, useResetStrategy: true, useNextPlan: true })
  const [allowAllGroups, setAllowAllGroups] = useState(true)
  const [allowAllInbounds, setAllowAllInbounds] = useState(true)

  useEffect(() => { fetchRoles() }, [])

  const fetchRoles = async () => {
    try { setLoading(true); const { data } = await apiGet('/api/v1/roles'); setRoles(data.roles || []) }
    catch (err: any) { setError(err?.response?.data?.error || 'Failed to load roles') }
    finally { setLoading(false) }
  }

  const openCreate = () => {
    setEditingRole(null); setRoleName(''); setPermissions({})
    setFeatures({ blockLimitedAdmins: false, disconnectUsersWhenLimited: true, disconnectUsersWhenDisabled: true, useResetStrategy: true, useNextPlan: true })
    setAllowAllGroups(true); setAllowAllInbounds(true); setShowEditor(true)
  }

  const openEdit = (role: AdminRole) => {
    if (role.ownerRole) return
    setEditingRole(role); setRoleName(role.name); setPermissions(role.permissions || {})
    setFeatures({ blockLimitedAdmins: role.features?.blockLimitedAdmins || false, disconnectUsersWhenLimited: role.features?.disconnectUsersWhenLimited ?? true, disconnectUsersWhenDisabled: role.features?.disconnectUsersWhenDisabled ?? true, useResetStrategy: role.features?.useResetStrategy ?? true, useNextPlan: role.features?.useNextPlan ?? true })
    setAllowAllGroups(role.access?.allowAllGroups ?? true); setAllowAllInbounds(role.access?.allowAllInbounds ?? true); setShowEditor(true)
  }

  const togglePermission = (resource: string, action: string) => {
    setPermissions((prev) => {
      const res = { ...(prev[resource] as Record<string, any> || {}) }
      if (res[action] === true) delete res[action]; else res[action] = true
      const next = { ...prev, [resource]: res }
      if (Object.keys(res).length === 0) delete next[resource]
      return next
    })
  }

  const getPermissionValue = (resource: string, action: string): boolean => {
    const res = permissions[resource] as Record<string, any> | undefined
    return !!res?.[action]
  }

  const saveRole = async () => {
    try {
      if (editingRole) await apiPut(`/api/v1/roles/${editingRole.id}`, { name: roleName, permissions, features, access: { allowAllGroups, allowAllInbounds }, limits: {} })
      else await apiPost('/api/v1/roles', { name: roleName, permissions, features, access: { allowAllGroups, allowAllInbounds }, limits: {} })
      setShowEditor(false); fetchRoles()
    } catch (err: any) { setError(err?.response?.data?.error || 'Failed to save role') }
  }

  const duplicateRole = async (id: number) => { try { await apiPost(`/api/v1/roles/${id}/duplicate`); fetchRoles() } catch (err: any) { setError(err?.response?.data?.error || 'Failed to duplicate role') } }
  const deleteRole = async (id: number) => { if (!confirm('Delete this role?')) return; try { await apiDelete(`/api/v1/roles/${id}`); fetchRoles() } catch (err: any) { setError(err?.response?.data?.error || 'Failed to delete role') } }

  return (
    <div className="page-enter max-w-[1200px] mx-auto py-6 px-4 space-y-6">
      <div className="glass-panel p-5 flex items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 flex items-center justify-center shadow-lg shrink-0">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
          </div>
          <div><h1 className="text-xl font-bold text-[var(--text-primary)]">Admin Roles</h1><p className="text-sm text-[var(--text-secondary)] mt-0.5">Manage RBAC roles and permissions</p></div>
        </div>
        <button onClick={openCreate} className="btn-primary text-sm">+ New Role</button>
      </div>

      {error && <div className="glass-card p-3 flex items-center justify-between border-red-500/20"><span className="text-sm text-red-400">{error}</span><button onClick={() => setError('')} className="text-red-400 hover:text-red-300">×</button></div>}

      {loading ? <div className="flex justify-center py-12"><div className="loading-spinner loading-spinner-lg" /></div> : (
        <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {roles.map((role) => (
            <div key={role.id} className={`glass-card p-5 ${role.ownerRole ? 'border-amber-500/20' : ''}`}>
              <div className="flex items-start justify-between mb-3">
                <div>
                  <h3 className="text-base font-bold text-[var(--text-primary)] flex items-center gap-2">
                    {role.name}
                    {role.builtIn && <span className="badge">Built-in</span>}
                    {role.ownerRole && <span className="badge badge-warning">Owner</span>}
                  </h3>
                  <span className="text-xs text-[var(--text-muted)]">@{role.slug}</span>
                </div>
                <span className="text-xs text-[var(--text-muted)] bg-[var(--bg-elevated)] px-2.5 py-1 rounded-md">{role.adminCount} admin{role.adminCount !== 1 ? 's' : ''}</span>
              </div>
              <div className="mb-4 flex flex-wrap gap-1">
                {Object.entries(role.permissions || {}).slice(0, 4).map(([res, actions]) => (
                  <span key={res} className="badge badge-purple text-[10px]">{res}: {Object.keys(actions as object).join(', ')}</span>
                ))}
                {Object.keys(role.permissions || {}).length > 4 && <span className="text-[10px] text-[var(--text-muted)] self-center">+{Object.keys(role.permissions).length - 4} more</span>}
              </div>
              <div className="flex gap-2">
                {!role.ownerRole && <><button onClick={() => openEdit(role)} className="btn-secondary text-xs py-1.5 px-3">Edit</button><button onClick={() => duplicateRole(role.id)} className="btn-ghost text-xs py-1.5 px-3">Duplicate</button>{!role.builtIn && <button onClick={() => deleteRole(role.id)} className="btn-ghost text-xs py-1.5 px-3 text-red-400 hover:text-red-300">Delete</button>}</>}
              </div>
            </div>
          ))}
        </div>
      )}

      {showEditor && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 z-50" onClick={() => setShowEditor(false)}>
          <div className="glass-card p-7 w-full max-w-[700px] max-h-[85vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-[var(--text-primary)] mb-5">{editingRole ? `Edit Role: ${editingRole.name}` : 'Create New Role'}</h2>
            <div className="mb-4"><label className="block text-xs font-medium text-[var(--text-secondary)] mb-1.5">Role Name</label>
              <input value={roleName} onChange={(e) => setRoleName(e.target.value)} className="input-modern text-sm" placeholder="e.g. Support Agent" /></div>
            <div className="mb-5"><h3 className="text-sm font-semibold text-[var(--text-secondary)] mb-3">Permissions</h3>
              <div className="grid gap-2">{RESOURCES.map(({ resource, actions }) => (
                <div key={resource} className="p-3 rounded-lg bg-[var(--bg-elevated)] border border-[var(--border-light)]">
                  <div className="text-xs font-bold text-[var(--accent-purple)] mb-2 uppercase tracking-wider">{resource}</div>
                  <div className="flex flex-wrap gap-2">{actions.map(({ key, label }) => (
                    <label key={key} className={`flex items-center gap-1.5 cursor-pointer text-xs ${getPermissionValue(resource, key) ? 'text-[var(--accent-purple)]' : 'text-[var(--text-muted)]'}`}>
                      <input type="checkbox" checked={getPermissionValue(resource, key)} onChange={() => togglePermission(resource, key)} className="accent-[var(--accent-purple)]" />{label}</label>
                  ))}</div></div>))}</div></div>
            <div className="mb-5"><h3 className="text-sm font-semibold text-[var(--text-secondary)] mb-3">Features</h3>
              <div className="flex flex-wrap gap-3">{Object.entries(features).map(([key, val]) => (
                <label key={key} className="flex items-center gap-1.5 cursor-pointer text-xs text-[var(--text-secondary)]">
                  <input type="checkbox" checked={val} onChange={() => setFeatures(f => ({ ...f, [key]: !f[key] }))} className="accent-[var(--accent-purple)]" />
                  {key.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase())}</label>))}</div></div>
            <div className="mb-5"><h3 className="text-sm font-semibold text-[var(--text-secondary)] mb-3">Access Scope</h3>
              <div className="flex gap-4"><label className="flex items-center gap-1.5 cursor-pointer text-xs text-[var(--text-secondary)]"><input type="checkbox" checked={allowAllGroups} onChange={() => setAllowAllGroups(!allowAllGroups)} className="accent-[var(--accent-purple)]" /> All Groups</label>
                <label className="flex items-center gap-1.5 cursor-pointer text-xs text-[var(--text-secondary)]"><input type="checkbox" checked={allowAllInbounds} onChange={() => setAllowAllInbounds(!allowAllInbounds)} className="accent-[var(--accent-purple)]" /> All Inbounds</label></div></div>
            <div className="flex justify-end gap-2.5">
              <button onClick={() => setShowEditor(false)} className="btn-ghost text-sm">Cancel</button>
              <button onClick={saveRole} className="btn-primary text-sm">{editingRole ? 'Save Changes' : 'Create Role'}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
