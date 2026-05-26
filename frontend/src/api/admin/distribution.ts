import { apiClient } from '@/api/client'
import type { BasePaginationResponse } from '@/types'

export type DistributionOrganizationType = 'platform' | 'reseller' | 'oem'
export type DistributionStatus = 'active' | 'inactive' | 'disabled'
export type DistributionMemberRole = 'manager' | 'agent' | 'kol1' | 'kol2'
export type DistributionCommissionStatus = 'frozen' | 'available' | 'settled' | 'cancelled' | 'reversed'
export type DistributionPromotionTargetType = 'registration' | 'oauth' | 'manual'

export interface DistributionOrganization {
  id: number
  type: DistributionOrganizationType
  name: string
  owner_user_id?: number | null
  status: DistributionStatus
  config: Record<string, unknown>
  brand_config: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface DistributionMember {
  member_id: number
  user_id: number
  user_email: string
  username: string
  channel_org_id: number
  role_type: DistributionMemberRole
  parent_member_id?: number | null
  level_code: string
  commission_rate: number
  status: DistributionStatus
  created_at: string
  updated_at: string
}

export interface DistributionAttribution {
  user_id: number
  user_email: string
  username: string
  channel_org_id: number
  referrer_member_id?: number | null
  promotion_link_id?: number | null
  bound_at: string
  bound_source: string
  bound_by: string
  created_at: string
  updated_at: string
}

export interface DistributionAttributionRecord {
  user_id: number
  channel_org_id: number
  referrer_member_id?: number | null
  promotion_link_id?: number | null
  bound_at: string
  bound_source: string
  bound_by: string
  audit_id?: number | null
  created_at: string
  updated_at: string
}

export interface DistributionAttributionAudit {
  id: number
  user_id: number
  user_email: string
  username: string
  previous_channel_org_id?: number | null
  previous_referrer_member_id?: number | null
  previous_promotion_link_id?: number | null
  previous_bound_source: string
  previous_bound_by: string
  new_channel_org_id: number
  new_referrer_member_id?: number | null
  new_promotion_link_id?: number | null
  new_bound_source: string
  new_bound_by: string
  note: string
  operator_user_id?: number | null
  operator_user_email: string
  operator_username: string
  created_at: string
}

export interface DistributionCommission {
  id: number
  channel_org_id: number
  member_id: number
  user_id: number
  user_email: string
  username: string
  usage_log_id?: number | null
  commission_type: string
  base_amount: number
  rate: number
  amount: number
  status: DistributionCommissionStatus | string
  settlement_method: string
  settlement_reference_no: string
  settlement_note: string
  frozen_until?: string | null
  settled_at?: string | null
  settled_by_user_id?: number | null
  reversed_from_id?: number | null
  created_at: string
  updated_at: string
}

export interface DistributionPromotionLink {
  id: number
  channel_org_id: number
  member_id: number
  code: string
  target_type: DistributionPromotionTargetType
  status: DistributionStatus
  user_id: number
  user_email: string
  username: string
  role_type: DistributionMemberRole
  created_at: string
  updated_at: string
}

export interface DistributionWallet {
  channel_org_id: number
  organization_name: string
  organization_type: DistributionOrganizationType
  prepaid_balance: number
  commission_reserved: number
  total_recharged: number
  total_consumed: number
  warning_threshold: number
  status: DistributionStatus
  created_at: string
  updated_at: string
}

export interface DistributionWalletTransaction {
  id: number
  channel_org_id: number
  organization_name: string
  organization_type: DistributionOrganizationType
  transaction_type: string
  amount: number
  prepaid_balance_before: number
  prepaid_balance_after: number
  commission_reserved_before: number
  commission_reserved_after: number
  reference_no: string
  note: string
  operator_user_id?: number | null
  created_at: string
}

export interface DistributionAlertEvent {
  id: number
  channel_org_id: number
  organization_name: string
  organization_type: DistributionOrganizationType
  alert_type: 'low_balance' | 'balance_exhausted' | 'consumption_warning' | 'consumption_exhausted' | string
  severity: 'info' | 'warning' | 'critical' | string
  status: 'active' | 'resolved' | string
  details: Record<string, unknown>
  triggered_at: string
  resolved_at?: string | null
  last_observed_at: string
  created_at: string
  updated_at: string
}

export interface DistributionWalletRequest {
  id: number
  channel_org_id: number
  organization_name: string
  organization_type: DistributionOrganizationType
  request_type: 'recharge' | 'refund' | string
  amount: number
  reference_no: string
  note: string
  status: 'pending' | 'approved' | 'rejected' | string
  created_by_user_id: number
  created_by_user_email: string
  created_by_username: string
  reviewed_by_user_id?: number | null
  reviewed_by_user_email: string
  reviewed_by_username: string
  review_note: string
  reviewed_at?: string | null
  created_at: string
}

export interface DistributionWalletRefundResult {
  wallet: DistributionWallet
  refund_amount: number
  fee_rate: number
  fee_amount: number
  net_amount: number
  reference_no: string
  note: string
  processed_mock: boolean
}

export interface DistributionAdminStats {
  organization_count: number
  platform_count: number
  reseller_count: number
  oem_count: number
  member_count: number
  agent_count: number
  kol1_count: number
  kol2_count: number
  promotion_link_count: number
  attribution_count: number
  commission_count: number
  wallet_count: number
  prepaid_balance_total: number
  commission_reserved_total: number
  total_recharged: number
  total_consumed: number
  frozen_commission_amount: number
  available_commission_amount: number
  settled_commission_amount: number
  commission_expense_ratio: number
  commission_upper_ratio: number
}

export interface DistributionChannelSummary {
  organization: DistributionOrganization
  wallet: DistributionWallet
  member_count: number
  agent_count: number
  kol1_count: number
  kol2_count: number
  promotion_link_count: number
  attribution_count: number
  commission_count: number
  frozen_commission_amount: number
  available_commission_amount: number
  settled_commission_amount: number
}

export interface DistributionListParams {
  page?: number
  page_size?: number
  channel_org_id?: number
  user_id?: number
  role_type?: DistributionMemberRole | string
  transaction_type?: string
  request_type?: string
  alert_type?: string
  severity?: string
  status?: string
  q?: string
}

export interface CreateDistributionOrganizationRequest {
  type: DistributionOrganizationType
  name: string
  owner_user_id?: number | null
  status?: DistributionStatus
  config?: Record<string, unknown>
  brand_config?: Record<string, unknown>
}

export interface UpdateDistributionOrganizationRequest extends CreateDistributionOrganizationRequest {}

export interface CreateDistributionMemberRequest {
  channel_org_id: number
  user_id: number
  role_type: DistributionMemberRole
  parent_member_id?: number | null
  level_code?: string
  commission_rate: number
  status?: DistributionStatus
}

export interface CreateDistributionPromotionLinkRequest {
  member_id: number
  code?: string
  target_type?: DistributionPromotionTargetType
  status?: DistributionStatus
}

export type DistributionOrganizationsResponse = BasePaginationResponse<DistributionOrganization>
export type DistributionMembersResponse = BasePaginationResponse<DistributionMember>
export type DistributionAttributionsResponse = BasePaginationResponse<DistributionAttribution>
export type DistributionAttributionAuditsResponse = BasePaginationResponse<DistributionAttributionAudit>
export type DistributionCommissionsResponse = BasePaginationResponse<DistributionCommission>
export type DistributionPromotionLinksResponse = BasePaginationResponse<DistributionPromotionLink>
export type DistributionWalletsResponse = BasePaginationResponse<DistributionWallet>
export type DistributionAlertEventsResponse = BasePaginationResponse<DistributionAlertEvent>
export type DistributionWalletRequestsResponse = BasePaginationResponse<DistributionWalletRequest>
export type DistributionWalletTransactionsResponse = BasePaginationResponse<DistributionWalletTransaction>

function cleanParams(params: DistributionListParams = {}) {
  return Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== undefined && value !== null && value !== ''),
  )
}

export async function listDistributionOrganizations(params: DistributionListParams = {}): Promise<DistributionOrganizationsResponse> {
  const { data } = await apiClient.get<DistributionOrganizationsResponse>('/admin/distribution/organizations', { params: cleanParams(params) })
  return data
}

export async function createDistributionOrganization(payload: CreateDistributionOrganizationRequest): Promise<DistributionOrganization> {
  const { data } = await apiClient.post<DistributionOrganization>('/admin/distribution/organizations', payload)
  return data
}

export async function updateDistributionOrganization(id: number, payload: UpdateDistributionOrganizationRequest): Promise<DistributionOrganization> {
  const { data } = await apiClient.put<DistributionOrganization>(`/admin/distribution/organizations/${id}`, payload)
  return data
}

export async function listDistributionMembers(params: DistributionListParams = {}): Promise<DistributionMembersResponse> {
  const { data } = await apiClient.get<DistributionMembersResponse>('/admin/distribution/members', { params: cleanParams(params) })
  return data
}

export async function createDistributionMember(payload: CreateDistributionMemberRequest): Promise<DistributionMember> {
  const { data } = await apiClient.post<DistributionMember>('/admin/distribution/members', payload)
  return data
}

export async function listDistributionAttributions(params: DistributionListParams = {}): Promise<DistributionAttributionsResponse> {
  const { data } = await apiClient.get<DistributionAttributionsResponse>('/admin/distribution/attributions', { params: cleanParams(params) })
  return data
}

export async function listDistributionAttributionAudits(params: DistributionListParams = {}): Promise<DistributionAttributionAuditsResponse> {
  const { data } = await apiClient.get<DistributionAttributionAuditsResponse>('/admin/distribution/attribution-audits', { params: cleanParams(params) })
  return data
}

export async function updateDistributionAttribution(userId: number, payload: {
  channel_org_id: number
  referrer_member_id?: number
  promotion_link_id?: number
  note?: string
}): Promise<DistributionAttributionRecord> {
  const { data } = await apiClient.put<DistributionAttributionRecord>(`/admin/distribution/attributions/${userId}`, payload)
  return data
}

export async function listDistributionCommissions(params: DistributionListParams = {}): Promise<DistributionCommissionsResponse> {
  const { data } = await apiClient.get<DistributionCommissionsResponse>('/admin/distribution/commissions', { params: cleanParams(params) })
  return data
}

export async function listDistributionPromotionLinks(params: DistributionListParams = {}): Promise<DistributionPromotionLinksResponse> {
  const { data } = await apiClient.get<DistributionPromotionLinksResponse>('/admin/distribution/promotion-links', { params: cleanParams(params) })
  return data
}

export async function createDistributionPromotionLink(payload: CreateDistributionPromotionLinkRequest): Promise<DistributionPromotionLink> {
  const { data } = await apiClient.post<DistributionPromotionLink>('/admin/distribution/promotion-links', payload)
  return data
}

export async function listDistributionWallets(params: DistributionListParams = {}): Promise<DistributionWalletsResponse> {
  const { data } = await apiClient.get<DistributionWalletsResponse>('/admin/distribution/wallets', { params: cleanParams(params) })
  return data
}

export async function listDistributionWalletRequests(params: DistributionListParams = {}): Promise<DistributionWalletRequestsResponse> {
  const { data } = await apiClient.get<DistributionWalletRequestsResponse>('/admin/distribution/wallet-requests', { params: cleanParams(params) })
  return data
}

export async function listDistributionAlertEvents(params: DistributionListParams = {}): Promise<DistributionAlertEventsResponse> {
  const { data } = await apiClient.get<DistributionAlertEventsResponse>('/admin/distribution/alert-events', { params: cleanParams(params) })
  return data
}

export async function listDistributionWalletTransactions(params: DistributionListParams = {}): Promise<DistributionWalletTransactionsResponse> {
  const { data } = await apiClient.get<DistributionWalletTransactionsResponse>('/admin/distribution/wallet-transactions', { params: cleanParams(params) })
  return data
}

export async function rechargeDistributionWallet(channelOrgId: number, payload: {
  amount: number
  reference_no?: string
  note?: string
}): Promise<DistributionWallet> {
  const { data } = await apiClient.post<DistributionWallet>(`/admin/distribution/wallets/${channelOrgId}/recharge`, payload)
  return data
}

export async function refundDistributionWallet(channelOrgId: number, payload: {
  amount: number
  reference_no?: string
  note?: string
}): Promise<DistributionWalletRefundResult> {
  const { data } = await apiClient.post<DistributionWalletRefundResult>(`/admin/distribution/wallets/${channelOrgId}/refund`, payload)
  return data
}

export async function reviewDistributionWalletRequest(requestId: number, payload: {
  action: 'approve' | 'reject'
  review_note?: string
}): Promise<DistributionWalletRequest> {
  const { data } = await apiClient.post<DistributionWalletRequest>(`/admin/distribution/wallet-requests/${requestId}/review`, payload)
  return data
}

export async function updateDistributionWalletWarningThreshold(channelOrgId: number, warning_threshold: number): Promise<DistributionWallet> {
  const { data } = await apiClient.put<DistributionWallet>(`/admin/distribution/wallets/${channelOrgId}/warning-threshold`, { warning_threshold })
  return data
}

export async function getDistributionStats(): Promise<DistributionAdminStats> {
  const { data } = await apiClient.get<DistributionAdminStats>('/admin/distribution/stats')
  return data
}

export async function settleDistributionCommission(commissionId: number, settlement_method: string = 'manual'): Promise<DistributionCommission> {
  const { data } = await apiClient.post<DistributionCommission>(`/admin/distribution/commissions/${commissionId}/settle`, { settlement_method })
  return data
}

export async function reverseDistributionCommission(commissionId: number): Promise<DistributionCommission> {
  const { data } = await apiClient.post<DistributionCommission>(`/admin/distribution/commissions/${commissionId}/reverse`)
  return data
}

export const adminDistributionAPI = {
  listDistributionOrganizations,
  createDistributionOrganization,
  updateDistributionOrganization,
  listDistributionMembers,
  createDistributionMember,
  listDistributionPromotionLinks,
  createDistributionPromotionLink,
  listDistributionWallets,
  listDistributionAlertEvents,
  listDistributionWalletRequests,
  listDistributionWalletTransactions,
  rechargeDistributionWallet,
  refundDistributionWallet,
  reviewDistributionWalletRequest,
  updateDistributionWalletWarningThreshold,
  getDistributionStats,
  settleDistributionCommission,
  reverseDistributionCommission,
  listDistributionAttributions,
  listDistributionAttributionAudits,
  updateDistributionAttribution,
  listDistributionCommissions,
}

export default adminDistributionAPI
