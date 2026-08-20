export default defineNuxtRouteMiddleware(async () => {
  if (import.meta.server) return
  const auth = useAuthStore()
  if (await auth.initialize()) return navigateTo(auth.isAdmin ? '/admin' : '/app')
})

