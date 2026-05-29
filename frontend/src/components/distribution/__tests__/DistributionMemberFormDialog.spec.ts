import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import DistributionMemberFormDialog from '../DistributionMemberFormDialog.vue'
import type { DistributionOrganization } from '@/api/admin/distribution'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const userViewSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../../../views/user/DistributionView.vue'),
  'utf8',
)
const adminViewSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../../../views/admin/distribution/AdminDistributionView.vue'),
  'utf8',
)

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<div v-if="show"><slot /></div>',
}

function buildProps(overrides: Record<string, unknown> = {}) {
  return {
    show: true,
    title: 'Create Member',
    saving: false,
    namespace: 'distribution',
    userSearchPlaceholderKey: 'admin.usage.searchUserPlaceholder',
    parentSearchPlaceholderKey: 'distribution.fields.parentMemberIdPlaceholder',
    roleFieldKey: 'fields.roleType',
    levelCodeDescriptionKey: undefined,
    showChannelOrgField: false,
    disableParentLookup: false,
    memberForm: {
      channel_org_id: 0,
      user_id: 0,
      role_type: 'kol1',
      parent_member_id: null,
      level_code: '',
      commission_rate: 0,
      status: 'active',
    },
    roleOptions: ['agent', 'kol1', 'kol2'],
    memberUserLookup: {
      keyword: '',
      loading: false,
      open: false,
      results: [],
      selected: null,
    },
    parentMemberLookup: {
      keyword: '',
      loading: false,
      open: false,
      results: [],
      selected: null,
    },
    ...overrides,
  }
}

describe('DistributionMemberFormDialog', () => {
  it('renders optional channel organization field only for admin mode', () => {
    const userWrapper = mount(DistributionMemberFormDialog, {
      props: buildProps(),
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    expect(userWrapper.text()).not.toContain('distribution.fields.channelOrgId')

    const adminWrapper = mount(DistributionMemberFormDialog, {
      props: buildProps({
        namespace: 'admin.distribution',
        roleFieldKey: 'fields.role',
        parentSearchPlaceholderKey: 'admin.usage.searchUserPlaceholder',
        levelCodeDescriptionKey: 'admin.distribution.fields.levelCodeDesc',
        showChannelOrgField: true,
        channelOrgLookup: {
          keyword: 'Demo Org · #1 · reseller',
          loading: false,
          open: false,
          results: [],
          selected: {
            id: 1,
            name: 'Demo Org',
            type: 'reseller',
            status: 'active',
            config: {},
            brand_config: {},
            created_at: '',
            updated_at: '',
          } satisfies DistributionOrganization,
        },
        memberForm: {
          channel_org_id: 1,
          user_id: 0,
          role_type: 'agent',
          parent_member_id: null,
          level_code: '',
          commission_rate: 0,
          status: 'active',
        },
      }),
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    expect(adminWrapper.text()).toContain('admin.distribution.fields.channelOrgId')
    expect(adminWrapper.find('input[placeholder="admin.distribution.fields.channelOrgIdPlaceholder"]').exists()).toBe(true)
    expect(adminWrapper.text()).toContain('admin.distribution.fields.levelCodeDesc')
  })

  it('hides level field for KOL roles', () => {
    const wrapper = mount(DistributionMemberFormDialog, {
      props: buildProps({
        memberForm: {
          channel_org_id: 0,
          user_id: 0,
          role_type: 'kol1',
          parent_member_id: 1,
          level_code: 'GOLD',
          commission_rate: 0.1,
          status: 'active',
        },
        levelOptions: [
          {
            code: 'GOLD',
            name: 'Gold',
            commission_rate: 12,
            source: 'channel',
            label: 'Gold',
          },
        ],
      }),
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).not.toContain('distribution.fields.levelCode')
  })

  it('renders level select when options are provided for agent role', () => {
    const wrapper = mount(DistributionMemberFormDialog, {
      props: buildProps({
        memberForm: {
          channel_org_id: 1,
          user_id: 0,
          role_type: 'agent',
          parent_member_id: null,
          level_code: '',
          commission_rate: 0,
          status: 'active',
        },
        levelOptions: [
          {
            code: 'GOLD',
            name: 'Gold',
            commission_rate: 12,
            source: 'channel',
            label: 'Gold (GOLD) · 12% · channel',
          },
        ],
      }),
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true,
        },
      },
    })

    const levelSelect = wrapper.findAll('select').find((node) =>
      node.findAll('option').some((option) => option.attributes('value') === 'GOLD'),
    )
    expect(levelSelect).toBeTruthy()
  })

  it('is used by both admin and user distribution views', () => {
    expect(userViewSource).toContain('DistributionMemberFormDialog')
    expect(adminViewSource).toContain('DistributionMemberFormDialog')
  })
})
