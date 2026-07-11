<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="space-y-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
                {{ t('mediaTasks.title') }}
              </h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ t('mediaTasks.description') }}
              </p>
            </div>
            <p class="pt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('mediaTasks.total', { count: total }) }}
            </p>
          </div>

          <div
            data-testid="media-task-toolbar"
            class="flex flex-wrap items-center gap-3"
          >
            <Select
              v-model="filterStatus"
              class="min-w-36 flex-1 sm:w-40 sm:flex-none"
              :options="statusOptions"
              @update:model-value="onFilterChange"
            />
            <Select
              v-model="filterMediaType"
              class="min-w-36 flex-1 sm:w-40 sm:flex-none"
              :options="mediaTypeOptions"
              @update:model-value="onFilterChange"
            />
            <DateRangePicker
              v-model:start-date="startDate"
              v-model:end-date="endDate"
              class="media-task-date-picker min-w-0"
              @change="onDateRangeChange"
            />
            <button
              data-testid="media-task-all-time"
              type="button"
              class="btn h-11 shrink-0"
              :class="allTime ? 'btn-primary' : 'btn-secondary'"
              @click="selectAllTime"
            >
              {{ t('mediaTasks.filters.allTime') }}
            </button>
            <button
              data-testid="media-task-refresh"
              class="btn btn-secondary ml-auto h-11 w-11 shrink-0 p-0"
              :disabled="loading"
              :title="t('common.refresh')"
              :aria-label="t('common.refresh')"
              @click="loadTasks"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div class="table-wrapper">
          <div data-testid="media-task-desktop" class="hidden lg:block">
            <table
              data-testid="media-task-table"
              class="min-w-full divide-y divide-gray-200 dark:divide-dark-700"
            >
            <thead>
              <tr>
                <th class="table-th">{{ t('mediaTasks.columns.task') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.type') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.status') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.cost') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.result') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.createdAt') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-if="loading">
                <td colspan="6" class="table-td py-12 text-center text-gray-500">
                  {{ t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="tasks.length === 0">
                <td colspan="6" class="table-td py-12 text-center text-gray-500">
                  {{ t('mediaTasks.empty') }}
                </td>
              </tr>
              <tr
                v-for="task in tasks"
                v-else
                :key="task.task_id"
                class="transition-colors hover:bg-gray-50/80 dark:hover:bg-dark-800/70"
              >
                <td class="table-td min-w-56">
                  <p class="font-medium text-gray-900 dark:text-white">{{ task.model }}</p>
                  <p class="mt-1 font-mono text-xs text-gray-400" :title="task.task_id">
                    {{ shortTaskId(task.task_id) }}
                  </p>
                </td>
                <td class="table-td whitespace-nowrap">{{ mediaTypeLabel(task.media_type) }}</td>
                <td class="table-td whitespace-nowrap">
                  <span class="badge" :class="statusClass(task.status)">
                    {{ statusLabel(task.status) }}
                  </span>
                </td>
                <td class="table-td min-w-32 whitespace-nowrap">
                  <p class="font-medium text-gray-900 dark:text-white">
                    {{ task.actual_cost != null ? formatCost(task.actual_cost) : t('mediaTasks.pendingSettlement') }}
                  </p>
                  <p class="mt-1 text-xs text-gray-400">
                    {{ t('mediaTasks.reservedCost', { amount: formatCost(task.reserved_cost) }) }}
                  </p>
                </td>
                <td class="table-td whitespace-nowrap">
                  <a
                    v-if="task.result_url"
                    :href="task.result_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="inline-flex min-h-11 items-center gap-1 text-primary-600 hover:text-primary-700 dark:text-primary-400"
                  >
                    <Icon name="externalLink" size="xs" />
                    {{ t('mediaTasks.openResult') }}
                  </a>
                  <span v-else class="text-gray-400">—</span>
                </td>
                <td class="table-td whitespace-nowrap">{{ formatDate(task.created_at) }}</td>
              </tr>
            </tbody>
            </table>
          </div>

          <div class="space-y-3 p-1 lg:hidden">
            <div v-if="loading" class="py-12 text-center text-sm text-gray-500">
              {{ t('common.loading') }}
            </div>
            <div v-else-if="tasks.length === 0" class="py-12 text-center text-sm text-gray-500">
              {{ t('mediaTasks.empty') }}
            </div>
            <article
              v-for="task in tasks"
              v-else
              :key="task.task_id"
              data-testid="media-task-card"
              class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <h2 class="truncate font-medium text-gray-900 dark:text-white">{{ task.model }}</h2>
                  <p class="mt-1 truncate font-mono text-xs text-gray-400" :title="task.task_id">
                    {{ task.task_id }}
                  </p>
                </div>
                <span class="badge shrink-0" :class="statusClass(task.status)">
                  {{ statusLabel(task.status) }}
                </span>
              </div>
              <dl class="mt-4 grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
                <div>
                  <dt class="text-xs text-gray-400">{{ t('mediaTasks.columns.type') }}</dt>
                  <dd class="mt-1 text-gray-700 dark:text-gray-200">{{ mediaTypeLabel(task.media_type) }}</dd>
                </div>
                <div>
                  <dt class="text-xs text-gray-400">{{ t('mediaTasks.columns.cost') }}</dt>
                  <dd class="mt-1 text-gray-700 dark:text-gray-200">
                    {{ task.actual_cost != null ? formatCost(task.actual_cost) : t('mediaTasks.pendingSettlement') }}
                  </dd>
                </div>
                <div class="col-span-2">
                  <dt class="text-xs text-gray-400">{{ t('mediaTasks.columns.createdAt') }}</dt>
                  <dd class="mt-1 text-gray-700 dark:text-gray-200">{{ formatDate(task.created_at) }}</dd>
                </div>
              </dl>
              <a
                v-if="task.result_url"
                :href="task.result_url"
                target="_blank"
                rel="noopener noreferrer"
                class="mt-4 inline-flex min-h-11 items-center gap-1 text-sm text-primary-600 dark:text-primary-400"
              >
                <Icon name="externalLink" size="xs" />
                {{ t('mediaTasks.openResult') }}
              </a>
            </article>
          </div>
        </div>
      </template>

      <template v-if="total > pageSize" #pagination>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <p class="text-sm text-gray-500">
            {{ t('pagination.of') }} {{ total }} {{ t('pagination.results') }}
          </p>
          <div class="flex gap-2">
            <button class="btn btn-secondary btn-sm" :disabled="page <= 1 || loading" @click="goPage(page - 1)">
              {{ t('pagination.previous') }}
            </button>
            <button class="btn btn-secondary btn-sm" :disabled="page >= pages || loading" @click="goPage(page + 1)">
              {{ t('pagination.next') }}
            </button>
          </div>
        </div>
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { mediaAPI, type MediaTask } from '@/api/media'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

function formatLocalDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function defaultDateRange() {
  const end = new Date()
  const start = new Date(end.getFullYear(), end.getMonth(), end.getDate())
  start.setDate(start.getDate() - 29)
  return { start: formatLocalDate(start), end: formatLocalDate(end) }
}

function localDateBounds(start: string, end: string) {
  return {
    from: new Date(`${start}T00:00:00`).toISOString(),
    to: new Date(`${end}T23:59:59.999`).toISOString()
  }
}

const initialRange = defaultDateRange()
const loading = ref(false)
const tasks = ref<MediaTask[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const pages = ref(1)
const filterStatus = ref('')
const filterMediaType = ref('')
const startDate = ref(initialRange.start)
const endDate = ref(initialRange.end)
const allTime = ref(false)

const statusOptions = computed(() => [
  { value: '', label: t('mediaTasks.filters.allStatus') },
  ...['pending', 'in_progress', 'completed', 'failed', 'expired'].map((value) => ({
    value,
    label: statusLabel(value)
  }))
])

const mediaTypeOptions = computed(() => [
  { value: '', label: t('mediaTasks.filters.allTypes') },
  ...['video', 'image', 'audio'].map((value) => ({ value, label: mediaTypeLabel(value) }))
])

function translatedEnum(prefix: string, value: string) {
  const key = `mediaTasks.${prefix}.${value}`
  const translated = t(key)
  return translated === key ? value : translated
}

function statusLabel(status: string) {
  return translatedEnum('statuses', status)
}

function mediaTypeLabel(mediaType: string) {
  return translatedEnum('types', mediaType)
}

function statusClass(status: string) {
  if (status === 'completed') return 'badge-success'
  if (status === 'failed' || status === 'expired') return 'badge-danger'
  if (status === 'in_progress' || status === 'pending') return 'badge-warning'
  return 'badge-secondary'
}

function shortTaskId(taskId: string) {
  if (taskId.length <= 20) return taskId
  return `${taskId.slice(0, 10)}…${taskId.slice(-6)}`
}

function formatCost(value: number) {
  return `$${value.toFixed(4)}`
}

function formatDate(value: string) {
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString()
}

async function loadTasks() {
  loading.value = true
  const bounds = allTime.value ? null : localDateBounds(startDate.value, endDate.value)
  try {
    const res = await mediaAPI.listTasks({
      page: page.value,
      page_size: pageSize.value,
      status: filterStatus.value || undefined,
      media_type: filterMediaType.value || undefined,
      created_from: bounds?.from,
      created_to: bounds?.to,
      sort_order: 'desc'
    })
    tasks.value = res.items
    total.value = res.total
    pages.value = res.pages
    page.value = res.page
  } catch {
    appStore.showError(t('mediaTasks.loadFailed'))
  } finally {
    loading.value = false
  }
}

function onFilterChange() {
  page.value = 1
  loadTasks()
}

function onDateRangeChange(range: { startDate: string; endDate: string }) {
  startDate.value = range.startDate
  endDate.value = range.endDate
  allTime.value = false
  page.value = 1
  loadTasks()
}

function selectAllTime() {
  allTime.value = true
  page.value = 1
  loadTasks()
}

function goPage(next: number) {
  page.value = next
  loadTasks()
}

onMounted(loadTasks)
</script>

<style scoped>
.media-task-date-picker :deep(.date-picker-trigger) {
  min-height: 44px;
}
</style>
