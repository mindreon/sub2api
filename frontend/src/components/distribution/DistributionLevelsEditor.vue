<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="min-w-0 flex-1">
        <p v-if="title" class="text-sm font-medium text-gray-900 dark:text-white">{{ title }}</p>
        <p v-if="description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">{{ description }}</p>
      </div>
      <button
        v-if="!disabled"
        type="button"
        class="btn btn-secondary btn-sm shrink-0"
        :disabled="levels.length >= maxLevels"
        @click="addLevel"
      >
        {{ t('distributionLevels.add') }}
      </button>
    </div>

    <p v-if="validationError" class="text-xs text-red-600 dark:text-red-400">
      {{ validationErrorMessage }}
    </p>

    <div
      v-if="levels.length === 0"
      class="rounded-lg border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400"
    >
      {{ emptyText || t('distributionLevels.empty') }}
    </div>

    <div v-else class="space-y-2">
      <div
        class="hidden gap-3 px-11 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400 md:grid"
        :style="{ gridTemplateColumns: levelGridColumns }"
      >
        <span>{{ t('distributionLevels.fields.code') }}</span>
        <span>{{ t('distributionLevels.fields.name') }}</span>
        <span>{{ t('distributionLevels.fields.commissionRate') }}</span>
        <span class="text-center">{{ t('distributionLevels.fields.active') }}</span>
      </div>

      <VueDraggable
        v-model="levels"
        :animation="200"
        handle=".level-drag-handle"
        item-key="code"
        class="space-y-2"
        :disabled="disabled"
      >
        <div
          v-for="(level, index) in levels"
          :key="levelRowKey(level, index)"
          class="rounded-lg border border-gray-200 bg-gray-50/80 p-3 dark:border-dark-700 dark:bg-dark-900/40"
        >
          <div class="flex items-start gap-2">
            <button
              type="button"
              class="level-drag-handle mt-8 flex shrink-0 cursor-grab text-gray-300 hover:text-gray-500 active:cursor-grabbing dark:text-dark-600 dark:hover:text-dark-400 md:mt-6"
              :disabled="disabled"
              :aria-label="t('distributionLevels.dragHandle')"
            >
              <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path d="M7 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4z" />
              </svg>
            </button>

            <div class="min-w-0 flex-1 space-y-3">
              <div class="level-fields-grid grid gap-3" :style="{ gridTemplateColumns: levelGridColumns }">
                <label class="block min-w-0">
                  <span class="input-label whitespace-nowrap md:sr-only">{{ t('distributionLevels.fields.code') }}</span>
                  <input
                    v-model="level.code"
                    class="input mt-1 font-mono text-sm uppercase md:mt-0"
                    :placeholder="t('distributionLevels.fields.code')"
                    :disabled="disabled"
                    @blur="normalizeLevelCode(level)"
                  />
                </label>
                <label class="block min-w-0">
                  <span class="input-label whitespace-nowrap md:sr-only">{{ t('distributionLevels.fields.name') }}</span>
                  <input
                    v-model="level.name"
                    class="input mt-1 md:mt-0"
                    :placeholder="t('distributionLevels.fields.name')"
                    :disabled="disabled"
                    @blur="maybeSuggestCode(level)"
                  />
                </label>
                <label class="block min-w-0">
                  <span class="input-label whitespace-nowrap md:sr-only">{{ t('distributionLevels.fields.commissionRate') }}</span>
                  <div class="relative mt-1 md:mt-0">
                    <input
                      v-model.number="level.commission_rate"
                      type="number"
                      min="0"
                      max="100"
                      step="0.01"
                      class="input w-full pr-8"
                      :disabled="disabled"
                    />
                    <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400">%</span>
                  </div>
                </label>
                <div class="flex items-center justify-start md:justify-center md:pt-0 pt-1">
                  <label class="inline-flex items-center gap-2">
                    <input
                      :id="`level-active-${index}`"
                      v-model="level.active"
                      type="checkbox"
                      class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                      :disabled="disabled"
                    />
                    <span class="text-sm text-gray-700 dark:text-dark-200 md:sr-only">{{ t('distributionLevels.fields.active') }}</span>
                    <span class="text-sm text-gray-700 dark:text-dark-200 md:hidden" aria-hidden="true">{{ t('distributionLevels.fields.active') }}</span>
                  </label>
                </div>
              </div>

              <label class="block">
                <span class="input-label whitespace-nowrap">{{ t('distributionLevels.fields.note') }}</span>
                <input
                  v-model="level.note"
                  class="input mt-1"
                  :placeholder="t('distributionLevels.fields.notePlaceholder')"
                  :disabled="disabled"
                />
              </label>
            </div>

            <button
              v-if="!disabled"
              type="button"
              class="mt-8 shrink-0 text-gray-400 hover:text-red-600 dark:hover:text-red-400 md:mt-6"
              :aria-label="t('common.delete')"
              @click="removeLevel(index)"
            >
              <Icon name="trash" size="sm" :stroke-width="2" />
            </button>
          </div>
        </div>
      </VueDraggable>
    </div>

    <details v-if="showAdvancedJson" class="rounded-lg border border-gray-200 dark:border-dark-700">
      <summary class="cursor-pointer px-3 py-2 text-xs font-medium text-gray-600 dark:text-dark-300">
        {{ t('distributionLevels.advancedJson') }}
      </summary>
      <div class="border-t border-gray-200 p-3 dark:border-dark-700">
        <textarea
          v-model="advancedJson"
          class="input min-h-[120px] font-mono text-xs"
          spellcheck="false"
          :disabled="disabled"
          @blur="applyAdvancedJson"
        />
        <p v-if="advancedJsonError" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ advancedJsonError }}</p>
      </div>
    </details>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { VueDraggable } from 'vue-draggable-plus'
import Icon from '@/components/icons/Icon.vue'
import type { DistributionLevelConfig } from '@/api/admin/settings'
import {
  cloneDistributionLevels,
  DISTRIBUTION_LEVELS_MAX,
  normalizeDistributionLevelConfigs,
  suggestLevelCodeFromName,
  validateDistributionLevelConfigs,
} from '@/utils/distributionLevels'

const levelGridColumns = 'minmax(7.5rem, 1fr) minmax(9rem, 1.6fr) minmax(6.5rem, 0.9fr) 4.5rem'

const props = withDefaults(defineProps<{
  modelValue: DistributionLevelConfig[]
  disabled?: boolean
  title?: string
  description?: string
  emptyText?: string
  showAdvancedJson?: boolean
  maxLevels?: number
}>(), {
  disabled: false,
  showAdvancedJson: true,
  maxLevels: DISTRIBUTION_LEVELS_MAX,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: DistributionLevelConfig[]): void
}>()

const { t } = useI18n()
const levels = ref<DistributionLevelConfig[]>([])
const advancedJson = ref('[]')
const advancedJsonError = ref('')
const validationError = ref<string | null>(null)

const validationErrorMessage = computed(() => {
  switch (validationError.value) {
    case 'required':
      return t('distributionLevels.errors.required')
    case 'duplicate':
      return t('distributionLevels.errors.duplicate')
    case 'rate':
      return t('distributionLevels.errors.rate')
    case 'max':
      return t('distributionLevels.errors.max', { max: props.maxLevels })
    default:
      return ''
  }
})

function levelsEqual(a: DistributionLevelConfig[], b: DistributionLevelConfig[]) {
  return JSON.stringify(a) === JSON.stringify(b)
}

watch(
  () => props.modelValue,
  (value) => {
    const incoming = cloneDistributionLevels(value || [])
    if (levelsEqual(incoming, levels.value)) return
    levels.value = incoming
    advancedJson.value = JSON.stringify(incoming, null, 2)
    validationError.value = validateDistributionLevelConfigs(levels.value)
  },
  { immediate: true, deep: true },
)

watch(
  levels,
  () => {
    levels.value.forEach((level, index) => {
      level.sort_order = index
    })
    validationError.value = validateDistributionLevelConfigs(levels.value)
    const next = cloneDistributionLevels(levels.value)
    advancedJson.value = JSON.stringify(next, null, 2)
    if (!levelsEqual(next, props.modelValue)) {
      emit('update:modelValue', next)
    }
  },
  { deep: true },
)

function levelRowKey(level: DistributionLevelConfig, index: number) {
  return `${level.code || 'new'}-${index}`
}

function normalizeLevelCode(level: DistributionLevelConfig) {
  level.code = String(level.code || '').trim().toUpperCase()
}

function maybeSuggestCode(level: DistributionLevelConfig) {
  if (!String(level.code || '').trim() && String(level.name || '').trim()) {
    level.code = suggestLevelCodeFromName(level.name)
  }
}

function addLevel() {
  if (levels.value.length >= props.maxLevels) return
  levels.value.push({
    code: '',
    name: '',
    commission_rate: 0,
    active: true,
    sort_order: levels.value.length,
    note: '',
  })
}

function removeLevel(index: number) {
  levels.value.splice(index, 1)
}

function applyAdvancedJson() {
  advancedJsonError.value = ''
  try {
    const parsed = normalizeDistributionLevelConfigs(JSON.parse(advancedJson.value || '[]'))
    if (!parsed) {
      throw new Error('invalid')
    }
    levels.value = parsed.map((level, index) => ({ ...level, sort_order: index }))
    validationError.value = validateDistributionLevelConfigs(levels.value)
    emit('update:modelValue', cloneDistributionLevels(levels.value))
  } catch {
    advancedJsonError.value = t('distributionLevels.errors.json')
  }
}

defineExpose({
  validate(): boolean {
    validationError.value = validateDistributionLevelConfigs(levels.value)
    return !validationError.value
  },
})
</script>

<style scoped>
@media (max-width: 767px) {
  .level-fields-grid {
    grid-template-columns: 1fr !important;
  }
}
</style>
