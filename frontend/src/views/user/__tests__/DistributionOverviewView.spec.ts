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

import DistributionOverviewView from '../DistributionOverviewView.vue'

const {
  getDistributionOverview,
  getMyDistributionAnalytics,
  updateMyDistributionOrganization,
} = vi.hoisted(() => ({
  getDistributionOverview: vi.fn(),
  getMyDistributionAnalytics: vi.fn(),
  updateMyDistributionOrganization: vi.fn(),
}))

const showError = vi.fn()
const showSuccess = vi.fn()
const routerReplace = vi.fn()
const routerPush = vi.fn()

vi.mock('@/api/distribution', () => ({
  getDistributionOverview,
  getMyDistributionAnalytics,
  updateMyDistributionOrganization,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
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
  useRouter: () => ({
    replace: routerReplace,
    push: routerPush,
  }),
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<div v-if="show"><slot /></div>',
}

describe('DistributionOverviewView', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    routerReplace.mockReset()
    routerPush.mockReset()
    getDistributionOverview.mockReset()
    getMyDistributionAnalytics.mockReset()
    updateMyDistributionOrganization.mockReset()

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
        promotion_link_count: 0,
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
  })

  it('shows channel warning banners and channel settings entry', async () => {
    const wrapper = mount(DistributionOverviewView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
          BaseDialog: BaseDialogStub,
          Icon: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('distribution.actions.manageChannel')
    expect(wrapper.text()).toContain('distribution.warnings.lowBalanceTitle')
    expect(wrapper.text()).toContain('distribution.warnings.consumptionTitle')
  })

  it('redirects non-managers back to distribution center', async () => {
    getDistributionOverview.mockResolvedValueOnce({
      user_id: 1,
      channel_org_id: 88,
      can_manage_channel: false,
      summary: {
        organization: null,
        wallet: null,
      },
    })

    mount(DistributionOverviewView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
          BaseDialog: BaseDialogStub,
          Icon: true,
          DateRangePicker: true,
          DistributionAnalyticsTrendChart: true,
        },
      },
    })

    await flushPromises()

    expect(routerReplace).toHaveBeenCalledWith('/distribution')
  })
})
