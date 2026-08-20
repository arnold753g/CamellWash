import { defineStore } from 'pinia'
import type { ActionResult, ApiEnvelope } from '~/types/api'
import type { SessionData, User } from '~/types/auth'
import { apiError } from '~/utils/api-error'
import { isAuthenticationError, isSessionExpired, isUnsafeMethod } from '~/utils/auth-session'

interface AuthState {
  user: User | null
  csrfToken: string | null
  sessionExpiresAt: number | null
  idleTimeoutSeconds: number
  lastActivityAt: number | null
  initialized: boolean
}

let initializePromise: Promise<boolean> | null = null
let expirationTimer: ReturnType<typeof setTimeout> | null = null

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    csrfToken: null,
    sessionExpiresAt: null,
    idleTimeoutSeconds: 30 * 60,
    lastActivityAt: null,
    initialized: false
  }),

  getters: {
    isAuthenticated: state => Boolean(state.user && state.csrfToken && state.sessionExpiresAt),
    isAdmin: state => state.user?.role === 'admin',
    fullName: state => state.user ? `${state.user.first_name} ${state.user.last_name}` : ''
  },

  actions: {
    async login(email: string, password: string): Promise<ActionResult<User>> {
      try {
        const config = useRuntimeConfig()
        const response = await $fetch<ApiEnvelope<SessionData>>(`${config.public.apiBase}/auth/login`, {
          method: 'POST', credentials: 'include', body: { email, password }
        })
        if (!response.data) return { ok: false, error: response.error }
        this.setSession(response.data)
        return { ok: true, data: response.data.user, message: response.message }
      } catch (error) {
        return { ok: false, error: apiError(error, 'No se pudo iniciar sesión') }
      }
    },

    async initialize(force = false): Promise<boolean> {
      if (!force && this.initialized) {
        if (!this.isAuthenticated) return false
        if (!this.expired()) return true
        this.clear()
        return false
      }
      if (initializePromise) return initializePromise
      initializePromise = (async () => {
        try {
          const config = useRuntimeConfig()
          const response = await $fetch<ApiEnvelope<SessionData>>(`${config.public.apiBase}/auth/session`, { credentials: 'include' })
          if (!response.data) return false
          this.setSession(response.data)
          return true
        } catch {
          this.clear()
          return false
        } finally {
          this.initialized = true
        }
      })().finally(() => { initializePromise = null })
      return initializePromise
    },

    async request<T>(path: string, options: Record<string, unknown> = {}): Promise<ApiEnvelope<T>> {
      if (!await this.initialize()) {
        const error = new Error('Authentication is required') as Error & { statusCode: number }
        error.statusCode = 401
        throw error
      }
      const config = useRuntimeConfig()
      const method = String(options.method || 'GET')
      const headers = new Headers(options.headers as HeadersInit | undefined)
      if (isUnsafeMethod(method) && this.csrfToken) headers.set('X-CSRF-Token', this.csrfToken)
      try {
        const response = await $fetch<ApiEnvelope<T>>(`${config.public.apiBase}${path}`, {
          ...options, method: method as 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE', credentials: 'include', headers
        })
        this.recordActivity()
        return response
      } catch (error) {
        if (isAuthenticationError(error)) {
          this.clear()
          this.redirectToLogin()
        }
        throw error
      }
    },

    async logout() {
      try {
        if (this.csrfToken) await this.request('/auth/logout', { method: 'POST' })
      } catch {
        // The session may already be invalid; local state must still be cleared.
      } finally {
        this.clear()
      }
    },

    setSession(data: SessionData) {
      this.user = data.user
      this.csrfToken = data.csrf_token
      this.sessionExpiresAt = Date.parse(data.session_expires_at)
      this.idleTimeoutSeconds = Number(data.idle_timeout_seconds) || 30 * 60
      this.lastActivityAt = Date.now()
      this.initialized = true
      this.scheduleExpiration()
    },

    clear() {
      if (expirationTimer) clearTimeout(expirationTimer)
      expirationTimer = null
      this.user = null
      this.csrfToken = null
      this.sessionExpiresAt = null
      this.lastActivityAt = null
      this.initialized = true
    },

    recordActivity() {
      this.lastActivityAt = Date.now()
      this.scheduleExpiration()
    },

    expired() {
      if (!this.sessionExpiresAt || !this.lastActivityAt) return true
      return isSessionExpired(this.sessionExpiresAt, this.lastActivityAt, this.idleTimeoutSeconds)
    },

    scheduleExpiration() {
      if (!import.meta.client || !this.sessionExpiresAt || !this.lastActivityAt) return
      if (expirationTimer) clearTimeout(expirationTimer)
      const deadline = Math.min(this.sessionExpiresAt, this.lastActivityAt + this.idleTimeoutSeconds * 1000)
      const remaining = deadline - Date.now()
      if (remaining <= 0) {
        this.clear()
        this.redirectToLogin()
        return
      }
      expirationTimer = setTimeout(() => {
        this.clear()
        this.redirectToLogin()
      }, remaining)
    },

    redirectToLogin() {
      if (!import.meta.client || window.location.pathname === '/login') return
      void navigateTo('/login?reason=session_expired')
    }
  }
})
