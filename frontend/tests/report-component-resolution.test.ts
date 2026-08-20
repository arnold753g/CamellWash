import assert from 'node:assert/strict'
import { access, readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import test from 'node:test'

const files = [
  'pages/admin/index.vue', 'pages/admin/reports/index.vue', 'pages/app/index.vue', 'pages/app/reports/index.vue',
  'components/reports/DashboardReportWorkspace.vue', 'components/app/AppShell.vue'
]

test('dashboard y reportes renderizan el panel real para cada rol autorizado', async () => {
  const sources = await Promise.all(files.map(path => readFile(resolve(path), 'utf8')))
  assert.match(sources[0]!, /<DashboardReportWorkspace show-financial/)
  assert.match(sources[1]!, /<DashboardReportWorkspace show-financial report-mode/)
  assert.match(sources[2]!, /auth\.user\?\.role === 'manager'/)
  assert.match(sources[3]!, /:show-financial="false" report-mode/)
  for (const contract of ['data-testid="dashboard-report-workspace"', '<LoadingState', '<EmptyState', 'severity="error"', 'Exportar operación', 'Exportar finanzas']) {
    assert.ok(sources[4]!.includes(contract), `Falta el contrato visual ${contract}`)
  }
  assert.match(sources[5]!, /\/admin\/reports/)
  assert.match(sources[5]!, /\/app\/reports/)
})

test('la marca Camell Wash usa el logo oficial y el token naranja aprobado', async () => {
  const [branding, theme, env] = await Promise.all([
    readFile(resolve('composables/useBranding.ts'), 'utf8'),
    readFile(resolve('assets/theme.css'), 'utf8'),
    readFile(resolve('../.env'), 'utf8')
  ])
  assert.match(branding, /camell-wash-logo\.png/)
  assert.match(theme, /--color-primary: #ff5500;/)
  assert.doesNotMatch(theme, /186,\s*254,\s*6|#bafe06/i)
  assert.match(env, /^APP_NAME=Camell Wash$/m)
  await access(resolve('public/branding/camell-wash-logo.png'))
})

test('los componentes de reportes están importados o registrados por Nuxt', async () => {
  const declarations = await readFile(resolve('.nuxt/components.d.ts'), 'utf8')
  const registered = new Set([...declarations.matchAll(/export const ([A-Z][A-Za-z0-9]*):/g)].map(match => match[1]!))
  const framework = new Set(['NuxtLink', 'NuxtPage', 'ClientOnly', 'Teleport', 'Transition'])
  const unresolved: string[] = []
  for (const path of files) {
    const source = await readFile(resolve(path), 'utf8')
    const template = source.match(/<template>([\s\S]*?)<\/template>/)?.[1] || ''
    const imports = new Set([...source.matchAll(/import\s+([A-Z][A-Za-z0-9]*)\s+from\s+['"]/g)].map(match => match[1]!))
    for (const component of new Set([...template.matchAll(/<([A-Z][A-Za-z0-9]*)\b/g)].map(match => match[1]!))) {
      if (!imports.has(component) && !registered.has(component) && !framework.has(component)) unresolved.push(`${path}: ${component}`)
    }
  }
  assert.deepEqual(unresolved, [])
})
