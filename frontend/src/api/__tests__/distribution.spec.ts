import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, put } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
    put,
  },
}))

import {
  createDistributionMember,
  createDistributionOrganization,
  createDistributionPromotionLink,
  getDistributionStats,
  listDistributionAttributions,
  listDistributionAttributionAudits,
  listDistributionAlertEvents,
  listDistributionCommissions,
  listDistributionMembers,
  listDistributionPromotionLinks,
  listDistributionWalletRequests,
  listDistributionWalletTransactions,
  listDistributionWallets,
  listDistributionOrganizations,
  rechargeDistributionWallet,
  reviewDistributionWalletRequest,
  refundDistributionWallet,
  reverseDistributionCommission,
  settleDistributionCommission,
  updateDistributionAttribution,
  updateDistributionOrganization,
  updateDistributionWalletWarningThreshold,
} from '@/api/admin/distribution'
import {
  createMyDistributionMember,
  createMyDistributionPromotionLink,
  getDistributionOverview,
  getMyDistributionAnalytics,
  getMyDistributionOrganization,
  listMyDistributionAttributions,
  listMyDistributionAlertEvents,
  listMyDistributionCommissions,
  listMyDistributionMembers,
  listMyDistributionPromotionLinks,
  listMyDistributionWalletRequests,
  listMyDistributionWholesalePricing,
  listMyDistributionWalletTransactions,
  settleMyDistributionCommission,
  submitMyDistributionWalletRequest,
  updateMyDistributionOrganization,
} from '@/api/distribution'

describe('distribution api clients', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
  })

  it('calls isolated admin distribution endpoints', async () => {
    get.mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 } })
    post.mockResolvedValue({ data: { id: 1 } })
    put.mockResolvedValue({ data: { channel_org_id: 8 } })

    await listDistributionOrganizations({ page: 2, page_size: 10 })
    await listDistributionMembers({ channel_org_id: 8, role_type: 'agent' })
    await listDistributionAttributions({ channel_org_id: 8 })
    await listDistributionAttributionAudits({ channel_org_id: 8, user_id: 7 })
    await listDistributionCommissions({ channel_org_id: 8 })
    await listDistributionPromotionLinks({ channel_org_id: 8, role_type: 'agent' })
    await listDistributionWallets({ channel_org_id: 8 })
    await listDistributionAlertEvents({ channel_org_id: 8, alert_type: 'low_balance', status: 'active', severity: 'warning' })
    await listDistributionWalletRequests({ channel_org_id: 8, request_type: 'recharge', status: 'pending' })
    await listDistributionWalletTransactions({ channel_org_id: 8, transaction_type: 'recharge' })
    await getDistributionStats()
    await createDistributionOrganization({ type: 'reseller', name: 'Agent' })
    await updateDistributionOrganization(88, { type: 'reseller', name: 'Agent Updated' })
    await createDistributionMember({ channel_org_id: 8, user_id: 9, role_type: 'agent', commission_rate: 0.1 })
    await createDistributionPromotionLink({ member_id: 11, code: 'LINK-1' })
    await updateDistributionAttribution(7, { channel_org_id: 88, referrer_member_id: 11, promotion_link_id: 12, note: 'manual reassignment' })
    await rechargeDistributionWallet(88, { amount: 120, reference_no: 'BANK-1' })
    await refundDistributionWallet(88, { amount: 80, reference_no: 'RF-1' })
    await reviewDistributionWalletRequest(31, { action: 'approve', review_note: '到账确认' })
    await updateDistributionWalletWarningThreshold(88, 50)
    await settleDistributionCommission(1001, 'manual')
    await reverseDistributionCommission(1002)

    expect(get).toHaveBeenNthCalledWith(1, '/admin/distribution/organizations', { params: { page: 2, page_size: 10 } })
    expect(get).toHaveBeenNthCalledWith(2, '/admin/distribution/members', { params: { channel_org_id: 8, role_type: 'agent' } })
    expect(get).toHaveBeenNthCalledWith(3, '/admin/distribution/attributions', { params: { channel_org_id: 8 } })
    expect(get).toHaveBeenNthCalledWith(4, '/admin/distribution/attribution-audits', { params: { channel_org_id: 8, user_id: 7 } })
    expect(get).toHaveBeenNthCalledWith(5, '/admin/distribution/commissions', { params: { channel_org_id: 8 } })
    expect(get).toHaveBeenNthCalledWith(6, '/admin/distribution/promotion-links', { params: { channel_org_id: 8, role_type: 'agent' } })
    expect(get).toHaveBeenNthCalledWith(7, '/admin/distribution/wallets', { params: { channel_org_id: 8 } })
    expect(get).toHaveBeenNthCalledWith(8, '/admin/distribution/alert-events', { params: { channel_org_id: 8, alert_type: 'low_balance', status: 'active', severity: 'warning' } })
    expect(get).toHaveBeenNthCalledWith(9, '/admin/distribution/wallet-requests', { params: { channel_org_id: 8, request_type: 'recharge', status: 'pending' } })
    expect(get).toHaveBeenNthCalledWith(10, '/admin/distribution/wallet-transactions', { params: { channel_org_id: 8, transaction_type: 'recharge' } })
    expect(get).toHaveBeenNthCalledWith(11, '/admin/distribution/stats')
    expect(post).toHaveBeenNthCalledWith(1, '/admin/distribution/organizations', { type: 'reseller', name: 'Agent' })
    expect(put).toHaveBeenNthCalledWith(1, '/admin/distribution/organizations/88', { type: 'reseller', name: 'Agent Updated' })
    expect(post).toHaveBeenNthCalledWith(2, '/admin/distribution/members', { channel_org_id: 8, user_id: 9, role_type: 'agent', commission_rate: 0.1 })
    expect(post).toHaveBeenNthCalledWith(3, '/admin/distribution/promotion-links', { member_id: 11, code: 'LINK-1' })
    expect(put).toHaveBeenNthCalledWith(2, '/admin/distribution/attributions/7', { channel_org_id: 88, referrer_member_id: 11, promotion_link_id: 12, note: 'manual reassignment' })
    expect(post).toHaveBeenNthCalledWith(4, '/admin/distribution/wallets/88/recharge', { amount: 120, reference_no: 'BANK-1' })
    expect(post).toHaveBeenNthCalledWith(5, '/admin/distribution/wallets/88/refund', { amount: 80, reference_no: 'RF-1' })
    expect(post).toHaveBeenNthCalledWith(6, '/admin/distribution/wallet-requests/31/review', { action: 'approve', review_note: '到账确认' })
    expect(put).toHaveBeenNthCalledWith(3, '/admin/distribution/wallets/88/warning-threshold', { warning_threshold: 50 })
    expect(post).toHaveBeenNthCalledWith(7, '/admin/distribution/commissions/1001/settle', { settlement_method: 'manual' })
    expect(post).toHaveBeenNthCalledWith(8, '/admin/distribution/commissions/1002/reverse')
  })

  it('calls current-user scoped distribution endpoints without channel params', async () => {
    get.mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 } })
    post.mockResolvedValue({ data: { id: 1 } })
    put.mockResolvedValue({ data: { id: 8, name: 'Channel A' } })

    await getDistributionOverview()
    await getMyDistributionAnalytics({ start_date: '2026-05-01', end_date: '2026-05-24', granularity: 'day', limit: 10 })
    await listMyDistributionMembers({ role_type: 'kol1' })
    await listMyDistributionAttributions({ page: 1 })
    await listMyDistributionCommissions({ page: 1 })
    await listMyDistributionPromotionLinks({ page: 1 })
    await listMyDistributionWholesalePricing({ page: 1, q: 'claude' })
    await listMyDistributionAlertEvents({ page: 1, alert_type: 'low_balance', status: 'active', severity: 'warning' })
    await listMyDistributionWalletRequests({ page: 1, request_type: 'refund', status: 'pending' })
    await listMyDistributionWalletTransactions({ page: 1, transaction_type: 'consume' })
    await getMyDistributionOrganization()
    await updateMyDistributionOrganization({ config: { commission_settlement_method: 'offline' } })
    await createMyDistributionMember({ user_id: 12, role_type: 'kol1', parent_member_id: 11, commission_rate: 0.1 })
    await createMyDistributionPromotionLink({ member_id: 11, target_type: 'registration' })
    await submitMyDistributionWalletRequest({ request_type: 'recharge', amount: 300, reference_no: 'BANK-1' })
    await settleMyDistributionCommission(1001, { settlement_method: 'offline', settlement_reference_no: 'VCH-1' })

    expect(get).toHaveBeenNthCalledWith(1, '/distribution/overview')
    expect(get).toHaveBeenNthCalledWith(2, '/distribution/analytics', { params: { start_date: '2026-05-01', end_date: '2026-05-24', granularity: 'day', limit: 10 } })
    expect(get).toHaveBeenNthCalledWith(3, '/distribution/members', { params: { role_type: 'kol1' } })
    expect(get).toHaveBeenNthCalledWith(4, '/distribution/attributions', { params: { page: 1 } })
    expect(get).toHaveBeenNthCalledWith(5, '/distribution/commissions', { params: { page: 1 } })
    expect(get).toHaveBeenNthCalledWith(6, '/distribution/promotion-links', { params: { page: 1 } })
    expect(get).toHaveBeenNthCalledWith(7, '/distribution/wholesale-pricing', { params: { page: 1, q: 'claude' } })
    expect(get).toHaveBeenNthCalledWith(8, '/distribution/alert-events', { params: { page: 1, alert_type: 'low_balance', status: 'active', severity: 'warning' } })
    expect(get).toHaveBeenNthCalledWith(9, '/distribution/wallet-requests', { params: { page: 1, request_type: 'refund', status: 'pending' } })
    expect(get).toHaveBeenNthCalledWith(10, '/distribution/wallet-transactions', { params: { page: 1, transaction_type: 'consume' } })
    expect(get).toHaveBeenNthCalledWith(11, '/distribution/organization')
    expect(put).toHaveBeenNthCalledWith(1, '/distribution/organization', { config: { commission_settlement_method: 'offline' } })
    expect(post).toHaveBeenNthCalledWith(1, '/distribution/members', { user_id: 12, role_type: 'kol1', parent_member_id: 11, commission_rate: 0.1 })
    expect(post).toHaveBeenNthCalledWith(2, '/distribution/promotion-links', { member_id: 11, target_type: 'registration' })
    expect(post).toHaveBeenNthCalledWith(3, '/distribution/wallet-requests', { request_type: 'recharge', amount: 300, reference_no: 'BANK-1' })
    expect(post).toHaveBeenNthCalledWith(4, '/distribution/commissions/1001/settle', { settlement_method: 'offline', settlement_reference_no: 'VCH-1' })
  })
})
