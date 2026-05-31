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

export interface VoucherAdminProduct {
  id: number
  kv_product_id?: number
  name: string
  denomination: number
  wholesale_price: number
  retail_price: number
  stock: number
  currency: string
  is_active: boolean
}

export interface VoucherB2BOrderLine {
  product_id: number
  name?: string
  denomination?: number
  quantity: number
  unit_price?: number
  line_total?: number
}

export interface VoucherB2BOrder {
  id: number
  kv_order_id: number
  order_no: string
  status: string
  subtotal: number
  fee_amount: number
  total_amount: number
  currency: string
  items: VoucherB2BOrderLine[]
  payment_ref?: string
  bank_account_id?: number
  merchant_notes?: string
  reject_reason?: string
  payment_info?: {
    reference?: string
    amount_due?: number
    currency?: string
    instructions?: string
    bank_accounts?: Array<{
      id: number
      bank_name: string
      account_name: string
      account_number: string
      type?: string
    }>
  }
  kv_last_request_id?: string
  kv_last_synced_at?: string
  created_by: string
  verified_at?: string
  pins_loaded_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export interface VoucherAuditEntry {
  id: number
  action: string
  operator: string
  metadata?: Record<string, unknown>
  created_at: string
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

  listProducts() {
    return apiClient.get<{ products: VoucherAdminProduct[] }>('/admin/voucher/products')
  },

  listB2BOrders(params?: { page?: number; per_page?: number; status?: string }) {
    return apiClient.get<{
      orders: VoucherB2BOrder[]
      pagination: { page: number; per_page: number; total: number; total_pages: number }
    }>('/admin/voucher/b2b/orders', { params })
  },

  createB2BOrder(data: {
    items: Array<{ product_id: number; quantity: number }>
    currency?: string
    merchant_notes?: string
    idempotency_key?: string
  }) {
    return apiClient.post<{ order: VoucherB2BOrder }>('/admin/voucher/b2b/orders', data)
  },

  getB2BOrder(id: number) {
    return apiClient.get<{ order: VoucherB2BOrder; audit: VoucherAuditEntry[] }>(`/admin/voucher/b2b/orders/${id}`)
  },

  syncB2BOrder(id: number) {
    return apiClient.post<{ order: VoucherB2BOrder }>(`/admin/voucher/b2b/orders/${id}/sync`)
  },

  submitB2BProof(id: number, form: FormData) {
    return apiClient.post<{ order: VoucherB2BOrder }>(`/admin/voucher/b2b/orders/${id}/payment-proof`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}
