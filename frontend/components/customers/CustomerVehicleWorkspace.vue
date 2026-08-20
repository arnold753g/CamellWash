<template>
  <section data-testid="customer-vehicle-workspace">
    <div class="page-heading">
      <h1>Clientes y vehículos</h1>
      <p>Registra a cada cliente y sus vehículos. El CI y la placa son opcionales; los datos ausentes nunca se sustituyen por identificadores ficticios.</p>
    </div>

    <div class="toolbar">
      <div class="cluster customer-filters">
        <div class="form-field">
          <label for="client-search">Buscar cliente</label>
          <InputText id="client-search" v-model.trim="search" placeholder="Nombre, CI o teléfono" @keyup.enter="applySearch" />
        </div>
        <Button label="Buscar" icon="pi pi-search" :loading="loading" @click="applySearch" />
        <Button v-if="search" label="Limpiar" icon="pi pi-filter-slash" severity="secondary" outlined @click="clearSearch" />
      </div>
      <Button label="Nuevo cliente" icon="pi pi-user-plus" @click="openClientCreate" />
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">
      {{ errorMessage }}
      <Button label="Reintentar" size="small" severity="secondary" outlined @click="loadClients" />
    </Message>
    <LoadingState v-if="loading" />
    <EmptyState
      v-else-if="clients.length === 0"
      title="No hay clientes para mostrar"
      :description="search ? 'No se encontraron coincidencias. Ajusta la búsqueda.' : 'Registra el primer cliente para asociar sus vehículos.'"
    />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="clients" data-key="id" striped-rows>
          <Column field="full_name" header="Cliente" />
          <Column header="CI"><template #body="{ data }"><span :class="{ muted: !data.ci_provided }">{{ data.ci }}</span></template></Column>
          <Column field="phone" header="Teléfono" />
          <Column header="Vehículos"><template #body="{ data }"><Tag :value="String(data.vehicle_count)" severity="secondary" /></template></Column>
          <Column header="Acciones"><template #body="{ data }"><Button label="Ver ficha" icon="pi pi-eye" size="small" outlined @click="openDetail(data.id)" /></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="clientDialogVisible" modal :header="editingClientID ? 'Editar cliente' : 'Registrar cliente'" :style="dialogStyle" @hide="resetClientForm">
      <Message v-if="clientFormError" severity="error" class="form-error" :closable="false">{{ clientFormError }}</Message>
      <form id="client-form" class="stack" novalidate @submit.prevent="saveClient">
        <div class="form-field">
          <label for="client-name">Nombre completo</label>
          <InputText id="client-name" v-model="clientForm.full_name" autocomplete="name" :invalid="Boolean(clientErrors.full_name)" autofocus />
          <small v-if="clientErrors.full_name" class="field-error" role="alert">{{ clientErrors.full_name }}</small>
        </div>
        <div class="detail-grid">
          <div class="form-field">
            <label for="client-ci">CI (opcional)</label>
            <InputText id="client-ci" v-model="clientForm.ci" autocomplete="off" :invalid="Boolean(clientErrors.ci)" placeholder="Dejar vacío si no fue proporcionado" />
            <small class="form-help">Si queda vacío, se guardará como dato ausente y en reportes se mostrará 00000000.</small>
            <small v-if="clientErrors.ci" class="field-error" role="alert">{{ clientErrors.ci }}</small>
          </div>
          <div class="form-field">
            <label for="client-phone">Teléfono</label>
            <InputText id="client-phone" v-model="clientForm.phone" inputmode="tel" autocomplete="tel" :invalid="Boolean(clientErrors.phone)" placeholder="+591 70000000" />
            <small v-if="clientErrors.phone" class="field-error" role="alert">{{ clientErrors.phone }}</small>
          </div>
        </div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="savingClient" @click="clientDialogVisible = false" />
        <Button form="client-form" type="submit" label="Guardar cliente" icon="pi pi-save" :loading="savingClient" />
      </template>
    </Dialog>

    <Dialog v-model:visible="detailVisible" modal header="Ficha del cliente" :style="detailDialogStyle" @hide="resetDetail">
      <LoadingState v-if="detailLoading" />
      <template v-else-if="selectedClient">
        <Message v-if="detailError" severity="error" class="form-error" :closable="false">{{ detailError }}</Message>
        <div class="detail-heading">
          <div><h2>{{ selectedClient.full_name }}</h2><p class="muted">Cliente interno #{{ selectedClient.id }}</p></div>
          <div class="cluster">
            <Button label="Historial de órdenes" icon="pi pi-history" outlined @click="openClientOrders" />
            <Button label="Editar cliente" icon="pi pi-pencil" outlined @click="openClientEdit" />
          </div>
        </div>
        <dl class="detail-grid customer-detail-grid">
          <div class="detail-item"><dt>CI</dt><dd>{{ selectedClient.ci }}</dd></div>
          <div class="detail-item"><dt>Teléfono</dt><dd>{{ selectedClient.phone }}</dd></div>
          <div class="detail-item"><dt>Vehículos registrados</dt><dd>{{ selectedClient.vehicle_count }}</dd></div>
        </dl>

        <div class="section-heading">
          <div><h3>Vehículos</h3><p>Los vehículos sin placa se identifican temporalmente por su ID interno.</p></div>
          <Button label="Agregar vehículo" icon="pi pi-plus" @click="openVehicleCreate" />
        </div>
        <Message v-if="vehicleListError" severity="error" :closable="false">{{ vehicleListError }}</Message>
        <LoadingState v-if="vehiclesLoading" />
        <EmptyState v-else-if="vehicles.length === 0" title="Este cliente aún no tiene vehículos" description="Agrega un vehículo; la placa puede quedar vacía, por ejemplo para vehículos de competición." />
        <div v-else class="table-wrap">
          <DataTable :value="vehicles" data-key="id" striped-rows>
            <Column header="Identificación"><template #body="{ data }"><strong>{{ vehicleLabel(data) }}</strong></template></Column>
            <Column header="Tipo"><template #body="{ data }">{{ data.vehicle_type.name }}</template></Column>
            <Column header="Descripción"><template #body="{ data }">{{ vehicleDescription(data) }}</template></Column>
            <Column header="Acciones"><template #body="{ data }"><div class="cluster"><Button label="Órdenes" icon="pi pi-history" size="small" outlined @click="openVehicleOrders(data.id)" /><Button label="Editar" icon="pi pi-pencil" size="small" outlined @click="openVehicleEdit(data)" /></div></template></Column>
          </DataTable>
        </div>
      </template>
    </Dialog>

    <Dialog v-model:visible="vehicleDialogVisible" modal :header="editingVehicleID ? 'Editar vehículo' : 'Agregar vehículo'" :style="dialogStyle" @hide="resetVehicleForm">
      <Message v-if="vehicleFormError" severity="error" class="form-error" :closable="false">{{ vehicleFormError }}</Message>
      <Message v-if="editingVehicleID" severity="info" :closable="false">El propietario no puede modificarse en esta versión.</Message>
      <form id="vehicle-form" class="stack" novalidate @submit.prevent="saveVehicle">
        <div class="form-field">
          <label for="vehicle-owner">Propietario</label>
          <InputText id="vehicle-owner" :model-value="selectedClient?.full_name || ''" disabled />
        </div>
        <div class="detail-grid">
          <div class="form-field">
            <label for="vehicle-type">Tipo de vehículo</label>
            <Select id="vehicle-type" v-model="vehicleForm.vehicle_type_id" :options="vehicleTypeOptions" option-label="label" option-value="value" placeholder="Selecciona un tipo" filter :invalid="Boolean(vehicleErrors.vehicle_type_id)" />
            <small v-if="vehicleErrors.vehicle_type_id" class="field-error" role="alert">{{ vehicleErrors.vehicle_type_id }}</small>
          </div>
          <div class="form-field">
            <label for="vehicle-plate">Placa (opcional)</label>
            <InputText id="vehicle-plate" v-model="vehicleForm.plate" autocomplete="off" :invalid="Boolean(vehicleErrors.plate)" placeholder="Sin placa" />
            <small class="form-help">No se generará una placa ficticia. El sistema usará el ID interno.</small>
            <small v-if="vehicleErrors.plate" class="field-error" role="alert">{{ vehicleErrors.plate }}</small>
          </div>
          <div class="form-field"><label for="vehicle-brand">Marca (opcional)</label><InputText id="vehicle-brand" v-model="vehicleForm.brand" :invalid="Boolean(vehicleErrors.brand)" /></div>
          <div class="form-field"><label for="vehicle-model">Modelo (opcional)</label><InputText id="vehicle-model" v-model="vehicleForm.model" :invalid="Boolean(vehicleErrors.model)" /></div>
          <div class="form-field"><label for="vehicle-color">Color (opcional)</label><InputText id="vehicle-color" v-model="vehicleForm.color" :invalid="Boolean(vehicleErrors.color)" /></div>
        </div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="savingVehicle" @click="vehicleDialogVisible = false" />
        <Button form="vehicle-form" type="submit" label="Guardar vehículo" icon="pi pi-save" :loading="savingVehicle" />
      </template>
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
import Paginator from 'primevue/paginator'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { z } from 'zod'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { CatalogItem } from '~/types/catalogs'
import type { ClientDetail, ClientListItem, ClientPage, Vehicle, VehiclePage } from '~/types/customers'
import { apiError } from '~/utils/api-error'

const auth = useAuthStore()
const toast = useToast()
const dialogStyle = { width: 'min(42rem, calc(100vw - 2rem))' }
const detailDialogStyle = { width: 'min(68rem, calc(100vw - 2rem))' }
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const clients = ref<ClientListItem[]>([])
const search = ref('')
const loading = ref(false)
const errorMessage = ref('')

const clientDialogVisible = ref(false)
const editingClientID = ref<number | null>(null)
const savingClient = ref(false)
const clientFormError = ref('')
const clientErrors = reactive<Record<string, string>>({})
const clientForm = reactive({ full_name: '', ci: '', phone: '' })

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const selectedClient = ref<ClientDetail | null>(null)
const vehicles = ref<Vehicle[]>([])
const vehiclesLoading = ref(false)
const vehicleListError = ref('')

const vehicleDialogVisible = ref(false)
const editingVehicleID = ref<number | null>(null)
const savingVehicle = ref(false)
const vehicleFormError = ref('')
const vehicleErrors = reactive<Record<string, string>>({})
const vehicleForm = reactive({ vehicle_type_id: null as number | null, plate: '', brand: '', model: '', color: '' })
const vehicleTypes = ref<CatalogItem[]>([])
const vehicleTypeOptions = computed(() => vehicleTypes.value.map(item => ({ label: item.name, value: item.id })))

const clientSchema = z.object({
  full_name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(150, 'Usa como máximo 150 caracteres'),
  ci: z.string().trim().refine(value => value === '' || /^(?!00000000$)[A-Za-z0-9 -]{4,30}$/.test(value), 'Usa 4 a 20 letras o números; 00000000 está reservado'),
  phone: z.string().trim().refine(value => /^\+?[0-9 ()-]{7,22}$/.test(value), 'Ingresa entre 7 y 15 dígitos')
})
const vehicleSchema = z.object({
  vehicle_type_id: z.number({ invalid_type_error: 'Selecciona un tipo de vehículo', required_error: 'Selecciona un tipo de vehículo' }).int().positive('Selecciona un tipo de vehículo'),
  plate: z.string().trim().refine(value => value === '' || /^[A-Za-z0-9 -]{3,20}$/.test(value), 'Usa entre 3 y 15 letras o números'),
  brand: z.string().trim().max(100, 'Usa como máximo 100 caracteres'),
  model: z.string().trim().max(100, 'Usa como máximo 100 caracteres'),
  color: z.string().trim().max(100, 'Usa como máximo 100 caracteres')
})

const loadClients = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const query = new URLSearchParams({ page: String(pagination.page), limit: String(pagination.limit) })
    if (search.value) query.set('search', search.value)
    const response = await auth.request<ClientPage>(`/clients?${query}`)
    clients.value = response.data?.clients || []
    Object.assign(pagination, response.data?.pagination)
  } catch (error) {
    clients.value = []
    errorMessage.value = apiError(error, 'No se pudieron cargar los clientes').message
  } finally { loading.value = false }
}
const applySearch = () => { pagination.page = 1; void loadClients() }
const clearSearch = () => { search.value = ''; pagination.page = 1; void loadClients() }
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void loadClients() }

const clearErrors = (target: Record<string, string>) => Object.keys(target).forEach(key => delete target[key])
const setRequestError = (error: unknown, message: { value: string }, fields: Record<string, string>) => {
  const parsed = apiError(error)
  message.value = parsed.message
  Object.assign(fields, parsed.fields || {})
}
const openClientCreate = () => { resetClientForm(); clientDialogVisible.value = true }
const openClientEdit = () => {
  if (!selectedClient.value) return
  editingClientID.value = selectedClient.value.id
  Object.assign(clientForm, { full_name: selectedClient.value.full_name, ci: selectedClient.value.ci_provided ? selectedClient.value.ci : '', phone: selectedClient.value.phone })
  clientDialogVisible.value = true
}
const resetClientForm = () => {
  editingClientID.value = null
  Object.assign(clientForm, { full_name: '', ci: '', phone: '' })
  clientFormError.value = ''
  clearErrors(clientErrors)
}
const saveClient = async () => {
  clientFormError.value = ''
  clearErrors(clientErrors)
  const parsed = clientSchema.safeParse(clientForm)
  if (!parsed.success) {
    for (const issue of parsed.error.issues) clientErrors[String(issue.path[0])] = issue.message
    return
  }
  savingClient.value = true
  try {
    const body = { ...parsed.data, ci: parsed.data.ci || null }
    const response = await auth.request<ClientDetail>(editingClientID.value ? `/clients/${editingClientID.value}` : '/clients', { method: editingClientID.value ? 'PATCH' : 'POST', body })
    toast.add({ severity: 'success', summary: editingClientID.value ? 'Cliente actualizado' : 'Cliente registrado', life: 3500 })
    clientDialogVisible.value = false
    if (response.data && selectedClient.value?.id === response.data.id) selectedClient.value = response.data
    await loadClients()
  } catch (error) { setRequestError(error, clientFormError, clientErrors) }
  finally { savingClient.value = false }
}

const loadVehicleTypes = async () => {
  if (vehicleTypes.value.length) return
  const response = await auth.request<{ vehicle_types: CatalogItem[] }>('/catalogs/vehicle-types')
  vehicleTypes.value = response.data?.vehicle_types || []
}
const loadVehicles = async () => {
  if (!selectedClient.value) return
  vehiclesLoading.value = true
  vehicleListError.value = ''
  try {
    const response = await auth.request<VehiclePage>(`/clients/${selectedClient.value.id}/vehicles?limit=100`)
    vehicles.value = response.data?.vehicles || []
  } catch (error) { vehicles.value = []; vehicleListError.value = apiError(error).message }
  finally { vehiclesLoading.value = false }
}
const openDetail = async (id: number) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  vehicles.value = []
  try {
    const [response] = await Promise.all([auth.request<ClientDetail>(`/clients/${id}`), loadVehicleTypes()])
    selectedClient.value = response.data || null
    await loadVehicles()
  } catch (error) { detailError.value = apiError(error).message }
  finally { detailLoading.value = false }
}
const resetDetail = () => { selectedClient.value = null; vehicles.value = []; detailError.value = ''; vehicleListError.value = '' }

const openVehicleCreate = () => { resetVehicleForm(); vehicleDialogVisible.value = true }
const openVehicleEdit = (vehicle: Vehicle) => {
  editingVehicleID.value = vehicle.id
  Object.assign(vehicleForm, { vehicle_type_id: vehicle.vehicle_type_id, plate: vehicle.plate || '', brand: vehicle.brand, model: vehicle.model, color: vehicle.color })
  vehicleDialogVisible.value = true
}
const resetVehicleForm = () => {
  editingVehicleID.value = null
  Object.assign(vehicleForm, { vehicle_type_id: null, plate: '', brand: '', model: '', color: '' })
  vehicleFormError.value = ''
  clearErrors(vehicleErrors)
}
const saveVehicle = async () => {
  if (!selectedClient.value) return
  vehicleFormError.value = ''
  clearErrors(vehicleErrors)
  const parsed = vehicleSchema.safeParse(vehicleForm)
  if (!parsed.success) {
    for (const issue of parsed.error.issues) vehicleErrors[String(issue.path[0])] = issue.message
    return
  }
  savingVehicle.value = true
  try {
    const body = { ...parsed.data, plate: parsed.data.plate || null }
    await auth.request(editingVehicleID.value ? `/vehicles/${editingVehicleID.value}` : `/clients/${selectedClient.value.id}/vehicles`, { method: editingVehicleID.value ? 'PATCH' : 'POST', body })
    toast.add({ severity: 'success', summary: editingVehicleID.value ? 'Vehículo actualizado' : 'Vehículo agregado', life: 3500 })
    vehicleDialogVisible.value = false
    await loadVehicles()
    const detail = await auth.request<ClientDetail>(`/clients/${selectedClient.value.id}`)
    if (detail.data) selectedClient.value = detail.data
    await loadClients()
  } catch (error) { setRequestError(error, vehicleFormError, vehicleErrors) }
  finally { savingVehicle.value = false }
}

const vehicleLabel = (vehicle: Vehicle) => vehicle.plate || `Sin placa · ID #${vehicle.id}`
const vehicleDescription = (vehicle: Vehicle) => [vehicle.brand, vehicle.model, vehicle.color].filter(Boolean).join(' · ') || 'Sin detalles adicionales'
const ordersPath = computed(() => auth.user?.role === 'admin' ? '/admin/orders' : '/app/orders')
const openClientOrders = () => { if (selectedClient.value) void navigateTo({ path: ordersPath.value, query: { client_id: String(selectedClient.value.id) } }) }
const openVehicleOrders = (vehicleID: number) => { void navigateTo({ path: ordersPath.value, query: { vehicle_id: String(vehicleID) } }) }

onMounted(loadClients)
</script>

<style scoped>
.customer-filters { align-items: end; }
.customer-filters .form-field { min-width: min(24rem, 100%); }
.detail-heading, .section-heading { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: var(--space-4); }
.detail-heading { margin-bottom: var(--space-5); }
.detail-heading h2, .detail-heading p, .section-heading h3, .section-heading p { margin-bottom: 0; }
.section-heading { margin: var(--space-8) 0 var(--space-4); padding-top: var(--space-5); border-top: 1px solid var(--color-border-soft); }
.section-heading p { color: var(--color-text-muted); }
.customer-detail-grid { margin-bottom: var(--space-5); }
.p-message .p-button { margin-left: var(--space-3); }
@media (max-width: 40rem) {
  .customer-filters, .customer-filters > *, .detail-heading > .p-button, .section-heading > .p-button { width: 100%; }
}
</style>
