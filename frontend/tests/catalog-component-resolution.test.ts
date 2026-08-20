import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import test from 'node:test'

const catalogViews = [
  'pages/admin/catalogs/index.vue',
  'components/catalogs/SimpleCatalogPanel.vue',
  'components/catalogs/ServiceCatalogPanel.vue'
]

const frameworkComponents = new Set([
  'ClientOnly',
  'KeepAlive',
  'NuxtLink',
  'NuxtPage',
  'Suspense',
  'Teleport',
  'Transition',
  'TransitionGroup'
])

test('catalog views only use components resolvable by Nuxt', async () => {
  const componentDeclarations = await readFile(resolve('.nuxt/components.d.ts'), 'utf8')
  const registeredComponents = new Set(
    [...componentDeclarations.matchAll(/export const ([A-Z][A-Za-z0-9]*):/g)].map(match => match[1]!)
  )
  const unresolved: string[] = []

  for (const view of catalogViews) {
    const source = await readFile(resolve(view), 'utf8')
    const template = source.match(/<template>([\s\S]*?)<\/template>/)?.[1] || ''
    const importedComponents = new Set(
      [...source.matchAll(/import\s+([A-Z][A-Za-z0-9]*)\s+from\s+['"]/g)].map(match => match[1]!)
    )
    const usedComponents = new Set(
      [...template.matchAll(/<([A-Z][A-Za-z0-9]*)\b/g)].map(match => match[1]!)
    )

    for (const component of usedComponents) {
      if (!importedComponents.has(component) && !registeredComponents.has(component) && !frameworkComponents.has(component)) {
        unresolved.push(`${view}: ${component}`)
      }
    }
  }

  assert.deepEqual(unresolved, [], `Unresolved catalog components:\n${unresolved.join('\n')}`)
})
