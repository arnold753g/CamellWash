import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import test from 'node:test'

const views = ['pages/admin/orders/index.vue', 'pages/app/orders/index.vue', 'components/orders/OrderWorkspace.vue', 'components/payments/PaymentPanel.vue', 'pages/admin/payment-reviews/index.vue']

test('las rutas de órdenes renderizan el panel y todos sus estados críticos', async () => {
  const sources = await Promise.all(views.map(path => readFile(resolve(path), 'utf8')))
  assert.match(sources[0]!, /<OrderWorkspace\s*\/>/)
  assert.match(sources[1]!, /<OrderWorkspace\s*\/>/)
  for (const contract of ['data-testid="order-workspace"', '<LoadingState', '<EmptyState', 'severity="error"', 'formVisible', 'detailVisible', 'cancelVisible']) assert.ok(sources[2]!.includes(contract), `Falta el panel o estado ${contract}`)
  for (const contract of ['data-testid="payment-panel"', 'Confirmar pago', 'Corregir pago', 'Anular pago', 'Entregar vehículo']) assert.ok(sources[3]!.includes(contract), `Falta el flujo ${contract}`)
  for (const contract of ['data-testid="payment-review-workspace"', '<LoadingState', '<EmptyState', 'severity="error"', 'Guardar revisión']) assert.ok(sources[4]!.includes(contract), `Falta el estado de revisión ${contract}`)
})

test('error, carga y vacío son estados mutuamente exclusivos', async () => {
  const source = await readFile(resolve('components/orders/OrderWorkspace.vue'), 'utf8')
  assert.match(source, /<Message v-if="errorMessage"[\s\S]*?<LoadingState v-else-if="loading"\s*\/>[\s\S]*?<EmptyState v-else-if="orders\.length === 0"/)
})

test('los paneles de pago renderizan estados exclusivos y un monto no editable', async () => {
  const payment = await readFile(resolve('components/payments/PaymentPanel.vue'), 'utf8')
  const reviews = await readFile(resolve('pages/admin/payment-reviews/index.vue'), 'utf8')
  assert.match(payment, /<Message v-if="errorMessage"[\s\S]*?<LoadingState v-else-if="loading"\s*\/>[\s\S]*?<template v-else>/)
  assert.match(reviews, /<Message v-if="errorMessage"[\s\S]*?<LoadingState v-else-if="loading"\s*\/>[\s\S]*?<EmptyState v-else-if="payments\.length === 0"/)
  assert.ok(payment.includes('<strong>{{ order.total }}</strong>'))
  assert.doesNotMatch(payment, /v-model="[^\"]*amount/)
})

test('los componentes usados por órdenes están importados o registrados', async () => {
  const declarations = await readFile(resolve('.nuxt/components.d.ts'), 'utf8')
  const registered = new Set([...declarations.matchAll(/export const ([A-Z][A-Za-z0-9]*):/g)].map(match => match[1]!))
  const framework = new Set(['NuxtLink', 'NuxtPage', 'ClientOnly', 'Teleport', 'Transition'])
  const unresolved: string[] = []
  for (const path of views) {
    const source = await readFile(resolve(path), 'utf8')
    const template = source.match(/<template>([\s\S]*?)<\/template>/)?.[1] || ''
    const imports = new Set([...source.matchAll(/import\s+([A-Z][A-Za-z0-9]*)\s+from\s+['"]/g)].map(match => match[1]!))
    for (const component of new Set([...template.matchAll(/<([A-Z][A-Za-z0-9]*)\b/g)].map(match => match[1]!))) {
      if (!imports.has(component) && !registered.has(component) && !framework.has(component)) unresolved.push(`${path}: ${component}`)
    }
  }
  assert.deepEqual(unresolved, [])
})
