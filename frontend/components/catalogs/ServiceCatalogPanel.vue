<template>
  <section class="stack" aria-labelledby="services-heading">
    <div class="catalog-intro">
      <div>
        <h2 id="services-heading">Servicios</h2>
        <p class="muted">Administra los servicios dentro de una categoría. Un mismo nombre puede existir en categorías diferentes.</p>
      </div>
      <Button label="Nuevo servicio" icon="pi pi-plus" :disabled="activeCategories.length === 0" @click="openCreate" />
    </div>
    <Message v-if="activeCategories.length === 0 && !loadingCategories" severity="info" :closable="false">Activa o crea una categoría antes de registrar servicios.</Message>

    <div class="toolbar catalog-toolbar">
      <div class="cluster catalog-filters">
        <div class="form-field"><label for="service-search">Buscar</label><InputText id="service-search" v-model.trim="search" placeholder="Nombre del servicio" @keyup.enter="applyFilters" /></div>
        <div class="form-field"><label for="service-category-filter">Categoría</label><Select id="service-category-filter" v-model="categoryFilter" :options="categoryOptions" option-label="label" option-value="value" placeholder="Todas" show-clear /></div>
        <div class="form-field"><label for="service-status-filter">Estado</label><Select id="service-status-filter" v-model="activeFilter" :options="statusFilterOptions" option-label="label" option-value="value" /></div>
        <Button label="Filtrar" icon="pi pi-search" :loading="loading" @click="applyFilters" />
      </div>
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <LoadingState v-if="loading" />
    <EmptyState v-else-if="services.length === 0" title="No hay servicios" description="Ajusta los filtros o crea el primer servicio." />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="services" data-key="id" striped-rows>
          <Column field="name" header="Servicio" />
          <Column header="Categoría"><template #body="{ data }">{{ data.category.name }}</template></Column>
          <Column header="Estado"><template #body="{ data }"><Tag :value="data.active ? 'Activo' : 'Inactivo'" :severity="data.active ? 'success' : 'danger'" /></template></Column>
          <Column header="Acciones"><template #body="{ data }"><div class="cluster row-actions">
            <Button label="Editar" icon="pi pi-pencil" size="small" outlined @click="openEdit(data)" />
            <Button :label="data.active ? 'Desactivar' : 'Activar'" :icon="data.active ? 'pi pi-ban' : 'pi pi-check-circle'" size="small" :severity="data.active ? 'danger' : 'success'" outlined @click="confirmStatus(data)" />
          </div></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Editar servicio' : 'Crear servicio'" :style="dialogStyle" @hide="resetForm">
      <Message v-if="formError" severity="error" class="form-error" :closable="false">{{ formError }}</Message>
      <form id="service-form" class="stack" novalidate @submit.prevent="save">
        <div class="form-field">
          <label for="service-category">Categoría</label>
          <Select id="service-category" v-model="form.category_id" :options="formCategoryOptions" option-label="label" option-value="value" option-disabled="disabled" :invalid="Boolean(errors.category_id)" />
          <small v-if="errors.category_id" class="field-error" role="alert">{{ errors.category_id }}</small>
        </div>
        <div class="form-field">
          <label for="service-name">Nombre</label>
          <InputText id="service-name" v-model.trim="form.name" autocomplete="off" :invalid="Boolean(errors.name)" />
          <small class="form-help">La unicidad se valida dentro de la categoría seleccionada.</small>
          <small v-if="errors.name" class="field-error" role="alert">{{ errors.name }}</small>
        </div>
      </form>
      <template #footer>
        <Button label="Cancelar" severity="secondary" outlined :disabled="saving" @click="dialogVisible = false" />
        <Button form="service-form" type="submit" label="Guardar" icon="pi pi-save" :loading="saving" />
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
import type { CatalogItem, CatalogPage, ServiceCatalogItem } from '~/types/catalogs'
import { apiError } from '~/utils/api-error'

const auth = useAuthStore()
const confirm = useConfirm()
const toast = useToast()
const dialogStyle = { width: 'min(34rem, calc(100vw - 2rem))' }
const statusFilterOptions = [{ label: 'Todos', value: 'all' }, { label: 'Activos', value: 'true' }, { label: 'Inactivos', value: 'false' }]
const schema = z.object({ category_id: z.number().int().positive('Selecciona una categoría'), name: z.string().trim().min(2, 'Ingresa al menos 2 caracteres').max(100, 'Usa como máximo 100 caracteres') })

const loading = ref(false)
const loadingCategories = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const formError = ref('')
const search = ref('')
const activeFilter = ref('all')
const categoryFilter = ref<number | null>(null)
const services = ref<ServiceCatalogItem[]>([])
const categories = ref<CatalogItem[]>([])
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const dialogVisible = ref(false)
const editing = ref<ServiceCatalogItem | null>(null)
const form = reactive({ category_id: 0, name: '' })
const errors = reactive<Record<string, string>>({})
const categoryOptions = computed(() => categories.value.map(item => ({ label: `${item.name}${item.active ? '' : ' — Inactiva'}`, value: item.id })))
const activeCategories = computed(() => categories.value.filter(item => item.active))
const formCategoryOptions = computed(() => categories.value.map(item => ({
  label: `${item.name}${item.active ? '' : ' — Inactiva'}`, value: item.id, disabled: !item.active && item.id !== editing.value?.category_id
})))

const loadCategories = async () => {
  loadingCategories.value = true
  try {
    const response = await auth.request<CatalogPage<CatalogItem>>('/admin/service-categories?active=all&limit=100')
    categories.value = response.data?.service_categories || []
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loadingCategories.value = false }
}
const load = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const query = new URLSearchParams({ page: String(pagination.page), limit: String(pagination.limit), active: activeFilter.value })
    if (search.value) query.set('search', search.value)
    if (categoryFilter.value) query.set('category_id', String(categoryFilter.value))
    const response = await auth.request<CatalogPage<ServiceCatalogItem>>(`/admin/services?${query}`)
    services.value = response.data?.services || []
    if (response.data?.pagination) Object.assign(pagination, response.data.pagination)
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loading.value = false }
}
const reload = async () => { await loadCategories(); await load() }
const applyFilters = () => { pagination.page = 1; void load() }
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void load() }
const resetForm = () => { editing.value = null; form.category_id = activeCategories.value[0]?.id || 0; form.name = ''; formError.value = ''; Object.keys(errors).forEach(key => delete errors[key]) }
const openCreate = () => { resetForm(); dialogVisible.value = true }
const openEdit = (item: ServiceCatalogItem) => { resetForm(); editing.value = item; form.category_id = item.category_id; form.name = item.name; dialogVisible.value = true }
const save = async () => {
  formError.value = ''
  Object.keys(errors).forEach(key => delete errors[key])
  const parsed = schema.safeParse(form)
  if (!parsed.success) { for (const issue of parsed.error.issues) errors[String(issue.path[0])] = issue.message; return }
  saving.value = true
  try {
    const path = editing.value ? `/admin/services/${editing.value.id}` : '/admin/services'
    await auth.request(path, { method: editing.value ? 'PATCH' : 'POST', body: parsed.data })
    toast.add({ severity: 'success', summary: 'Servicio guardado', life: 3500 })
    dialogVisible.value = false
    await load()
  } catch (error) { formError.value = apiError(error).message }
  finally { saving.value = false }
}
const confirmStatus = (item: ServiceCatalogItem) => {
  const activating = !item.active
  confirm.require({
    message: activating ? `¿Deseas activar “${item.name}”?` : `¿Deseas desactivar “${item.name}”? Dejará de aparecer en selectores nuevos, sin eliminarse.`,
    header: activating ? 'Activar servicio' : 'Desactivar servicio', icon: activating ? 'pi pi-check-circle' : 'pi pi-exclamation-triangle',
    acceptLabel: activating ? 'Activar' : 'Desactivar', rejectLabel: 'Cancelar',
    accept: async () => {
      try {
        await auth.request(`/admin/services/${item.id}/status`, { method: 'PATCH', body: { active: activating } })
        toast.add({ severity: 'success', summary: `Servicio ${activating ? 'activado' : 'desactivado'}`, life: 3500 })
        await load()
      } catch (error) { errorMessage.value = apiError(error).message }
    }
  })
}

onMounted(reload)
defineExpose({ reload })
</script>

<style scoped>
.catalog-intro { display: flex; flex-wrap: wrap; align-items: start; justify-content: space-between; gap: var(--space-4); }
.catalog-intro h2 { margin-bottom: var(--space-2); }
.catalog-intro p { max-width: 68ch; margin-bottom: 0; }
.catalog-toolbar { margin-bottom: 0; }
.catalog-filters { align-items: end; }
.catalog-filters .form-field:first-child { min-width: min(19rem, 100%); }
.row-actions { flex-wrap: nowrap; }
@media (max-width: 40rem) {
  .catalog-intro > *, .catalog-filters > * { width: 100%; }
  .row-actions { flex-wrap: wrap; }
}
</style>
