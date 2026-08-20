<template>
  <section>
    <div class="page-heading">
      <h1>Matriz de precios</h1>
      <p>Configura una tarifa por cada combinación activa de servicio y tipo de vehículo. Las celdas sin precio permanecen visibles hasta que el Administrador las complete.</p>
    </div>

    <div class="stat-grid pricing-summary" aria-label="Resumen de la matriz">
      <Card>
        <template #content>
          <i class="pi pi-th-large" aria-hidden="true" />
          <h2>Combinaciones activas</h2>
          <p class="stat-value">{{ matrix.summary.total }}</p>
        </template>
      </Card>
      <Card>
        <template #content>
          <i class="pi pi-check-circle" aria-hidden="true" />
          <h2>Con precio</h2>
          <p class="stat-value">{{ matrix.summary.configured }}</p>
        </template>
      </Card>
      <Card>
        <template #content>
          <i class="pi pi-exclamation-circle" aria-hidden="true" />
          <h2>Sin precio</h2>
          <p class="stat-value">{{ matrix.summary.missing }}</p>
        </template>
      </Card>
    </div>

    <div class="toolbar pricing-toolbar">
      <div class="cluster pricing-filters">
        <div class="form-field">
          <label for="price-category-filter">Categoría</label>
          <Select id="price-category-filter" v-model="filters.category_id" :options="categoryOptions" option-label="label" option-value="value" placeholder="Todas" show-clear />
        </div>
        <div class="form-field">
          <label for="price-service-filter">Servicio</label>
          <Select id="price-service-filter" v-model="filters.service_id" :options="serviceOptions" option-label="label" option-value="value" placeholder="Todos" show-clear filter />
        </div>
        <div class="form-field">
          <label for="price-vehicle-filter">Tipo de vehículo</label>
          <Select id="price-vehicle-filter" v-model="filters.vehicle_type_id" :options="vehicleTypeOptions" option-label="label" option-value="value" placeholder="Todos" show-clear filter />
        </div>
        <div class="form-field">
          <label for="price-state-filter">Cobertura</label>
          <Select id="price-state-filter" v-model="coverageFilter" :options="coverageOptions" option-label="label" option-value="value" />
        </div>
        <Button label="Filtrar" icon="pi pi-search" :loading="loading" @click="loadMatrix" />
        <Button label="Limpiar" icon="pi pi-filter-slash" severity="secondary" outlined :disabled="loading" @click="clearFilters" />
      </div>
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <LoadingState v-if="loading" />
    <EmptyState
      v-else-if="visibleRows.length === 0"
      title="No hay combinaciones para mostrar"
      description="Ajusta los filtros o activa al menos un servicio y un tipo de vehículo en Catálogos."
    />
    <div v-else class="table-wrap pricing-matrix-wrap">
      <table class="pricing-matrix">
        <thead>
          <tr>
            <th scope="col">Categoría</th>
            <th scope="col">Servicio</th>
            <th v-for="vehicleType in visibleVehicleTypes" :key="vehicleType.id" scope="col">{{ vehicleType.name }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in visibleRows" :key="row.service_id">
            <td><Tag :value="row.category_name" severity="secondary" /></td>
            <th scope="row">{{ row.service_name }}</th>
            <td v-for="vehicleType in visibleVehicleTypes" :key="vehicleType.id">
              <button
                v-if="cellFor(row, vehicleType.id)"
                class="price-cell"
                :class="{ 'is-missing': !cellFor(row, vehicleType.id)?.configured }"
                type="button"
                :aria-label="cellLabel(cellFor(row, vehicleType.id)!)"
                @click="openEditor(cellFor(row, vehicleType.id)!)"
              >
                <strong v-if="cellFor(row, vehicleType.id)?.configured">{{ cellFor(row, vehicleType.id)?.price }}</strong>
                <span v-else><i class="pi pi-exclamation-circle" aria-hidden="true" /> Sin precio</span>
                <small>{{ cellFor(row, vehicleType.id)?.configured ? 'Editar' : 'Cargar' }}</small>
              </button>
              <span v-else class="outside-filter">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Dialog v-model:visible="dialogVisible" modal :header="editingCell?.configured ? 'Editar precio' : 'Cargar precio'" :style="dialogStyle" @hide="resetEditor">
      <Message v-if="formError" severity="error" class="form-error" :closable="false">{{ formError }}</Message>
      <dl v-if="editingCell" class="detail-grid price-context">
        <div class="detail-item"><dt>Servicio</dt><dd>{{ editingCell.service_name }}</dd></div>
        <div class="detail-item"><dt>Tipo de vehículo</dt><dd>{{ editingCell.vehicle_type_name }}</dd></div>
      </dl>
      <form id="price-form" class="stack" novalidate @submit.prevent="savePrice">
        <div class="form-field">
          <label for="service-price">Precio</label>
          <InputText id="service-price" v-model.trim="priceInput" inputmode="decimal" autocomplete="off" placeholder="0.00" :invalid="Boolean(priceError)" autofocus />
          <small class="form-help">Ingresa un importe mayor que 0.00, con un máximo de dos decimales. No se aceptan precios vacíos o gratuitos.</small>
          <small v-if="priceError" class="field-error" role="alert">{{ priceError }}</small>
        </div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="saving" @click="dialogVisible = false" />
        <Button form="price-form" type="submit" label="Guardar precio" icon="pi pi-save" :loading="saving" />
      </template>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { CatalogItem, ServiceCatalogItem } from '~/types/catalogs'
import type { PricingMatrixCell, PricingMatrixData, PricingMatrixRow } from '~/types/pricing'
import { apiError } from '~/utils/api-error'
import { normalizePriceInput } from '~/utils/money'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Matriz de precios' })

const auth = useAuthStore()
const toast = useToast()
const dialogStyle = { width: 'min(36rem, calc(100vw - 2rem))' }
const coverageOptions = [
  { label: 'Todas', value: 'all' },
  { label: 'Con faltantes', value: 'missing' },
  { label: 'Completas', value: 'complete' }
]

const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const formError = ref('')
const priceError = ref('')
const priceInput = ref('')
const coverageFilter = ref('all')
const dialogVisible = ref(false)
const editingCell = ref<PricingMatrixCell | null>(null)
const categories = ref<CatalogItem[]>([])
const services = ref<ServiceCatalogItem[]>([])
const vehicleTypes = ref<CatalogItem[]>([])
const matrix = reactive<PricingMatrixData>({ cells: [], summary: { total: 0, configured: 0, missing: 0 } })
const filters = reactive<{ category_id: number | null, service_id: number | null, vehicle_type_id: number | null }>({
  category_id: null, service_id: null, vehicle_type_id: null
})

const categoryOptions = computed(() => categories.value.map(item => ({ label: item.name, value: item.id })))
const serviceOptions = computed(() => services.value
  .filter(item => !filters.category_id || item.category_id === filters.category_id)
  .map(item => ({ label: `${item.name} — ${item.category.name}`, value: item.id })))
const vehicleTypeOptions = computed(() => vehicleTypes.value.map(item => ({ label: item.name, value: item.id })))
const visibleVehicleTypes = computed(() => {
  const identifiers = new Set(matrix.cells.map(cell => cell.vehicle_type_id))
  return vehicleTypes.value.filter(item => identifiers.has(item.id))
})
const rows = computed<PricingMatrixRow[]>(() => {
  const grouped = new Map<number, PricingMatrixRow>()
  for (const cell of matrix.cells) {
    const current = grouped.get(cell.service_id) || {
      service_id: cell.service_id,
      service_name: cell.service_name,
      category_id: cell.category_id,
      category_name: cell.category_name,
      cells: []
    }
    current.cells.push(cell)
    grouped.set(cell.service_id, current)
  }
  return [...grouped.values()]
})
const visibleRows = computed(() => rows.value.filter((row) => {
  if (coverageFilter.value === 'missing') return row.cells.some(cell => !cell.configured)
  if (coverageFilter.value === 'complete') return row.cells.length > 0 && row.cells.every(cell => cell.configured)
  return true
}))

watch(() => filters.category_id, () => {
  if (filters.service_id && !serviceOptions.value.some(option => option.value === filters.service_id)) filters.service_id = null
})

const loadCatalogs = async () => {
  const [categoryResponse, serviceResponse, vehicleResponse] = await Promise.all([
    auth.request<{ service_categories: CatalogItem[] }>('/catalogs/service-categories'),
    auth.request<{ services: ServiceCatalogItem[] }>('/catalogs/services'),
    auth.request<{ vehicle_types: CatalogItem[] }>('/catalogs/vehicle-types')
  ])
  categories.value = categoryResponse.data?.service_categories || []
  services.value = serviceResponse.data?.services || []
  vehicleTypes.value = vehicleResponse.data?.vehicle_types || []
}

const loadMatrix = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const query = new URLSearchParams()
    if (filters.category_id) query.set('category_id', String(filters.category_id))
    if (filters.service_id) query.set('service_id', String(filters.service_id))
    if (filters.vehicle_type_id) query.set('vehicle_type_id', String(filters.vehicle_type_id))
    const response = await auth.request<PricingMatrixData>(`/admin/service-prices/matrix?${query}`)
    matrix.cells = response.data?.cells || []
    Object.assign(matrix.summary, response.data?.summary || { total: 0, configured: 0, missing: 0 })
  } catch (error) {
    errorMessage.value = apiError(error).message
    matrix.cells = []
    Object.assign(matrix.summary, { total: 0, configured: 0, missing: 0 })
  } finally { loading.value = false }
}

const initialize = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    await loadCatalogs()
    await loadMatrix()
  } catch (error) {
    errorMessage.value = apiError(error).message
    loading.value = false
  }
}

const clearFilters = () => {
  filters.category_id = null
  filters.service_id = null
  filters.vehicle_type_id = null
  coverageFilter.value = 'all'
  void loadMatrix()
}
const cellFor = (row: PricingMatrixRow, vehicleTypeID: number) => row.cells.find(cell => cell.vehicle_type_id === vehicleTypeID)
const cellLabel = (cell: PricingMatrixCell) => cell.configured
  ? `${cell.service_name}, ${cell.vehicle_type_name}: precio ${cell.price}. Editar.`
  : `${cell.service_name}, ${cell.vehicle_type_name}: sin precio. Cargar.`
const openEditor = (cell: PricingMatrixCell) => {
  editingCell.value = cell
  priceInput.value = cell.price || ''
  priceError.value = ''
  formError.value = ''
  dialogVisible.value = true
}
const resetEditor = () => {
  editingCell.value = null
  priceInput.value = ''
  priceError.value = ''
  formError.value = ''
}
const savePrice = async () => {
  if (!editingCell.value) return
  priceError.value = ''
  formError.value = ''
  const normalized = normalizePriceInput(priceInput.value)
  if (!normalized) {
    priceError.value = 'Ingresa un precio mayor que 0.00 con un máximo de dos decimales.'
    return
  }
  saving.value = true
  try {
    const cell = editingCell.value
    const path = cell.configured
      ? `/admin/service-prices/${cell.service_id}/${cell.vehicle_type_id}`
      : '/admin/service-prices'
    const body = cell.configured
      ? { price: normalized }
      : { service_id: cell.service_id, vehicle_type_id: cell.vehicle_type_id, price: normalized }
    await auth.request(path, { method: cell.configured ? 'PATCH' : 'POST', body })
    toast.add({ severity: 'success', summary: cell.configured ? 'Precio actualizado' : 'Precio cargado', life: 3500 })
    dialogVisible.value = false
    await loadMatrix()
  } catch (error) { formError.value = apiError(error).message }
  finally { saving.value = false }
}

onMounted(initialize)
</script>

<style scoped>
.pricing-summary { margin-bottom: var(--space-8); }
.pricing-summary .stat-value { font-size: clamp(1.75rem, 3vw, 2.5rem); }
.pricing-toolbar { margin-bottom: var(--space-6); }
.pricing-filters { align-items: end; }
.pricing-filters .form-field { min-width: min(13rem, 100%); }
.pricing-matrix-wrap { border: 1px solid var(--color-border-soft); background: var(--color-surface); }
.pricing-matrix { width: 100%; border-collapse: separate; border-spacing: 0; }
.pricing-matrix th, .pricing-matrix td { min-width: 10.5rem; padding: var(--space-3); border-right: 1px solid var(--color-border-soft); border-bottom: 1px solid var(--color-border-soft); text-align: left; vertical-align: middle; }
.pricing-matrix thead th { position: sticky; top: 0; z-index: 2; background: var(--color-background-deep); color: var(--color-text-muted); font-size: 0.75rem; letter-spacing: 0.06em; text-transform: uppercase; }
.pricing-matrix thead th:first-child, .pricing-matrix tbody td:first-child { position: sticky; left: 0; z-index: 3; min-width: 9rem; background: var(--color-surface); }
.pricing-matrix thead th:nth-child(2), .pricing-matrix tbody th { position: sticky; left: 9rem; z-index: 3; min-width: 13rem; background: var(--color-surface); }
.pricing-matrix thead th:first-child, .pricing-matrix thead th:nth-child(2) { z-index: 4; background: var(--color-background-deep); }
.pricing-matrix tbody th { color: var(--color-text); font-weight: var(--font-weight-semibold); }
.price-cell { width: 100%; min-height: 4.5rem; display: grid; place-items: center; gap: var(--space-1); padding: var(--space-2); border: 1px solid rgba(142, 229, 93, 0.25); border-radius: var(--radius-sm); background: var(--color-success-soft); color: var(--color-text); cursor: pointer; font: inherit; }
.price-cell:hover, .price-cell:focus-visible { border-color: var(--color-primary); box-shadow: 0 0 0 3px var(--color-primary-soft); outline: 0; }
.price-cell strong { font-size: 1.0625rem; font-variant-numeric: tabular-nums; }
.price-cell small { color: var(--color-text-muted); }
.price-cell.is-missing { border-color: rgba(245, 196, 81, 0.32); border-style: dashed; background: var(--color-warning-soft); }
.price-cell.is-missing span { display: flex; align-items: center; gap: var(--space-2); color: var(--color-warning); font-weight: var(--font-weight-semibold); }
.outside-filter { display: block; text-align: center; color: var(--color-text-disabled); }
.price-context { margin-bottom: var(--space-5); }
@media (max-width: 40rem) {
  .pricing-filters > * { width: 100%; }
  .pricing-matrix th, .pricing-matrix td { min-width: 9.5rem; }
  .pricing-matrix thead th:first-child, .pricing-matrix tbody td:first-child { position: static; min-width: 8rem; }
  .pricing-matrix thead th:nth-child(2), .pricing-matrix tbody th { left: 0; min-width: 11rem; }
}
</style>
