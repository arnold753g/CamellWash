import type { Role } from '../types/auth'
import { safeInternalRedirect } from './auth-session.ts'

export const roleOptions = [
  { label: 'Administrador', value: 'admin' },
  { label: 'Encargado', value: 'manager' },
  { label: 'Personal de lavado', value: 'wash_staff' },
  { label: 'Contador', value: 'accountant' }
] satisfies { label: string, value: Role }[]

export const roleLabel = (role: Role) => roleOptions.find(option => option.value === role)?.label || role

export const roleHomePath = (role: string | null | undefined) => {
  if (role === 'admin') return '/admin'
  if (role === 'accountant') return '/accountant'
  if (role === 'manager' || role === 'wash_staff') return '/app'
  return '/403'
}

export type RoleArea = 'admin' | 'operations' | 'manager' | 'accountant' | 'cashier'

export const canAccessRoleArea = (role: string | null | undefined, area: RoleArea) => {
  switch (area) {
    case 'cashier': return role === 'admin' || role === 'manager'
    case 'admin': return role === 'admin'
    case 'operations': return role === 'manager' || role === 'wash_staff'
    case 'manager': return role === 'manager'
    case 'accountant': return role === 'accountant'
  }
}

export interface NavigationItem { label: string, to: string, icon: string }

const profileItem = { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }

export const roleNavigation = (role: string | null | undefined): NavigationItem[] => {
  if (role === 'admin') return [
    { label: 'Dashboard', to: '/admin', icon: 'pi pi-home' },
    { label: 'Reportes', to: '/admin/reports', icon: 'pi pi-chart-bar' },
    { label: 'Usuarios', to: '/admin/users', icon: 'pi pi-users' },
    { label: 'Catálogos', to: '/admin/catalogs', icon: 'pi pi-list' },
    { label: 'Precios', to: '/admin/prices', icon: 'pi pi-dollar' },
    { label: 'Clientes', to: '/admin/clients', icon: 'pi pi-id-card' },
    { label: 'Órdenes', to: '/admin/orders', icon: 'pi pi-clipboard' },
    { label: 'Revisión de pagos', to: '/admin/payment-reviews', icon: 'pi pi-verified' },
    profileItem
  ]
  if (role === 'accountant') return [
    { label: 'Reportes', to: '/accountant', icon: 'pi pi-chart-bar' },
    profileItem
  ]
  if (role === 'manager' || role === 'wash_staff') return [
    { label: 'Inicio', to: '/app', icon: 'pi pi-home' },
    ...(role === 'manager' ? [
      { label: 'Reportes', to: '/app/reports', icon: 'pi pi-chart-bar' },
      { label: 'Clientes', to: '/app/clients', icon: 'pi pi-id-card' }
    ] : []),
    { label: 'Órdenes', to: '/app/orders', icon: 'pi pi-clipboard' },
    profileItem
  ]
  return []
}

export const roleRedirect = (role: string | null | undefined, requested: unknown) => {
  const fallback = roleHomePath(role)
  const target = safeInternalRedirect(requested, fallback)
  const path = target.split(/[?#]/, 1)[0]?.replace(/\/+$/, '') || '/'
  if (!roleOptions.some(option => option.value === role)) return fallback
  if (path === '/app/profile') return target
  if (canAccessRoleArea(role, 'cashier') && /^\/print\/orders\/[1-9]\d*$/.test(path)) return target
  if (role === 'admin' && (path === '/admin' || path.startsWith('/admin/'))) return target
  if (role === 'accountant' && path === '/accountant') return target
  if (canAccessRoleArea(role, 'operations') && (path === '/app' || path === '/app/orders')) return target
  if (role === 'manager' && (path === '/app/clients' || path === '/app/reports')) return target
  return fallback
}
