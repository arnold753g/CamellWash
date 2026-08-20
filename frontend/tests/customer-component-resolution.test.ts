import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import test from 'node:test'

const views = [
  'pages/admin/clients/index.vue',
  'pages/app/clients/index.vue',
  'components/customers/CustomerVehicleWorkspace.vue'
]

test('las rutas de clientes renderizan el panel compartido y sus estados críticos', async () => {
  const sources = await Promise.all(views.map(path => readFile(resolve(path), 'utf8')))
  assert.match(sources[0]!, /<CustomerVehicleWorkspace\s*\/>/)
  assert.match(sources[1]!, /<CustomerVehicleWorkspace\s*\/>/)

  const workspace = sources[2]!
  assert.match(workspace, /data-testid="customer-vehicle-workspace"/)
  for (const state of ['<LoadingState', '<EmptyState', 'severity="error"', 'clientDialogVisible', 'detailVisible', 'vehicleDialogVisible']) {
    assert.ok(workspace.includes(state), `Falta el estado o panel ${state}`)
  }
})

test('todos los componentes del módulo de clientes son resolubles', async () => {
  const declarations = await readFile(resolve('.nuxt/components.d.ts'), 'utf8')
  const registered = new Set([...declarations.matchAll(/export const ([A-Z][A-Za-z0-9]*):/g)].map(match => match[1]!))
  const framework = new Set(['NuxtLink', 'NuxtPage', 'ClientOnly', 'Teleport', 'Transition'])
  const unresolved: string[] = []

  for (const [index, source] of (await Promise.all(views.map(path => readFile(resolve(path), 'utf8')))).entries()) {
    const template = source.match(/<template>([\s\S]*?)<\/template>/)?.[1] || ''
    const imports = new Set([...source.matchAll(/import\s+([A-Z][A-Za-z0-9]*)\s+from\s+['"]/g)].map(match => match[1]!))
    const used = new Set([...template.matchAll(/<([A-Z][A-Za-z0-9]*)\b/g)].map(match => match[1]!))
    for (const component of used) {
      if (!imports.has(component) && !registered.has(component) && !framework.has(component)) unresolved.push(`${views[index]}: ${component}`)
    }
  }
  assert.deepEqual(unresolved, [])
})
