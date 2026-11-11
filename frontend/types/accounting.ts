import type { PaginationData } from './catalogs'
import type { ReportRange } from './reports'

export interface AccountingOrder {
  id: number
  received_at: string
  status: string
  invoice_requested: boolean
  client_id: number
  client_name: string
  client_ci: string
  client_phone: string
  client_email: string | null
  plate: string | null
  total: string
  payment_id: number | null
  payment_method: string | null
  paid_at: string | null
}

export interface AccountingOrdersPage {
  range: ReportRange
  orders: AccountingOrder[]
  pagination: PaginationData
}
