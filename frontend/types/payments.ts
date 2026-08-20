import type { User } from './auth'
import type { PaginationData } from './catalogs'
import type { Order } from './orders'

export type PaymentMethod = 'CASH' | 'QR' | 'BANK_TRANSFER' | 'CARD'
export type PaymentStatus = 'VALID' | 'VOIDED'
export type PaymentReviewStatus = 'NOT_REQUIRED' | 'PENDING_REVIEW' | 'REVIEWED' | 'OBSERVED'

export interface Payment {
  id: number
  order_id: number
  order?: Order
  amount: string
  method: PaymentMethod
  external_reference: string | null
  status: PaymentStatus
  delivery_requested: boolean
  collected_by_user_id: number
  collected_by: User
  paid_at: string
  voided_by_user_id: number | null
  voided_by?: User | null
  voided_at: string | null
  void_reason: string | null
  review_status: PaymentReviewStatus
  reviewed_by_user_id: number | null
  reviewed_by?: User | null
  reviewed_at: string | null
  review_note: string | null
  replaces_payment_id: number | null
}

export interface PaymentResult { payment: Payment, order: Order }
export interface PaymentList { payments: Payment[] }
export interface PaymentReviewPage { payments: Payment[], pagination: PaginationData }
