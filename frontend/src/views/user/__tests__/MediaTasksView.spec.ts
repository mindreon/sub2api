import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import MediaTasksView from '../MediaTasksView.vue'

const { listTasks, showError } = vi.hoisted(() => ({
  listTasks: vi.fn(),
  showError: vi.fn(),
}))

const messages: Record<string, string> = {
  'common.refresh': '刷新',
  'common.loading': '加载中',
  'mediaTasks.title': '媒体任务',
  'mediaTasks.description': '查看异步生成任务',
  'mediaTasks.empty': '暂无媒体生成任务',
  'mediaTasks.loadFailed': '加载媒体任务失败',
  'mediaTasks.openResult': '打开',
  'mediaTasks.pendingSettlement': '待结算',
  'mediaTasks.total': '共 {count} 条',
  'mediaTasks.filters.allStatus': '全部状态',
  'mediaTasks.filters.allTypes': '全部类型',
  'mediaTasks.filters.allTime': '全部时间',
  'mediaTasks.statuses.pending': '等待中',
  'mediaTasks.statuses.in_progress': '处理中',
  'mediaTasks.statuses.completed': '已完成',
  'mediaTasks.statuses.failed': '失败',
  'mediaTasks.statuses.expired': '已过期',
  'mediaTasks.types.video': '视频',
  'mediaTasks.types.image': '图片',
  'mediaTasks.types.audio': '音频',
  'mediaTasks.columns.task': '任务',
  'mediaTasks.columns.type': '类型',
  'mediaTasks.columns.status': '状态',
  'mediaTasks.columns.cost': '费用',
  'mediaTasks.columns.result': '结果',
  'mediaTasks.columns.createdAt': '创建时间',
  'pagination.of': '共',
  'pagination.results': '条',
  'pagination.previous': '上一页',
  'pagination.next': '下一页',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        const value = messages[key] ?? key
        return params ? value.replace('{count}', String(params.count)) : value
      },
    }),
  }
})

vi.mock('@/api/media', () => ({
  mediaAPI: { listTasks },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

const tablePageLayoutStub = {
  template: `
    <section>
      <div data-testid="layout-actions"><slot name="actions" /></div>
      <div data-testid="layout-filters"><slot name="filters" /></div>
      <div data-testid="layout-table"><slot name="table" /></div>
      <div data-testid="layout-pagination"><slot name="pagination" /></div>
    </section>
  `,
}

const selectStub = {
  props: ['modelValue', 'options'],
  template: '<button class="select-stub">{{ options?.[0]?.label }}</button>',
}

const dateRangePickerStub = {
  props: ['startDate', 'endDate'],
  emits: ['update:startDate', 'update:endDate', 'change'],
  template: `
    <button
      class="date-range-stub"
      @click="$emit('change', { startDate: '2026-07-01', endDate: '2026-07-02', preset: null })"
    >
      date range
    </button>
  `,
}

function formatLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function localDayBounds(startDate: string, endDate: string) {
  return {
    from: new Date(`${startDate}T00:00:00`).toISOString(),
    to: new Date(`${endDate}T23:59:59.999`).toISOString(),
  }
}

describe('MediaTasksView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 6, 11, 12, 0, 0))
    listTasks.mockReset()
    showError.mockReset()
    listTasks.mockResolvedValue({
      items: [
        {
          task_id: 'de69106d-f5e7-4870-bc25-2d921b537c73',
          model: 'seedance2.0-fast-p5',
          media_type: 'video',
          status: 'completed',
          reserved_cost: 0.24883166,
          actual_cost: 0.26230484,
          billing_currency: 'USD',
          result_url: 'https://example.com/result.mp4',
          expires_at: '2026-07-11T12:04:11+08:00',
          created_at: '2026-07-11T11:34:11+08:00',
          updated_at: '2026-07-11T11:47:24+08:00',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  function mountView() {
    return mount(MediaTasksView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: tablePageLayoutStub,
          DateRangePicker: dateRangePickerStub,
          Icon: true,
          Select: selectStub,
        },
      },
    })
  }

  it('loads the last 30 local days and renders localized task rows in the table slot', async () => {
    const wrapper = mountView()

    await flushPromises()
    await nextTick()

    const start = new Date(2026, 6, 11)
    start.setDate(start.getDate() - 29)
    const bounds = localDayBounds(formatLocalDate(start), '2026-07-11')
    expect(listTasks).toHaveBeenCalledWith(expect.objectContaining({
      page: 1,
      page_size: 20,
      created_from: bounds.from,
      created_to: bounds.to,
    }))
    expect(wrapper.find('[data-testid="media-task-table"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="media-task-desktop"]').classes()).toEqual(
      expect.arrayContaining(['hidden', 'lg:block'])
    )
    expect(wrapper.get('[data-testid="layout-table"]').text()).toContain('seedance2.0-fast-p5')
    expect(wrapper.text()).toContain('已完成')
    expect(wrapper.text()).toContain('视频')
  })

  it('keeps refresh and all filters in one toolbar', async () => {
    const wrapper = mountView()
    await flushPromises()

    const toolbar = wrapper.get('[data-testid="media-task-toolbar"]')
    expect(toolbar.findAll('.select-stub')).toHaveLength(2)
    expect(toolbar.find('.date-range-stub').exists()).toBe(true)
    expect(toolbar.find('[data-testid="media-task-refresh"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="layout-actions"]').text()).toBe('')
  })

  it('reloads page one with inclusive custom date bounds', async () => {
    const wrapper = mountView()
    await flushPromises()
    listTasks.mockClear()

    await wrapper.get('.date-range-stub').trigger('click')
    await flushPromises()

    const bounds = localDayBounds('2026-07-01', '2026-07-02')
    expect(listTasks).toHaveBeenCalledWith(expect.objectContaining({
      page: 1,
      created_from: bounds.from,
      created_to: bounds.to,
    }))
  })

  it('can clear date bounds for all-time results', async () => {
    const wrapper = mountView()
    await flushPromises()
    listTasks.mockClear()

    await wrapper.get('[data-testid="media-task-all-time"]').trigger('click')
    await flushPromises()

    expect(listTasks).toHaveBeenCalledWith(expect.objectContaining({
      page: 1,
      created_from: undefined,
      created_to: undefined,
    }))
  })
})
