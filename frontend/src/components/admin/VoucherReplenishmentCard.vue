<template>
  <div class="card">
    <div class="border-b border-gray-100 p-6 dark:border-dark-700">
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.voucher.replenishmentCardTitle') }}</h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.voucher.replenishmentCardHint') }}</p>
    </div>

    <div class="space-y-6 p-6">
      <div class="flex flex-wrap gap-3">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="syncing" @click="syncCatalog">
          {{ t('admin.settings.voucher.syncCatalog') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="syncing" @click="syncStock">
          {{ t('admin.settings.voucher.syncStock') }}
        </button>
        <button type="button" class="btn btn-primary btn-sm" :disabled="creating" @click="openCreateModal">
          {{ t('admin.voucher.createB2BOrder') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadAll">
          {{ t('common.refresh') }}
        </button>
      </div>

      <p v-if="syncMessage" class="text-sm text-green-700 dark:text-green-300">{{ syncMessage }}</p>
      <p v-if="actionError" class="text-sm text-red-600 dark:text-red-400">{{ actionError }}</p>

      <div v-if="products.length" class="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <div v-for="p in products" :key="p.id" class="rounded-xl border border-gray-200 p-3 dark:border-dark-600">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ p.name }}</p>
          <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">{{ p.stock }}</p>
          <p class="text-[10px] text-gray-400">{{ t('admin.voucher.stockLabel') }}</p>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-10">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent" />
      </div>

      <div v-else-if="orders.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.voucher.noB2BOrders') }}
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 text-left text-xs uppercase tracking-wide text-gray-500 dark:border-dark-600">
              <th class="px-3 py-2">{{ t('voucher.orderNo') }}</th>
              <th class="px-3 py-2">{{ t('voucher.product') }}</th>
              <th class="px-3 py-2">{{ t('voucher.totalDue') }}</th>
              <th class="px-3 py-2">{{ t('payment.orders.status') }}</th>
              <th class="px-3 py-2">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in orders" :key="row.id" class="border-b border-gray-100 dark:border-dark-700">
              <td class="px-3 py-3 font-mono text-xs">{{ row.order_no }}</td>
              <td class="px-3 py-3">{{ summarizeItems(row.items) }}</td>
              <td class="px-3 py-3">{{ row.total_amount.toFixed(2) }} {{ row.currency }}</td>
              <td class="px-3 py-3">
                <span class="badge badge-info">{{ b2bStatusLabel(row.status) }}</span>
              </td>
              <td class="px-3 py-3">
                <div class="flex flex-wrap gap-2">
                  <button
                    v-if="row.status === 'pending_payment'"
                    type="button"
                    class="btn btn-primary btn-xs"
                    @click="openProofModal(row)"
                  >
                    {{ t('admin.voucher.uploadB2BProof') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary btn-xs"
                    :disabled="actingId === row.id"
                    @click="syncOrder(row.id)"
                  >
                    {{ t('admin.voucher.syncB2BOrder') }}
                  </button>
                  <button type="button" class="btn btn-secondary btn-xs" @click="openDetail(row.id)">
                    {{ t('common.details') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create modal -->
    <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="showCreate = false">
      <div class="card w-full max-w-lg p-6 space-y-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.voucher.createB2BOrder') }}</h3>
        <div v-if="catalogProducts.length === 0" class="text-sm text-amber-700 dark:text-amber-300">
          {{ t('admin.voucher.syncCatalogFirst') }}
        </div>
        <div v-for="(line, idx) in createLines" :key="idx" class="grid grid-cols-1 gap-2 sm:grid-cols-3">
          <select v-model.number="line.product_id" class="input text-sm sm:col-span-2">
            <option :value="0">{{ t('admin.voucher.selectProduct') }}</option>
            <option v-for="p in catalogProducts" :key="p.id" :value="p.kv_product_id || 0" :disabled="!p.kv_product_id">
              {{ p.name }} (KV #{{ p.kv_product_id }})
            </option>
          </select>
          <input v-model.number="line.quantity" type="number" min="1" class="input text-sm" :placeholder="t('voucher.quantity')" />
        </div>
        <textarea v-model="merchantNotes" rows="2" class="input w-full text-sm" :placeholder="t('admin.voucher.merchantNotes')" />
        <p v-if="createError" class="text-xs text-red-600">{{ createError }}</p>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="showCreate = false">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="creating" @click="submitCreate">
            {{ creating ? t('common.processing') : t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Proof modal -->
    <div v-if="proofOrder" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="proofOrder = null">
      <div class="card w-full max-w-lg p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ t('admin.voucher.uploadB2BProof') }}</h3>
        <p class="font-mono text-sm">{{ proofOrder.order_no }}</p>
        <div v-if="proofOrder.payment_info?.bank_accounts?.length" class="space-y-2 text-sm">
          <p class="font-medium text-gray-700 dark:text-gray-300">{{ t('admin.voucher.kvBankAccounts') }}</p>
          <div v-for="acc in proofOrder.payment_info.bank_accounts" :key="acc.id" class="rounded-lg border p-3 dark:border-dark-600">
            <p>{{ acc.bank_name }} · {{ acc.account_name }}</p>
            <p class="font-mono">{{ acc.account_number }}</p>
          </div>
        </div>
        <input v-model="proofRef" type="text" class="input w-full text-sm" :placeholder="t('voucher.paymentRefPlaceholder')" />
        <input type="file" accept="image/jpeg,image/png,application/pdf" @change="onProofFile" />
        <p v-if="proofError" class="text-xs text-red-600">{{ proofError }}</p>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="proofOrder = null">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="submittingProof" @click="submitProof">
            {{ submittingProof ? t('common.processing') : t('voucher.submitProof') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Detail modal -->
    <div v-if="detailOrder" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="detailOrder = null">
      <div class="card max-h-[80vh] w-full max-w-2xl overflow-y-auto p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ detailOrder.order_no }}</h3>
        <p class="text-sm text-gray-500">{{ b2bStatusLabel(detailOrder.status) }} · KV #{{ detailOrder.kv_order_id }}</p>
        <div v-if="detailAudit.length" class="space-y-2">
          <p class="text-sm font-medium">{{ t('admin.voucher.auditTrail') }}</p>
          <div v-for="entry in detailAudit" :key="entry.id" class="rounded-lg bg-gray-50 p-2 text-xs dark:bg-dark-800">
            <span class="font-mono">{{ entry.action }}</span>
            <span class="mx-2 text-gray-400">{{ entry.operator }}</span>
            <span class="text-gray-500">{{ formatTime(entry.created_at) }}</span>
          </div>
        </div>
        <button type="button" class="btn btn-secondary" @click="detailOrder = null">{{ t('common.close') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { voucherAdminAPI, type VoucherAdminProduct, type VoucherB2BOrder, type VoucherAuditEntry } from '@/api/admin/voucher'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()

const loading = ref(true)
const syncing = ref(false)
const creating = ref(false)
const actingId = ref<number | null>(null)
const submittingProof = ref(false)
const products = ref<VoucherAdminProduct[]>([])
const orders = ref<VoucherB2BOrder[]>([])
const syncMessage = ref('')
const actionError = ref('')
const showCreate = ref(false)
const createError = ref('')
const merchantNotes = ref('')
const createLines = ref([{ product_id: 0, quantity: 1 }])
const catalogProducts = ref<VoucherAdminProduct[]>([])
const proofOrder = ref<VoucherB2BOrder | null>(null)
const proofRef = ref('')
const proofFile = ref<File | null>(null)
const proofError = ref('')
const detailOrder = ref<VoucherB2BOrder | null>(null)
const detailAudit = ref<VoucherAuditEntry[]>([])

function summarizeItems(items: VoucherB2BOrder['items']): string {
  if (!items?.length) return '—'
  return items.map((i) => `${i.name || i.product_id}×${i.quantity}`).join(', ')
}

function b2bStatusLabel(status: string): string {
  const key = `admin.voucher.b2bStatus.${status}`
  const translated = t(key)
  return translated === key ? status : translated
}

function formatTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

async function loadProducts() {
  const res = await voucherAdminAPI.listProducts()
  products.value = res.data.products || []
  catalogProducts.value = products.value.filter((p) => p.kv_product_id && p.is_active)
}

async function loadOrders() {
  const res = await voucherAdminAPI.listB2BOrders({ page: 1, per_page: 50 })
  orders.value = res.data.orders || []
}

async function loadAll() {
  loading.value = true
  actionError.value = ''
  try {
    await Promise.all([loadProducts(), loadOrders()])
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    loading.value = false
  }
}

async function syncCatalog() {
  syncing.value = true
  syncMessage.value = ''
  actionError.value = ''
  try {
    const res = await voucherAdminAPI.syncCatalog()
    syncMessage.value = t('admin.settings.voucher.syncCatalogDone', { count: res.data.synced })
    await loadProducts()
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    syncing.value = false
  }
}

async function syncStock() {
  syncing.value = true
  syncMessage.value = ''
  actionError.value = ''
  try {
    const res = await voucherAdminAPI.syncStock()
    syncMessage.value = t('admin.settings.voucher.syncStockDone', { count: res.data.updated })
    await loadProducts()
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    syncing.value = false
  }
}

function openCreateModal() {
  createError.value = ''
  merchantNotes.value = ''
  createLines.value = [{ product_id: 0, quantity: 1 }]
  showCreate.value = true
}

function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) return crypto.randomUUID()
  return `b2b-${Date.now()}`
}

async function submitCreate() {
  const items = createLines.value.filter((l) => l.product_id > 0 && l.quantity > 0)
  if (!items.length) {
    createError.value = t('admin.voucher.selectProduct')
    return
  }
  creating.value = true
  createError.value = ''
  try {
    await voucherAdminAPI.createB2BOrder({
      items,
      merchant_notes: merchantNotes.value.trim() || undefined,
      idempotency_key: newIdempotencyKey(),
    })
    showCreate.value = false
    syncMessage.value = t('admin.voucher.b2bOrderCreated')
    await loadOrders()
  } catch (err: unknown) {
    createError.value = extractApiErrorMessage(err)
  } finally {
    creating.value = false
  }
}

function openProofModal(order: VoucherB2BOrder) {
  proofOrder.value = order
  proofRef.value = ''
  proofFile.value = null
  proofError.value = ''
}

function onProofFile(e: Event) {
  const input = e.target as HTMLInputElement
  proofFile.value = input.files?.[0] ?? null
}

async function submitProof() {
  if (!proofOrder.value) return
  if (!proofRef.value.trim() && !proofFile.value) {
    proofError.value = t('voucher.proofRequired')
    return
  }
  submittingProof.value = true
  proofError.value = ''
  try {
    const form = new FormData()
    if (proofRef.value.trim()) form.append('payment_ref', proofRef.value.trim())
    if (proofFile.value) form.append('payment_proof', proofFile.value)
    await voucherAdminAPI.submitB2BProof(proofOrder.value.id, form)
    proofOrder.value = null
    syncMessage.value = t('admin.voucher.b2bProofSubmitted')
    await loadOrders()
  } catch (err: unknown) {
    proofError.value = extractApiErrorMessage(err)
  } finally {
    submittingProof.value = false
  }
}

async function syncOrder(id: number) {
  actingId.value = id
  actionError.value = ''
  try {
    await voucherAdminAPI.syncB2BOrder(id)
    syncMessage.value = t('admin.voucher.b2bOrderSynced')
    await loadAll()
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  } finally {
    actingId.value = null
  }
}

async function openDetail(id: number) {
  try {
    const res = await voucherAdminAPI.getB2BOrder(id)
    detailOrder.value = res.data.order
    detailAudit.value = res.data.audit || []
  } catch (err: unknown) {
    actionError.value = extractApiErrorMessage(err)
  }
}

onMounted(() => {
  void loadAll()
})
</script>
