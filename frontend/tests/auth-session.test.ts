import assert from 'node:assert/strict'
import test from 'node:test'
import { isSessionExpired, isUnsafeMethod, safeInternalRedirect, sessionDeadline } from '../utils/auth-session.ts'

test('session deadline uses the earliest absolute or idle deadline', () => {
  assert.equal(sessionDeadline(10_000, 1_000, 3), 4_000)
  assert.equal(isSessionExpired(10_000, 1_000, 3, 4_000), true)
})

test('unsafe HTTP methods require CSRF', () => {
  assert.equal(isUnsafeMethod('GET'), false)
  assert.equal(isUnsafeMethod('POST'), true)
})

test('redirects remain internal', () => {
  assert.equal(safeInternalRedirect('/admin?page=1', '/app'), '/admin?page=1')
  assert.equal(safeInternalRedirect('//evil.example', '/app'), '/app')
  assert.equal(safeInternalRedirect('https://evil.example', '/app'), '/app')
})

