<template>
  <section data-testid="order-workspace">
    <div class="page-heading">
      <h1>Órdenes de lavado</h1>
      <p v-if="canManage">Gestiona el trabajo diario, asignaciones, precios congelados y estados de cada orden.</p>
      <p v-else>Consulta únicamente tus órdenes asignadas y registra las transiciones operativas permitidas.</p>
    </div>

    <Message v-if="isWashStaff" severity="info" :closable="false">Solo se muestran órdenes donde estás asignado.</Message>
    <div class="toolbar order-toolbar">
      <div class="cluster order-filters">
        <div class="form-field"><label for="order-status-filter">Estado</label><Select id="order-status-filter" v-model="filters.status" :options="statusOptions" option-label="label" option-value="value" placeholder="Todos" show-clear /></div>
        <div class="form-field"><label for="order-date-from">Desde</label><InputText id="order-date-from" v-model="filters.date_from" type="date" /></div>
        <div class="form-field"><label for="order-date-to">Hasta</label><InputText id="order-date-to" v-model="filters.date_to" type="date" /></div>
        <Button label="Filtrar" icon="pi pi-search" :loading="loading" @click="applyFilters" />
        <Button label="Hoy" icon="pi pi-calendar" severity="secondary" outlined @click="filterToday" />
      </div>
      <Button v-if="canManage" label="Nueva orden" icon="pi pi-plus" @click="openCreate" />
    </div>

    <Message v-if="activeContext" severity="secondary" :closable="false">
      Historial filtrado por {{ activeContext }}.
      <Button label="Quitar filtro" size="small" outlined @click="clearContext" />
    </Message>
    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }} <Button label="Reintentar" size="small" outlined @click="loadOrders" /></Message>
    <LoadingState v-else-if="loading" />
    <EmptyState v-else-if="orders.length === 0" title="No hay órdenes para mostrar" description="Ajusta los filtros o registra la primera orden del período." />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="orders" data-key="id" striped-rows>
          <Column header="Orden"><template #body="{ data }"><strong>#{{ data.id }}</strong><small class="table-secondary">{{ formatDate(data.received_at) }}</small></template></Column>
          <Column header="Cliente"><template #body="{ data }">{{ data.vehicle.client_name || `Cliente #${data.vehicle.client_id}` }}</template></Column>
          <Column header="Vehículo"><template #body="{ data }">{{ vehicleLabel(data.vehicle) }}<small class="table-secondary">{{ data.vehicle.vehicle_type.name }}</small></template></Column>
          <Column header="Estado"><template #body="{ data }"><Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" /></template></Column>
          <Column header="Entrega estimada"><template #body="{ data }">{{ formatDate(data.estimated_delivery_at) }}</template></Column>
          <Column header="Total"><template #body="{ data }"><strong>{{ data.total }}</strong></template></Column>
          <Column header="Acciones"><template #body="{ data }"><Button label="Ver detalle" icon="pi pi-eye" size="small" outlined @click="openDetail(data.id)" /></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="formVisible" modal :header="editingOrder ? `Editar orden #${editingOrder.id}` : 'Crear orden'" :style="formDialogStyle" @hide="resetForm">
      <Message v-if="formError" severity="error" class="form-error" :closable="false">{{ formError }}</Message>
      <Message v-if="editingOrder?.status === 'IN_PROGRESS'" severity="info" :closable="false">En proceso solo permite ajustar entrega estimada y personal asignado.</Message>
      <Message v-if="editingOrderPaid" severity="info" :closable="false">La orden está pagada: vehículo y servicios quedan congelados. Solo puedes ajustar la entrega estimada y el personal.</Message>
      <form id="order-form" class="stack" novalidate @submit.prevent="saveOrder">
        <div v-if="(!editingOrder || editingOrder.status === 'PENDING') && !editingOrderPaid" class="detail-grid">
          <div class="form-field">
            <label for="order-client">Cliente</label>
            <Select id="order-client" v-model="form.client_id" :options="clientOptions" option-label="label" option-value="value" placeholder="Selecciona un cliente" filter :invalid="Boolean(formErrors.client_id)" @change="onClientChange" />
            <small v-if="formErrors.client_id" class="field-error">{{ formErrors.client_id }}</small>
          </div>
          <div class="form-field">
            <label for="order-vehicle">Vehículo</label>
            <Select id="order-vehicle" v-model="form.vehicle_id" :options="vehicleOptions" option-label="label" option-value="value" placeholder="Selecciona un vehículo" filter :disabled="!form.client_id" :invalid="Boolean(formErrors.vehicle_id)" @change="onVehicleChange" />
            <small v-if="formErrors.vehicle_id" class="field-error">{{ formErrors.vehicle_id }}</small>
          </div>
        </div>
        <div v-if="(!editingOrder || editingOrder.status === 'PENDING') && !editingOrderPaid" class="form-field">
          <label for="order-services">Servicios con precio vigente</label>
          <MultiSelect id="order-services" v-model="form.service_ids" :options="serviceOptions" option-label="label" option-value="value" placeholder="Selecciona uno o varios servicios" filter display="chip" :disabled="!form.vehicle_id" :invalid="Boolean(formErrors.service_ids)" />
          <small v-if="serviceOptions.length === 0 && form.vehicle_id" class="form-help">Este tipo de vehículo no tiene precios configurados.</small>
          <small v-if="formErrors.service_ids" class="field-error">{{ formErrors.service_ids }}</small>
        </div>
        <div class="detail-grid">
          <div class="form-field">
            <label for="order-staff">Personal de lavado</label>
            <MultiSelect id="order-staff" v-model="form.staff_ids" :options="staffOptions" option-label="label" option-value="value" placeholder="Asigna al menos una persona" filter display="chip" :invalid="Boolean(formErrors.staff_ids)" />
            <small v-if="formErrors.staff_ids" class="field-error">{{ formErrors.staff_ids }}</small>
          </div>
          <div class="form-field">
            <label for="order-estimated">Entrega estimada</label>
            <InputText id="order-estimated" v-model="form.estimated_delivery_at" type="datetime-local" :invalid="Boolean(formErrors.estimated_delivery_at)" />
            <small class="form-help">Hora operativa de La Paz.</small>
            <small v-if="formErrors.estimated_delivery_at" class="field-error">{{ formErrors.estimated_delivery_at }}</small>
          </div>
        </div>
        <div v-if="(!editingOrder || editingOrder.status === 'PENDING') && !editingOrderPaid" class="order-total-preview"><span>Total calculado por el servidor</span><strong>{{ selectedTotal }}</strong></div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="saving" @click="formVisible = false" />
        <Button form="order-form" type="submit" label="Guardar orden" icon="pi pi-save" :loading="saving" />
      </template>
    </Dialog>

    <Dialog v-model:visible="detailVisible" modal :header="selected ? `Orden #${selected.id}` : 'Detalle de orden'" :style="detailDialogStyle" @hide="selected = null; hasValidPayment = null">
      <LoadingState v-if="detailLoading" />
      <template v-else-if="selected">
        <Message v-if="detailError" severity="error" class="form-error" :closable="false">{{ detailError }}</Message>
        <div class="order-detail-header">
          <Tag :value="statusLabel(selected.status)" :severity="statusSeverity(selected.status)" />
          <div class="cluster">
            <Button v-if="canEditSelected" label="Editar" icon="pi pi-pencil" outlined @click="openEdit" />
            <Button v-if="canCancelSelected" label="Cancelar orden" icon="pi pi-times" severity="danger" outlined @click="cancelVisible = true" />
            <Button v-if="canAdvanceSelected" :label="advanceLabel" icon="pi pi-arrow-right" @click="confirmAdvance" />
          </div>
        </div>
        <dl class="detail-grid order-summary">
          <div class="detail-item"><dt>Cliente</dt><dd>{{ selected.vehicle.client_name || `Cliente #${selected.vehicle.client_id}` }}</dd></div>
          <div class="detail-item"><dt>Vehículo</dt><dd>{{ vehicleLabel(selected.vehicle) }} · {{ selected.vehicle.vehicle_type.name }}</dd></div>
          <div class="detail-item"><dt>Recepción</dt><dd>{{ formatDate(selected.received_at) }}</dd></div>
          <div class="detail-item"><dt>Entrega estimada</dt><dd>{{ formatDate(selected.estimated_delivery_at) }}</dd></div>
          <div class="detail-item"><dt>Registrada por</dt><dd>{{ userName(selected.registered_by) }}</dd></div>
          <div class="detail-item"><dt>Versión</dt><dd>{{ selected.version }}</dd></div>
        </dl>
        <div class="detail-section"><h3>Servicios congelados</h3><div class="table-wrap"><DataTable :value="selected.services"><Column field="service_name_snapshot" header="Servicio" /><Column field="price_snapshot" header="Precio congelado" /></DataTable></div><div class="detail-total"><span>Total</span><strong>{{ selected.total }}</strong></div></div>
        <PaymentPanel v-if="canManage" :order="selected" @order-updated="handlePaymentOrderUpdated" @payment-state="hasValidPayment = $event" />
        <div class="detail-section"><h3>Personal asignado</h3><div class="cluster"><Tag v-for="person in selected.staff" :key="person.id" :value="userName(person)" severity="secondary" /></div></div>
        <div class="detail-section"><h3>Historial de estados</h3><ol class="history-list"><li v-for="entry in selected.history" :key="entry.id"><span class="history-dot" /><div><strong>{{ entry.from_status ? `${statusLabel(entry.from_status)} → ` : '' }}{{ statusLabel(entry.to_status) }}</strong><p>{{ formatDate(entry.changed_at) }} · {{ userName(entry.actor) }}</p><p v-if="entry.reason">{{ entry.reason }}</p></div></li></ol></div>
      </template>
    </Dialog>

    <Dialog v-model:visible="cancelVisible" modal header="Cancelar orden" :style="smallDialogStyle">
      <Message severity="warn" :closable="false">La cancelación es terminal y quedará registrada en el historial.</Message>
      <div class="form-field"><label for="cancel-reason">Motivo obligatorio</label><Textarea id="cancel-reason" v-model="cancelReason" rows="4" maxlength="500" auto-resize /><small v-if="cancelError" class="field-error">{{ cancelError }}</small></div>
      <template #footer><Button label="Volver" severity="secondary" outlined @click="cancelVisible = false" /><Button label="Confirmar cancelación" severity="danger" :loading="transitioning" @click="cancelOrder" /></template>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import MultiSelect from 'primevue/multiselect'
import Paginator from 'primevue/paginator'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import PaymentPanel from '~/components/payments/PaymentPanel.vue'
import type { User } from '~/types/auth'
import type { ClientListItem, ClientPage, Vehicle, VehiclePage } from '~/types/customers'
import type { Order, OrderPage, OrderServiceOption, OrderStatus } from '~/types/orders'
import { apiError } from '~/utils/api-error'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const confirm = useConfirm()
const toast = useToast()
const canManage = computed(() => auth.user?.role === 'admin' || auth.user?.role === 'manager')
const isWashStaff = computed(() => auth.user?.role === 'wash_staff')
const orders = ref<Order[]>([])
const loading = ref(false)
const errorMessage = ref('')
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const today = laPazDateInput(new Date())
const initialClientID = numericQuery(route.query.client_id)
const initialVehicleID = numericQuery(route.query.vehicle_id)
const hasInitialContext = Boolean(initialClientID || initialVehicleID)
const filters = reactive<{ status: OrderStatus | null, date_from: string, date_to: string, client_id: number | null, vehicle_id: number | null }>({
  status: null, date_from: hasInitialContext ? '' : today, date_to: hasInitialContext ? '' : today, client_id: initialClientID, vehicle_id: initialVehicleID
})
const statusOptions = [
  { label: 'Pendiente', value: 'PENDING' }, { label: 'En proceso', value: 'IN_PROGRESS' },
  { label: 'Control de calidad', value: 'QUALITY_CONTROL' }, { label: 'Terminado', value: 'FINISHED' },
  { label: 'Entregado', value: 'DELIVERED' }, { label: 'Cancelado', value: 'CANCELLED' }
] satisfies { label: string, value: OrderStatus }[]
const activeContext = computed(() => filters.client_id ? `cliente #${filters.client_id}` : filters.vehicle_id ? `vehículo #${filters.vehicle_id}` : '')

const formVisible = ref(false)
const saving = ref(false)
const formError = ref('')
const formErrors = reactive<Record<string, string>>({})
const editingOrder = ref<Order | null>(null)
const form = reactive({ client_id: null as number | null, vehicle_id: null as number | null, service_ids: [] as number[], staff_ids: [] as number[], estimated_delivery_at: '' })
const clients = ref<ClientListItem[]>([])
const vehicles = ref<Vehicle[]>([])
const staff = ref<User[]>([])
const availableServices = ref<OrderServiceOption[]>([])
const clientOptions = computed(() => clients.value.map(item => ({ label: item.full_name, value: item.id })))
const vehicleOptions = computed(() => vehicles.value.map(item => ({ label: `${vehicleLabel(item)} · ${item.vehicle_type.name}`, value: item.id })))
const staffOptions = computed(() => staff.value.map(item => ({ label: userName(item), value: item.id })))
const serviceOptions = computed(() => availableServices.value.map(item => ({ label: `${item.service_name} · ${item.category_name} · ${item.price}`, value: item.service_id })))
const selectedTotal = computed(() => centsToPrice(form.service_ids.reduce((sum, id) => sum + priceCents(availableServices.value.find(item => item.service_id === id)?.price || '0.00'), 0)))

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const selected = ref<Order | null>(null)
const hasValidPayment = ref<boolean | null>(null)
const transitioning = ref(false)
const cancelVisible = ref(false)
const cancelReason = ref('')
const cancelError = ref('')
const formDialogStyle = { width: 'min(58rem, calc(100vw - 2rem))' }
const detailDialogStyle = { width: 'min(70rem, calc(100vw - 2rem))' }
const smallDialogStyle = { width: 'min(36rem, calc(100vw - 2rem))' }

const loadOrders = async () => {
  loading.value = true; errorMessage.value = ''
  try {
    const query = new URLSearchParams({ page: String(pagination.page), limit: String(pagination.limit) })
    if (filters.status) query.set('status', filters.status)
    if (filters.date_from) query.set('date_from', filters.date_from)
    if (filters.date_to) query.set('date_to', filters.date_to)
    if (filters.client_id) query.set('client_id', String(filters.client_id))
    if (filters.vehicle_id) query.set('vehicle_id', String(filters.vehicle_id))
    const response = await auth.request<OrderPage>(`/orders?${query}`)
    orders.value = response.data?.orders || []
    Object.assign(pagination, response.data?.pagination)
  } catch (error) { orders.value = []; errorMessage.value = apiError(error, 'No se pudieron cargar las órdenes').message }
  finally { loading.value = false }
}
const applyFilters = () => { pagination.page = 1; void loadOrders() }
const filterToday = () => { filters.date_from = today; filters.date_to = today; pagination.page = 1; void loadOrders() }
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void loadOrders() }
const clearContext = async () => { filters.client_id = null; filters.vehicle_id = null; await router.replace({ query: {} }); applyFilters() }

const loadFormOptions = async () => {
  const [clientResponse, staffResponse] = await Promise.all([
    auth.request<ClientPage>('/clients?limit=100'), auth.request<{ staff: User[] }>('/operations/wash-staff')
  ])
  clients.value = clientResponse.data?.clients || []
  staff.value = staffResponse.data?.staff || []
}
const openCreate = async () => {
  resetForm(); formVisible.value = true
  try { await loadFormOptions() } catch (error) { formError.value = apiError(error).message }
}
const onClientChange = async () => {
  form.vehicle_id = null; form.service_ids = []; vehicles.value = []; availableServices.value = []
  if (!form.client_id) return
  try {
    const response = await auth.request<VehiclePage>(`/clients/${form.client_id}/vehicles?limit=100`)
    vehicles.value = response.data?.vehicles || []
  } catch (error) { formError.value = apiError(error).message }
}
const onVehicleChange = async () => {
  form.service_ids = []; availableServices.value = []
  const vehicle = vehicles.value.find(item => item.id === form.vehicle_id)
  if (!vehicle) return
  try {
    const response = await auth.request<{ services: OrderServiceOption[] }>(`/operations/service-options?vehicle_type_id=${vehicle.vehicle_type_id}`)
    availableServices.value = response.data?.services || []
  } catch (error) { formError.value = apiError(error).message }
}
const resetForm = () => {
  editingOrder.value = null; Object.assign(form, { client_id: null, vehicle_id: null, service_ids: [], staff_ids: [], estimated_delivery_at: defaultEstimatedInput() })
  clients.value = []; vehicles.value = []; staff.value = []; availableServices.value = []; formError.value = ''; clearErrors(formErrors)
}
const openEdit = async () => {
  if (!selected.value || !canManage.value) return
  editingOrder.value = selected.value; formVisible.value = true; formError.value = ''; clearErrors(formErrors)
  try {
    await loadFormOptions()
    Object.assign(form, { client_id: selected.value.vehicle.client_id, vehicle_id: selected.value.vehicle_id, service_ids: selected.value.services.map(item => item.service_id), staff_ids: selected.value.staff.map(item => item.id), estimated_delivery_at: laPazDateTimeInput(new Date(selected.value.estimated_delivery_at)) })
    if (selected.value.status === 'PENDING' && !editingOrderPaid.value) {
      const vehicleResponse = await auth.request<VehiclePage>(`/clients/${form.client_id}/vehicles?limit=100`)
      vehicles.value = vehicleResponse.data?.vehicles || []
      const serviceResponse = await auth.request<{ services: OrderServiceOption[] }>(`/operations/service-options?vehicle_type_id=${selected.value.vehicle.vehicle_type_id}`)
      availableServices.value = serviceResponse.data?.services || []
    }
  } catch (error) { formError.value = apiError(error).message }
}
const saveOrder = async () => {
  clearErrors(formErrors); formError.value = ''
  if (!editingOrder.value || (editingOrder.value.status === 'PENDING' && !editingOrderPaid.value)) {
    if (!form.client_id) formErrors.client_id = 'Selecciona un cliente'
    if (!form.vehicle_id) formErrors.vehicle_id = 'Selecciona un vehículo'
    if (!form.service_ids.length) formErrors.service_ids = 'Selecciona al menos un servicio con precio'
  }
  if (!form.staff_ids.length) formErrors.staff_ids = 'Asigna al menos una persona'
  if (!form.estimated_delivery_at) formErrors.estimated_delivery_at = 'Selecciona una fecha y hora'
  if (Object.keys(formErrors).length) return
  saving.value = true
  try {
    const estimated = `${form.estimated_delivery_at}:00-04:00`
    const body = editingOrder.value
      ? editingOrder.value.status === 'PENDING' && !editingOrderPaid.value
        ? { version: editingOrder.value.version, vehicle_id: form.vehicle_id, service_ids: form.service_ids, staff_ids: form.staff_ids, estimated_delivery_at: estimated }
        : { version: editingOrder.value.version, staff_ids: form.staff_ids, estimated_delivery_at: estimated }
      : { vehicle_id: form.vehicle_id, service_ids: form.service_ids, staff_ids: form.staff_ids, estimated_delivery_at: estimated }
    const path = editingOrder.value ? `/orders/${editingOrder.value.id}` : '/orders'
    const response = await auth.request<Order>(path, { method: editingOrder.value ? 'PATCH' : 'POST', body })
    toast.add({ severity: 'success', summary: editingOrder.value ? 'Orden actualizada' : 'Orden creada', life: 3500 })
    formVisible.value = false
    if (response.data && selected.value?.id === response.data.id) selected.value = response.data
    await loadOrders()
  } catch (error) { const parsed = apiError(error); formError.value = parsed.message; Object.assign(formErrors, parsed.fields || {}) }
  finally { saving.value = false }
}

const openDetail = async (id: number) => {
  detailVisible.value = true; detailLoading.value = true; detailError.value = ''; hasValidPayment.value = null
  try { const response = await auth.request<Order>(`/orders/${id}`); selected.value = response.data || null }
  catch (error) { detailError.value = apiError(error).message }
  finally { detailLoading.value = false }
}
const nextStatus = computed<OrderStatus | null>(() => selected.value ? ({ PENDING: 'IN_PROGRESS', IN_PROGRESS: 'QUALITY_CONTROL', QUALITY_CONTROL: 'FINISHED', FINISHED: 'DELIVERED' } as Partial<Record<OrderStatus, OrderStatus>>)[selected.value.status] || null : null)
const editingOrderPaid = computed(() => Boolean(editingOrder.value && hasValidPayment.value === true))
const canEditSelected = computed(() => canManage.value && hasValidPayment.value !== null && (selected.value?.status === 'PENDING' || selected.value?.status === 'IN_PROGRESS'))
const canCancelSelected = computed(() => canManage.value && hasValidPayment.value === false && ['PENDING', 'IN_PROGRESS', 'QUALITY_CONTROL'].includes(selected.value?.status || ''))
const canAdvanceSelected = computed(() => {
  if (!selected.value || !nextStatus.value || nextStatus.value === 'DELIVERED') return false
  if (canManage.value) return true
  return isWashStaff.value && (nextStatus.value === 'IN_PROGRESS' || nextStatus.value === 'QUALITY_CONTROL')
})
const advanceLabel = computed(() => nextStatus.value ? `Pasar a ${statusLabel(nextStatus.value)}` : '')
const confirmAdvance = () => {
  if (!selected.value || !nextStatus.value) return
  confirm.require({ message: `La orden pasará a ${statusLabel(nextStatus.value)} y se agregará una entrada al historial.`, header: 'Confirmar transición', icon: 'pi pi-arrow-right', acceptLabel: 'Confirmar', rejectLabel: 'Volver', accept: advanceOrder })
}
const advanceOrder = async () => {
  if (!selected.value || !nextStatus.value) return
  transitioning.value = true; detailError.value = ''
  try {
    const response = await auth.request<Order>(`/orders/${selected.value.id}/transitions`, { method: 'POST', body: { version: selected.value.version, target: nextStatus.value, reason: '' } })
    if (response.data) selected.value = response.data
    toast.add({ severity: 'success', summary: 'Estado actualizado', life: 3000 }); await loadOrders()
  } catch (error) { detailError.value = apiError(error).message; await reloadSelected() }
  finally { transitioning.value = false }
}
const cancelOrder = async () => {
  if (!selected.value) return
  cancelError.value = ''
  if (cancelReason.value.trim().length < 5) { cancelError.value = 'Describe el motivo con al menos 5 caracteres.'; return }
  transitioning.value = true
  try {
    const response = await auth.request<Order>(`/orders/${selected.value.id}/cancel`, { method: 'POST', body: { version: selected.value.version, reason: cancelReason.value } })
    if (response.data) selected.value = response.data
    cancelVisible.value = false; cancelReason.value = ''; toast.add({ severity: 'success', summary: 'Orden cancelada', life: 3000 }); await loadOrders()
  } catch (error) { cancelError.value = apiError(error).message; await reloadSelected() }
  finally { transitioning.value = false }
}
const handlePaymentOrderUpdated = (order: Order) => { selected.value = order; void loadOrders() }
const reloadSelected = async () => { if (selected.value) { const response = await auth.request<Order>(`/orders/${selected.value.id}`); if (response.data) selected.value = response.data } }

function clearErrors(target: Record<string, string>) { Object.keys(target).forEach(key => delete target[key]) }
function numericQuery(value: unknown) { const raw = Array.isArray(value) ? value[0] : value; const parsed = Number(raw); return Number.isInteger(parsed) && parsed > 0 ? parsed : null }
function statusLabel(status: OrderStatus) { return statusOptions.find(item => item.value === status)?.label || status }
function statusSeverity(status: OrderStatus) { return status === 'CANCELLED' ? 'danger' : status === 'DELIVERED' ? 'success' : status === 'FINISHED' ? 'info' : status === 'QUALITY_CONTROL' ? 'warn' : status === 'IN_PROGRESS' ? 'secondary' : 'contrast' }
function vehicleLabel(vehicle: Vehicle) { return vehicle.plate || `Sin placa · ID #${vehicle.id}` }
function userName(user: User) { return `${user.first_name} ${user.last_name}` }
function formatDate(value: string) { return new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value)) }
function laPazParts(date: Date) { return Object.fromEntries(new Intl.DateTimeFormat('en-CA', { timeZone: 'America/La_Paz', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }).formatToParts(date).map(part => [part.type, part.value])) }
function laPazDateInput(date: Date) { const p = laPazParts(date); return `${p.year}-${p.month}-${p.day}` }
function laPazDateTimeInput(date: Date) { const p = laPazParts(date); return `${p.year}-${p.month}-${p.day}T${p.hour}:${p.minute}` }
function defaultEstimatedInput() { return laPazDateTimeInput(new Date(Date.now() + 2 * 60 * 60 * 1000)) }
function priceCents(value: string) { const [whole = '0', fraction = '00'] = value.split('.'); return Number(whole) * 100 + Number(fraction.padEnd(2, '0').slice(0, 2)) }
function centsToPrice(value: number) { return `${Math.floor(value / 100)}.${String(value % 100).padStart(2, '0')}` }

onMounted(loadOrders)
</script>

<style scoped>
.order-toolbar { margin-top: var(--space-5); }
.order-filters { align-items: end; }
.order-filters .form-field { min-width: 10rem; }
.table-secondary { display: block; color: var(--color-text-muted); }
.order-total-preview, .detail-total { display: flex; align-items: center; justify-content: space-between; gap: var(--space-4); padding: var(--space-4); border: 1px solid rgba(255, 85, 0, .22); border-radius: var(--radius-md); background: var(--color-primary-soft); }
.order-total-preview strong, .detail-total strong { color: var(--color-primary); font-size: 1.4rem; }
.order-detail-header { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: var(--space-4); margin-bottom: var(--space-5); }
.order-summary { margin: var(--space-5) 0; }
.detail-section { margin-top: var(--space-8); padding-top: var(--space-5); border-top: 1px solid var(--color-border-soft); }
.detail-section h3 { margin-bottom: var(--space-4); }
.detail-total { margin-top: var(--space-3); }
.history-list { list-style: none; margin: 0; padding: 0; }
.history-list li { position: relative; display: grid; grid-template-columns: 1rem 1fr; gap: var(--space-3); padding-bottom: var(--space-5); }
.history-list li:not(:last-child)::before { content: ''; position: absolute; left: .35rem; top: 1rem; bottom: 0; width: 2px; background: var(--color-border); }
.history-dot { width: .8rem; height: .8rem; margin-top: .35rem; border-radius: 50%; background: var(--color-primary); box-shadow: 0 0 12px rgba(255,85,0,.35); }
.history-list p { margin: .2rem 0 0; color: var(--color-text-muted); }
@media (max-width: 40rem) { .order-filters > *, .order-toolbar > .p-button { width: 100%; } }
</style>
