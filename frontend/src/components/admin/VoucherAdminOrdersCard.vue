<template>
  <div class="card">
    <div class="border-b border-gray-100 p-6 dark:border-dark-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.voucher.retailOrdersTitle') }}</h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.voucher.retailOrdersDescription') }}</p>
    </div>

    <div class="space-y-4 p-6">
      <div class="flex flex-wrap gap-3">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadOrders">
          {{ t('common.refresh') }}
        </button>
        <select v-model="statusFilter" class="input w-44 text-sm" @change="loadOrders">
          <option value="">{{ t('admin.settings.voucher.allStatuses') }}</option>
          <option value="payment_submitted">{{ t('voucher.status.payment_submitted') }}</option>
          <option value="fulfilling">{{ t('voucher.status.fulfilling') }}</option>
          <option value="completed">{{ t('voucher.status.completed') }}</option>
          <option value="rejected">{{ t('voucher.status.rejected') }}</option>
        </select>
      </div>

      <p v-if="actionError" class="text-sm text-red-600 dark:text-red-400">{{ actionError }}</p>

      <div v-if="loading" class="flex justify-center py-10">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent" />
      </div>

      <div v-else-if="orders.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.voucher.noOrders') }}
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-600 dark:text-gray-400">
              <th class="px-3 py-2">{{ t('voucher.orderNo') }}</th>
              <th class="px-3 py-2">{{ t('voucher.product') }}</th>
              <th class="px-3 py-2">{{ t('voucher.quantity') }}</th>
              <th class="px-3 py-2">{{ t('voucher.totalDue') }}</th>
              <th class="px-3 py-2">{{ t('payment.orders.status') }}</th>
              <th class="px-3 py-2">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in orders" :key="row.id" class="border-b border-gray-100 dark:border-dark-700">
              <td class="px-3 py-3 font-mono text-xs">{{ row.order_no }}</td>
              <td class="px-3 py-3">{{ row.product_name }}</td>
              <td class="px-3 py-3">{{ row.quantity }}</td>
              <td class="px-3 py-3">{{ row.total_amount.toFixed(2) }} {{ row.currency }}</td>
              <td class="px-3 py-3">
                <span class="badge badge-info">{{ t(`voucher.status.${row.status}`) }}</span>
              </td>
              <td class="px-3 py-3">
                <div class="flex flex-wrap gap-2">
                  <button
                    v-if="row.status === 'payment_submitted'"
                    type="button"
                    class="btn btn-primary btn-xs"
                    :disabled="actingId === row.id"
                    @click="verify(row.id)"
                  >
                    {{ t('admin.settings.voucher.verify') }}
                  </button>
                  <button
                    v-if="row.status === 'payment_submitted'"
                    type="button"
                    class="btn btn-secondary btn-xs"
                    :disabled="actingId === row.id"
                    @click="reject(row.id)"
                  >
                    {{ t('admin.settings.voucher.reject') }}
                  </button>
                  <button
                    v-if="row.status === 'fulfilling'"
                    type="button"
                    class="btn btn-secondary btn-xs"
                    :disabled="actingId === row.id"
                    @click="retryFulfill(row.id)"
                  >
                    {{ t('admin.settings.voucher.retryFulfill') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { voucherAdminAPI } from '@/api/admin/voucher'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { VoucherOrder } from '@/types/voucher'

const { t } = useI18n()

const loading = ref(true)
const actingId = ref<number | null>(null)
const orders = ref<VoucherOrder[]>([])
const statusFilter = ref('payment_submitted')
const actionError = ref('')

async function loadOrders() {
  loading.value = true
  actionError.value = ''
  try {
    const res = await voucherAdminAPI.listOrders({
      page: 1,
      per_page: 50,
      status: statusFilter.value || undefined,
    })
    orders.value = res.data.orders || []
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    loading.value = false
  }
}

async function verify(id: number) {
  actingId.value = id
  actionError.value = ''
  try {
    await voucherAdminAPI.verifyOrder(id)
    await loadOrders()
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    actingId.value = null
  }
}

async function reject(id: number) {
  const reason = window.prompt(t('admin.settings.voucher.rejectReasonPrompt')) || ''
  actingId.value = id
  actionError.value = ''
  try {
    await voucherAdminAPI.rejectOrder(id, reason)
    await loadOrders()
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    actingId.value = null
  }
}

async function retryFulfill(id: number) {
  actingId.value = id
  actionError.value = ''
  try {
    await voucherAdminAPI.retryFulfill(id)
    await loadOrders()
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    actingId.value = null
  }
}

onMounted(() => {
  void loadOrders()
})
</script>
