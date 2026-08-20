import type { CatalogItem, PaginationData } from './catalogs'

export interface ClientListItem {
  id: number
  full_name: string
  ci: string
  ci_provided: boolean
  phone: string
  vehicle_count: number
}

export interface ClientDetail extends ClientListItem {
  created_at: string
  updated_at: string
}

export interface Vehicle {
  id: number
  client_id: number
  vehicle_type_id: number
  vehicle_type: CatalogItem
  plate: string | null
  brand: string
  model: string
  color: string
  created_at: string
  updated_at: string
}

export interface ClientPage { clients: ClientListItem[], pagination: PaginationData }
export interface VehiclePage { vehicles: Vehicle[], pagination: PaginationData }
