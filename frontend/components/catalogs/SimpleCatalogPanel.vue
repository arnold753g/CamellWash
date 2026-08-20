<template>
  <section class="stack" :aria-labelledby="headingId">
    <div class="catalog-intro">
      <div>
        <h2 :id="headingId">{{ title }}</h2>
        <p class="muted">{{ description }}</p>
      </div>
      <Button :label="`Nuevo ${singular.toLowerCase()}`" icon="pi pi-plus" @click="openCreate" />
    </div>

    <div class="toolbar catalog-toolbar">
      <div class="cluster catalog-filters">
        <div class="form-field">
          <label :for="`${kind}-search`">Buscar</label>
          <InputText :id="`${kind}-search`" v-model.trim="search" :placeholder="`Nombre de ${singular.toLowerCase()}`" @keyup.enter="applyFilters" />
        </div>
        <div class="form-field">
          <label :for="`${kind}-status`">Estado</label>
          <Select :id="`${kind}-status`" v-model="activeFilter" :options="statusFilterOptions" option-label="label" option-value="value" />
        </div>
        <Button label="Filtrar" icon="pi pi-search" :loading="loading" @click="applyFilters" />
      </div>
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <LoadingState v-if="loading" />
    <EmptyState v-else-if="items.length === 0" :title="`No hay ${title.toLowerCase()}`" description="Ajusta los filtros o crea el primer registro." />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="items" data-key="id" striped-rows>
          <Column field="name" header="Nombre" />
          <Column header="Estado">
            <template #body="{ data }"><Tag :value="data.active ? 'Activo' : 'Inactivo'" :severity="data.active ? 'success' : 'danger'" /></template>
          </Column>
          <Column header="Acciones">
            <template #body="{ data }">
              <div class="cluster row-actions">
                <Button label="Editar" icon="pi pi-pencil" size="small" outlined @click="openEdit(data)" />
                <Button
                  :label="data.active ? 'Desactivar' : 'Activar'"
                  :icon="data.active ? 'pi pi-ban' : 'pi pi-check-circle'"
                  size="small"
                  :severity="data.active ? 'danger' : 'success'"
                  outlined
                  @click="confirmStatus(data)"
                />
              </div>
            </template>
          </Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? `Editar ${singular.toLowerCase()}` : `Crear ${singular.toLowerCase()}`" :style="dialogStyle" @hide="resetForm">
      <Message v-if="formError" severity="error" class="form-error" :closable="false">{{ formError }}</Message>
      <form :id="formId" class="stack" novalidate @submit.prevent="save">
        <div class="form-field">
          <label :for="`${kind}-name`">Nombre</label>
          <InputText :id="`${kind}-name`" v-model.trim="form.name" autocomplete="off" :invalid="Boolean(nameError)" autofocus />
          <small class="form-help">Se ignorarán mayúsculas, espacios iniciales y espacios repetidos al detectar duplicados.</small>
          <small v-if="nameError" class="field-error" role="alert">{{ nameError }}</small>
        </div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="saving" @click="dialogVisible = false" />
        <Button :form="formId" type="submit" label="Guardar" icon="pi pi-save" :loading="saving" />
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
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { z } from 'zod'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { CatalogItem, CatalogPage } from '~/types/catalogs'
import { apiError } from '~/utils/api-error'

const props = defineProps<{
  kind: 'vehicle-types' | 'service-categories'
  responseKey: 'vehicle_types' | 'service_categories'
  title: string
  singular: string
  description: string
  inactiveWarning?: string
}>()
const emit = defineEmits<{ changed: [] }>()

const auth = useAuthStore()
const confirm = useConfirm()
const toast = useToast()
const headingId = `${props.kind}-heading`
const formId = `${props.kind}-form`
const dialogStyle = { width: 'min(34rem, calc(100vw - 2rem))' }
const statusFilterOptions = [
  { label: 'Todos', value: 'all' },
  { label: 'Activos', value: 'true' },
  { label: 'Inactivos', value: 'false' }
]
const schema = z.object({ name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(100, 'Usa como máximo 100 caracteres') })

const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const formError = ref('')
const nameError = ref('')
const search = ref('')
const activeFilter = ref('all')
const items = ref<CatalogItem[]>([])
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const dialogVisible = ref(false)
const editing = ref<CatalogItem | null>(null)
const form = reactive({ name: '' })

const load = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const query = new URLSearchParams({ page: String(pagination.page), limit: String(pagination.limit), active: activeFilter.value })
    if (search.value) query.set('search', search.value)
    const response = await auth.request<CatalogPage<CatalogItem>>(`/admin/${props.kind}?${query}`)
    items.value = response.data?.[props.responseKey] || []
    if (response.data?.pagination) Object.assign(pagination, response.data.pagination)
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loading.value = false }
}

const applyFilters = () => { pagination.page = 1; void load() }
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void load() }
const openCreate = () => { resetForm(); dialogVisible.value = true }
const openEdit = (item: CatalogItem) => { resetForm(); editing.value = item; form.name = item.name; dialogVisible.value = true }
const resetForm = () => { editing.value = null; form.name = ''; nameError.value = ''; formError.value = '' }
const save = async () => {
  nameError.value = ''
  formError.value = ''
  const parsed = schema.safeParse(form)
  if (!parsed.success) { nameError.value = parsed.error.issues[0]?.message || 'Revisa el nombre'; return }
  saving.value = true
  try {
    const path = editing.value ? `/admin/${props.kind}/${editing.value.id}` : `/admin/${props.kind}`
    await auth.request(path, { method: editing.value ? 'PATCH' : 'POST', body: parsed.data })
    toast.add({ severity: 'success', summary: `${props.singular} guardado`, life: 3500 })
    dialogVisible.value = false
    await load()
    emit('changed')
  } catch (error) { formError.value = apiError(error).message }
  finally { saving.value = false }
}
const confirmStatus = (item: CatalogItem) => {
  const activating = !item.active
  const detail = activating
    ? `¿Deseas activar “${item.name}”? Volverá a estar disponible en los selectores operativos.`
    : `¿Deseas desactivar “${item.name}”? ${props.inactiveWarning || 'Dejará de estar disponible para registros nuevos, sin eliminar su historial.'}`
  confirm.require({
    message: detail,
    header: activating ? `Activar ${props.singular.toLowerCase()}` : `Desactivar ${props.singular.toLowerCase()}`,
    icon: activating ? 'pi pi-check-circle' : 'pi pi-exclamation-triangle',
    acceptLabel: activating ? 'Activar' : 'Desactivar', rejectLabel: 'Cancelar',
    accept: async () => {
      try {
        await auth.request(`/admin/${props.kind}/${item.id}/status`, { method: 'PATCH', body: { active: activating } })
        toast.add({ severity: 'success', summary: `${props.singular} ${activating ? 'activado' : 'desactivado'}`, life: 3500 })
        await load()
        emit('changed')
      } catch (error) { errorMessage.value = apiError(error).message }
    }
  })
}

onMounted(load)
defineExpose({ load })
</script>

<style scoped>
.catalog-intro { display: flex; flex-wrap: wrap; align-items: start; justify-content: space-between; gap: var(--space-4); }
.catalog-intro h2 { margin-bottom: var(--space-2); }
.catalog-intro p { max-width: 68ch; margin-bottom: 0; }
.catalog-toolbar { margin-bottom: 0; }
.catalog-filters { align-items: end; }
.catalog-filters .form-field:first-child { min-width: min(22rem, 100%); }
.row-actions { flex-wrap: nowrap; }
@media (max-width: 40rem) {
  .catalog-intro > *, .catalog-filters > * { width: 100%; }
  .row-actions { flex-wrap: wrap; }
}
</style>
