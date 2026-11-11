<template>
  <section class="accounting-orders" aria-labelledby="accounting-orders-title" data-testid="accounting-orders-report">
    <div class="section-heading">
      <div><h2 id="accounting-orders-title">Órdenes y solicitudes de factura</h2><p>Consulta los datos del cliente y el pago vigente de cada orden del periodo.</p></div>
      <Button label="Exportar órdenes" icon="pi pi-download" severity="secondary" outlined :loading="exporting" :disabled="loading || !pagination.total" @click="exportOrders" />
    </div>
    <div class="toolbar">
      <div class="form-field">
        <label for="accounting-invoice">Solicitud de factura</label>
        <Select id="accounting-invoice" v-model="invoiceFilter" :options="invoiceOptions" option-label="label" option-value="value" />
      </div>
      <Button label="Actualizar" icon="pi pi-refresh" severity="secondary" outlined :loading="loading" @click="load(1)" />
    </div>
    <Message v-if="exportError" severity="error" :closable="false">{{ exportError }}</Message>
    <Message v-if="error" severity="error" :closable="false">{{ error }} <Button label="Reintentar" size="small" outlined @click="load(1)" /></Message>
    <LoadingState v-else-if="loading" />
    <EmptyState v-else-if="!orders.length" title="No hay órdenes con estos filtros" description="Prueba otro periodo o cambia la solicitud de factura." />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="orders" data-key="id" striped-rows>
          <Column header="Orden"><template #body="{ data }"><strong>#{{ data.id }}</strong><small class="table-secondary">{{ date(data.received_at) }}</small><small class="table-secondary">{{ statusLabel(data.status) }}</small></template></Column>
          <Column header="Factura"><template #body="{ data }"><Tag :value="data.invoice_requested ? 'Solicitada' : 'No solicitada'" :severity="data.invoice_requested ? 'info' : 'secondary'" /></template></Column>
          <Column field="client_name" header="Cliente" />
          <Column field="client_ci" header="CI" />
          <Column field="client_phone" header="Teléfono" />
          <Column header="Correo electrónico"><template #body="{ data }">{{ data.client_email || 'No proporcionado' }}</template></Column>
          <Column header="Placa"><template #body="{ data }">{{ data.plate || 'Sin placa' }}</template></Column>
          <Column header="Total"><template #body="{ data }">Bs {{ data.total }}</template></Column>
          <Column header="Pago"><template #body="{ data }">{{ data.payment_id ? methodLabel(data.payment_method) : 'Pendiente' }}<small v-if="data.paid_at" class="table-secondary">{{ date(data.paid_at) }}</small></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="load($event.page + 1)" />
    </template>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { AccountingOrder, AccountingOrdersPage } from '~/types/accounting'
import { apiError } from '~/utils/api-error'
import { downloadCSV } from '~/utils/csv'

const props = defineProps<{ dateFrom: string, dateTo: string }>()
const auth = useAuthStore()
const orders = ref<AccountingOrder[]>([])
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const invoiceFilter = ref('')
const invoiceOptions = [{ label: 'Todas', value: '' }, { label: 'Solicitada', value: 'true' }, { label: 'No solicitada', value: 'false' }]
const loading = ref(false)
const exporting = ref(false)
const error = ref('')
const exportError = ref('')
let requestVersion = 0

const parameters = (page: number, limit: number) => {
  const query = new URLSearchParams({ date_from: props.dateFrom, date_to: props.dateTo, page: String(page), limit: String(limit) })
  if (invoiceFilter.value) query.set('invoice_requested', invoiceFilter.value)
  return query
}
async function load(page: number) {
  const version = ++requestVersion
  loading.value = true
  error.value = ''
  try {
    const response = await auth.request<AccountingOrdersPage>(`/reports/orders?${parameters(page, 20)}`)
    if (version !== requestVersion) return
    orders.value = response.data?.orders || []
    Object.assign(pagination, response.data?.pagination)
  } catch (cause) {
    if (version !== requestVersion) return
    orders.value = []
    error.value = apiError(cause, 'No se pudieron cargar las órdenes contables').message
  } finally {
    if (version === requestVersion) loading.value = false
  }
}
const date = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'short', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const statusLabel = (value: string) => ({ PENDING: 'Pendiente', IN_PROGRESS: 'En proceso', QUALITY_CONTROL: 'Control de calidad', FINISHED: 'Terminado', DELIVERED: 'Entregado', CANCELLED: 'Cancelado' }[value] || value)
const methodLabel = (value: string | null) => ({ CASH: 'Efectivo', QR: 'QR', BANK_TRANSFER: 'Transferencia', CARD: 'Tarjeta' }[value || ''] || value || '')
async function exportOrders() {
  if (exporting.value) return
  exporting.value = true
  exportError.value = ''
  const query = parameters(1, 100)
  try {
    const items: AccountingOrder[] = []
    let pages = 1
    for (let page = 1; page <= pages; page++) {
      query.set('page', String(page))
      const response = await auth.request<AccountingOrdersPage>(`/reports/orders?${query}`)
      if (!response.data) throw new Error('No se pudo completar la exportación')
      items.push(...response.data.orders)
      pages = response.data.pagination.total_pages
    }
    downloadCSV(`camell-wash-ordenes-${query.get('date_from')}-${query.get('date_to')}.csv`, [
      ['Orden', 'Recepción', 'Estado', 'Solicitud de factura', 'Cliente', 'CI', 'Teléfono', 'Correo', 'Placa', 'Total BOB', 'Pago', 'Método', 'Fecha de pago'],
      ...items.map(item => [item.id, date(item.received_at), statusLabel(item.status), item.invoice_requested ? 'Solicitada' : 'No solicitada', item.client_name, item.client_ci, item.client_phone, item.client_email, item.plate, item.total, item.payment_id ? 'Pagado' : 'Pendiente', methodLabel(item.payment_method), item.paid_at ? date(item.paid_at) : ''])
    ])
  } catch (cause) { exportError.value = apiError(cause, 'No se pudo exportar el periodo completo').message }
  finally { exporting.value = false }
}
watch(() => [props.dateFrom, props.dateTo, invoiceFilter.value], () => { void load(1) }, { immediate: true })
</script>

<style scoped>
.accounting-orders { margin-top: var(--space-8); }
.section-heading { display: flex; justify-content: space-between; align-items: center; gap: var(--space-4); flex-wrap: wrap; }
.section-heading p { color: var(--color-text-muted); }
.toolbar { justify-content: flex-start; align-items: end; }
.table-wrap { overflow-x: auto; }
.table-wrap :deep(td) { overflow-wrap: anywhere; min-width: 7rem; }
</style>
