import { apiClient } from './client'
import type { VoucherCheckoutConfig, VoucherOrder } from '@/types/voucher'

export interface VoucherCheckoutInfoResponse {
  enabled: boolean
  checkout_ready: boolean
  currency: string
  products: Array<{
    id: number
    name: string
    denomination: number
    retail_price: number
    stock: number
    currency?: string
  }>
  bank_accounts: Array<{
    id: number
    bank_name: string
    account_name: string
    account_number: string
    type?: string
  }>
  order_timeout_hours: number
  max_quantity: number
  review_sla_hours: number
  fee_rate: number
  help_text: string
  sandbox?: boolean
}

export function mapCheckoutInfo(data: VoucherCheckoutInfoResponse): VoucherCheckoutConfig {
  return {
    enabled: data.enabled,
    checkoutReady: data.checkout_ready,
    currency: data.currency || 'MYR',
    products: (data.products || []).map((p) => ({
      id: p.id,
      name: p.name,
      denomination: p.denomination,
      retail_price: p.retail_price,
      stock: p.stock,
    })),
    bankAccounts: data.bank_accounts || [],
    orderTimeoutHours: data.order_timeout_hours || 24,
    maxQuantityPerOrder: data.max_quantity || 10,
    reviewSlaHours: data.review_sla_hours || 24,
    feeRate: data.fee_rate || 0,
    helpText: data.help_text || '',
  }
}

export const voucherAPI = {
  getCheckoutInfo() {
    return apiClient.get<VoucherCheckoutInfoResponse>('/voucher/checkout-info')
  },

  createOrder(data: { product_id: number; quantity: number; idempotency_key?: string }) {
    return apiClient.post<{ order: VoucherOrder }>('/voucher/orders', data)
  },

  getOrder(id: number, includePins = false) {
    return apiClient.get<{ order: VoucherOrder }>(`/voucher/orders/${id}`, {
      params: includePins ? { include_pins: '1' } : undefined,
    })
  },

  listMyOrders(params?: { page?: number; per_page?: number }) {
    return apiClient.get<{ orders: VoucherOrder[]; pagination: { page: number; per_page: number; total: number; total_pages: number } }>(
      '/voucher/orders',
      { params },
    )
  },

  submitPaymentProof(orderId: number, form: FormData) {
    return apiClient.post<{ order: VoucherOrder }>(`/voucher/orders/${orderId}/payment-proof`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  cancelOrder(orderId: number) {
    return apiClient.post<{ order: VoucherOrder }>(`/voucher/orders/${orderId}/cancel`)
  },
}
