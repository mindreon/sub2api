import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

vi.hoisted(() => {
  Object.defineProperty(globalThis, 'localStorage', {
    value: {
      getItem: vi.fn(() => null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
    },
    configurable: true,
  })
})

import AdminDistributionView from '../distribution/AdminDistributionView.vue'

const {
  createDistributionMember,
  createDistributionOrganization,
  createDistributionPromotionLink,
  getDistributionStats,
  listDistributionAttributions,
  listDistributionAttributionAudits,
  listDistributionAlertEvents,
  listDistributionCommissions,
  listDistributionMembers,
  listDistributionOrganizations,
  listDistributionPromotionLinks,
  listDistributionWalletRequests,
  listDistributionWalletTransactions,
  listDistributionWallets,
  rechargeDistributionWallet,
  reviewDistributionWalletRequest,
  refundDistributionWallet,
  reverseDistributionCommission,
  settleDistributionCommission,
  updateDistributionAttribution,
  updateDistributionOrganization,
  updateDistributionWalletWarningThreshold,
} = vi.hoisted(() => ({
  createDistributionMember: vi.fn(),
  createDistributionOrganization: vi.fn(),
  createDistributionPromotionLink: vi.fn(),
  getDistributionStats: vi.fn(),
  listDistributionAttributions: vi.fn(),
  listDistributionAttributionAudits: vi.fn(),
  listDistributionAlertEvents: vi.fn(),
  listDistributionCommissions: vi.fn(),
  listDistributionMembers: vi.fn(),
  listDistributionOrganizations: vi.fn(),
  listDistributionPromotionLinks: vi.fn(),
  listDistributionWalletRequests: vi.fn(),
  listDistributionWalletTransactions: vi.fn(),
  listDistributionWallets: vi.fn(),
  rechargeDistributionWallet: vi.fn(),
  reviewDistributionWalletRequest: vi.fn(),
  refundDistributionWallet: vi.fn(),
  reverseDistributionCommission: vi.fn(),
  settleDistributionCommission: vi.fn(),
  updateDistributionAttribution: vi.fn(),
  updateDistributionOrganization: vi.fn(),
  updateDistributionWalletWarningThreshold: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()
const { listUsers, getUserById } = vi.hoisted(() => ({
  listUsers: vi.fn(),
  getUserById: vi.fn(),
}))

vi.mock('@/api/admin/distribution', () => ({
  createDistributionMember,
  createDistributionOrganization,
  createDistributionPromotionLink,
  getDistributionStats,
  listDistributionAttributions,
  listDistributionAttributionAudits,
  listDistributionAlertEvents,
  listDistributionCommissions,
  listDistributionMembers,
  listDistributionOrganizations,
  listDistributionPromotionLinks,
  listDistributionWalletRequests,
  listDistributionWalletTransactions,
  listDistributionWallets,
  rechargeDistributionWallet,
  reviewDistributionWalletRequest,
  refundDistributionWallet,
  reverseDistributionCommission,
  settleDistributionCommission,
  updateDistributionAttribution,
  updateDistributionOrganization,
  updateDistributionWalletWarningThreshold,
  adminDistributionAPI: {
    createDistributionMember,
    createDistributionOrganization,
    createDistributionPromotionLink,
    getDistributionStats,
    listDistributionAttributions,
    listDistributionAttributionAudits,
    listDistributionAlertEvents,
    listDistributionCommissions,
    listDistributionMembers,
    listDistributionOrganizations,
    listDistributionPromotionLinks,
    listDistributionWalletRequests,
    listDistributionWalletTransactions,
    listDistributionWallets,
    rechargeDistributionWallet,
    reviewDistributionWalletRequest,
    refundDistributionWallet,
    reverseDistributionCommission,
    settleDistributionCommission,
    updateDistributionAttribution,
    updateDistributionOrganization,
    updateDistributionWalletWarningThreshold,
  },
  default: {
    createDistributionMember,
    createDistributionOrganization,
    createDistributionPromotionLink,
    getDistributionStats,
    listDistributionAttributions,
    listDistributionAttributionAudits,
    listDistributionCommissions,
    listDistributionMembers,
    listDistributionOrganizations,
    listDistributionPromotionLinks,
    listDistributionWalletRequests,
    listDistributionWalletTransactions,
    listDistributionWallets,
    rechargeDistributionWallet,
    reviewDistributionWalletRequest,
    refundDistributionWallet,
    reverseDistributionCommission,
    settleDistributionCommission,
    updateDistributionAttribution,
    updateDistributionOrganization,
    updateDistributionWalletWarningThreshold,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/api/admin', () => ({
  usersAPI: {
    list: listUsers,
    getById: getUserById,
  },
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'error',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const mockRoute = vi.hoisted(() => ({
  path: '/admin/distribution/organizations',
  params: { tab: 'organizations' },
}))

vi.mock('vue-router', () => ({
  useRoute: () => mockRoute,
  useRouter: () => ({
    replace: vi.fn(),
  }),
}))

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <div v-for="row in data" :key="row.channel_org_id || row.id" class="wallet-row">
        <slot name="cell-actions" :row="row" />
      </div>
    </div>
  `,
}

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<div v-if="show"><slot /></div>',
}

describe('AdminDistributionView', () => {
  beforeEach(() => {
    mockRoute.path = '/admin/distribution/organizations'
    mockRoute.params = { tab: 'organizations' }

    showError.mockReset()
    showSuccess.mockReset()
    listUsers.mockReset()
    getUserById.mockReset()
    createDistributionMember.mockReset()
    createDistributionOrganization.mockReset()
    createDistributionPromotionLink.mockReset()
    getDistributionStats.mockReset()
    listDistributionAttributions.mockReset()
    listDistributionAttributionAudits.mockReset()
    listDistributionAlertEvents.mockReset()
    listDistributionCommissions.mockReset()
    listDistributionMembers.mockReset()
    listDistributionOrganizations.mockReset()
    listDistributionPromotionLinks.mockReset()
    listDistributionWalletRequests.mockReset()
    listDistributionWalletTransactions.mockReset()
    listDistributionWallets.mockReset()
    rechargeDistributionWallet.mockReset()
    reviewDistributionWalletRequest.mockReset()
    refundDistributionWallet.mockReset()
    reverseDistributionCommission.mockReset()
    settleDistributionCommission.mockReset()
    updateDistributionAttribution.mockReset()
    updateDistributionOrganization.mockReset()
    updateDistributionWalletWarningThreshold.mockReset()

    getDistributionStats.mockResolvedValue({
      organization_count: 1,
      platform_count: 0,
      reseller_count: 1,
      oem_count: 0,
      member_count: 0,
      agent_count: 0,
      kol1_count: 0,
      kol2_count: 0,
      promotion_link_count: 0,
      attribution_count: 0,
      commission_count: 0,
      wallet_count: 1,
      prepaid_balance_total: 120,
      commission_reserved_total: 0,
      total_recharged: 120,
      total_consumed: 0,
      frozen_commission_amount: 0,
      available_commission_amount: 0,
      settled_commission_amount: 0,
    })
    listDistributionOrganizations.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listDistributionWallets.mockResolvedValue({
      items: [
        {
          channel_org_id: 88,
          organization_name: 'Independent Agent',
          organization_type: 'reseller',
          prepaid_balance: 120,
          commission_reserved: 0,
          total_recharged: 200,
          total_consumed: 80,
          warning_threshold: 50,
          status: 'active',
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listDistributionAlertEvents.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listDistributionAttributions.mockResolvedValue({
      items: [
        {
          user_id: 7,
          user_email: 'user@example.com',
          username: 'example-user',
          channel_org_id: 88,
          referrer_member_id: 66,
          promotion_link_id: 33,
          bound_at: '2026-05-24T00:00:00Z',
          bound_source: 'link',
          bound_by: 'system',
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listDistributionAttributionAudits.mockResolvedValue({
      items: [
        {
          id: 1,
          user_id: 7,
          user_email: 'user@example.com',
          username: 'example-user',
          previous_channel_org_id: 66,
          previous_referrer_member_id: 67,
          previous_promotion_link_id: 68,
          previous_bound_source: 'registration',
          previous_bound_by: 'system',
          new_channel_org_id: 88,
          new_referrer_member_id: 77,
          new_promotion_link_id: 44,
          new_bound_source: 'manual',
          new_bound_by: 'admin',
          note: 'manual attribution fix',
          operator_user_id: 9,
          operator_user_email: 'admin@example.com',
          operator_username: 'admin-user',
          created_at: '2026-05-24T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listDistributionWalletTransactions.mockResolvedValue({
      items: [
        {
          id: 9001,
          channel_org_id: 88,
          organization_name: 'Independent Agent',
          organization_type: 'reseller',
          transaction_type: 'refund',
          amount: 80,
          prepaid_balance_before: 120,
          prepaid_balance_after: 40,
          commission_reserved_before: 0,
          commission_reserved_after: 0,
          reference_no: 'RF-1',
          note: 'manual refund',
          operator_user_id: 9,
          created_at: '2026-05-24T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    listDistributionWalletRequests.mockResolvedValue({
      items: [
        {
          id: 31,
          channel_org_id: 88,
          organization_name: 'Independent Agent',
          organization_type: 'reseller',
          request_type: 'recharge',
          amount: 120,
          reference_no: 'BANK-1',
          note: 'bank transfer',
          status: 'pending',
          created_by_user_id: 7,
          created_by_user_email: 'manager@example.com',
          created_by_username: 'manager',
          reviewed_by_user_id: null,
          reviewed_by_user_email: '',
          reviewed_by_username: '',
          review_note: '',
          reviewed_at: null,
          created_at: '2026-05-24T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    refundDistributionWallet.mockResolvedValue({
      wallet: {
        channel_org_id: 88,
        organization_name: 'Independent Agent',
        organization_type: 'reseller',
        prepaid_balance: 40,
        commission_reserved: 0,
        total_recharged: 200,
        total_consumed: 80,
        warning_threshold: 50,
        status: 'active',
        created_at: '2026-05-24T00:00:00Z',
        updated_at: '2026-05-24T00:00:00Z',
      },
      refund_amount: 80,
      fee_rate: 0.1,
      fee_amount: 8,
      net_amount: 72,
      reference_no: 'RF-1',
      note: 'manual refund',
      processed_mock: true,
    })
    reviewDistributionWalletRequest.mockResolvedValue({
      id: 31,
      channel_org_id: 88,
      organization_name: 'Independent Agent',
      organization_type: 'reseller',
      request_type: 'recharge',
      amount: 120,
      reference_no: 'BANK-1',
      note: 'bank transfer',
      status: 'approved',
      created_by_user_id: 7,
      created_by_user_email: 'manager@example.com',
      created_by_username: 'manager',
      reviewed_by_user_id: 9,
      reviewed_by_user_email: 'admin@example.com',
      reviewed_by_username: 'admin',
      review_note: '到账确认',
      reviewed_at: '2026-05-24T01:00:00Z',
      created_at: '2026-05-24T00:00:00Z',
    })
    updateDistributionAttribution.mockResolvedValue({
      user_id: 7,
      channel_org_id: 99,
      referrer_member_id: 77,
      promotion_link_id: 44,
      bound_at: '2026-05-24T00:00:00Z',
      bound_source: 'manual',
      bound_by: 'admin',
      audit_id: 1,
      created_at: '2026-05-24T00:00:00Z',
      updated_at: '2026-05-24T00:00:00Z',
    })
    listUsers.mockResolvedValue({
      items: [
        {
          id: 42,
          username: 'alice',
          email: 'alice@example.com',
        },
      ],
      total: 1,
      page: 1,
      page_size: 10,
      pages: 1,
    })
    getUserById.mockResolvedValue({
      id: 42,
      username: 'alice',
      email: 'alice@example.com',
    })
  })

  it('filters members by selecting a searched user', async () => {
    vi.useFakeTimers()
    mockRoute.path = '/admin/distribution/members'
    mockRoute.params = { tab: 'members' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const filterSearch = wrapper.find('input[placeholder="admin.usage.searchUserPlaceholder"]')
    await filterSearch.trigger('focus')
    await filterSearch.setValue('alice')
    vi.advanceTimersByTime(300)
    await flushPromises()

    const option = wrapper.findAll('button').find((button) => button.text().includes('alice@example.com'))
    expect(option).toBeDefined()
    await option!.trigger('click')
    await flushPromises()

    expect(listDistributionMembers).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      channel_org_id: undefined,
      user_id: 42,
      role_type: undefined,
      alert_type: undefined,
      severity: undefined,
      request_type: undefined,
      status: undefined,
      transaction_type: undefined,
    })
    vi.useRealTimers()
  })

  it('submits member creation with the selected user instead of a raw id input', async () => {
    vi.useFakeTimers()
    mockRoute.path = '/admin/distribution/members'
    mockRoute.params = { tab: 'members' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const createButton = wrapper.findAll('button').find((button) => button.text() === 'admin.distribution.actions.createMember')
    expect(createButton).toBeDefined()
    await createButton!.trigger('click')
    await flushPromises()

    const forms = wrapper.findAll('form')
    const memberForm = forms[0]
    const numberInputs = memberForm.findAll('input[type="number"]')
    await numberInputs[0].setValue('88')
    await numberInputs[numberInputs.length - 1].setValue('0.2')

    const memberUserSearch = memberForm.find('input[placeholder="admin.usage.searchUserPlaceholder"]')
    await memberUserSearch.trigger('focus')
    await memberUserSearch.setValue('alice')
    vi.advanceTimersByTime(300)
    await flushPromises()

    const option = wrapper.findAll('button').find((button) => button.text().includes('alice@example.com'))
    expect(option).toBeDefined()
    await option!.trigger('click')
    await flushPromises()

    await memberForm.trigger('submit.prevent')
    await flushPromises()

    expect(createDistributionMember).toHaveBeenCalledWith({
      channel_org_id: 88,
      user_id: 42,
      role_type: 'agent',
      parent_member_id: undefined,
      level_code: undefined,
      commission_rate: 0.2,
      status: 'active',
    })
    vi.useRealTimers()
  })

  it('opens wallet refund dialog and submits mock refund request', async () => {
    mockRoute.path = '/admin/distribution/wallets'
    mockRoute.params = { tab: 'wallets' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const refundButton = wrapper.findAll('button').find((button) => button.text() === 'admin.distribution.actions.refundWallet')
    expect(refundButton).toBeDefined()

    await refundButton!.trigger('click')
    await flushPromises()

    const amountInput = wrapper.find('input[type="number"][min="0.0001"]')
    const textInputs = wrapper.findAll('input:not([type="number"])')
    const noteInput = wrapper.find('textarea')

    await amountInput.setValue('80')
    await textInputs[0].setValue('RF-1')
    await noteInput.setValue('manual refund')

    const forms = wrapper.findAll('form')
    await forms[forms.length - 1].trigger('submit.prevent')
    await flushPromises()

    expect(refundDistributionWallet).toHaveBeenCalledWith(88, {
      amount: 80,
      reference_no: 'RF-1',
      note: 'manual refund',
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.distribution.messages.walletRefunded')
  })

  it('loads wallet transactions from the route segment and does not render top tabs', async () => {
    mockRoute.path = '/admin/distribution/wallet-transactions'
    mockRoute.params = { tab: 'wallet-transactions' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(listDistributionWalletTransactions).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      channel_org_id: undefined,
      role_type: undefined,
      transaction_type: undefined,
    })
    expect(wrapper.text()).not.toContain('admin.distribution.tabs.walletTransactions')
    expect(wrapper.text()).not.toContain('admin.distribution.stats.organizations')
    expect(wrapper.text()).not.toContain('admin.distribution.stats.commissionUpperRatio')
  })

  it('reviews a pending wallet request from the wallet requests tab', async () => {
    mockRoute.path = '/admin/distribution/wallet-requests'
    mockRoute.params = { tab: 'wallet-requests' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const approveButton = wrapper.findAll('button').find((button) => button.text() === 'admin.distribution.actions.approveWalletRequest')
    expect(approveButton).toBeDefined()

    await approveButton!.trigger('click')
    await flushPromises()

    expect(reviewDistributionWalletRequest).toHaveBeenCalledWith(31, {
      action: 'approve',
      review_note: '',
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.distribution.messages.walletRequestReviewed')
  })

  it('opens attribution dialog and submits manual attribution adjustment', async () => {
    vi.useFakeTimers()
    mockRoute.path = '/admin/distribution/attributions'
    mockRoute.params = { tab: 'attributions' }
    listDistributionMembers.mockResolvedValueOnce({
      items: [
        {
          member_id: 77,
          user_id: 12,
          user_email: 'ref@example.com',
          username: 'referrer',
          channel_org_id: 99,
          role_type: 'agent',
          parent_member_id: null,
          level_code: '',
          commission_rate: 0.2,
          status: 'active',
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const adjustButton = wrapper.findAll('button').find((button) => button.text() === 'admin.distribution.actions.adjustAttribution')
    expect(adjustButton).toBeDefined()

    await adjustButton!.trigger('click')
    await flushPromises()

    const forms = wrapper.findAll('form')
    const attributionForm = forms[forms.length - 1]
    const numberInputs = attributionForm.findAll('input[type="number"]')
    const referrerSearch = attributionForm.find('input[placeholder="admin.usage.searchUserPlaceholder"]')
    const noteInput = attributionForm.find('textarea')

    await numberInputs[0].setValue('99')
    await referrerSearch.trigger('focus')
    await referrerSearch.setValue('referrer')
    vi.advanceTimersByTime(300)
    await flushPromises()

    const option = wrapper.findAll('button').find((button) => button.text().includes('ref@example.com'))
    expect(option).toBeDefined()
    await option!.trigger('click')
    await flushPromises()

    await numberInputs[1].setValue('44')
    await noteInput.setValue('manual attribution fix')

    await attributionForm.trigger('submit.prevent')
    await flushPromises()

    expect(updateDistributionAttribution).toHaveBeenCalledWith(7, {
      channel_org_id: 99,
      referrer_member_id: 77,
      promotion_link_id: 44,
      note: 'manual attribution fix',
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.distribution.messages.attributionUpdated')
    vi.useRealTimers()
  })

  it('loads attribution audits for the selected user', async () => {
    mockRoute.path = '/admin/distribution/attributions'
    mockRoute.params = { tab: 'attributions' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const auditButton = wrapper.findAll('button').find((button) => button.text() === 'admin.distribution.actions.viewAttributionAudits')
    expect(auditButton).toBeDefined()

    await auditButton!.trigger('click')
    await flushPromises()

    expect(listDistributionAttributionAudits).toHaveBeenCalledWith({
      user_id: 7,
      page: 1,
      page_size: 20,
    })
  })

  it('loads distribution alert events from the route segment', async () => {
    mockRoute.path = '/admin/distribution/alert-events'
    mockRoute.params = { tab: 'alert-events' }

    const wrapper = mount(AdminDistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          DataTable: DataTableStub,
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(listDistributionAlertEvents).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      channel_org_id: undefined,
      alert_type: undefined,
      severity: undefined,
      status: undefined,
      request_type: undefined,
      role_type: undefined,
      transaction_type: undefined,
    })
  })
})
