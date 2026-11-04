import assert from 'node:assert/strict'
import test from 'node:test'
import { canAccessRoleArea, roleRedirect } from '../utils/role-access.ts'

test('receipt printing remains limited to administrator and manager', () => {
  for (const role of ['admin', 'manager', 'accountant', 'wash_staff', 'unknown']) {
    const allowed = role === 'admin' || role === 'manager'
    assert.equal(canAccessRoleArea(role, 'cashier'), allowed)
    const target = '/print/orders/25?payment_id=12'
    assert.equal(roleRedirect(role, target) === target, allowed)
  }
})
