import assert from 'node:assert/strict'
import test from 'node:test'
import { normalizePriceInput } from '../utils/money.ts'

test('normaliza importes decimales sin convertirlos a punto flotante', () => {
  assert.equal(normalizePriceInput('1'), '1.00')
  assert.equal(normalizePriceInput('001,5'), '1.50')
  assert.equal(normalizePriceInput('9999999999.99'), '9999999999.99')
  assert.equal(normalizePriceInput(' 12.30 '), '12.30')
})

test('rechaza cero, negativos, exponentes y más de dos decimales', () => {
  for (const value of ['0', '0.00', '-1', '1e2', '1.234', '10000000000', 'abc']) {
    assert.equal(normalizePriceInput(value), null, value)
  }
})
