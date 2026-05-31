import apiClient from '../client'
import type { VoucherBankAccount, VoucherOrder } from '@/types/voucher'

export interface VoucherAdminSettings {
  enabled: boolean
  ui_enabled: boolean
  sandbox: boolean
  sandbox_from_key: boolean
  api_base: string
  api_key_masked: string
  secret_configured: boolean
  bank_accounts: VoucherBankAccount[]
  order_timeout_hours: number
  max_quantity_per_order: number
  review_sla_hours: number
  fee_rate: number
  help_text: string
  retail_markup_percent: number
}

export interface VoucherAdminSettingsUpdate {
  enabled?: boolean
  ui_enabled?: boolean
  sandbox?: boolean
  api_key?: string
  api_secret?: string
  api_base?: string
  bank_accounts?: VoucherBankAccount[]
  order_timeout_hours?: number
  max_quantity_per_order?: number
  review_sla_hours?: number
  fee_rate?: number
  help_text?: string
  retail_markup_percent?: number
}

export interface VoucherTestConnectionResult {
  ok: boolean
  request_id?: string
  message: string
  configured: boolean
  account?: {
    merchant_id: number
    company_name: string
    currency: string
    api_key_scope: string
    is_test_mode: boolean
  }
}

export const voucherAdminAPI = {
  getSettings() {
    return apiClient.get<VoucherAdminSettings>('/admin/voucher/settings')
  },

  updateSettings(data: VoucherAdminSettingsUpdate) {
    return apiClient.put<VoucherAdminSettings>('/admin/voucher/settings', data)
  },

  testConnection() {
    return apiClient.post<VoucherTestConnectionResult>('/admin/voucher/test-connection')
  },

  syncCatalog() {
    return apiClient.post<{ synced: number }>('/admin/voucher/sync-catalog')
  },

  syncStock() {
    return apiClient.post<{ updated: number }>('/admin/voucher/sync-stock')
  },

  listOrders(params?: { page?: number; per_page?: number; status?: string }) {
    return apiClient.get<{
      orders: VoucherOrder[]
      pagination: { page: number; per_page: number; total: number; total_pages: number }
    }>('/admin/voucher/orders', { params })
  },

  getOrder(id: number) {
    return apiClient.get<{ order: VoucherOrder }>(`/admin/voucher/orders/${id}`)
  },

  verifyOrder(id: number) {
    return apiClient.post<{ order: VoucherOrder }>(`/admin/voucher/orders/${id}/verify`)
  },

  rejectOrder(id: number, reason?: string) {
    return apiClient.post<{ order: VoucherOrder }>(`/admin/voucher/orders/${id}/reject`, { reason })
  },

  retryFulfill(id: number) {
    return apiClient.post<{ order: VoucherOrder }>(`/admin/voucher/orders/${id}/retry-fulfill`)
  },
}
