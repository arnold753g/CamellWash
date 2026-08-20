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

interface NavigationItem { label: string, to: string, icon: string }

const props = defineProps<{ mode: 'admin' | 'staff' }>()
const auth = useAuthStore()
const branding = useBranding()
const menuOpen = ref(false)
const loggingOut = ref(false)
const hydrated = ref(false)
const homePath = computed(() => props.mode === 'admin' ? '/admin' : '/app')
const roleLabels = { admin: 'Administrador', manager: 'Encargado', wash_staff: 'Personal de lavado' } as const
const displayName = computed(() => hydrated.value ? auth.fullName : '')
const roleLabel = computed(() => hydrated.value && auth.user ? roleLabels[auth.user.role] : '')
const userInitials = computed(() => {
  if (!hydrated.value) return 'U'
  const names = [auth.user?.first_name, auth.user?.last_name].filter(Boolean)
  return names.map(value => String(value).charAt(0).toUpperCase()).join('').slice(0, 2) || 'U'
})
const items = computed<NavigationItem[]>(() => props.mode === 'admin'
  ? [
      { label: 'Dashboard', to: '/admin', icon: 'pi pi-home' },
      { label: 'Reportes', to: '/admin/reports', icon: 'pi pi-chart-bar' },
      { label: 'Usuarios', to: '/admin/users', icon: 'pi pi-users' },
      { label: 'Catálogos', to: '/admin/catalogs', icon: 'pi pi-list' },
      { label: 'Precios', to: '/admin/prices', icon: 'pi pi-dollar' },
      { label: 'Clientes', to: '/admin/clients', icon: 'pi pi-id-card' },
      { label: 'Órdenes', to: '/admin/orders', icon: 'pi pi-clipboard' },
      { label: 'Revisión de pagos', to: '/admin/payment-reviews', icon: 'pi pi-verified' },
      { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }
    ]
  : [
      { label: 'Inicio', to: '/app', icon: 'pi pi-home' },
      ...(hydrated.value && auth.user?.role === 'manager' ? [{ label: 'Reportes', to: '/app/reports', icon: 'pi pi-chart-bar' }] : []),
      ...(hydrated.value && auth.user?.role === 'manager' ? [{ label: 'Clientes', to: '/app/clients', icon: 'pi pi-id-card' }] : []),
      { label: 'Órdenes', to: '/app/orders', icon: 'pi pi-clipboard' },
      { label: 'Mi perfil', to: '/app/profile', icon: 'pi pi-user' }
    ])

const logout = async () => {
  loggingOut.value = true
  await auth.logout()
  await navigateTo('/login')
}

onMounted(() => { hydrated.value = true })
</script>
