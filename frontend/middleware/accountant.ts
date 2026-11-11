import { canAccessRoleArea } from '~/utils/role-access'

export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) return
  const auth = useAuthStore()
  if (!await auth.initialize()) return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  if (!canAccessRoleArea(auth.user?.role, 'accountant')) return navigateTo('/403')
})
