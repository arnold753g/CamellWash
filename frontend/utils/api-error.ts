import type { ApiErrorData } from '~/types/api'

export const apiError = (error: unknown, fallback = 'No se pudo completar la operación'): ApiErrorData => {
  const value = error as { data?: { error?: ApiErrorData } }
  return value?.data?.error || { code: 'REQUEST_FAILED', message: fallback }
}

