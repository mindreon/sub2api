<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 sm:max-w-64">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索 Model ID / 名称"
              class="input"
              @input="handleSearch"
            />
          </div>
          <Select
            v-model="filters.vendor"
            :options="vendorOptions"
            class="w-36"
            @change="() => loadModels()"
          />
          <Select
            v-model="filters.category"
            :options="categoryOptions"
            class="w-36"
            @change="() => loadModels()"
          />
          <Select
            v-model="enabledFilter"
            :options="enabledOptions"
            class="w-32"
            @change="() => loadModels()"
          />
          <div class="flex flex-1 items-center justify-end gap-2">
            <button @click="() => loadModels()" :disabled="loading" class="btn btn-secondary" title="刷新">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button @click="handleSeed" :disabled="seeding" class="btn btn-secondary">
              {{ seeding ? '重置中...' : '重置种子数据' }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="models"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="model_id"
          default-sort-order="asc"
          @sort="handleSort"
        >
          <template #cell-model_id="{ value, row }">
            <div class="min-w-0">
              <div class="font-mono text-sm font-medium text-gray-900 dark:text-white">{{ value }}</div>
              <div class="mt-0.5 text-xs text-gray-500">{{ row.name }}</div>
            </div>
          </template>

          <template #cell-vendor="{ value }">
            <span class="badge badge-gray">{{ value || '—' }}</span>
          </template>

          <template #cell-category="{ value }">
            <span class="badge badge-gray">{{ value }}</span>
          </template>

          <template #cell-input_price="{ value }">
            <span class="text-sm tabular-nums">{{ value > 0 ? `$${value}` : '—' }}</span>
          </template>

          <template #cell-output_price="{ value }">
            <span class="text-sm tabular-nums">{{ value > 0 ? `$${value}` : '—' }}</span>
          </template>

          <template #cell-is_enabled="{ value, row }">
            <Toggle
              :model-value="value"
              @update:model-value="handleToggle(row)"
            />
          </template>

          <template #cell-actions="{ row }">
            <button @click="openEdit(row)" class="btn btn-ghost btn-sm">
              <Icon name="edit" size="sm" />
            </button>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          :page="currentPage"
          :total="total"
          :page-size="pageSize"
          @update:page="handlePageChange"
        />
      </template>
    </TablePageLayout>

    <CatalogModelEditDialog
      v-if="editingModel"
      :show="true"
      :model="editingModel"
      @close="editingModel = null"
      @saved="onSaved"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'
import CatalogModelEditDialog from '@/components/admin/CatalogModelEditDialog.vue'
import { catalogModels } from '@/api/admin'
import type { CatalogModel } from '@/api/admin/catalogModels'
import type { Column } from '@/components/common/types'

const appStore = useAppStore()

const models = ref<CatalogModel[]>([])
const loading = ref(false)
const seeding = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const searchQuery = ref('')
const editingModel = ref<CatalogModel | null>(null)

const filters = ref({ vendor: '', category: '' })
const enabledFilter = ref('')

const vendorOptions = [
  { label: '全部厂家', value: '' },
  { label: 'OpenAI', value: 'OpenAI' },
  { label: 'Anthropic', value: 'Anthropic' },
  { label: 'Google', value: 'Google' },
  { label: 'Meta', value: 'Meta' },
  { label: 'Mistral', value: 'Mistral' },
  { label: 'xAI', value: 'xAI' },
  { label: 'Other', value: 'Other' },
]

const categoryOptions = [
  { label: '全部分类', value: '' },
  { label: 'chat', value: 'chat' },
  { label: 'embedding', value: 'embedding' },
  { label: 'image', value: 'image' },
  { label: 'audio', value: 'audio' },
  { label: 'video', value: 'video' },
]

const enabledOptions = [
  { label: '全部状态', value: '' },
  { label: '上架', value: 'true' },
  { label: '下架', value: 'false' },
]

const columns: Column[] = [
  { key: 'model_id', label: 'Model ID', sortable: true },
  { key: 'vendor', label: '厂家', sortable: true },
  { key: 'category', label: '分类', sortable: true },
  { key: 'input_price', label: '输入价格', sortable: true },
  { key: 'output_price', label: '输出价格', sortable: true },
  { key: 'is_enabled', label: '上架', sortable: false },
  { key: 'actions', label: '', sortable: false },
]

let searchTimer: ReturnType<typeof setTimeout>
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadModels()
  }, 300)
}

function handleSort(sortBy: string, sortOrder: 'asc' | 'desc') {
  loadModels(sortBy, sortOrder)
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadModels()
}

async function loadModels(sortBy = 'model_id', sortOrder: 'asc' | 'desc' = 'asc') {
  loading.value = true
  try {
    const resp = await catalogModels.list(currentPage.value, pageSize.value, {
      vendor: filters.value.vendor || undefined,
      category: filters.value.category || undefined,
      enabled: enabledFilter.value !== '' ? enabledFilter.value === 'true' : undefined,
      q: searchQuery.value || undefined,
      sort_by: sortBy,
      sort_order: sortOrder,
    })
    models.value = resp.items
    total.value = resp.total
  } catch {
    appStore.showError('加载模型列表失败')
  } finally {
    loading.value = false
  }
}

async function handleToggle(model: CatalogModel) {
  try {
    const updated = await catalogModels.toggle(model.id)
    const idx = models.value.findIndex((m) => m.id === model.id)
    if (idx !== -1) models.value[idx] = updated
  } catch {
    appStore.showError('操作失败，请重试')
    loadModels()
  }
}

async function handleSeed() {
  if (!confirm('确认重置种子数据？这将覆盖所有模型字段（不影响上架/下架状态）。')) return
  seeding.value = true
  try {
    const resp = await catalogModels.seed()
    appStore.showSuccess(`已重置 ${resp.seeded} 条模型数据`)
    loadModels()
  } catch {
    appStore.showError('重置失败，请查看日志')
  } finally {
    seeding.value = false
  }
}

function openEdit(model: CatalogModel) {
  editingModel.value = { ...model }
}

function onSaved(updated: CatalogModel) {
  const idx = models.value.findIndex((m) => m.id === updated.id)
  if (idx !== -1) models.value[idx] = updated
  editingModel.value = null
}

onMounted(loadModels)
</script>
