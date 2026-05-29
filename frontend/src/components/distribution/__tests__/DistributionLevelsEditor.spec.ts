import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import DistributionLevelsEditor from '../DistributionLevelsEditor.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('vue-draggable-plus', () => ({
  VueDraggable: {
    name: 'VueDraggable',
    props: ['modelValue'],
    template: '<div class="draggable"><slot /></div>',
  },
}))

describe('DistributionLevelsEditor', () => {
  it('renders empty state and adds a row', async () => {
    const wrapper = mount(DistributionLevelsEditor, {
      props: {
        modelValue: [],
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('distributionLevels.empty')

    await wrapper.get('button').trigger('click')

    const updates = wrapper.emitted('update:modelValue')
    expect(updates?.length).toBeGreaterThan(0)
    const last = updates?.[updates.length - 1]?.[0] as Array<Record<string, unknown>>
    expect(last).toHaveLength(1)
    expect(last[0]).toMatchObject({ active: true, sort_order: 0 })
  })

  it('exposes validate()', async () => {
    const wrapper = mount(DistributionLevelsEditor, {
      props: {
        modelValue: [{ code: '', name: '', commission_rate: 0, active: true, sort_order: 0, note: '' }],
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    const exposed = wrapper.vm as { validate: () => boolean }
    expect(exposed.validate()).toBe(false)
  })
})
