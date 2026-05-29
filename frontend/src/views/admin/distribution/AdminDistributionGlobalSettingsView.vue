<template>
  <AdminDistributionLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
          {{ t('admin.distribution.globalSettings.title') }}
        </h1>
        <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.distribution.globalSettings.description') }}
        </p>
      </div>

      <form class="card" @submit.prevent="handleSave">
        <div class="space-y-5 p-6">
          <div>
            <label class="input-label">
              {{ t('admin.distribution.globalSettings.freezeHours') }}
            </label>
            <input
              v-model.number="form.distribution_freeze_hours"
              type="number"
              step="1"
              min="0"
              max="720"
              class="input"
              :disabled="loading || saving"
            />
            <p class="mt-1 text-xs text-gray-400">
              {{ t('admin.distribution.globalSettings.freezeHoursDesc') }}
            </p>
          </div>

          <div>
            <label class="input-label">
              {{ t('admin.distribution.globalSettings.kol2Rate') }}
            </label>
            <div class="relative">
              <input
                v-model.number="form.distribution_kol2_rate"
                type="number"
                step="0.01"
                min="0"
                max="100"
                class="input pr-8"
                :disabled="loading || saving"
              />
              <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400">%</span>
            </div>
            <p class="mt-1 text-xs text-gray-400">
              {{ t('admin.distribution.globalSettings.kol2RateDesc') }}
            </p>
          </div>

          <div>
            <label class="input-label">
              {{ t('admin.distribution.globalSettings.commissionUpperRatio') }}
            </label>
            <div class="relative">
              <input
                v-model.number="form.distribution_commission_upper_ratio"
                type="number"
                step="0.01"
                min="0"
                max="100"
                class="input pr-8"
                :disabled="loading || saving"
              />
              <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400">%</span>
            </div>
            <p class="mt-1 text-xs text-gray-400">
              {{ t('admin.distribution.globalSettings.commissionUpperRatioDesc') }}
            </p>
          </div>

          <DistributionLevelsEditor
            ref="levelsEditorRef"
            v-model="distributionGlobalLevels"
            :disabled="loading || saving"
            :title="t('admin.distribution.globalSettings.levelsTitle')"
            :description="t('admin.distribution.globalSettings.levelsDesc')"
          />
        </div>

        <div class="flex justify-end border-t border-gray-100 px-6 py-4 dark:border-dark-700">
          <button type="submit" class="btn btn-primary" :disabled="loading || saving || loadFailed">
            <svg
              v-if="saving"
              class="h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </form>
    </div>
  </AdminDistributionLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AdminDistributionLayout from '@/components/layout/AdminDistributionLayout.vue'
import DistributionLevelsEditor from '@/components/distribution/DistributionLevelsEditor.vue'
import { getSettings, updateSettings, type DistributionLevelConfig } from '@/api/admin/settings'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { normalizeDistributionLevelConfigs } from '@/utils/distributionLevels'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const loadFailed = ref(false)

const form = reactive({
  distribution_freeze_hours: 168,
  distribution_kol2_rate: 5,
  distribution_commission_upper_ratio: 35,
})

const distributionGlobalLevels = ref<DistributionLevelConfig[]>([])
const levelsEditorRef = ref<InstanceType<typeof DistributionLevelsEditor> | null>(null)

async function loadSettings() {
  loading.value = true
  loadFailed.value = false
  try {
    const settings = await getSettings()
    form.distribution_freeze_hours = settings.distribution_freeze_hours ?? 168
    form.distribution_kol2_rate = settings.distribution_kol2_rate ?? 5
    form.distribution_commission_upper_ratio = settings.distribution_commission_upper_ratio ?? 35
    distributionGlobalLevels.value =
      normalizeDistributionLevelConfigs(settings.distribution_global_levels ?? []) ?? []
  } catch (error) {
    loadFailed.value = true
    appStore.showError(extractApiErrorMessage(error, t('admin.settings.failedToLoad')))
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  if (!levelsEditorRef.value?.validate()) {
    appStore.showError(t('admin.distribution.globalSettings.levelsValidationError'))
    return
  }

  const parsedLevels = normalizeDistributionLevelConfigs(distributionGlobalLevels.value) ?? []

  saving.value = true
  try {
    await updateSettings({
      distribution_freeze_hours: Math.max(
        0,
        Math.min(720, Math.floor(Number(form.distribution_freeze_hours) || 0)),
      ),
      distribution_kol2_rate: Math.max(0, Number(form.distribution_kol2_rate) || 0),
      distribution_commission_upper_ratio: Math.max(
        0,
        Number(form.distribution_commission_upper_ratio) || 0,
      ),
      distribution_global_levels: parsedLevels,
    })
    appStore.showSuccess(t('admin.distribution.globalSettings.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.settings.failedToSave')))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadSettings()
})
</script>
