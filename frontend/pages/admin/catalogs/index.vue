<template>
  <section>
    <div class="page-heading">
      <h1>Catálogos</h1>
      <p>Configura tipos de vehículo, categorías y servicios extensibles. Los registros inactivos conservan sus referencias y no aparecen en operaciones nuevas.</p>
    </div>

    <Tabs v-model:value="activeTab">
      <TabList>
        <Tab value="vehicle-types"><i class="pi pi-car" aria-hidden="true" /> Tipos de vehículo</Tab>
        <Tab value="categories"><i class="pi pi-folder" aria-hidden="true" /> Categorías</Tab>
        <Tab value="services"><i class="pi pi-sparkles" aria-hidden="true" /> Servicios</Tab>
      </TabList>
      <TabPanels>
        <TabPanel value="vehicle-types">
          <SimpleCatalogPanel
            kind="vehicle-types"
            response-key="vehicle_types"
            title="Tipos de vehículo"
            singular="Tipo de vehículo"
            description="Define los tipos que estarán disponibles al registrar vehículos y configurar precios."
          />
        </TabPanel>
        <TabPanel value="categories">
          <SimpleCatalogPanel
            kind="service-categories"
            response-key="service_categories"
            title="Categorías de servicio"
            singular="Categoría"
            description="Organiza los servicios sin limitar el catálogo a una lista fija."
            inactive-warning="Sus servicios dejarán de aparecer en listados operativos mientras permanezca inactiva, aunque conservarán su propio estado."
            @changed="servicePanel?.reload()"
          />
        </TabPanel>
        <TabPanel value="services">
          <ServiceCatalogPanel ref="servicePanel" />
        </TabPanel>
      </TabPanels>
    </Tabs>
  </section>
</template>

<script setup lang="ts">
import Tab from 'primevue/tab'
import TabList from 'primevue/tablist'
import TabPanel from 'primevue/tabpanel'
import TabPanels from 'primevue/tabpanels'
import Tabs from 'primevue/tabs'
import ServiceCatalogPanel from '~/components/catalogs/ServiceCatalogPanel.vue'
import SimpleCatalogPanel from '~/components/catalogs/SimpleCatalogPanel.vue'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Catálogos' })

const activeTab = ref('vehicle-types')
const servicePanel = ref<InstanceType<typeof ServiceCatalogPanel> | null>(null)
</script>

<style scoped>
:deep(.p-tablist-tab-list) { overflow-x: auto; border-color: var(--color-border-soft); background: transparent; }
:deep(.p-tab) { min-height: 3rem; gap: var(--space-2); color: var(--color-text-muted); white-space: nowrap; }
:deep(.p-tab-active) { color: var(--color-primary); }
:deep(.p-tabpanels) { padding: var(--space-6) 0 0; background: transparent; color: var(--color-text); }
@media (max-width: 40rem) { :deep(.p-tabpanels) { padding-top: var(--space-5); } }
</style>
