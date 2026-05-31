export interface VoucherProduct {
  id: number
  name: string
  denomination: number
  retail_price: number
  stock: number
}

export interface VoucherBankAccount {
  id: number
  bank_name: string
  account_name: string
  account_number: string
  type?: string
}

export interface VoucherCheckoutConfig {
  enabled: boolean
  checkoutReady: boolean
  currency: string
  products: VoucherProduct[]
  bankAccounts: VoucherBankAccount[]
  orderTimeoutHours: number
  maxQuantityPerOrder: number
  reviewSlaHours: number
  feeRate: number
  helpText: string
}

export type VoucherOrderStatus =
  | 'pending_payment'
  | 'payment_submitted'
  | 'payment_verified'
  | 'fulfilling'
  | 'completed'
  | 'rejected'
  | 'expired'

export interface VoucherPinDelivery {
  pin_code: string
  serial: string
  denomination: number
  expires_at: string
  masked?: boolean
}

export interface VoucherOrder {
  id: number
  order_no: string
  status: VoucherOrderStatus
  product_name: string
  denomination: number
  quantity: number
  subtotal?: number
  fee_amount?: number
  total_amount: number
  currency: string
  payment_ref?: string
  reject_reason?: string
  fulfill_error?: string
  expires_at: string
  created_at: string
  completed_at?: string
  pins?: VoucherPinDelivery[]
}

/** @deprecated use VoucherOrder */
export type VoucherMockOrder = VoucherOrder

export type VoucherWizardStep = 'select' | 'confirm' | 'pay' | 'proof' | 'waiting' | 'completed'

export const EMPTY_VOUCHER_CHECKOUT: VoucherCheckoutConfig = {
  enabled: false,
  checkoutReady: false,
  currency: 'MYR',
  products: [],
  bankAccounts: [],
  orderTimeoutHours: 24,
  maxQuantityPerOrder: 10,
  reviewSlaHours: 24,
  feeRate: 0,
  helpText: '',
}
