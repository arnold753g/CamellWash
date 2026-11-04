export interface DocumentSnapshot {
  business_name: string
  business_address?: string
  business_phone?: string
  client_name: string
  vehicle: string
  plate: string
  received_at: string
  estimated_delivery_at: string
  services: { name: string, price: string }[]
  total: string
  currency: string
  method?: string
  paid_at?: string
  collected_by?: string
  external_reference?: string
}
export interface PrintAttempt {
  id: number
  document_id: number
  copy_number: number
  actor_user_id: number
  requested_at: string
}
export interface ThermalDocument {
  id: number
  number: string
  kind: 'RECEIPT' | 'RECEPTION'
  order_id: number
  payment_id: number | null
  snapshot: DocumentSnapshot
  issued_at: string
  state: string
  replacement_number?: string
  replaces_number?: string
  attempts: PrintAttempt[]
}
export interface PrintResult { document: ThermalDocument, attempt: PrintAttempt }
