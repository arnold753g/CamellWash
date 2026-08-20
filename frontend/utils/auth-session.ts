const authenticationCodes = new Set(['UNAUTHORIZED', 'INVALID_SESSION', 'SESSION_EXPIRED'])

export const isUnsafeMethod = (method = 'GET') =>
  !['GET', 'HEAD', 'OPTIONS'].includes(String(method).toUpperCase())

export const sessionDeadline = (absoluteExpiration: number, lastActivity: number, idleSeconds: number) =>
  Math.min(absoluteExpiration, lastActivity + idleSeconds * 1000)

export const isSessionExpired = (absoluteExpiration: number, lastActivity: number, idleSeconds: number, now = Date.now()) =>
  now >= sessionDeadline(absoluteExpiration, lastActivity, idleSeconds)

export const isAuthenticationError = (error: unknown) => {
  const value = error as { status?: number, statusCode?: number, data?: { error?: { code?: string } } }
  const status = Number(value?.statusCode || value?.status)
  const code = String(value?.data?.error?.code || '')
  return status === 401 && authenticationCodes.has(code)
}

export const safeInternalRedirect = (value: unknown, fallback: string) => {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//') || value.includes('\\')) return fallback
  return value
}

