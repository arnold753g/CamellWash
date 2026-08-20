export interface CatalogItem {
  id: number
  name: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface ServiceCatalogItem extends CatalogItem {
  category_id: number
  category: CatalogItem
}

export interface PaginationData {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface CatalogPage<T> {
  pagination: PaginationData
  vehicle_types?: T[]
  service_categories?: T[]
  services?: T[]
}
