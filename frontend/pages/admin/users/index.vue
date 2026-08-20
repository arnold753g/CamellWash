<template>
  <section>
    <div class="page-heading">
      <h1>Usuarios</h1>
      <p>Crea cuentas de trabajo y administra sus datos, roles y estado. Solo el Administrador tiene acceso a estas acciones.</p>
    </div>

    <div class="toolbar">
      <div class="cluster user-filters">
        <div class="form-field">
          <label for="search">Buscar</label>
          <InputText id="search" v-model.trim="search" placeholder="Correo, nombre o apellido" @keyup.enter="applyFilters" />
        </div>
        <div class="form-field">
          <label for="role-filter">Rol</label>
          <Select id="role-filter" v-model="roleFilter" :options="roleOptions" option-label="label" option-value="value" placeholder="Todos" show-clear />
        </div>
        <div class="form-field">
          <label for="status-filter">Estado</label>
          <Select id="status-filter" v-model="statusFilter" :options="statusOptions" option-label="label" option-value="value" placeholder="Todos" show-clear />
        </div>
        <Button label="Filtrar" icon="pi pi-search" :loading="loading" @click="applyFilters" />
      </div>
      <Button label="Nuevo usuario" icon="pi pi-user-plus" @click="openCreate" />
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <LoadingState v-if="loading" />
    <EmptyState v-else-if="users.length === 0" title="No hay usuarios" description="Ajusta los filtros o crea la primera cuenta de trabajo." />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="users" data-key="id" striped-rows>
          <Column field="email" header="Correo" />
          <Column header="Nombre"><template #body="{ data }">{{ data.first_name }} {{ data.last_name }}</template></Column>
          <Column header="Rol"><template #body="{ data }"><Tag :value="roleLabel(data.role)" :severity="roleSeverity(data.role)" /></template></Column>
          <Column header="Estado"><template #body="{ data }"><Tag :value="statusLabel(data.status)" :severity="data.status === 'active' ? 'success' : 'danger'" /></template></Column>
          <Column header="Acciones"><template #body="{ data }"><Button label="Ver y editar" icon="pi pi-pencil" size="small" outlined @click="openUser(data.id)" /></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="createVisible" modal header="Crear usuario" :style="dialogStyle" @hide="resetCreate">
      <Message v-if="createError" severity="error" class="form-error" :closable="false">{{ createError }}</Message>
      <form id="create-user-form" class="stack" novalidate @submit.prevent="createUser">
        <div class="detail-grid">
          <div class="form-field">
            <label for="create-first-name">Nombre</label>
            <InputText id="create-first-name" v-model.trim="createForm.first_name" autocomplete="given-name" :invalid="Boolean(createErrors.first_name)" />
            <small v-if="createErrors.first_name" class="field-error" role="alert">{{ createErrors.first_name }}</small>
          </div>
          <div class="form-field">
            <label for="create-last-name">Apellido</label>
            <InputText id="create-last-name" v-model.trim="createForm.last_name" autocomplete="family-name" :invalid="Boolean(createErrors.last_name)" />
            <small v-if="createErrors.last_name" class="field-error" role="alert">{{ createErrors.last_name }}</small>
          </div>
        </div>
        <div class="form-field">
          <label for="create-email">Correo de acceso</label>
          <InputText id="create-email" v-model.trim="createForm.email" type="email" autocomplete="off" :invalid="Boolean(createErrors.email)" />
          <small v-if="createErrors.email" class="field-error" role="alert">{{ createErrors.email }}</small>
        </div>
        <div class="form-field">
          <label for="create-password">Contraseña inicial</label>
          <Password input-id="create-password" v-model="createForm.password" autocomplete="new-password" :feedback="false" toggle-mask :invalid="Boolean(createErrors.password)" />
          <small class="form-help">Entre 12 y 72 caracteres, sin espacios al inicio o al final.</small>
          <small v-if="createErrors.password" class="field-error" role="alert">{{ createErrors.password }}</small>
        </div>
        <div class="detail-grid">
          <div class="form-field">
            <label for="create-role">Rol</label>
            <Select id="create-role" v-model="createForm.role" :options="roleOptions" option-label="label" option-value="value" :invalid="Boolean(createErrors.role)" />
            <small v-if="createErrors.role" class="field-error" role="alert">{{ createErrors.role }}</small>
          </div>
          <div class="form-field">
            <label for="create-status">Estado</label>
            <Select id="create-status" v-model="createForm.status" :options="statusOptions" option-label="label" option-value="value" />
          </div>
        </div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="creating" @click="createVisible = false" />
        <Button form="create-user-form" type="submit" label="Crear usuario" icon="pi pi-user-plus" :loading="creating" />
      </template>
    </Dialog>

    <Dialog v-model:visible="detailVisible" modal header="Detalle de usuario" :style="dialogStyle">
      <LoadingState v-if="detailLoading" />
      <template v-else-if="selected">
        <Message v-if="detailError" severity="error" class="form-error" :closable="false">{{ detailError }}</Message>
        <div class="detail-grid">
          <div class="form-field"><label for="edit-first-name">Nombre</label><InputText id="edit-first-name" v-model.trim="selected.first_name" :disabled="isSelf" /></div>
          <div class="form-field"><label for="edit-last-name">Apellido</label><InputText id="edit-last-name" v-model.trim="selected.last_name" :disabled="isSelf" /></div>
          <div class="form-field"><label for="edit-email">Correo de acceso</label><InputText id="edit-email" v-model.trim="selected.email" type="email" :disabled="isSelf" /></div>
          <div class="form-field"><label>Último ingreso</label><InputText :model-value="formatDate(selected.last_login_at)" disabled /></div>
          <div class="form-field"><label for="edit-role">Rol</label><Select id="edit-role" v-model="selected.role" :options="roleOptions" option-label="label" option-value="value" :disabled="isSelf" /></div>
          <div class="form-field"><label for="edit-status">Estado</label><Select id="edit-status" v-model="selected.status" :options="statusOptions" option-label="label" option-value="value" :disabled="isSelf" /></div>
        </div>
        <Message v-if="isSelf" severity="info" :closable="false">Edita tu nombre o correo desde Mi perfil. Tu propio rol y estado no pueden cambiarse aquí.</Message>
        <div class="cluster dialog-actions">
          <Button label="Cerrar" severity="secondary" outlined @click="detailVisible = false" />
          <Button label="Guardar cambios" icon="pi pi-save" :disabled="isSelf" :loading="saving" @click="confirmSave" />
        </div>
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
import Password from 'primevue/password'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { z } from 'zod'
import type { Role, User, UserStatus } from '~/types/auth'
import { apiError } from '~/utils/api-error'

interface UserPage { users: User[], pagination: { page: number, limit: number, total: number, total_pages: number } }

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Usuarios' })

const auth = useAuthStore()
const confirm = useConfirm()
const toast = useToast()
const dialogStyle = { width: 'min(40rem, calc(100vw - 2rem))' }
const roleOptions = [
  { label: 'Administrador', value: 'admin' },
  { label: 'Encargado', value: 'manager' },
  { label: 'Personal de lavado', value: 'wash_staff' }
] satisfies { label: string, value: Role }[]
const statusOptions = [
  { label: 'Activo', value: 'active' },
  { label: 'Inactivo', value: 'inactive' }
] satisfies { label: string, value: UserStatus }[]

const loading = ref(false)
const detailLoading = ref(false)
const saving = ref(false)
const creating = ref(false)
const errorMessage = ref('')
const detailError = ref('')
const createError = ref('')
const search = ref('')
const roleFilter = ref<Role | null>(null)
const statusFilter = ref<UserStatus | null>(null)
const users = ref<User[]>([])
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const detailVisible = ref(false)
const createVisible = ref(false)
const selected = ref<User | null>(null)
const original = ref<User | null>(null)
const createErrors = reactive<Record<string, string>>({})
const createForm = reactive<{ email: string, password: string, first_name: string, last_name: string, role: Role, status: UserStatus }>({
  email: '', password: '', first_name: '', last_name: '', role: 'wash_staff', status: 'active'
})
const isSelf = computed(() => selected.value?.id === auth.user?.id)
const createSchema = z.object({
  email: z.string().trim().email('Ingresa un correo válido'),
  password: z.string().min(12, 'Usa al menos 12 caracteres').max(72, 'Usa como máximo 72 caracteres').refine(value => value.trim() === value, 'No uses espacios al inicio o al final'),
  first_name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(100),
  last_name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(100),
  role: z.enum(['admin', 'manager', 'wash_staff']),
  status: z.enum(['active', 'inactive'])
})

const load = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const query = new URLSearchParams({ page: String(pagination.page), limit: String(pagination.limit) })
    if (search.value) query.set('search', search.value)
    if (roleFilter.value) query.set('role', roleFilter.value)
    if (statusFilter.value) query.set('status', statusFilter.value)
    const response = await auth.request<UserPage>(`/admin/users?${query}`)
    users.value = response.data?.users || []
    Object.assign(pagination, response.data?.pagination)
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loading.value = false }
}

const applyFilters = () => { pagination.page = 1; void load() }
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void load() }
const openCreate = () => { resetCreate(); createVisible.value = true }
const resetCreate = () => {
  Object.assign(createForm, { email: '', password: '', first_name: '', last_name: '', role: 'wash_staff', status: 'active' })
  Object.keys(createErrors).forEach(key => delete createErrors[key])
  createError.value = ''
}
const createUser = async () => {
  Object.keys(createErrors).forEach(key => delete createErrors[key])
  createError.value = ''
  const parsed = createSchema.safeParse(createForm)
  if (!parsed.success) {
    for (const issue of parsed.error.issues) createErrors[String(issue.path[0])] = issue.message
    return
  }
  creating.value = true
  try {
    await auth.request('/admin/users', { method: 'POST', body: parsed.data })
    toast.add({ severity: 'success', summary: 'Usuario creado', detail: 'La cuenta ya puede iniciar sesión.', life: 3500 })
    createVisible.value = false
    pagination.page = 1
    await load()
  } catch (error) { createError.value = apiError(error).message }
  finally { creating.value = false }
}

const openUser = async (id: number) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    const response = await auth.request<User>(`/admin/users/${id}`)
    selected.value = response.data ? { ...response.data } : null
    original.value = response.data ? { ...response.data } : null
  } catch (error) { errorMessage.value = apiError(error).message; detailVisible.value = false }
  finally { detailLoading.value = false }
}
const confirmSave = () => confirm.require({
  message: 'Cambiar el correo, rol o estado puede cerrar las sesiones existentes. ¿Deseas continuar?',
  header: 'Confirmar cambios', icon: 'pi pi-exclamation-triangle', acceptLabel: 'Guardar', rejectLabel: 'Cancelar',
  accept: save
})
const save = async () => {
  if (!selected.value || !original.value || isSelf.value) return
  detailError.value = ''
  const general = z.object({
    email: z.string().trim().email(), first_name: z.string().trim().min(2).max(100), last_name: z.string().trim().min(2).max(100)
  }).safeParse(selected.value)
  if (!general.success) {
    detailError.value = 'Revisa el nombre, apellido y correo antes de guardar.'
    return
  }
  saving.value = true
  try {
    await auth.request(`/admin/users/${selected.value.id}`, { method: 'PATCH', body: general.data })
    if (selected.value.role !== original.value.role) await auth.request(`/admin/users/${selected.value.id}/role`, { method: 'PATCH', body: { role: selected.value.role } })
    if (selected.value.status !== original.value.status) await auth.request(`/admin/users/${selected.value.id}/status`, { method: 'PATCH', body: { status: selected.value.status } })
    toast.add({ severity: 'success', summary: 'Usuario actualizado', life: 3500 })
    detailVisible.value = false
    await load()
  } catch (error) { detailError.value = apiError(error).message }
  finally { saving.value = false }
}

const formatDate = (value: string | null) => value ? new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value)) : 'Sin registro'
const roleLabel = (role: Role) => roleOptions.find(option => option.value === role)?.label || role
const roleSeverity = (role: Role) => role === 'admin' ? 'info' : role === 'manager' ? 'warn' : 'secondary'
const statusLabel = (status: UserStatus) => status === 'active' ? 'Activo' : 'Inactivo'

onMounted(load)
</script>

<style scoped>
.user-filters { align-items: end; }
.user-filters .form-field:first-child { min-width: min(22rem, 100%); }
.dialog-actions { justify-content: flex-end; margin-top: var(--space-6); }
</style>
