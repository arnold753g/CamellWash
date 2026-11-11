<template>
  <div class="app-shell">
    <a class="skip-link" href="#main-content">Saltar al contenido</a>
    <button v-if="menuOpen" class="sidebar-overlay" type="button" aria-label="Cerrar navegación" @click="menuOpen = false" />
    <aside class="app-sidebar" :class="{ 'is-open': menuOpen }" aria-label="Navegación principal">
      <NuxtLink class="sidebar-brand" :to="homePath" :aria-label="branding.name" @click="menuOpen = false">
        <img :src="branding.logo" :alt="branding.name">
      </NuxtLink>
      <nav class="sidebar-nav">
        <NuxtLink v-for="item in items" :key="item.to" :to="item.to" @click="menuOpen = false">
          <i :class="item.icon" aria-hidden="true" />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </nav>
      <div class="sidebar-footer">
        <Button label="Cerrar sesión" icon="pi pi-sign-out" severity="secondary" outlined fluid :loading="loggingOut" @click="logout" />
      </div>
    </aside>

    <div class="app-content">
      <header class="app-header">
        <button class="icon-button mobile-only" type="button" aria-label="Abrir navegación" @click="menuOpen = true">
          <i class="pi pi-bars" aria-hidden="true" />
        </button>
        <div class="header-spacer" />
        <div class="header-account">
          <div class="header-user">
            <strong>{{ displayName }}</strong>
            <span>{{ roleLabel }}</span>
          </div>
          <div class="header-avatar" aria-hidden="true">{{ userInitials }}</div>
        </div>
      </header>
      <main id="main-content" class="app-main" tabindex="-1">
        <div class="page-container"><slot /></div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import { roleHomePath, roleLabel as labelForRole, roleNavigation } from '~/utils/role-access'

const auth = useAuthStore()
const branding = useBranding()
const menuOpen = ref(false)
const loggingOut = ref(false)
const hydrated = ref(false)
const homePath = computed(() => hydrated.value ? roleHomePath(auth.user?.role) : '/')
const displayName = computed(() => hydrated.value ? auth.fullName : '')
const roleLabel = computed(() => hydrated.value && auth.user ? labelForRole(auth.user.role) : '')
const userInitials = computed(() => {
  if (!hydrated.value) return 'U'
  const names = [auth.user?.first_name, auth.user?.last_name].filter(Boolean)
  return names.map(value => String(value).charAt(0).toUpperCase()).join('').slice(0, 2) || 'U'
})
const items = computed(() => hydrated.value ? roleNavigation(auth.user?.role) : [])

const logout = async () => {
  loggingOut.value = true
  await auth.logout()
  await navigateTo('/login')
}

onMounted(() => { hydrated.value = true })
</script>
