import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'

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

import DistributionView from '../DistributionView.vue'

const {
  createMyDistributionMember,
  createMyDistributionPromotionLink,
  getDistributionOverview,
  getMyDistributionAnalytics,
  listMyDistributionAttributions,
  listMyDistributionAlertEvents,
  listMyDistributionCommissions,
  listMyDistributionMembers,
  listMyDistributionPromotionLinks,
  listMyDistributionWalletRequests,
  listMyDistributionWalletTransactions,
  listMyDistributionWholesalePricing,
  settleMyDistributionCommission,
  submitMyDistributionWalletRequest,
  updateMyDistributionOrganization,
} = vi.hoisted(() => ({
  createMyDistributionMember: vi.fn(),
  createMyDistributionPromotionLink: vi.fn(),
  getDistributionOverview: vi.fn(),
  getMyDistributionAnalytics: vi.fn(),
  listMyDistributionAttributions: vi.fn(),
  listMyDistributionAlertEvents: vi.fn(),
  listMyDistributionCommissions: vi.fn(),
  listMyDistributionMembers: vi.fn(),
  listMyDistributionPromotionLinks: vi.fn(),
  listMyDistributionWalletRequests: vi.fn(),
  listMyDistributionWalletTransactions: vi.fn(),
  listMyDistributionWholesalePricing: vi.fn(),
  settleMyDistributionCommission: vi.fn(),
  submitMyDistributionWalletRequest: vi.fn(),
  updateMyDistributionOrganization: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()
const listAdminUsers = vi.hoisted(() => vi.fn())
const routeState = reactive({
  path: '/distribution',
  hash: '',
})

vi.mock('@/api/distribution', () => ({
  createMyDistributionMember,
  createMyDistributionPromotionLink,
  getDistributionOverview,
  getMyDistributionAnalytics,
  listMyDistributionAttributions,
  listMyDistributionAlertEvents,
  listMyDistributionCommissions,
  listMyDistributionMembers,
  listMyDistributionPromotionLinks,
  listMyDistributionWalletRequests,
  listMyDistributionWalletTransactions,
  listMyDistributionWholesalePricing,
  settleMyDistributionCommission,
  submitMyDistributionWalletRequest,
  updateMyDistributionOrganization,
  distributionAPI: {
    createMyDistributionMember,
    createMyDistributionPromotionLink,
    getDistributionOverview,
    getMyDistributionAnalytics,
    listMyDistributionAttributions,
    listMyDistributionAlertEvents,
    listMyDistributionCommissions,
    listMyDistributionMembers,
    listMyDistributionPromotionLinks,
    listMyDistributionWalletRequests,
    listMyDistributionWalletTransactions,
    listMyDistributionWholesalePricing,
    settleMyDistributionCommission,
    submitMyDistributionWalletRequest,
    updateMyDistributionOrganization,
  },
  default: {
    createMyDistributionMember,
    createMyDistributionPromotionLink,
    getDistributionOverview,
    getMyDistributionAnalytics,
    listMyDistributionAttributions,
    listMyDistributionAlertEvents,
    listMyDistributionCommissions,
    listMyDistributionMembers,
    listMyDistributionPromotionLinks,
    listMyDistributionWalletRequests,
    listMyDistributionWalletTransactions,
    listMyDistributionWholesalePricing,
    settleMyDistributionCommission,
    submitMyDistributionWalletRequest,
    updateMyDistributionOrganization,
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
    list: listAdminUsers,
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

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({
    push: vi.fn((to: any) => {
      if (typeof to === 'object' && to && 'hash' in to) {
        routeState.hash = typeof to.hash === 'string' ? to.hash : ''
      }
    }),
  }),
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<div v-if="show"><slot /></div>',
}

describe('DistributionView', () => {
  beforeEach(() => {
    routeState.path = '/distribution'
    routeState.hash = ''
    showError.mockReset()
    showSuccess.mockReset()
    createMyDistributionMember.mockReset()
    createMyDistributionPromotionLink.mockReset()
    getDistributionOverview.mockReset()
    getMyDistributionAnalytics.mockReset()
    listMyDistributionAttributions.mockReset()
    listMyDistributionAlertEvents.mockReset()
    listMyDistributionCommissions.mockReset()
    listMyDistributionMembers.mockReset()
    listMyDistributionPromotionLinks.mockReset()
    listMyDistributionWalletRequests.mockReset()
    listMyDistributionWalletTransactions.mockReset()
    listMyDistributionWholesalePricing.mockReset()
    settleMyDistributionCommission.mockReset()
    submitMyDistributionWalletRequest.mockReset()
    updateMyDistributionOrganization.mockReset()
    listAdminUsers.mockReset().mockResolvedValue({ items: [] })

    getDistributionOverview.mockResolvedValue({
      user_id: 1,
      channel_org_id: 88,
      can_manage_channel: true,
      summary: {
        organization: {
          id: 88,
          type: 'reseller',
          name: 'Independent Agent',
          status: 'active',
          config: {
            commission_settlement_method: 'manual',
            distribution_levels: [],
            wholesale_discount_rate: 0.5,
            refund_fee_rate: 0.1,
            first_recharge_min_amount: 100,
            recharge_min_amount: 50,
            consumption_limit: 1000,
            consumption_warning_threshold: 200,
            recharge_lead_time_days: 2,
            recharge_deadline_note: 'Recharge two business days in advance.',
          },
          brand_config: {
            logo_url: 'https://example.com/logo.png',
            primary_color: '#112233',
            domain: 'reseller.example.com',
            api_domain: 'api.reseller.example.com',
          },
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
        wallet: {
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
        member_count: 0,
        agent_count: 0,
        kol1_count: 0,
        kol2_count: 0,
        promotion_link_count: 0,
        attribution_count: 0,
        commission_count: 0,
        frozen_commission_amount: 0,
        available_commission_amount: 0,
        settled_commission_amount: 0,
      },
    })
    getMyDistributionAnalytics.mockResolvedValue({
      can_manage_channel: true,
      filter: {
        start_date: '2026-05-01',
        end_date: '2026-05-24',
        granularity: 'day',
        limit: 10,
      },
      channel: null,
      personal: null,
    })
    listMyDistributionWalletRequests.mockResolvedValue({
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
          created_by_user_id: 1,
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
    listMyDistributionAlertEvents.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    submitMyDistributionWalletRequest.mockResolvedValue({
      id: 32,
      channel_org_id: 88,
      organization_name: 'Independent Agent',
      organization_type: 'reseller',
      request_type: 'recharge',
      amount: 300,
      reference_no: 'BANK-2',
      note: '',
      status: 'pending',
      created_by_user_id: 1,
      created_by_user_email: 'manager@example.com',
      created_by_username: 'manager',
      reviewed_by_user_id: null,
      reviewed_by_user_email: '',
      reviewed_by_username: '',
      review_note: '',
      reviewed_at: null,
      created_at: '2026-05-24T00:00:00Z',
    })
  })

  it('shows platform-imposed warning and recharge timing fields in channel settings', async () => {
    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    const settingsButton = wrapper.findAll('button').find((button) => button.text() === 'distribution.actions.manageChannel')
    expect(settingsButton).toBeUndefined()
    expect(wrapper.text()).not.toContain('distribution.tabs.analytics')
  })

  it('does not render a second in-page distribution menu', async () => {
    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('aside').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('distribution.actions.viewChannelOverview')
    expect(wrapper.text()).not.toContain('distribution.groups.organizationManagement')
    expect(wrapper.text()).not.toContain('distribution.groups.promotionManagement')
    expect(wrapper.text()).not.toContain('distribution.groups.commissionSettlement')
    expect(wrapper.text()).not.toContain('distribution.groups.riskAlerts')
  })

  it('uses lookup inputs instead of raw numeric fields in the member dialog', async () => {
    routeState.hash = '#members'

    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    const createButton = wrapper.findAll('button').find((button) => button.text() === 'distribution.actions.createMember')
    expect(createButton).toBeDefined()

    await createButton!.trigger('click')
    await flushPromises()

    const forms = wrapper.findAll('form')
    const memberForm = forms[forms.length - 1]

    expect(memberForm.findAll('input[type="number"]')).toHaveLength(1)
    expect(memberForm.text()).toContain('distribution.fields.userId')
    expect(memberForm.text()).toContain('distribution.columns.parentMemberId')
  })

  it('keeps warning banners out of the distribution center page', async () => {
    getDistributionOverview.mockResolvedValueOnce({
      user_id: 1,
      channel_org_id: 88,
      can_manage_channel: true,
      summary: {
        organization: {
          id: 88,
          type: 'reseller',
          name: 'Independent Agent',
          status: 'active',
          config: {
            commission_settlement_method: 'manual',
            distribution_levels: [],
            wholesale_discount_rate: 0.5,
            refund_fee_rate: 0.1,
            first_recharge_min_amount: 100,
            recharge_min_amount: 50,
            consumption_limit: 1000,
            consumption_warning_threshold: 200,
            recharge_lead_time_days: 2,
            recharge_deadline_note: 'Recharge two business days in advance.',
          },
          brand_config: {},
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
        wallet: {
          channel_org_id: 88,
          organization_name: 'Independent Agent',
          organization_type: 'reseller',
          prepaid_balance: 40,
          commission_reserved: 0,
          total_recharged: 200,
          total_consumed: 850,
          warning_threshold: 50,
          status: 'active',
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
        member_count: 0,
        agent_count: 0,
        kol1_count: 0,
        kol2_count: 0,
        promotion_link_count: 0,
        attribution_count: 0,
        commission_count: 0,
        frozen_commission_amount: 0,
        available_commission_amount: 0,
        settled_commission_amount: 0,
      },
    })

    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('distribution.warnings.lowBalanceTitle')
    expect(wrapper.text()).not.toContain('distribution.warnings.consumptionTitle')
  })

  it('keeps suspended warning banners out of the distribution center page', async () => {
    getDistributionOverview.mockResolvedValueOnce({
      user_id: 1,
      channel_org_id: 88,
      can_manage_channel: true,
      summary: {
        organization: {
          id: 88,
          type: 'reseller',
          name: 'Independent Agent',
          status: 'active',
          config: {
            commission_settlement_method: 'manual',
            consumption_limit: 1000,
            recharge_deadline_note: 'Recharge two business days in advance.',
          },
          brand_config: {},
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
        wallet: {
          channel_org_id: 88,
          organization_name: 'Independent Agent',
          organization_type: 'reseller',
          prepaid_balance: 0,
          commission_reserved: 0,
          total_recharged: 200,
          total_consumed: 100,
          warning_threshold: 50,
          status: 'inactive',
          created_at: '2026-05-24T00:00:00Z',
          updated_at: '2026-05-24T00:00:00Z',
        },
        member_count: 0,
        agent_count: 0,
        kol1_count: 0,
        kol2_count: 0,
        promotion_link_count: 0,
        attribution_count: 0,
        commission_count: 0,
        frozen_commission_amount: 0,
        available_commission_amount: 0,
        settled_commission_amount: 0,
      },
    })

    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('distribution.warnings.suspendedTitle')
  })

  it('submits a wallet recharge request from the wallet requests tab', async () => {
    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    routeState.hash = '#wallet-requests'
    await flushPromises()

    const requestButton = wrapper.findAll('button').find((button) => button.text() === 'distribution.actions.requestWalletRecharge')
    expect(requestButton).toBeDefined()

    await requestButton!.trigger('click')
    await flushPromises()

    const forms = wrapper.findAll('form')
    const requestForm = forms[forms.length - 1]
    const numberInput = requestForm.find('input[type="number"]')
    const textInputs = requestForm.findAll('input:not([type="number"])')

    await numberInput.setValue('300')
    await textInputs[0].setValue('BANK-2')

    await requestForm.trigger('submit.prevent')
    await flushPromises()

    expect(submitMyDistributionWalletRequest).toHaveBeenCalledWith({
      request_type: 'recharge',
      amount: 300,
      reference_no: 'BANK-2',
      note: undefined,
    })
    expect(showSuccess).toHaveBeenCalledWith('distribution.messages.walletRequestSubmitted')
  })

  it('loads distribution alert events when the alert tab is selected', async () => {
    const wrapper = mount(DistributionView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
          Pagination: true,
          BaseDialog: BaseDialogStub,
          Icon: true,
          DataTable: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    routeState.hash = '#alert-events'
    await flushPromises()

    expect(listMyDistributionAlertEvents).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      role_type: undefined,
      alert_type: undefined,
      severity: undefined,
      request_type: undefined,
      status: undefined,
      transaction_type: undefined,
      q: undefined,
    })
  })
})
