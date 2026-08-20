<template>
  <section data-testid="dashboard-report-workspace">
    <div class="report-heading">
      <div class="page-heading">
        <h1>{{ reportMode ? 'Reportes' : 'Dashboard' }}</h1>
        <p>{{ subtitle }}</p>
      </div>
      <span class="timezone-badge"><i class="pi pi-clock" aria-hidden="true" /> Hora de Bolivia</span>
    </div>

    <form class="report-filters" aria-label="Periodo del reporte" @submit.prevent="loadReports">
      <div class="form-field">
        <label for="report-date-from">Desde</label>
        <input id="report-date-from" v-model="filters.dateFrom" class="p-inputtext" type="date" :max="today" required>
      </div>
      <div class="form-field">
        <label for="report-date-to">Hasta</label>
        <input id="report-date-to" v-model="filters.dateTo" class="p-inputtext" type="date" :max="today" required>
      </div>
      <Button type="submit" label="Aplicar periodo" icon="pi pi-filter" :loading="loading" />
    </form>

    <Message v-if="errorMessage" severity="error" :closable="false">
      <div class="error-recovery"><span>{{ errorMessage }}</span><Button label="Reintentar" icon="pi pi-refresh" severity="secondary" outlined @click="loadReports" /></div>
    </Message>
    <LoadingState v-else-if="loading" />
    <EmptyState v-else-if="!operational || !hasReportData" title="Todavía no hay actividad en este periodo" description="Cambia el rango o registra nuevas órdenes para comenzar a ver indicadores." />
    <div v-else class="report-content">
      <section aria-labelledby="operational-title">
        <div class="section-heading">
          <div><span class="section-kicker">Operación</span><h2 id="operational-title">Flujo del lavadero</h2></div>
          <Button v-if="reportMode" label="Exportar operación" icon="pi pi-download" severity="secondary" outlined @click="exportOperational" />
        </div>
        <div class="metric-grid">
          <article class="metric-card"><i class="pi pi-inbox" aria-hidden="true" /><span>Órdenes recibidas</span><strong>{{ operational.summary.orders_received }}</strong><small>En el periodo</small></article>
          <article class="metric-card"><i class="pi pi-check-circle" aria-hidden="true" /><span>Entregadas</span><strong>{{ operational.summary.orders_delivered }}</strong><small>Transiciones a entregada</small></article>
          <article class="metric-card"><i class="pi pi-car" aria-hidden="true" /><span>En operación</span><strong>{{ operational.summary.active_orders }}</strong><small>Estado actual</small></article>
          <article class="metric-card" :class="{ 'metric-card--attention': operational.summary.ready_for_pickup > 0 }"><i class="pi pi-flag" aria-hidden="true" /><span>Listas para retiro</span><strong>{{ operational.summary.ready_for_pickup }}</strong><small>Terminadas y no entregadas</small></article>
          <article class="metric-card" :class="{ 'metric-card--danger': operational.summary.overdue_orders > 0 }"><i class="pi pi-exclamation-triangle" aria-hidden="true" /><span>Con demora</span><strong>{{ operational.summary.overdue_orders }}</strong><small>Superaron hora estimada</small></article>
          <article class="metric-card"><i class="pi pi-stopwatch" aria-hidden="true" /><span>Tiempo promedio</span><strong>{{ cycleTime }}</strong><small>Recepción a entrega</small></article>
        </div>

        <div class="report-grid">
          <Card>
            <template #title>Órdenes activas por estado</template>
            <template #content>
              <div class="bar-list" aria-label="Distribución actual de órdenes activas">
                <div v-for="item in operational.status" :key="item.status" class="bar-row">
                  <div class="bar-label"><span>{{ statusLabel(item.status) }}</span><strong>{{ item.count }}</strong></div>
                  <div class="bar-track" aria-hidden="true"><span :style="{ width: barWidth(item.count, maxStatus) }" /></div>
                </div>
              </div>
            </template>
          </Card>
          <Card>
            <template #title>Servicios más solicitados</template>
            <template #content>
              <div v-if="operational.top_services.length" class="bar-list" aria-label="Servicios más solicitados en el periodo">
                <div v-for="item in operational.top_services" :key="item.service_name" class="bar-row">
                  <div class="bar-label"><span>{{ item.service_name }}</span><strong>{{ item.uses }}</strong></div>
                  <div class="bar-track" aria-hidden="true"><span :style="{ width: barWidth(item.uses, maxServices) }" /></div>
                </div>
              </div>
              <p v-else class="muted">No hay servicios completados para comparar.</p>
            </template>
          </Card>
        </div>

        <Card v-if="reportMode" class="data-card">
          <template #title>Movimiento diario</template>
          <template #content>
            <div class="table-wrap">
              <DataTable :value="operational.daily" data-key="date" striped-rows>
                <Column field="date" header="Fecha"><template #body="{ data }">{{ formatDate(data.date) }}</template></Column>
                <Column field="received" header="Recibidas" />
                <Column field="delivered" header="Entregadas" />
              </DataTable>
            </div>
          </template>
        </Card>
      </section>

      <section v-if="financial" aria-labelledby="financial-title">
        <div class="section-heading">
          <div><span class="section-kicker">Solo Administrador</span><h2 id="financial-title">Resumen financiero</h2></div>
          <Button v-if="reportMode" label="Exportar finanzas" icon="pi pi-download" severity="secondary" outlined @click="exportFinancial" />
        </div>
        <div class="metric-grid metric-grid--financial">
          <article class="metric-card"><i class="pi pi-wallet" aria-hidden="true" /><span>Ingresos válidos</span><strong class="money">{{ money(financial.summary.revenue) }}</strong><small>Pagos vigentes del periodo</small></article>
          <article class="metric-card"><i class="pi pi-receipt" aria-hidden="true" /><span>Pagos registrados</span><strong>{{ financial.summary.payments }}</strong><small>Sin anulados</small></article>
          <article class="metric-card"><i class="pi pi-chart-line" aria-hidden="true" /><span>Ticket promedio</span><strong class="money">{{ money(financial.summary.average_ticket) }}</strong><small>Promedio por pago válido</small></article>
          <article class="metric-card metric-card--muted"><i class="pi pi-ban" aria-hidden="true" /><span>Anulaciones</span><strong>{{ financial.summary.voided_count }}</strong><small>{{ money(financial.summary.voided_amount) }} auditados aparte</small></article>
        </div>
        <Card>
          <template #title>Ingresos por método de pago</template>
          <template #content>
            <div class="payment-methods">
              <div v-for="item in financial.methods" :key="item.method" class="payment-method">
                <div><span>{{ methodLabel(item.method) }}</span><small>{{ item.count }} pago{{ item.count === 1 ? '' : 's' }}</small></div>
                <strong>{{ money(item.amount) }}</strong>
              </div>
            </div>
          </template>
        </Card>
        <Card v-if="reportMode" class="data-card">
          <template #title>Ingresos diarios</template>
          <template #content>
            <div class="table-wrap">
              <DataTable :value="financial.daily" data-key="date" striped-rows>
                <Column field="date" header="Fecha"><template #body="{ data }">{{ formatDate(data.date) }}</template></Column>
                <Column field="payments" header="Pagos" />
                <Column field="amount" header="Ingreso"><template #body="{ data }"><span class="money-cell">{{ money(data.amount) }}</span></template></Column>
              </DataTable>
            </div>
          </template>
        </Card>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Message from 'primevue/message'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { DashboardReport, FinancialReport, OperationalReport } from '~/types/reports'
import { apiError } from '~/utils/api-error'

const props = defineProps<{ showFinancial: boolean, reportMode?: boolean }>()
const auth = useAuthStore()
const loading = ref(true)
const errorMessage = ref('')
const operational = ref<OperationalReport | null>(null)
const financial = ref<FinancialReport | null>(null)
const now = new Date()
const todayParts = boliviaDateParts(now)
const today = `${todayParts.year}-${todayParts.month}-${todayParts.day}`
const firstDayOfMonth = `${todayParts.year}-${todayParts.month}-01`
const filters = reactive({ dateFrom: props.reportMode ? firstDayOfMonth : today, dateTo: today })
const subtitle = computed(() => props.reportMode
  ? 'Analiza el desempeño por periodo. Los límites usan la zona horaria America/La_Paz.'
  : 'La operación actual y los resultados del día, sin estimaciones ni datos duplicados.')
const hasReportData = computed(() => Boolean(operational.value && (
  operational.value.summary.orders_received || operational.value.summary.orders_delivered || operational.value.summary.orders_cancelled || operational.value.summary.active_orders
)))
const maxStatus = computed(() => Math.max(1, ...(operational.value?.status.map(item => item.count) || [0])))
const maxServices = computed(() => Math.max(1, ...(operational.value?.top_services.map(item => item.uses) || [0])))
const cycleTime = computed(() => {
  const minutes = operational.value?.summary.average_cycle_minutes || 0
  if (!minutes) return '—'
  const hours = Math.floor(minutes / 60)
  const remainder = minutes % 60
  return hours ? `${hours} h ${remainder} min` : `${remainder} min`
})

async function loadReports () {
  errorMessage.value = ''
  if (!filters.dateFrom || !filters.dateTo || filters.dateFrom > filters.dateTo) {
    errorMessage.value = 'Selecciona un periodo válido: la fecha inicial no puede superar la final.'
    return
  }
  loading.value = true
  const query = new URLSearchParams({ date_from: filters.dateFrom, date_to: filters.dateTo }).toString()
  try {
    if (props.showFinancial) {
      const response = await auth.request<DashboardReport>(`/admin/dashboard?${query}`)
      operational.value = response.data?.operational || null
      financial.value = response.data?.financial || null
    } else {
      const response = await auth.request<OperationalReport>(`/reports/operations?${query}`)
      operational.value = response.data || null
      financial.value = null
    }
  } catch (error) {
    operational.value = null
    financial.value = null
    errorMessage.value = apiError(error, 'No se pudieron cargar los indicadores').message
  } finally { loading.value = false }
}

function boliviaDateParts (date: Date) {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: 'America/La_Paz',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).formatToParts(date)
  return Object.fromEntries(parts.map(part => [part.type, part.value])) as Record<'year' | 'month' | 'day', string>
}
const statusLabels: Record<string, string> = { PENDING: 'Pendientes', IN_PROGRESS: 'En proceso', QUALITY_CONTROL: 'Control de calidad', FINISHED: 'Listas para retiro' }
const methodLabels: Record<string, string> = { CASH: 'Efectivo', QR: 'QR', BANK_TRANSFER: 'Transferencia', CARD: 'Tarjeta' }
const statusLabel = (status: string) => statusLabels[status] || status
const methodLabel = (method: string) => methodLabels[method] || method
const barWidth = (value: number, maximum: number) => `${Math.max(value ? 8 : 0, Math.round((value / maximum) * 100))}%`
const money = (value: string) => new Intl.NumberFormat('es-BO', { style: 'currency', currency: 'BOB', minimumFractionDigits: 2 }).format(Number(value || 0))
const formatDate = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeZone: 'UTC' }).format(new Date(`${value}T00:00:00Z`))
const csvCell = (value: unknown) => `"${String(value ?? '').replaceAll('"', '""')}"`
function downloadCSV (filename: string, rows: unknown[][]) {
  const body = '\uFEFF' + rows.map(row => row.map(csvCell).join(';')).join('\r\n')
  const url = URL.createObjectURL(new Blob([body], { type: 'text/csv;charset=utf-8' }))
  const anchor = document.createElement('a')
  anchor.href = url; anchor.download = filename; anchor.click(); URL.revokeObjectURL(url)
}
function exportOperational () {
  if (!operational.value) return
  downloadCSV(`camell-wash-operacion-${filters.dateFrom}-${filters.dateTo}.csv`, [
    ['Fecha', 'Órdenes recibidas', 'Órdenes entregadas'],
    ...operational.value.daily.map(item => [item.date, item.received, item.delivered])
  ])
}
function exportFinancial () {
  if (!financial.value) return
  downloadCSV(`camell-wash-finanzas-${filters.dateFrom}-${filters.dateTo}.csv`, [
    ['Fecha', 'Pagos válidos', 'Ingreso BOB'],
    ...financial.value.daily.map(item => [item.date, item.payments, item.amount])
  ])
}

onMounted(loadReports)
</script>

<style scoped>
.report-heading, .section-heading, .error-recovery { display: flex; align-items: center; justify-content: space-between; gap: var(--space-4); }
.timezone-badge, .section-kicker { color: var(--color-primary); font-size: .78rem; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; }
.timezone-badge { display: inline-flex; align-items: center; gap: var(--space-2); padding: .55rem .75rem; border: 1px solid rgba(255,85,0,.22); border-radius: 999px; background: var(--color-primary-soft); white-space: nowrap; }
.report-filters { display: grid; grid-template-columns: repeat(2, minmax(10rem, 13rem)) auto; align-items: end; gap: var(--space-4); margin-bottom: var(--space-6); padding: var(--space-4); border: 1px solid var(--color-border-soft); border-radius: var(--radius-lg); background: var(--color-surface); }
.report-content { display: grid; gap: var(--space-10); }
.section-heading { margin-bottom: var(--space-5); }
.section-heading h2 { margin: .2rem 0 0; font-size: clamp(1.45rem, 3vw, 2rem); }
.metric-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-4); margin-bottom: var(--space-5); }
.metric-card { min-width: 0; padding: var(--space-5); border: 1px solid var(--color-border-soft); border-radius: var(--radius-lg); background: linear-gradient(145deg, rgba(24,28,29,.96), rgba(17,21,23,.94)); box-shadow: var(--shadow-sm); }
.metric-card > i { display: block; margin-bottom: var(--space-4); color: var(--color-primary); font-size: 1.25rem; }
.metric-card span, .metric-card small { display: block; color: var(--color-text-muted); }
.metric-card strong { display: block; margin: .35rem 0; color: var(--color-text); font-size: clamp(1.75rem, 4vw, 2.5rem); font-variant-numeric: tabular-nums; line-height: 1.05; }
.metric-card small { font-size: .78rem; }
.metric-card--attention { border-color: rgba(245,196,81,.32); }
.metric-card--danger { border-color: rgba(255,116,116,.34); }
.metric-card--muted strong { color: var(--color-text-muted); }
.metric-card .money { color: var(--color-primary); font-size: clamp(1.45rem, 3vw, 2.2rem); }
.report-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-5); margin-bottom: var(--space-5); }
.bar-list { display: grid; gap: var(--space-4); }
.bar-label { display: flex; justify-content: space-between; gap: var(--space-3); margin-bottom: var(--space-2); }
.bar-label span { overflow-wrap: anywhere; }
.bar-label strong { font-variant-numeric: tabular-nums; }
.bar-track { height: .55rem; overflow: hidden; border-radius: 999px; background: rgba(248,249,242,.07); }
.bar-track span { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, var(--color-primary), var(--color-secondary)); }
.payment-methods { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-3); }
.payment-method { display: flex; justify-content: space-between; gap: var(--space-4); padding: var(--space-4); border: 1px solid var(--color-border-soft); border-radius: var(--radius-md); background: rgba(248,249,242,.025); }
.payment-method span, .payment-method small { display: block; }
.payment-method small { margin-top: .2rem; color: var(--color-text-muted); }
.payment-method strong, .money-cell { color: var(--color-primary); font-variant-numeric: tabular-nums; white-space: nowrap; }
.data-card { margin-top: var(--space-5); }
@media (max-width: 64rem) { .metric-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 48rem) {
  .report-heading, .section-heading, .error-recovery { align-items: flex-start; flex-direction: column; }
  .report-filters, .report-grid, .payment-methods { grid-template-columns: 1fr; }
  .report-filters .p-button { width: 100%; }
}
@media (max-width: 34rem) { .metric-grid { grid-template-columns: 1fr; } }
</style>
