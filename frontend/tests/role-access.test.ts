import assert from 'node:assert/strict'
import test from 'node:test'
import { canAccessRoleArea, roleHomePath, roleLabel, roleNavigation, roleOptions, roleRedirect } from '../utils/role-access.ts'
import type { RoleArea } from '../utils/role-access.ts'

test('each role has its own workspace and unknown roles have no workspace', () => {
  const expected: Record<string, { home: string, areas: RoleArea[] }> = {
    admin: { home: '/admin', areas: ['admin'] },
    manager: { home: '/app', areas: ['operations', 'manager'] },
    wash_staff: { home: '/app', areas: ['operations'] },
    accountant: { home: '/accountant', areas: ['accountant'] },
    unknown: { home: '/403', areas: [] }
  }
  const areas = ['admin', 'operations', 'manager', 'accountant'] as const
  for (const [role, access] of Object.entries(expected)) {
    assert.equal(roleHomePath(role), access.home)
    for (const area of areas) assert.equal(canAccessRoleArea(role, area), access.areas.includes(area), `${role}: ${area}`)
  }
  assert.deepEqual(roleNavigation(undefined), [])
  assert.deepEqual(roleNavigation('unknown'), [])
})

test('accountant navigation contains reports and own profile only', () => {
  assert.equal(roleLabel('accountant'), 'Contador')
  assert.ok(roleOptions.some(option => option.value === 'accountant'))
  assert.deepEqual(roleNavigation('accountant').map(item => item.to), ['/accountant', '/app/profile'])
  assert.ok(roleNavigation('admin').some(item => item.to === '/admin/users'))
  assert.ok(roleNavigation('manager').some(item => item.to === '/app/reports'))
  assert.deepEqual(roleNavigation('wash_staff').map(item => item.to), ['/app', '/app/orders', '/app/profile'])
})

test('login redirect preserves allowed destinations and rejects routes from other roles', () => {
  assert.equal(roleRedirect('accountant', '/accountant?date_from=2026-11-01'), '/accountant?date_from=2026-11-01')
  assert.equal(roleRedirect('accountant', '/app/profile'), '/app/profile')
  for (const requested of ['/admin', '/admin/users', '/admin/payment-reviews', '/app', '/app/orders', '/app/clients', '/app/reports', '//evil.example', 'https://evil.example', '/accountant/../admin']) {
    assert.equal(roleRedirect('accountant', requested), '/accountant', requested)
  }
  assert.equal(roleRedirect('manager', '/admin/reports'), '/app')
  assert.equal(roleRedirect('manager', '/accountant'), '/app')
  assert.equal(roleRedirect('manager', '/app/reports?period=month'), '/app/reports?period=month')
  assert.equal(roleRedirect('wash_staff', '/app/reports'), '/app')
  assert.equal(roleRedirect('wash_staff', '/app/orders'), '/app/orders')
  assert.equal(roleRedirect('admin', '/admin/users'), '/admin/users')
  assert.equal(roleRedirect('admin', '/accountant'), '/admin')
  assert.equal(roleRedirect('unknown', '/app/profile'), '/403')
})
