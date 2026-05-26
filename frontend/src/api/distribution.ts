import { apiClient } from '@/api/client'
import type { BasePaginationResponse } from '@/types'
import type {
  DistributionAttribution,
  DistributionChannelSummary,
  DistributionCommission,
  DistributionOrganization,
  DistributionPromotionLink,
  DistributionListParams,
  DistributionMember,
  DistributionAlertEvent,
  DistributionWalletTransaction,
  DistributionWalletRequest,
} from '@/api/admin/distribution'

export interface DistributionOverview {
  user_id: number
  channel_org_id: number
  can_manage_channel: boolean
  summary: DistributionChannelSummary
}

export type MyDistributionMembersResponse = BasePaginationResponse<DistributionMember>
export type MyDistributionAttributionsResponse = BasePaginationResponse<DistributionAttribution>
export type MyDistributionCommissionsResponse = BasePaginationResponse<DistributionCommission>
export type MyDistributionPromotionLinksResponse = BasePaginationResponse<DistributionPromotionLink>
export type MyDistributionAlertEventsResponse = BasePaginationResponse<DistributionAlertEvent>
export type MyDistributionWalletRequestsResponse = BasePaginationResponse<DistributionWalletRequest>
export type MyDistributionWalletTransactionsResponse = BasePaginationResponse<DistributionWalletTransaction>
export interface DistributionWholesalePricingItem {
  model: string
  provider: string
  billing_mode: string
  official_input_price: number
  official_output_price: number
  official_cache_write_price: number
  official_cache_read_price: number
  official_image_price: number
  wholesale_input_price: number
  wholesale_output_price: number
  wholesale_cache_write_price: number
  wholesale_cache_read_price: number
  wholesale_image_price: number
}

export interface MyDistributionWholesalePricingResponse extends BasePaginationResponse<DistributionWholesalePricingItem> {
  discount_rate: number
}

export interface DistributionAnalyticsSummary {
  registered_users: number
  recharge_amount: number
  consumption_amount: number
  commission_amount: number
  settled_commission_amount: number
  member_count: number
  agent_count: number
  kol1_count: number
  kol2_count: number
  commission_expense_ratio: number
  commission_upper_ratio: number
}

export interface DistributionAnalyticsTrendPoint {
  date: string
  registered_users: number
  recharge_amount: number
  consumption_amount: number
  commission_amount: number
  settled_commission_amount: number
}

export interface DistributionAnalyticsRankingItem {
  member_id: number
  user_id: number
  user_email: string
  username: string
  role_type: string
  registered_users: number
  recharge_amount: number
  consumption_amount: number
  commission_amount: number
  settled_commission_amount: number
}

export interface DistributionAnalyticsFilter {
  start_date: string
  end_date: string
  granularity: string
  limit: number
}

export interface DistributionAnalyticsChannel {
  summary: DistributionAnalyticsSummary
  trend: DistributionAnalyticsTrendPoint[]
  member_ranking: DistributionAnalyticsRankingItem[]
}

export interface DistributionAnalyticsPersonal {
  role_types: string[]
  summary: DistributionAnalyticsSummary
  child_member_ranking: DistributionAnalyticsRankingItem[]
}

export interface MyDistributionAnalyticsResponse {
  can_manage_channel: boolean
  filter: DistributionAnalyticsFilter
  channel?: DistributionAnalyticsChannel | null
  personal?: DistributionAnalyticsPersonal | null
}

export interface DistributionAnalyticsParams {
  start_date?: string
  end_date?: string
  granularity?: 'hour' | 'day' | 'week' | 'month' | string
  limit?: number
}

function cleanParams<T extends object>(params: T) {
  return Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== undefined && value !== null && value !== ''),
  )
}

export async function getDistributionOverview(): Promise<DistributionOverview> {
  const { data } = await apiClient.get<DistributionOverview>('/distribution/overview')
  return data
}

export async function listMyDistributionMembers(params: DistributionListParams = {}): Promise<MyDistributionMembersResponse> {
  const { data } = await apiClient.get<MyDistributionMembersResponse>('/distribution/members', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionAttributions(params: DistributionListParams = {}): Promise<MyDistributionAttributionsResponse> {
  const { data } = await apiClient.get<MyDistributionAttributionsResponse>('/distribution/attributions', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionCommissions(params: DistributionListParams = {}): Promise<MyDistributionCommissionsResponse> {
  const { data } = await apiClient.get<MyDistributionCommissionsResponse>('/distribution/commissions', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionPromotionLinks(params: DistributionListParams = {}): Promise<MyDistributionPromotionLinksResponse> {
  const { data } = await apiClient.get<MyDistributionPromotionLinksResponse>('/distribution/promotion-links', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionWalletRequests(params: DistributionListParams = {}): Promise<MyDistributionWalletRequestsResponse> {
  const { data } = await apiClient.get<MyDistributionWalletRequestsResponse>('/distribution/wallet-requests', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionAlertEvents(params: DistributionListParams = {}): Promise<MyDistributionAlertEventsResponse> {
  const { data } = await apiClient.get<MyDistributionAlertEventsResponse>('/distribution/alert-events', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionWalletTransactions(params: DistributionListParams = {}): Promise<MyDistributionWalletTransactionsResponse> {
  const { data } = await apiClient.get<MyDistributionWalletTransactionsResponse>('/distribution/wallet-transactions', { params: cleanParams(params) })
  return data
}

export async function listMyDistributionWholesalePricing(params: DistributionListParams = {}): Promise<MyDistributionWholesalePricingResponse> {
  const { data } = await apiClient.get<MyDistributionWholesalePricingResponse>('/distribution/wholesale-pricing', { params: cleanParams(params) })
  return data
}

export async function getMyDistributionAnalytics(params: DistributionAnalyticsParams = {}): Promise<MyDistributionAnalyticsResponse> {
  const { data } = await apiClient.get<MyDistributionAnalyticsResponse>('/distribution/analytics', { params: cleanParams(params) })
  return data
}

export async function getMyDistributionOrganization(): Promise<DistributionOrganization> {
  const { data } = await apiClient.get<DistributionOrganization>('/distribution/organization')
  return data
}

export async function updateMyDistributionOrganization(payload: {
  name?: string
  config?: Record<string, unknown>
  brand_config?: Record<string, unknown>
}): Promise<DistributionOrganization> {
  const { data } = await apiClient.put<DistributionOrganization>('/distribution/organization', payload)
  return data
}

export async function createMyDistributionMember(payload: {
  user_id: number
  role_type: 'agent' | 'kol1' | 'kol2'
  parent_member_id?: number
  level_code?: string
  commission_rate: number
  status?: 'active' | 'inactive' | 'disabled'
}): Promise<DistributionMember> {
  const { data } = await apiClient.post<DistributionMember>('/distribution/members', payload)
  return data
}

export async function createMyDistributionPromotionLink(payload: {
  member_id: number
  code?: string
  target_type?: 'registration' | 'oauth' | 'manual'
  status?: 'active' | 'inactive' | 'disabled'
}): Promise<DistributionPromotionLink> {
  const { data } = await apiClient.post<DistributionPromotionLink>('/distribution/promotion-links', payload)
  return data
}

export async function submitMyDistributionWalletRequest(payload: {
  request_type: 'recharge' | 'refund'
  amount: number
  reference_no?: string
  note?: string
}): Promise<DistributionWalletRequest> {
  const { data } = await apiClient.post<DistributionWalletRequest>('/distribution/wallet-requests', payload)
  return data
}

export async function settleMyDistributionCommission(commissionId: number, payload: {
  settlement_method: 'manual' | 'offline' | 'balance' | 'auto'
  settlement_reference_no?: string
  settlement_note?: string
}): Promise<DistributionCommission> {
  const { data } = await apiClient.post<DistributionCommission>(`/distribution/commissions/${commissionId}/settle`, payload)
  return data
}

export const distributionAPI = {
  getDistributionOverview,
  getMyDistributionOrganization,
  getMyDistributionAnalytics,
  updateMyDistributionOrganization,
  listMyDistributionMembers,
  listMyDistributionAttributions,
  listMyDistributionCommissions,
  listMyDistributionPromotionLinks,
  listMyDistributionAlertEvents,
  listMyDistributionWalletRequests,
  listMyDistributionWholesalePricing,
  listMyDistributionWalletTransactions,
  settleMyDistributionCommission,
  createMyDistributionMember,
  createMyDistributionPromotionLink,
  submitMyDistributionWalletRequest,
}

export default distributionAPI
