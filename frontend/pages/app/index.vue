<template>
  <DashboardReportWorkspace v-if="auth.user?.role === 'manager'" :show-financial="false" />
  <section v-else>
    <div class="page-heading"><h1>Mi operación</h1><p>Consulta las órdenes que tienes asignadas y registra sus avances.</p></div>
    <div class="stat-grid">
      <Card><template #content><i class="pi pi-clipboard" aria-hidden="true" /><h2>Órdenes asignadas</h2><p class="muted">Tu lista muestra únicamente los vehículos donde formas parte del equipo.</p><Button as="router-link" to="/app/orders" label="Ver mis órdenes" icon="pi pi-arrow-right" /></template></Card>
      <Card><template #content><i class="pi pi-shield" aria-hidden="true" /><h2>Datos protegidos</h2><p class="muted">Los cobros y agregados financieros permanecen disponibles solo para los roles autorizados.</p></template></Card>
    </div>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import DashboardReportWorkspace from '~/components/reports/DashboardReportWorkspace.vue'

definePageMeta({ layout: 'app', middleware: 'user' })
useHead({ title: 'Inicio' })
const auth = useAuthStore()
</script>
