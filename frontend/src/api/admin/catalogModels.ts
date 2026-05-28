/**
 * Admin Catalog Models API endpoints
 */

import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'

export interface CatalogModel {
  id: number
  model_id: string
  name: string
  vendor: string
  category: string
  description: string
  tags: string[]
  doc_url: string
  icon_url: string
  context_window: number
  max_output_tokens: number
  input_modalities: string[]
  output_modalities: string[]
  features: string[]
  input_price: number
  output_price: number
  cache_write_price: number | null
  cache_read_price: number | null
  currency: string
  is_enabled: boolean
  created_at: number
  updated_at: number
}

export interface UpdateCatalogModelRequest {
  name: string
  vendor: string
  category: string
  description: string
  tags: string[]
  doc_url: string
  icon_url: string
  context_window: number
  max_output_tokens: number
  input_modalities: string[]
  output_modalities: string[]
  features: string[]
  input_price: number
  output_price: number
  cache_write_price: number | null
  cache_read_price: number | null
  currency: string
}

export interface CatalogModelListFilters {
  vendor?: string
  category?: string
  enabled?: boolean
  q?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export async function list(
  page = 1,
  pageSize = 20,
  filters?: CatalogModelListFilters,
  options?: { signal?: AbortSignal }
): Promise<BasePaginationResponse<CatalogModel>> {
  const { data } = await apiClient.get<BasePaginationResponse<CatalogModel>>(
    '/admin/catalog/models',
    { params: { page, page_size: pageSize, ...filters }, signal: options?.signal }
  )
  return data
}

export async function update(id: number, request: UpdateCatalogModelRequest): Promise<CatalogModel> {
  const { data } = await apiClient.put<CatalogModel>(`/admin/catalog/models/${id}`, request)
  return data as unknown as CatalogModel
}

export async function toggle(id: number): Promise<CatalogModel> {
  const { data } = await apiClient.patch<CatalogModel>(`/admin/catalog/models/${id}/toggle`)
  return data as unknown as CatalogModel
}

export async function seed(): Promise<{ seeded: number }> {
  const { data } = await apiClient.post<{ seeded: number }>('/admin/catalog/models/seed')
  return data as unknown as { seeded: number }
}

const catalogModelsAPI = { list, update, toggle, seed }
export default catalogModelsAPI
