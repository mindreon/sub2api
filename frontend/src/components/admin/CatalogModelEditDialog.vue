<template>
  <BaseDialog :show="show" :title="`编辑模型 — ${model.model_id}`" width="wide" @close="$emit('close')">
    <!-- Tab bar -->
    <div class="mb-4 flex border-b border-gray-200 dark:border-dark-600">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="[
          'px-4 py-2 text-sm font-medium transition-colors',
          activeTab === tab.id
            ? 'border-b-2 border-primary-500 text-primary-600 dark:text-primary-400'
            : 'text-gray-500 hover:text-gray-700 dark:text-dark-400 dark:hover:text-dark-200'
        ]"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
    </div>

    <form @submit.prevent="handleSave">
      <!-- Tab 1: 基础信息 -->
      <div v-show="activeTab === 'basic'" class="space-y-4">
        <div>
          <label class="form-label">Model ID（只读）</label>
          <input :value="form.model_id" disabled class="input opacity-60" />
        </div>
        <div>
          <label class="form-label">名称 *</label>
          <input v-model="form.name" class="input" required />
        </div>
        <div>
          <label class="form-label">厂家</label>
          <Select v-model="form.vendor" :options="vendorOptions" />
        </div>
        <div>
          <label class="form-label">分类</label>
          <Select v-model="form.category" :options="categoryOptions" />
        </div>
        <div>
          <label class="form-label">描述</label>
          <textarea v-model="form.description" class="input" rows="3" />
        </div>
        <div>
          <label class="form-label">标签（逗号分隔）</label>
          <input v-model="tagsText" class="input" placeholder="tag1, tag2, tag3" />
        </div>
        <div>
          <label class="form-label">文档链接</label>
          <input v-model="form.doc_url" class="input" />
        </div>
        <div>
          <label class="form-label">图标链接</label>
          <input v-model="form.icon_url" class="input" />
          <img v-if="form.icon_url" :src="form.icon_url" class="mt-2 h-8 w-8 object-contain" />
        </div>
      </div>

      <!-- Tab 2: 规格参数 -->
      <div v-show="activeTab === 'specs'" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="form-label">上下文窗口（tokens）</label>
            <input v-model.number="form.context_window" type="number" min="0" class="input" />
          </div>
          <div>
            <label class="form-label">最大输出 tokens</label>
            <input v-model.number="form.max_output_tokens" type="number" min="0" class="input" />
          </div>
        </div>
        <div>
          <label class="form-label">输入模态</label>
          <div class="flex flex-wrap gap-3">
            <label v-for="m in modalityOptions" :key="m" class="flex items-center gap-1 text-sm cursor-pointer">
              <input type="checkbox" :value="m" v-model="form.input_modalities" class="rounded" />
              {{ m }}
            </label>
          </div>
        </div>
        <div>
          <label class="form-label">输出模态</label>
          <div class="flex flex-wrap gap-3">
            <label v-for="m in outputModalityOptions" :key="m" class="flex items-center gap-1 text-sm cursor-pointer">
              <input type="checkbox" :value="m" v-model="form.output_modalities" class="rounded" />
              {{ m }}
            </label>
          </div>
        </div>
        <div>
          <label class="form-label">能力特性</label>
          <div class="flex flex-wrap gap-3">
            <label v-for="f in featureOptions" :key="f" class="flex items-center gap-1 text-sm cursor-pointer">
              <input type="checkbox" :value="f" v-model="form.features" class="rounded" />
              {{ f }}
            </label>
          </div>
        </div>
      </div>

      <!-- Tab 3: 价格 -->
      <div v-show="activeTab === 'pricing'" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="form-label">输入价格（$/1M tokens）</label>
            <input v-model.number="form.input_price" type="number" step="0.0001" min="0" class="input" />
          </div>
          <div>
            <label class="form-label">输出价格（$/1M tokens）</label>
            <input v-model.number="form.output_price" type="number" step="0.0001" min="0" class="input" />
          </div>
          <div>
            <label class="form-label">缓存写入价格（可选）</label>
            <input v-model.number="form.cache_write_price" type="number" step="0.0001" min="0" class="input" placeholder="留空表示不适用" />
          </div>
          <div>
            <label class="form-label">缓存读取价格（可选）</label>
            <input v-model.number="form.cache_read_price" type="number" step="0.0001" min="0" class="input" placeholder="留空表示不适用" />
          </div>
        </div>
        <div>
          <label class="form-label">货币</label>
          <Select v-model="form.currency" :options="currencyOptions" />
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-6 flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="$emit('close')">取消</button>
        <button type="submit" class="btn btn-primary" :disabled="saving">
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </form>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import { catalogModels } from '@/api/admin'
import type { CatalogModel } from '@/api/admin/catalogModels'

const props = defineProps<{ show: boolean; model: CatalogModel }>()
const emit = defineEmits<{
  close: []
  saved: [updated: CatalogModel]
}>()

const appStore = useAppStore()
const saving = ref(false)
const activeTab = ref<'basic' | 'specs' | 'pricing'>('basic')

const tabs = [
  { id: 'basic' as const, label: '基础信息' },
  { id: 'specs' as const, label: '规格参数' },
  { id: 'pricing' as const, label: '价格' },
]

const form = reactive({
  ...props.model,
  input_modalities: [...(props.model.input_modalities ?? [])],
  output_modalities: [...(props.model.output_modalities ?? [])],
  features: [...(props.model.features ?? [])],
  tags: [...(props.model.tags ?? [])],
})

const tagsText = computed({
  get: () => form.tags.join(', '),
  set: (v: string) => {
    form.tags = v.split(',').map((t) => t.trim()).filter(Boolean)
  },
})

const vendorOptions = [
  { label: 'OpenAI', value: 'OpenAI' },
  { label: 'Anthropic', value: 'Anthropic' },
  { label: 'Google', value: 'Google' },
  { label: 'Meta', value: 'Meta' },
  { label: 'Mistral', value: 'Mistral' },
  { label: 'xAI', value: 'xAI' },
  { label: 'Other', value: 'Other' },
]

const categoryOptions = [
  { label: 'chat', value: 'chat' },
  { label: 'embedding', value: 'embedding' },
  { label: 'image', value: 'image' },
  { label: 'audio', value: 'audio' },
  { label: 'video', value: 'video' },
]

const currencyOptions = [{ label: 'USD', value: 'USD' }]
const modalityOptions = ['text', 'image', 'audio', 'video']
const outputModalityOptions = ['text', 'image', 'audio']
const featureOptions = ['streaming', 'function_calling', 'vision', 'json_mode', 'file_upload']

async function handleSave() {
  saving.value = true
  try {
    const updated = await catalogModels.update(props.model.id, {
      name: form.name,
      vendor: form.vendor,
      category: form.category,
      description: form.description,
      tags: form.tags,
      doc_url: form.doc_url,
      icon_url: form.icon_url,
      context_window: form.context_window,
      max_output_tokens: form.max_output_tokens,
      input_modalities: form.input_modalities,
      output_modalities: form.output_modalities,
      features: form.features,
      input_price: form.input_price,
      output_price: form.output_price,
      cache_write_price: form.cache_write_price ?? null,
      cache_read_price: form.cache_read_price ?? null,
      currency: form.currency || 'USD',
    })
    appStore.showSuccess('保存成功')
    emit('saved', updated)
  } catch {
    appStore.showError('保存失败，请重试')
  } finally {
    saving.value = false
  }
}
</script>
