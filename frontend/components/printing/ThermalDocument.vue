<template>
  <article class="thermal-document" :data-document-id="document.id">
    <header>
      <img src="/branding/camell-wash-logo.png" alt="" class="receipt-logo">
      <h1>{{ data.business_name }}</h1>
      <p v-if="data.business_address">{{ data.business_address }}</p>
      <p v-if="data.business_phone">Tel. {{ data.business_phone }}</p>
      <h2>{{ document.kind === 'RECEIPT' ? 'RECIBO DE PAGO' : 'FICHA DE RECEPCIÓN Y RETIRO' }}</h2>
      <strong>{{ document.number }}</strong>
      <p class="copy-label">{{ attempt ? (attempt.copy_number ? `REIMPRESIÓN ${attempt.copy_number}` : 'ORIGINAL') : 'VISTA PREVIA' }}</p>
    </header>
    <p v-if="document.state === 'VOIDED' || document.state === 'REPLACED' || document.state === 'CANCELLED'" class="state-warning">
      {{ document.state === 'REPLACED' ? 'PAGO REEMPLAZADO' : document.state === 'CANCELLED' ? 'ORDEN CANCELADA' : 'PAGO ANULADO' }}
      <span v-if="document.replacement_number">Nuevo recibo: {{ document.replacement_number }}</span>
    </p>
    <p v-if="document.replaces_number">Reemplaza al recibo {{ document.replaces_number }}</p>
    <p class="order-number">ORDEN #{{ document.order_id }}</p>
    <dl>
      <div><dt>Cliente</dt><dd>{{ data.client_name }}</dd></div>
      <div><dt>Solicitud de factura</dt><dd>{{ data.invoice_requested === undefined ? 'Sin registro previo' : data.invoice_requested ? 'Solicitada' : 'No solicitada' }}</dd></div>
      <div><dt>Placa</dt><dd>{{ data.plate || 'Sin placa registrada' }}</dd></div>
      <div><dt>Vehículo</dt><dd>{{ data.vehicle }}</dd></div>
      <div><dt>Recepción</dt><dd>{{ date(data.received_at) }}</dd></div>
      <div v-if="document.kind === 'RECEPTION'"><dt>Entrega estimada</dt><dd>{{ date(data.estimated_delivery_at) }}</dd></div>
    </dl>
    <table><thead><tr><th>Servicio</th><th>BOB</th></tr></thead><tbody>
      <tr v-for="(line, index) in data.services" :key="index"><td>{{ line.name }}</td><td>{{ money(line.price) }}</td></tr>
    </tbody></table>
    <p class="total"><span>{{ document.kind === 'RECEIPT' ? 'TOTAL PAGADO' : 'TOTAL DE LA ORDEN' }}</span><strong>Bs {{ money(data.total) }}</strong></p>
    <dl v-if="document.kind === 'RECEIPT'">
      <div><dt>Método</dt><dd>{{ method(data.method) }}</dd></div>
      <div><dt>Fecha de pago</dt><dd>{{ date(data.paid_at) }}</dd></div>
      <div><dt>Cobrado por</dt><dd>{{ data.collected_by }}</dd></div>
      <div v-if="data.external_reference"><dt>Referencia</dt><dd>{{ data.external_reference }}</dd></div>
    </dl>
    <p v-else class="pickup-note">{{ document.state === 'DELIVERED' ? 'Vehículo entregado.' : 'Conserve esta ficha y preséntela al retirar su vehículo.' }} Esta ficha no acredita un pago.</p>
    <footer>
      <p>Emitido: {{ date(document.issued_at) }}</p>
      <p v-if="attempt">Solicitud: {{ date(attempt.requested_at) }}</p>
      <p>Hora de Bolivia · BOB</p><p>Gracias por su preferencia.</p>
    </footer>
  </article>
</template>

<script setup lang="ts">
import type { PrintAttempt, ThermalDocument } from '~/types/documents'
const props = defineProps<{ document: ThermalDocument, attempt?: PrintAttempt }>()
const data = computed(() => props.document.snapshot)
const date = (value?: string) => value ? new Intl.DateTimeFormat('es-BO', { dateStyle: 'short', timeStyle: 'short', timeZone: 'America/La_Paz' }).format(new Date(value)) : '—'
const money = (value: string) => Number(value).toLocaleString('es-BO', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const method = (value?: string) => ({ CASH: 'Efectivo', QR: 'QR', BANK_TRANSFER: 'Transferencia', CARD: 'Tarjeta' }[value || ''] || value)
</script>

<style scoped>
.thermal-document { width: 72mm; padding: 0; margin: 0 auto; background: #fff; color: #000; font-family: Arial, sans-serif; font-size: 10pt; line-height: 1.3; overflow-wrap: anywhere; }
.thermal-document * { color: #000; box-shadow: none; font-family: Arial, sans-serif; }
header, footer, .pickup-note { text-align: center; }
.receipt-logo { width: 48mm; height: auto; display: block; margin: 0 auto 2mm; filter: grayscale(1) invert(1); }
h1 { font-size: 14pt; margin: 0 0 1mm; } h2 { font-size: 11pt; margin: 3mm 0 1mm; }
p { margin: 1.5mm 0; } .copy-label { font-weight: bold; }
.order-number { font-size: 16pt; font-weight: bold; text-align: center; border-block: 1px dashed #000; padding: 2mm 0; }
dl { margin: 2mm 0; } dl div { margin: 1mm 0; } dt { font-weight: bold; } dd { margin: 0; }
table { width: 100%; border-collapse: collapse; font-size: inherit; margin-top: 3mm; table-layout: fixed; }
th { border-bottom: 1px dashed #000; text-align: left; } th:last-child, td:last-child { width: 19mm; text-align: right; vertical-align: top; }
td { padding: 1mm 0; } .total { display: flex; gap: 2mm; justify-content: space-between; border-block: 1px dashed #000; padding: 2mm 0; font-weight: bold; }
.state-warning { border: 2px solid #000; padding: 2mm; text-align: center; font-weight: bold; } .state-warning span { display: block; }
.pickup-note, footer { font-size: 9pt; } footer { border-top: 1px dashed #000; margin-top: 3mm; padding-top: 2mm; }
</style>
