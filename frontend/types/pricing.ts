export interface PricingMatrixCell {
  service_id: number
  service_name: string
  category_id: number
  category_name: string
  vehicle_type_id: number
  vehicle_type_name: string
  price: string | null
  configured: boolean
  updated_at?: string
}

export interface PricingMatrixSummary {
  total: number
  configured: number
  missing: number
}

export interface PricingMatrixData {
  cells: PricingMatrixCell[]
  summary: PricingMatrixSummary
}

export interface PricingMatrixRow {
  service_id: number
  service_name: string
  category_id: number
  category_name: string
  cells: PricingMatrixCell[]
}
