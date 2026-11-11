import assert from 'node:assert/strict'
import test from 'node:test'
import { csvCell } from '../utils/csv.ts'

test('CSV neutralizes client values that spreadsheets treat as formulas', () => {
  for (const value of ['=HYPERLINK("https://example.test")', '+59170000000', '@SUM(A1)', '-1+2', ' =1+2', '\t=1+2']) {
    assert.ok(csvCell(value).startsWith('"\''), value)
  }
  assert.equal(csvCell('Cliente "Dos"'), '"Cliente ""Dos"""')
  assert.equal(csvCell(null), '""')
  assert.equal(csvCell('cliente@example.test'), '"cliente@example.test"')
})
