<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <Select
            v-model="filterStatus"
            class="w-40"
            :options="statusOptions"
            @update:model-value="onFilterChange"
          />
          <Select
            v-model="filterMediaType"
            class="w-40"
            :options="mediaTypeOptions"
            @update:model-value="onFilterChange"
          />
        </div>
      </template>

      <template #actions>
        <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="loadTasks">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </template>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="table-th">{{ t('mediaTasks.columns.taskId') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.model') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.type') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.status') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.reserved') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.actual') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.result') }}</th>
                <th class="table-th">{{ t('mediaTasks.columns.createdAt') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-if="loading">
                <td colspan="8" class="table-td text-center text-gray-500">{{ t('common.loading') }}</td>
              </tr>
              <tr v-else-if="tasks.length === 0">
                <td colspan="8" class="table-td text-center text-gray-500">{{ t('mediaTasks.empty') }}</td>
              </tr>
              <tr v-for="task in tasks" :key="task.task_id" class="hover:bg-gray-50 dark:hover:bg-dark-800">
                <td class="table-td font-mono text-xs">{{ task.task_id }}</td>
                <td class="table-td">{{ task.model }}</td>
                <td class="table-td">{{ task.media_type }}</td>
                <td class="table-td">
                  <span class="badge" :class="statusClass(task.status)">{{ task.status }}</span>
                </td>
                <td class="table-td">${{ task.reserved_cost.toFixed(4) }}</td>
                <td class="table-td">
                  {{ task.actual_cost != null ? `$${task.actual_cost.toFixed(4)}` : '—' }}
                </td>
                <td class="table-td">
                  <a
                    v-if="task.result_url"
                    :href="task.result_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="inline-flex items-center gap-1 text-primary-600 hover:text-primary-700 dark:text-primary-400"
                  >
                    <Icon name="externalLink" size="xs" />
                    {{ t('mediaTasks.openResult') }}
                  </a>
                  <span v-else>—</span>
                </td>
                <td class="table-td whitespace-nowrap">{{ formatDate(task.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="total > pageSize" class="flex items-center justify-between border-t border-gray-200 px-4 py-3 dark:border-dark-700">
          <p class="text-sm text-gray-500">{{ t('pagination.of') }} {{ total }} {{ t('pagination.results') }}</p>
          <div class="flex gap-2">
            <button class="btn btn-secondary btn-sm" :disabled="page <= 1 || loading" @click="goPage(page - 1)">
              {{ t('pagination.previous') }}
            </button>
            <button class="btn btn-secondary btn-sm" :disabled="page >= pages || loading" @click="goPage(page + 1)">
              {{ t('pagination.next') }}
            </button>
          </div>
        </div>
      </div>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { mediaAPI, type MediaTask } from '@/api/media'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const tasks = ref<MediaTask[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const pages = ref(1)
const filterStatus = ref('')
const filterMediaType = ref('')

const statusOptions = computed(() => [
  { value: '', label: t('mediaTasks.filters.allStatus') },
  { value: 'pending', label: 'pending' },
  { value: 'in_progress', label: 'in_progress' },
  { value: 'completed', label: 'completed' },
  { value: 'failed', label: 'failed' },
  { value: 'expired', label: 'expired' }
])

const mediaTypeOptions = computed(() => [
  { value: '', label: t('mediaTasks.filters.allTypes') },
  { value: 'video', label: 'video' },
  { value: 'image', label: 'image' },
  { value: 'audio', label: 'audio' }
])

function statusClass(status: string) {
  if (status === 'completed') return 'badge-success'
  if (status === 'failed' || status === 'expired') return 'badge-danger'
  if (status === 'in_progress' || status === 'pending') return 'badge-warning'
  return 'badge-secondary'
}

function formatDate(value: string) {
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}

async function loadTasks() {
  loading.value = true
  try {
    const res = await mediaAPI.listTasks({
      page: page.value,
      page_size: pageSize.value,
      status: filterStatus.value || undefined,
      media_type: filterMediaType.value || undefined,
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

function goPage(next: number) {
  page.value = next
  loadTasks()
}

onMounted(loadTasks)
</script>
