<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.media.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.media.description') }}
      </p>
    </div>

    <form class="space-y-6 p-6" @submit.prevent="handleSave">
      <div class="rounded-lg border border-blue-100 bg-blue-50 p-4 text-sm text-blue-900 dark:border-blue-900/40 dark:bg-blue-950/30 dark:text-blue-100">
        {{ t('admin.settings.media.accountHint') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.settings.media.cnyToUsdRate') }}</label>
        <input
          v-model.number="form.cny_to_usd_rate"
          type="number"
          step="0.0001"
          min="0.0001"
          class="input mt-1 max-w-xs"
          required
        />
        <p class="input-hint">{{ t('admin.settings.media.cnyToUsdRateHint') }}</p>
      </div>

      <div>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <label class="input-label">{{ t('admin.settings.media.pricingOverrides') }}</label>
          <button type="button" class="btn btn-secondary inline-flex items-center gap-2" @click="addPricingOverride">
            <Icon name="plus" size="sm" :stroke-width="2" />
            {{ t('admin.settings.media.addPricingOverride') }}
          </button>
        </div>
        <p class="input-hint">{{ t('admin.settings.media.pricingOverridesHint') }}</p>

        <div
          v-if="pricingOverrideRows.length > 0"
          class="mt-3 overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700"
        >
          <table class="min-w-[980px] w-full divide-y divide-gray-200 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 text-left text-xs font-medium uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              <tr>
                <th class="px-3 py-2">{{ t('admin.settings.media.model') }}</th>
                <th class="px-3 py-2">{{ t('admin.settings.media.metric') }}</th>
                <th class="px-3 py-2">{{ t('admin.settings.media.pricePerMillion') }}</th>
                <th class="px-3 py-2">{{ t('admin.settings.media.currency') }}</th>
                <th class="px-3 py-2">{{ t('admin.settings.media.resolutions') }}</th>
                <th class="px-3 py-2">{{ t('admin.settings.media.videoInput') }}</th>
                <th class="px-3 py-2">{{ t('admin.settings.media.audio') }}</th>
                <th class="w-14 px-3 py-2 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-for="(row, index) in pricingOverrideRows" :key="row.id">
                <td class="px-3 py-2">
                  <input
                    v-model.trim="row.model"
                    type="text"
                    class="input h-9 min-w-48"
	                    placeholder="dreamina-seedance-2-0-260128"
                  />
                </td>
                <td class="px-3 py-2">
                  <select v-model="row.metric" class="input h-9 min-w-36">
                    <option value="video_token">video_token</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <input
                    v-model.number="row.price_per_million"
                    type="number"
                    step="0.0001"
                    min="0.0001"
                    class="input h-9 min-w-36"
                  />
                </td>
                <td class="px-3 py-2">
                  <select v-model="row.currency" class="input h-9 min-w-24">
                    <option value="CNY">CNY</option>
                    <option value="USD">USD</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <input
                    v-model.trim="row.resolutionsText"
                    type="text"
                    class="input h-9 min-w-40"
                    placeholder="720p, 1080p"
                  />
                </td>
                <td class="px-3 py-2">
                  <select v-model="row.hasVideoInput" class="input h-9 min-w-28">
                    <option value="any">{{ t('common.all') }}</option>
                    <option value="true">{{ t('common.yes') }}</option>
                    <option value="false">{{ t('common.no') }}</option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <select v-model="row.hasAudio" class="input h-9 min-w-28">
                    <option value="any">{{ t('common.all') }}</option>
                    <option value="true">{{ t('common.yes') }}</option>
                    <option value="false">{{ t('common.no') }}</option>
                  </select>
                </td>
                <td class="px-3 py-2 text-right">
                  <button
                    type="button"
                    class="inline-flex h-9 w-9 items-center justify-center rounded-md text-red-600 hover:bg-red-50 disabled:opacity-50 dark:text-red-400 dark:hover:bg-red-950/30"
                    :title="t('common.delete')"
                    :aria-label="t('common.delete')"
                    @click="removePricingOverride(index)"
                  >
                    <Icon name="trash" size="sm" :stroke-width="2" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div
          v-else
          class="mt-3 rounded-lg border border-dashed border-gray-300 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400"
        >
          {{ t('admin.settings.media.noPricingOverrides') }}
        </div>
      </div>

      <div class="flex justify-end">
        <button type="submit" class="btn btn-primary" :disabled="loading || saving">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminMediaAPI, type MediaPricingOverride } from '@/api/admin/media'
import { Icon } from '@/components/icons'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
let nextPricingOverrideID = 1

type BoolFilter = 'any' | 'true' | 'false'

interface MediaPricingOverrideRow {
  id: number
  model: string
  metric: string
  price_per_million: number
  currency: 'CNY' | 'USD'
  resolutionsText: string
  hasVideoInput: BoolFilter
  hasAudio: BoolFilter
}

const form = reactive({
  cny_to_usd_rate: 0.14
})

const pricingOverrideRows = ref<MediaPricingOverrideRow[]>([])

async function loadSettings() {
  loading.value = true
  try {
    const settings = await adminMediaAPI.getSettings()
    form.cny_to_usd_rate = settings.cny_to_usd_rate || 0.14
    pricingOverrideRows.value = (settings.pricing_overrides || []).map(toPricingOverrideRow)
  } catch {
    appStore.showError(t('admin.settings.media.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const pricingOverrides = parsePricingOverrides()
    await adminMediaAPI.updateSettings({
      cny_to_usd_rate: form.cny_to_usd_rate,
      pricing_overrides: pricingOverrides
    })
    appStore.showSuccess(t('admin.settings.media.saveSuccess'))
  } catch (err) {
    appStore.showError(err instanceof Error ? err.message : t('admin.settings.media.saveFailed'))
  } finally {
    saving.value = false
  }
}

function parsePricingOverrides(): MediaPricingOverride[] {
  return pricingOverrideRows.value.map((row) => {
    if (!row.model.trim()) {
      throw new Error(t('admin.settings.media.invalidPricingOverrideModel'))
    }
    const price = Number(row.price_per_million)
    if (!Number.isFinite(price) || price <= 0) {
      throw new Error(t('admin.settings.media.invalidPricingOverridePrice'))
    }
    const override: MediaPricingOverride = {
      model: row.model.trim(),
      metric: row.metric || 'video_token',
      price_per_million: price,
      currency: row.currency,
      resolutions: parseResolutions(row.resolutionsText)
    }
    if (override.resolutions?.length === 0) {
      delete override.resolutions
    }
    const hasVideoInput = parseBoolFilter(row.hasVideoInput)
    if (hasVideoInput !== undefined) {
      override.has_video_input = hasVideoInput
    }
    const hasAudio = parseBoolFilter(row.hasAudio)
    if (hasAudio !== undefined) {
      override.has_audio = hasAudio
    }
    return override
  })
}

function toPricingOverrideRow(override: MediaPricingOverride): MediaPricingOverrideRow {
  return {
    id: nextPricingOverrideID++,
    model: override.model || '',
    metric: override.metric || 'video_token',
    price_per_million: override.price_per_million || 0,
    currency: override.currency || 'CNY',
    resolutionsText: (override.resolutions || []).join(', '),
    hasVideoInput: toBoolFilter(override.has_video_input),
    hasAudio: toBoolFilter(override.has_audio)
  }
}

function addPricingOverride() {
  pricingOverrideRows.value.push(toPricingOverrideRow({
    model: '',
    metric: 'video_token',
    price_per_million: 0,
    currency: 'CNY'
  }))
}

function removePricingOverride(index: number) {
  pricingOverrideRows.value.splice(index, 1)
}

function parseResolutions(raw: string): string[] {
  return raw
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function toBoolFilter(value: boolean | undefined): BoolFilter {
  if (value === true) return 'true'
  if (value === false) return 'false'
  return 'any'
}

function parseBoolFilter(value: BoolFilter): boolean | undefined {
  if (value === 'true') return true
  if (value === 'false') return false
  return undefined
}

onMounted(loadSettings)
</script>
