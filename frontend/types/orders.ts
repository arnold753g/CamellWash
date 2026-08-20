import type { User } from './auth'
import type { PaginationData } from './catalogs'
import type { Vehicle } from './customers'

export type OrderStatus = 'PENDING' | 'IN_PROGRESS' | 'QUALITY_CONTROL' | 'FINISHED' | 'DELIVERED' | 'CANCELLED'

export interface OrderLine {
  id: number
  order_id: number
  service_id: number
  service_name_snapshot: string
  price_snapshot: string
  created_at: string
}

export interface OrderHistory {
  id: number
  order_id: number
  from_status: OrderStatus | null
  to_status: OrderStatus
  changed_at: string
  actor_user_id: number
  actor: User
  reason: string | null
}

export interface Order {
  id: number
  vehicle_id: number
  vehicle: Vehicle & { client_name?: string }
  registered_by_user_id: number
  registered_by: User
  received_at: string
  estimated_delivery_at: string
  status: OrderStatus
  total: string
  version: number
  services: OrderLine[]
  staff: User[]
  history: OrderHistory[]
  created_at: string
  updated_at: string
}

export interface OrderPage { orders: Order[], pagination: PaginationData }

export interface OrderServiceOption {
  service_id: number
  service_name: string
  category_name: string
  price: string
}
