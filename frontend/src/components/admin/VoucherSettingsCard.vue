<template>
  <div class="card">
    <div class="border-b border-gray-100 p-6 dark:border-dark-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.settings.voucher.title') }}</h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.voucher.description') }}</p>
    </div>

    <div class="space-y-6 p-6">
      <div v-if="loading" class="flex justify-center py-8">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent" />
      </div>

      <template v-else-if="form">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.voucher.panelHint') }}</p>
        <p v-if="loadError" class="text-sm text-red-600 dark:text-red-400">{{ loadError }}</p>
        <p v-if="saveError" class="text-sm text-red-600 dark:text-red-400">{{ saveError }}</p>
        <p v-if="saveOk" class="text-sm text-green-700 dark:text-green-300">{{ t('admin.settings.voucher.saved') }}</p>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <label class="flex items-center gap-2 text-sm">
            <input v-model="form.enabled" type="checkbox" class="rounded border-gray-300" />
            {{ t('admin.settings.voucher.backendEnabled') }}
          </label>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="form.ui_enabled" type="checkbox" class="rounded border-gray-300" />
            {{ t('admin.settings.voucher.frontendTab') }}
          </label>
          <label class="flex items-center gap-2 text-sm">
            <input v-model="form.sandbox" type="checkbox" class="rounded border-gray-300" />
            {{ t('admin.settings.voucher.sandbox') }}
          </label>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.settings.voucher.apiKey') }}</label>
            <input v-model="form.api_key" type="text" class="input w-full font-mono text-sm" :placeholder="settings?.api_key_masked || 'kvm_test_...'" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.settings.voucher.apiSecret') }}</label>
            <input v-model="form.api_secret" type="password" class="input w-full text-sm" :placeholder="settings?.secret_configured ? t('admin.settings.voucher.secretKeepPlaceholder') : t('admin.settings.voucher.apiSecret')" />
          </div>
          <div class="md:col-span-2">
            <label class="mb-1 block text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.settings.voucher.apiBase') }}</label>
            <input v-model="form.api_base" type="text" class="input w-full font-mono text-sm" />
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500">{{ t('admin.settings.voucher.orderTimeoutHours') }}</label>
            <input v-model.number="form.order_timeout_hours" type="number" min="1" class="input w-full" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500">{{ t('admin.settings.voucher.maxQuantity') }}</label>
            <input v-model.number="form.max_quantity_per_order" type="number" min="1" class="input w-full" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500">{{ t('admin.settings.voucher.reviewSlaHours') }}</label>
            <input v-model.number="form.review_sla_hours" type="number" min="1" class="input w-full" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500">{{ t('admin.settings.voucher.feeRate') }}</label>
            <input v-model.number="form.fee_rate" type="number" min="0" step="0.01" class="input w-full" />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-500">{{ t('admin.settings.voucher.retailMarkup') }}</label>
            <input v-model.number="form.retail_markup_percent" type="number" min="0" step="0.1" class="input w-full" />
          </div>
        </div>

        <div>
          <label class="mb-1 block text-xs font-medium text-gray-500">{{ t('admin.settings.voucher.helpText') }}</label>
          <textarea v-model="form.help_text" rows="2" class="input w-full text-sm" />
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.settings.voucher.bankAccountsTitle') }}</h4>
            <button type="button" class="btn btn-secondary btn-sm" @click="addBankAccount">{{ t('admin.settings.voucher.addBankAccount') }}</button>
          </div>
          <div v-if="form.bank_accounts.length === 0" class="text-sm text-gray-500">{{ t('admin.settings.voucher.noBankAccounts') }}</div>
          <div v-for="(account, idx) in form.bank_accounts" :key="idx" class="grid grid-cols-1 gap-3 rounded-xl border border-gray-200 p-4 md:grid-cols-2 dark:border-dark-600">
            <input v-model="account.bank_name" type="text" class="input text-sm" :placeholder="t('admin.settings.voucher.bankName')" />
            <input v-model="account.account_name" type="text" class="input text-sm" :placeholder="t('admin.settings.voucher.accountName')" />
            <input v-model="account.account_number" type="text" class="input font-mono text-sm md:col-span-2" :placeholder="t('admin.settings.voucher.accountNumber')" />
            <button type="button" class="btn btn-secondary btn-sm md:col-span-2" @click="removeBankAccount(idx)">{{ t('common.delete') }}</button>
          </div>
        </div>

        <div v-if="testResult" class="rounded-xl border p-4 text-sm" :class="testResult.ok ? 'border-green-200 bg-green-50 text-green-900 dark:border-green-800/40 dark:bg-green-900/20 dark:text-green-100' : 'border-red-200 bg-red-50 text-red-900 dark:border-red-800/40 dark:bg-red-900/20 dark:text-red-100'">
          <p class="font-medium">{{ testResult.message }}</p>
          <p v-if="testResult.request_id" class="mt-1 font-mono text-xs opacity-80">request_id: {{ testResult.request_id }}</p>
        </div>

        <div class="flex flex-wrap gap-3">
          <button type="button" class="btn btn-primary" :disabled="saving" @click="save">
            <span v-if="saving">{{ t('common.processing') }}</span>
            <span v-else>{{ t('common.save') }}</span>
          </button>
          <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">{{ t('common.refresh') }}</button>
          <button type="button" class="btn btn-secondary" :disabled="testing" @click="testConnection">{{ t('admin.settings.voucher.testConnection') }}</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { voucherAdminAPI, type VoucherAdminSettings, type VoucherAdminSettingsUpdate, type VoucherTestConnectionResult } from '@/api/admin/voucher'
import type { VoucherBankAccount } from '@/types/voucher'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()

interface VoucherSettingsForm {
  enabled: boolean
  ui_enabled: boolean
  sandbox: boolean
  api_key: string
  api_secret: string
  api_base: string
  bank_accounts: VoucherBankAccount[]
  order_timeout_hours: number
  max_quantity_per_order: number
  review_sla_hours: number
  fee_rate: number
  help_text: string
  retail_markup_percent: number
}

const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const settings = ref<VoucherAdminSettings | null>(null)
const form = ref<VoucherSettingsForm | null>(null)
const testResult = ref<VoucherTestConnectionResult | null>(null)
const loadError = ref('')
const saveError = ref('')
const saveOk = ref(false)

function applySettings(data: VoucherAdminSettings) {
  settings.value = data
  form.value = {
    enabled: data.enabled,
    ui_enabled: data.ui_enabled,
    sandbox: data.sandbox,
    api_key: '',
    api_secret: '',
    api_base: data.api_base,
    bank_accounts: data.bank_accounts?.length ? data.bank_accounts.map((b, i) => ({ ...b, id: b.id || i + 1 })) : [],
    order_timeout_hours: data.order_timeout_hours,
    max_quantity_per_order: data.max_quantity_per_order,
    review_sla_hours: data.review_sla_hours,
    fee_rate: data.fee_rate,
    help_text: data.help_text,
    retail_markup_percent: data.retail_markup_percent,
  }
}

async function load() {
  loading.value = true
  loadError.value = ''
  saveOk.value = false
  try {
    const res = await voucherAdminAPI.getSettings()
    applySettings(res.data)
  } catch (err: unknown) {
    loadError.value = extractApiErrorMessage(err)
  } finally {
    loading.value = false
  }
}

function buildUpdatePayload(): VoucherAdminSettingsUpdate {
  if (!form.value) return {}
  const payload: VoucherAdminSettingsUpdate = {
    enabled: form.value.enabled,
    ui_enabled: form.value.ui_enabled,
    sandbox: form.value.sandbox,
    api_base: form.value.api_base,
    bank_accounts: form.value.bank_accounts.map((b, i) => ({ ...b, id: b.id || i + 1 })),
    order_timeout_hours: form.value.order_timeout_hours,
    max_quantity_per_order: form.value.max_quantity_per_order,
    review_sla_hours: form.value.review_sla_hours,
    fee_rate: form.value.fee_rate,
    help_text: form.value.help_text,
    retail_markup_percent: form.value.retail_markup_percent,
  }
  if (form.value.api_key.trim()) payload.api_key = form.value.api_key.trim()
  if (form.value.api_secret.trim()) payload.api_secret = form.value.api_secret.trim()
  return payload
}

async function save() {
  if (!form.value) return
  saving.value = true
  saveError.value = ''
  saveOk.value = false
  try {
    const res = await voucherAdminAPI.updateSettings(buildUpdatePayload())
    applySettings(res.data)
    saveOk.value = true
  } catch (err: unknown) {
    saveError.value = extractApiErrorMessage(err)
  } finally {
    saving.value = false
  }
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  try {
    if (form.value?.api_key.trim() || form.value?.api_secret.trim()) {
      await save()
    }
    const res = await voucherAdminAPI.testConnection()
    testResult.value = res.data
  } catch (err: unknown) {
    testResult.value = { ok: false, configured: true, message: extractApiErrorMessage(err) }
  } finally {
    testing.value = false
  }
}

function addBankAccount() {
  if (!form.value) return
  const nextId = form.value.bank_accounts.length + 1
  form.value.bank_accounts.push({ id: nextId, bank_name: '', account_name: '', account_number: '' })
}

function removeBankAccount(index: number) {
  form.value?.bank_accounts.splice(index, 1)
}

onMounted(() => {
  void load()
})
</script>
