<template>
  <section class="detail-section payment-panel" data-testid="payment-panel" aria-labelledby="payment-heading">
    <div class="payment-heading-row">
      <div>
        <h3 id="payment-heading">Pago y entrega</h3>
        <p class="section-help">El importe siempre proviene del total congelado de la orden.</p>
      </div>
      <Tag :value="currentPayment ? 'Pagado' : 'Pendiente de pago'" :severity="currentPayment ? 'success' : 'warn'" />
    </div>

    <Message v-if="errorMessage" severity="error" :closable="false">
      {{ errorMessage }} <Button label="Reintentar" size="small" outlined @click="loadPayments" />
    </Message>
    <LoadingState v-else-if="loading" />
    <template v-else>
      <div v-if="currentPayment" class="payment-card">
        <dl class="detail-grid">
          <div class="detail-item"><dt>Monto pagado</dt><dd class="payment-amount">{{ currentPayment.amount }}</dd></div>
          <div class="detail-item"><dt>Método</dt><dd>{{ methodLabel(currentPayment.method) }}</dd></div>
          <div class="detail-item"><dt>Fecha</dt><dd>{{ formatDate(currentPayment.paid_at) }}</dd></div>
          <div class="detail-item"><dt>Cobrado por</dt><dd>{{ userName(currentPayment.collected_by) }}</dd></div>
          <div v-if="currentPayment.external_reference" class="detail-item"><dt>Referencia</dt><dd>{{ currentPayment.external_reference }}</dd></div>
        </dl>
        <Message v-if="order.status !== 'FINISHED' && order.status !== 'DELIVERED'" severity="info" :closable="false">Pago registrado al recibir el vehículo. La orden continuará su flujo normal.</Message>
        <div class="cluster payment-actions">
          <Button v-if="order.status === 'FINISHED'" label="Entregar vehículo" icon="pi pi-check-circle" :loading="submitting" @click="confirmDelivery" />
          <Button label="Corregir pago" icon="pi pi-sync" severity="secondary" outlined :disabled="submitting" @click="openCorrection" />
          <Button v-if="order.status !== 'DELIVERED'" label="Anular pago" icon="pi pi-times" severity="danger" outlined :disabled="submitting" @click="openVoid" />
        </div>
      </div>
      <div v-else class="payment-empty">
        <i class="pi pi-wallet" aria-hidden="true" />
        <div><strong>Esta orden todavía no tiene un pago válido</strong><p>Puede cobrarse ahora o al retirar el vehículo.</p></div>
        <Button v-if="canRegister" label="Registrar pago" icon="pi pi-wallet" @click="openRegistration" />
      </div>

      <details v-if="voidedPayments.length" class="payment-history">
        <summary>Ver pagos anulados ({{ voidedPayments.length }})</summary>
        <div v-for="payment in voidedPayments" :key="payment.id" class="voided-payment">
          <div class="cluster"><Tag value="Anulado" severity="danger" /><Tag :value="reviewLabel(payment.review_status)" :severity="reviewSeverity(payment.review_status)" /></div>
          <p>{{ payment.amount }} · {{ methodLabel(payment.method) }} · {{ formatDate(payment.paid_at) }}</p>
          <p>{{ payment.void_reason }}</p>
        </div>
      </details>
    </template>

    <Dialog v-model:visible="paymentVisible" modal :header="paymentMode === 'register' ? 'Registrar pago total' : 'Corregir pago'" :style="smallDialogStyle" @hide="resetPaymentForm">
      <Message :severity="paymentMode === 'register' ? 'info' : 'warn'" :closable="false">
        {{ paymentMode === 'register' ? 'Confirma el método antes de registrar. El monto no puede modificarse.' : 'El pago anterior quedará anulado y el reemplazo se registrará en la misma transacción.' }}
      </Message>
      <Message v-if="formError" severity="error" :closable="false">{{ formError }}</Message>
      <form id="payment-form" class="stack" novalidate @submit.prevent="submitPayment">
        <div class="readonly-amount"><span>Total a cobrar</span><strong>{{ order.total }}</strong></div>
        <div class="form-field">
          <label for="payment-method">Método de pago</label>
          <Select id="payment-method" v-model="paymentForm.method" :options="methodOptions" option-label="label" option-value="value" :invalid="Boolean(formErrors.method)" />
          <small v-if="formErrors.method" class="field-error">{{ formErrors.method }}</small>
        </div>
        <div class="form-field">
          <label for="payment-reference">Referencia externa (opcional)</label>
          <InputText id="payment-reference" v-model="paymentForm.external_reference" maxlength="120" :invalid="Boolean(formErrors.external_reference)" />
          <small class="form-help">Puede usarse para el número de operación QR, transferencia o tarjeta.</small>
          <small v-if="formErrors.external_reference" class="field-error">{{ formErrors.external_reference }}</small>
        </div>
        <div v-if="paymentMode === 'correct'" class="form-field">
          <label for="payment-correction-reason">Motivo de la corrección</label>
          <Textarea id="payment-correction-reason" v-model="paymentForm.reason" rows="3" maxlength="500" auto-resize :invalid="Boolean(formErrors.reason)" />
          <small v-if="formErrors.reason" class="field-error">{{ formErrors.reason }}</small>
        </div>
        <div v-if="paymentMode === 'register' && order.status === 'FINISHED'" class="checkbox-row">
          <Checkbox v-model="paymentForm.deliver" input-id="payment-deliver" binary />
          <label for="payment-deliver">Entregar el vehículo al confirmar este pago</label>
        </div>
      </form>
      <template #footer>
        <Button label="Volver" severity="secondary" outlined :disabled="submitting" @click="paymentVisible = false" />
        <Button form="payment-form" type="submit" :label="paymentMode === 'register' ? 'Confirmar pago' : 'Confirmar corrección'" icon="pi pi-check" :loading="submitting" />
      </template>
    </Dialog>

    <Dialog v-model:visible="voidVisible" modal header="Anular pago" :style="smallDialogStyle">
      <Message severity="warn" :closable="false">La orden volverá a quedar pendiente de pago. Si eres Encargado, el Administrador revisará esta anulación posteriormente.</Message>
      <Message v-if="formError" severity="error" :closable="false">{{ formError }}</Message>
      <div class="form-field">
        <label for="payment-void-reason">Motivo obligatorio</label>
        <Textarea id="payment-void-reason" v-model="voidReason" rows="4" maxlength="500" auto-resize :invalid="Boolean(formErrors.reason)" />
        <small v-if="formErrors.reason" class="field-error">{{ formErrors.reason }}</small>
      </div>
      <template #footer>
        <Button label="Volver" severity="secondary" outlined :disabled="submitting" @click="voidVisible = false" />
        <Button label="Confirmar anulación" icon="pi pi-times" severity="danger" :loading="submitting" @click="submitVoid" />
      </template>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import LoadingState from '~/components/ui/LoadingState.vue'
import type { User } from '~/types/auth'
import type { Order } from '~/types/orders'
import type { Payment, PaymentList, PaymentMethod, PaymentResult, PaymentReviewStatus } from '~/types/payments'
import { apiError } from '~/utils/api-error'

const props = defineProps<{ order: Order }>()
const emit = defineEmits<{ 'order-updated': [order: Order], 'payment-state': [paid: boolean] }>()
const auth = useAuthStore()
const confirm = useConfirm()
const toast = useToast()
const payments = ref<Payment[]>([])
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const formError = ref('')
const formErrors = reactive<Record<string, string>>({})
const paymentVisible = ref(false)
const voidVisible = ref(false)
const voidReason = ref('')
const paymentMode = ref<'register' | 'correct'>('register')
const paymentForm = reactive({ method: 'CASH' as PaymentMethod, external_reference: '', reason: '', deliver: false, idempotency_key: '' })
const smallDialogStyle = { width: 'min(36rem, calc(100vw - 2rem))' }
const methodOptions = [
  { label: 'Efectivo', value: 'CASH' }, { label: 'QR', value: 'QR' },
  { label: 'Transferencia', value: 'BANK_TRANSFER' }, { label: 'Tarjeta', value: 'CARD' }
] satisfies { label: string, value: PaymentMethod }[]
const currentPayment = computed(() => payments.value.find(item => item.status === 'VALID') || null)
const voidedPayments = computed(() => payments.value.filter(item => item.status === 'VOIDED'))
const canRegister = computed(() => !['CANCELLED', 'DELIVERED'].includes(props.order.status) && !currentPayment.value)

const loadPayments = async () => {
  loading.value = true; errorMessage.value = ''
  try {
    const response = await auth.request<PaymentList>(`/orders/${props.order.id}/payments`)
    payments.value = response.data?.payments || []
    emit('payment-state', Boolean(currentPayment.value))
  } catch (error) { errorMessage.value = apiError(error, 'No se pudo cargar el pago').message }
  finally { loading.value = false }
}
const newOperationKey = () => `payment_${crypto.randomUUID().replaceAll('-', '_')}`
const openRegistration = () => {
  paymentMode.value = 'register'; resetPaymentForm(); paymentForm.deliver = props.order.status === 'FINISHED'; paymentForm.idempotency_key = newOperationKey(); paymentVisible.value = true
}
const openCorrection = () => {
  paymentMode.value = 'correct'; resetPaymentForm(); paymentForm.method = currentPayment.value?.method || 'CASH'; paymentForm.external_reference = currentPayment.value?.external_reference || ''; paymentForm.idempotency_key = newOperationKey(); paymentVisible.value = true
}
const openVoid = () => { clearErrors(); voidReason.value = ''; formError.value = ''; voidVisible.value = true }
const resetPaymentForm = () => { Object.assign(paymentForm, { method: 'CASH', external_reference: '', reason: '', deliver: false, idempotency_key: '' }); clearErrors(); formError.value = '' }
const clearErrors = () => { Object.keys(formErrors).forEach(key => delete formErrors[key]) }

const submitPayment = async () => {
  clearErrors(); formError.value = ''
  if (!paymentForm.method) formErrors.method = 'Selecciona un método'
  if (paymentForm.external_reference.trim() && paymentForm.external_reference.trim().length < 3) formErrors.external_reference = 'Usa al menos 3 caracteres'
  if (paymentMode.value === 'correct' && paymentForm.reason.trim().length < 5) formErrors.reason = 'Describe el motivo con al menos 5 caracteres'
  if (Object.keys(formErrors).length || !currentPayment.value && paymentMode.value === 'correct') return
  submitting.value = true
  try {
    const path = paymentMode.value === 'register' ? `/orders/${props.order.id}/payments` : `/payments/${currentPayment.value!.id}/correct`
    const body = paymentMode.value === 'register'
      ? { version: props.order.version, method: paymentForm.method, external_reference: paymentForm.external_reference, idempotency_key: paymentForm.idempotency_key, deliver: paymentForm.deliver }
      : { version: props.order.version, method: paymentForm.method, external_reference: paymentForm.external_reference, idempotency_key: paymentForm.idempotency_key, reason: paymentForm.reason }
    const response = await auth.request<PaymentResult>(path, { method: 'POST', body })
    if (response.data?.order) emit('order-updated', response.data.order)
    paymentVisible.value = false; toast.add({ severity: 'success', summary: paymentMode.value === 'register' ? 'Pago registrado' : 'Pago corregido', life: 3500 })
    await loadPayments()
  } catch (error) { const parsed = apiError(error); formError.value = parsed.message; Object.assign(formErrors, parsed.fields || {}) }
  finally { submitting.value = false }
}
const submitVoid = async () => {
  clearErrors(); formError.value = ''
  if (voidReason.value.trim().length < 5) { formErrors.reason = 'Describe el motivo con al menos 5 caracteres'; return }
  if (!currentPayment.value) return
  submitting.value = true
  try {
    const response = await auth.request<PaymentResult>(`/payments/${currentPayment.value.id}/void`, { method: 'POST', body: { version: props.order.version, reason: voidReason.value } })
    if (response.data?.order) emit('order-updated', response.data.order)
    voidVisible.value = false; toast.add({ severity: 'success', summary: 'Pago anulado', life: 3500 }); await loadPayments()
  } catch (error) { const parsed = apiError(error); formError.value = parsed.message; Object.assign(formErrors, parsed.fields || {}) }
  finally { submitting.value = false }
}
const confirmDelivery = () => confirm.require({ message: 'La orden está pagada. Al confirmar, el vehículo quedará entregado y el estado será terminal.', header: 'Confirmar entrega', icon: 'pi pi-check-circle', acceptLabel: 'Entregar vehículo', rejectLabel: 'Volver', accept: deliverOrder })
const deliverOrder = async () => {
  submitting.value = true; errorMessage.value = ''
  try {
    const response = await auth.request<Order>(`/orders/${props.order.id}/transitions`, { method: 'POST', body: { version: props.order.version, target: 'DELIVERED', reason: 'Pago verificado al entregar' } })
    if (response.data) emit('order-updated', response.data)
    toast.add({ severity: 'success', summary: 'Vehículo entregado', life: 3500 })
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { submitting.value = false }
}

function methodLabel(method: PaymentMethod) { return methodOptions.find(item => item.value === method)?.label || method }
function reviewLabel(status: PaymentReviewStatus) { return ({ NOT_REQUIRED: 'Sin revisión', PENDING_REVIEW: 'Revisión pendiente', REVIEWED: 'Revisado', OBSERVED: 'Observado' } as const)[status] }
function reviewSeverity(status: PaymentReviewStatus) { return status === 'OBSERVED' ? 'danger' : status === 'PENDING_REVIEW' ? 'warn' : 'secondary' }
function userName(user: User) { return `${user.first_name} ${user.last_name}` }
function formatDate(value: string) { return new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value)) }

watch(() => props.order.id, loadPayments, { immediate: true })
</script>

<style scoped>
.payment-heading-row, .payment-empty { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: var(--space-4); }
.payment-heading-row h3, .section-help, .payment-empty p { margin: 0; }
.section-help, .payment-empty p { color: var(--color-text-muted); }
.payment-card, .payment-empty { padding: var(--space-5); border: 1px solid var(--color-border-soft); border-radius: var(--radius-lg); background: var(--color-surface-elevated); }
.payment-amount, .readonly-amount strong { color: var(--color-primary); font-size: 1.35rem; font-variant-numeric: tabular-nums; }
.payment-actions { margin-top: var(--space-4); }
.payment-empty > i { color: var(--color-primary); font-size: 1.75rem; }
.readonly-amount { display: flex; align-items: center; justify-content: space-between; gap: var(--space-4); padding: var(--space-4); border: 1px solid rgba(255, 85, 0, .22); border-radius: var(--radius-md); background: var(--color-primary-soft); }
.checkbox-row { display: flex; align-items: center; gap: var(--space-3); min-height: 2.75rem; }
.payment-history { margin-top: var(--space-4); }
.payment-history summary { cursor: pointer; color: var(--color-text-muted); }
.voided-payment { margin-top: var(--space-3); padding: var(--space-4); border-left: 3px solid var(--color-danger); background: var(--color-surface-elevated); }
.voided-payment p { margin: .35rem 0 0; color: var(--color-text-muted); }
@media (max-width: 40rem) { .payment-actions > *, .payment-empty > .p-button { width: 100%; } }
</style>
