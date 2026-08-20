<template>
  <section data-testid="payment-review-workspace">
    <div class="page-heading">
      <h1>Revisión de anulaciones</h1>
      <p>Revisa posteriormente las anulaciones realizadas por el Encargado sin detener la operación diaria.</p>
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }} <Button label="Reintentar" size="small" outlined @click="loadReviews" /></Message>
    <LoadingState v-else-if="loading" />
    <EmptyState v-else-if="payments.length === 0" title="No hay anulaciones pendientes" description="Las nuevas anulaciones del Encargado aparecerán aquí para su revisión." />
    <template v-else>
      <div class="table-wrap review-table">
        <DataTable :value="payments" data-key="id" striped-rows>
          <Column header="Orden"><template #body="{ data }"><strong>#{{ data.order_id }}</strong><small class="table-secondary">{{ data.order?.vehicle?.client_name || 'Cliente' }}</small></template></Column>
          <Column header="Pago"><template #body="{ data }">{{ data.amount }}<small class="table-secondary">{{ methodLabel(data.method) }}</small></template></Column>
          <Column header="Anulado por"><template #body="{ data }">{{ data.voided_by ? userName(data.voided_by) : '—' }}<small class="table-secondary">{{ data.voided_at ? formatDate(data.voided_at) : '—' }}</small></template></Column>
          <Column field="void_reason" header="Motivo" />
          <Column header="Acciones"><template #body="{ data }"><Button label="Revisar" icon="pi pi-search" size="small" @click="openReview(data)" /></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="reviewVisible" modal header="Revisar anulación" :style="dialogStyle" @hide="resetReview">
      <Message severity="info" :closable="false">La revisión no modifica el pago vigente ni el estado operativo de la orden; deja constancia de control administrativo.</Message>
      <Message v-if="formError" severity="error" :closable="false">{{ formError }}</Message>
      <dl v-if="selected" class="detail-grid review-summary">
        <div class="detail-item"><dt>Orden</dt><dd>#{{ selected.order_id }}</dd></div>
        <div class="detail-item"><dt>Pago anulado</dt><dd>{{ selected.amount }} · {{ methodLabel(selected.method) }}</dd></div>
        <div class="detail-item"><dt>Anulado por</dt><dd>{{ selected.voided_by ? userName(selected.voided_by) : '—' }}</dd></div>
        <div class="detail-item"><dt>Motivo</dt><dd>{{ selected.void_reason }}</dd></div>
      </dl>
      <div class="form-field">
        <label for="payment-review-decision">Resultado</label>
        <Select id="payment-review-decision" v-model="decision" :options="decisionOptions" option-label="label" option-value="value" />
      </div>
      <div class="form-field">
        <label for="payment-review-note">Nota {{ decision === 'OBSERVED' ? 'obligatoria' : 'opcional' }}</label>
        <Textarea id="payment-review-note" v-model="note" rows="4" maxlength="500" auto-resize :invalid="Boolean(noteError)" />
        <small v-if="noteError" class="field-error">{{ noteError }}</small>
      </div>
      <template #footer>
        <Button label="Volver" severity="secondary" outlined :disabled="saving" @click="reviewVisible = false" />
        <Button label="Guardar revisión" icon="pi pi-check" :loading="saving" @click="submitReview" />
      </template>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import { useToast } from 'primevue/usetoast'
import EmptyState from '~/components/ui/EmptyState.vue'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { User } from '~/types/auth'
import type { Payment, PaymentMethod, PaymentReviewPage, PaymentReviewStatus } from '~/types/payments'
import { apiError } from '~/utils/api-error'

definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Revisión de pagos' })

const auth = useAuthStore()
const toast = useToast()
const payments = ref<Payment[]>([])
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const formError = ref('')
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const reviewVisible = ref(false)
const selected = ref<Payment | null>(null)
const decision = ref<Extract<PaymentReviewStatus, 'REVIEWED' | 'OBSERVED'>>('REVIEWED')
const note = ref('')
const noteError = ref('')
const dialogStyle = { width: 'min(42rem, calc(100vw - 2rem))' }
const decisionOptions = [{ label: 'Revisada y conforme', value: 'REVIEWED' }, { label: 'Observada', value: 'OBSERVED' }]
const methodOptions: Record<PaymentMethod, string> = { CASH: 'Efectivo', QR: 'QR', BANK_TRANSFER: 'Transferencia', CARD: 'Tarjeta' }

const loadReviews = async () => {
  loading.value = true; errorMessage.value = ''
  try {
    const response = await auth.request<PaymentReviewPage>(`/admin/payment-reviews?page=${pagination.page}&limit=${pagination.limit}`)
    payments.value = response.data?.payments || []
    Object.assign(pagination, response.data?.pagination)
  } catch (error) { payments.value = []; errorMessage.value = apiError(error, 'No se pudieron cargar las anulaciones').message }
  finally { loading.value = false }
}
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void loadReviews() }
const openReview = (payment: Payment) => { selected.value = payment; decision.value = 'REVIEWED'; note.value = ''; noteError.value = ''; formError.value = ''; reviewVisible.value = true }
const resetReview = () => { selected.value = null; note.value = ''; noteError.value = ''; formError.value = '' }
const submitReview = async () => {
  if (!selected.value) return
  noteError.value = ''; formError.value = ''
  if (decision.value === 'OBSERVED' && note.value.trim().length < 5) { noteError.value = 'Explica la observación con al menos 5 caracteres'; return }
  saving.value = true
  try {
    await auth.request<Payment>(`/admin/payment-reviews/${selected.value.id}`, { method: 'POST', body: { decision: decision.value, note: note.value } })
    reviewVisible.value = false; toast.add({ severity: 'success', summary: 'Revisión registrada', life: 3500 }); await loadReviews()
  } catch (error) { formError.value = apiError(error).message }
  finally { saving.value = false }
}

function methodLabel(method: PaymentMethod) { return methodOptions[method] }
function userName(user: User) { return `${user.first_name} ${user.last_name}` }
function formatDate(value: string) { return new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value)) }

onMounted(loadReviews)
</script>

<style scoped>
.review-table { margin-top: var(--space-5); }
.table-secondary { display: block; color: var(--color-text-muted); }
.review-summary { margin: var(--space-5) 0; }
</style>
