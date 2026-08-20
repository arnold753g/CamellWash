export interface ApiErrorData {
  code: string
  message: string
  fields?: Record<string, string>
}

export interface ApiEnvelope<T> {
  success: boolean
  data?: T
  message?: string
  error?: ApiErrorData
  request_id?: string
  timestamp: string
}

export interface ActionResult<T = undefined> {
  ok: boolean
  data?: T
  message?: string
  error?: ApiErrorData
}

