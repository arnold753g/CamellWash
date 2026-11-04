import { roleHomePath } from '~/utils/role-access'

export default defineNuxtRouteMiddleware(async () => {
  if (import.meta.server) return
  const auth = useAuthStore()
  if (await auth.initialize()) return navigateTo(roleHomePath(auth.user?.role))
})
