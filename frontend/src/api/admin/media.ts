import apiClient from '../client'
import type { PaginatedResponse } from '@/types'
import type { MediaTask } from '@/api/media'

export interface MediaBillingSettings {
  cny_to_usd_rate: number
  pricing_overrides: MediaPricingOverride[]
}

export interface MediaPricingOverride {
  model: string
  metric?: string
  price_per_million: number
  currency: 'CNY' | 'USD'
  resolutions?: string[]
  has_video_input?: boolean
  has_audio?: boolean
}

export interface UpdateMediaBillingSettingsRequest {
  cny_to_usd_rate?: number
  pricing_overrides?: MediaPricingOverride[]
}

export async function getSettings(): Promise<MediaBillingSettings> {
  const { data } = await apiClient.get<MediaBillingSettings>('/admin/media/settings')
  return data
}

export async function updateSettings(
  payload: UpdateMediaBillingSettingsRequest
): Promise<MediaBillingSettings> {
  const { data } = await apiClient.put<MediaBillingSettings>('/admin/media/settings', payload)
  return data
}

export async function listTasks(params?: Record<string, string | number>): Promise<PaginatedResponse<MediaTask>> {
  const { data } = await apiClient.get<PaginatedResponse<MediaTask>>('/admin/media/tasks', {
    params
  })
  return data
}

export const adminMediaAPI = {
  getSettings,
  updateSettings,
  listTasks
}
