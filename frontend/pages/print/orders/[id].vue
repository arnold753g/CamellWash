<template>
  <main class="thermal-print-page" :class="{ 'print-authorized': armed }">
    <div class="print-controls">
      <h1>Comprobantes · Orden #{{ orderID }}</h1>
      <p>Papel de 80 mm. Selecciona tu impresora USB, escala 100 %, sin encabezados ni pies del navegador.</p>
      <p>Se registra cada solicitud del botón. El diálogo puede cancelarse; el sistema no confirma la salida física del papel.</p>
      <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
      <div class="controls-row">
        <Button label="Volver a órdenes" severity="secondary" outlined @click="goBack" />
        <Button v-if="error && !documents.length" label="Reintentar" @click="load" />
        <Button :label="documents.length > 1 ? 'Imprimir recibo y ficha' : 'Imprimir ficha'" icon="pi pi-print" :loading="busy" :disabled="!documents.length" @click="printDocuments" />
      </div>
      <details v-if="documents.length"><summary>Historial de solicitudes</summary>
        <div v-for="document in documents" :key="document.id"><strong>{{ document.number }}</strong>
          <p v-if="!document.attempts.length">Sin solicitudes de impresión.</p>
          <p v-for="attempt in document.attempts" :key="attempt.id">{{ attempt.copy_number ? `Reimpresión ${attempt.copy_number}` : 'Original' }} · Usuario #{{ attempt.actor_user_id }} · {{ formatDate(attempt.requested_at) }}</p>
        </div>
      </details>
    </div>
    <p class="print-instructions">Para imprimir estos documentos, usa el botón «Imprimir» de la aplicación.</p>
    <div v-if="documents.length" ref="paper" class="thermal-pages">
      <template v-for="(document, index) in documents" :key="document.id">
        <div v-if="index" class="cut-line">✂ — — — Corte por esta línea — — —</div>
        <ThermalDocument :document="document" :attempt="attempts[document.id]" />
      </template>
    </div>
  </main>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Message from 'primevue/message'
import ThermalDocument from '~/components/printing/ThermalDocument.vue'
import type { PrintAttempt, PrintResult, ThermalDocument as DocumentData } from '~/types/documents'
import { apiError } from '~/utils/api-error'

definePageMeta({ layout: false, middleware: 'cashier', key: route => route.fullPath })
const auth = useAuthStore()
const route = useRoute()
const orderID = Number(route.params.id)
const documents = ref<DocumentData[]>([])
const attempts = ref<Record<number, PrintAttempt>>({})
const busy = ref(false)
const error = ref('')
const armed = ref(false)
const paper = ref<HTMLElement | null>(null)
const pageHeight = ref(300)
let keys: Record<number, string> = {}
useHead(() => ({ title: `Comprobantes orden ${orderID}`, style: [{ textContent: `@media print {
  @page { size: 80mm ${pageHeight.value}mm; margin: 4mm; }
  html, body, #__nuxt { background: #fff !important; color-scheme: light !important; min-width: 0 !important; min-height: 0 !important; margin: 0 !important; padding: 0 !important; }
  .p-toast, .p-dialog-mask, .p-dialog, .p-tooltip, #teleports, nuxt-route-announcer { display: none !important; }
}` }] }))
const goBack = () => navigateTo(auth.user?.role === 'admin' ? '/admin/orders' : '/app/orders')
const formatDate = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'short', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value))
const load = async () => {
  busy.value = true; error.value = ''; armed.value = false
  try {
    if (!Number.isSafeInteger(orderID) || orderID < 1) throw new Error('Número de orden inválido')
    const loaded: DocumentData[] = []
    if (route.query.payment_id !== undefined) {
      const paymentID = Number(route.query.payment_id)
      if (!Number.isSafeInteger(paymentID) || paymentID < 1) throw new Error('Número de pago inválido')
      const receipt = await auth.request<DocumentData>(`/payments/${paymentID}/documents/receipt`, { method: 'POST' })
      if (!receipt.data || receipt.data.order_id !== orderID) throw new Error('El pago no corresponde a esta orden')
      loaded.push(receipt.data)
    }
    const reception = await auth.request<DocumentData>(`/reception/orders/${orderID}/document`, { method: 'POST' })
    if (!reception.data) throw new Error('No se pudo cargar la ficha')
    loaded.push(reception.data)
    documents.value = loaded
  } catch (cause) { error.value = apiError(cause, 'No se pudieron cargar los comprobantes').message }
  finally { busy.value = false }
}
const printDocuments = async () => {
  if (busy.value || !documents.value.length) return
  busy.value = true; error.value = ''; armed.value = false
  try {
    const loaded: DocumentData[] = []
    for (const document of documents.value) {
      keys[document.id] ||= `print_${crypto.randomUUID().replaceAll('-', '_')}`
      const response = await auth.request<PrintResult>(`/documents/${document.id}/print-attempts`, { method: 'POST', body: { idempotency_key: keys[document.id] } })
      if (!response.data) throw new Error('No se pudo registrar la solicitud')
      loaded.push(response.data.document)
      attempts.value[document.id] = response.data.attempt
    }
    documents.value = loaded
    await document.fonts.ready
    await nextTick()
    // CSS px use 96 dpi. Measure after copy/status labels render; allow rounding and margins.
    pageHeight.value = Math.max(60, Math.ceil((paper.value?.scrollHeight || 900) * 25.4 / 96) + 10)
    armed.value = true
    await nextTick()
    window.print()
    keys = {}
  } catch (cause) { error.value = apiError(cause, 'No se pudo solicitar la impresión').message }
  finally { busy.value = false }
}
const afterPrint = () => { armed.value = false }
onMounted(() => { window.addEventListener('afterprint', afterPrint); void load() })
onBeforeUnmount(() => window.removeEventListener('afterprint', afterPrint))
</script>

<style scoped>
.thermal-print-page { min-height: 100vh; padding: 1.5rem 1rem; background: #e9e9e9; color: #111; }
.print-controls { max-width: 48rem; margin: 0 auto 1.5rem; color: #111; } .print-controls h1 { color: #111; }
.controls-row { display: flex; flex-wrap: wrap; gap: .75rem; margin: 1rem 0; }
.thermal-pages { width: 80mm; margin: auto; padding: 4mm; background: white; box-sizing: border-box; }
.cut-line { margin: 5mm 0; text-align: center; font: 9pt Arial, sans-serif; color: #000; }
.print-instructions { display: none; }
@media print {
  :global(html), :global(body), :global(#__nuxt) { background: white !important; min-width: 0 !important; margin: 0 !important; padding: 0 !important; }
  .thermal-print-page { min-height: 0; padding: 0; background: white; }
  .print-controls, .thermal-pages { display: none; }
  .print-instructions { display: block; color: #000; }
  .print-authorized .print-instructions { display: none; }
  .print-authorized .thermal-pages { display: block; width: 72mm; padding: 0; margin: 0; }
  .cut-line { break-inside: avoid; }
}
</style>
