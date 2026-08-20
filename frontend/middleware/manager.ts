export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) return
  const auth = useAuthStore()
  if (!await auth.initialize()) return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  if (auth.user?.role === 'admin') return navigateTo('/admin/clients')
  if (auth.user?.role !== 'manager') return navigateTo('/403')
})
