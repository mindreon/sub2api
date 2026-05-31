<template>
  <div class="space-y-6">
    <!-- Step indicator -->
    <div class="card p-4">
      <div class="flex flex-wrap items-center gap-2">
        <template v-for="(step, index) in stepItems" :key="step.key">
          <div class="flex items-center gap-2">
            <div
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold transition-colors"
              :class="stepCircleClass(step.key)"
            >
              {{ index + 1 }}
            </div>
            <span class="text-xs font-medium" :class="stepLabelClass(step.key)">{{ step.label }}</span>
          </div>
          <div
            v-if="index < stepItems.length - 1"
            class="mx-1 hidden h-px w-6 bg-gray-200 sm:block dark:bg-dark-600"
          />
        </template>
      </div>
    </div>

    <!-- Account -->
    <div v-if="wizardStep === 'select'" class="card p-5">
      <p class="text-xs font-medium text-gray-400 dark:text-gray-500">{{ t('voucher.account') }}</p>
      <p class="mt-1 text-base font-semibold text-gray-900 dark:text-white">{{ username }}</p>
      <p v-if="config.helpText" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ config.helpText }}</p>
    </div>

    <!-- Step: Select -->
    <template v-if="wizardStep === 'select'">
      <div v-if="config.products.length === 0" class="card py-16 text-center space-y-2">
        <p class="text-gray-500 dark:text-gray-400">{{ t('voucher.noProducts') }}</p>
        <p v-if="!config.checkoutReady" class="text-sm text-amber-700 dark:text-amber-300">{{ t('voucher.setupIncomplete') }}</p>
      </div>
      <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <button
          v-for="product in config.products"
          :key="product.id"
          type="button"
          class="card group relative p-5 text-left transition-all hover:ring-2 hover:ring-primary-500/40"
          :class="selectedProduct?.id === product.id ? 'ring-2 ring-primary-500' : ''"
          :disabled="product.stock <= 0"
          @click="selectProduct(product)"
        >
          <div class="flex items-start justify-between gap-2">
            <div>
              <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ product.name }}</p>
              <p class="mt-1 text-2xl font-bold text-primary-600 dark:text-primary-400">
                {{ formatMoney(product.retail_price) }}
              </p>
            </div>
            <span
              class="rounded-full px-2 py-0.5 text-[10px] font-medium"
              :class="product.stock > 0 ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'"
            >
              {{ product.stock > 0 ? t('voucher.inStock', { count: product.stock }) : t('voucher.outOfStock') }}
            </span>
          </div>
          <p class="mt-3 text-xs text-gray-400 dark:text-gray-500">
            {{ t('voucher.denominationLabel', { amount: formatMoney(product.denomination, false) }) }}
          </p>
        </button>
      </div>

      <div v-if="selectedProduct" class="card p-6">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('voucher.selectedProduct') }}</p>
            <p class="text-base font-semibold text-gray-900 dark:text-white">{{ selectedProduct.name }}</p>
          </div>
          <div class="flex items-center gap-3">
            <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('voucher.quantity') }}</span>
            <div class="flex items-center rounded-lg border border-gray-200 dark:border-dark-600">
              <button type="button" class="px-3 py-2 text-gray-600 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-dark-700" :disabled="quantity <= 1" @click="quantity -= 1">−</button>
              <span class="min-w-[2rem] text-center text-sm font-semibold text-gray-900 dark:text-white">{{ quantity }}</span>
              <button type="button" class="px-3 py-2 text-gray-600 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-dark-700" :disabled="quantity >= maxQuantity" @click="quantity += 1">+</button>
            </div>
          </div>
        </div>
        <div class="mt-4 flex items-center justify-between border-t border-gray-100 pt-4 dark:border-dark-700">
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('voucher.subtotal') }}</span>
          <span class="text-lg font-bold text-gray-900 dark:text-white">{{ formatMoney(subtotal) }}</span>
        </div>
        <button class="btn btn-primary mt-5 w-full py-3" :disabled="!canProceedSelect" @click="wizardStep = 'confirm'">
          {{ t('voucher.nextConfirm') }}
        </button>
      </div>
    </template>

    <!-- Step: Confirm -->
    <template v-else-if="wizardStep === 'confirm' && selectedProduct">
      <div class="card p-6 space-y-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('voucher.confirmTitle') }}</h3>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('voucher.product') }}</span>
            <span class="text-gray-900 dark:text-white">{{ selectedProduct.name }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('voucher.quantity') }}</span>
            <span class="text-gray-900 dark:text-white">{{ quantity }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('voucher.unitPrice') }}</span>
            <span class="text-gray-900 dark:text-white">{{ formatMoney(selectedProduct.retail_price) }}</span>
          </div>
          <div v-if="feeAmount > 0" class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('voucher.fee', { rate: config.feeRate }) }}</span>
            <span class="text-gray-900 dark:text-white">{{ formatMoney(feeAmount) }}</span>
          </div>
          <div class="flex justify-between border-t border-gray-200 pt-2 dark:border-dark-600">
            <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('voucher.totalDue') }}</span>
            <span class="text-xl font-bold text-primary-600 dark:text-primary-400">{{ formatMoney(totalDue) }}</span>
          </div>
        </div>
        <p class="rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
          {{ t('voucher.bankTransferHint') }}
        </p>
        <p v-if="orderError" class="text-xs text-red-600 dark:text-red-400">{{ orderError }}</p>
        <div class="flex gap-3">
          <button class="btn btn-secondary flex-1" @click="wizardStep = 'select'">{{ t('common.back') }}</button>
          <button class="btn btn-primary flex-1" :disabled="creatingOrder" @click="createOrder">
            <span v-if="creatingOrder" class="flex items-center justify-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              {{ t('common.processing') }}
            </span>
            <span v-else>{{ t('voucher.createOrder') }}</span>
          </button>
        </div>
      </div>
    </template>

    <!-- Step: Pay -->
    <template v-else-if="wizardStep === 'pay' && order">
      <div class="card p-6 space-y-5">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-gray-500">{{ t('voucher.orderNo') }}</p>
            <p class="mt-1 font-mono text-lg font-bold text-gray-900 dark:text-white">{{ order.order_no }}</p>
          </div>
          <span class="badge badge-warning">{{ t('voucher.status.pending_payment') }}</span>
        </div>

        <div class="rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-800/40 dark:bg-primary-900/20">
          <p class="text-sm text-primary-800 dark:text-primary-200">{{ t('voucher.transferReferenceHint') }}</p>
          <div class="mt-3 flex flex-wrap items-center gap-2">
            <code class="rounded-lg bg-white px-3 py-1.5 font-mono text-sm text-gray-900 dark:bg-dark-800 dark:text-white">{{ order.order_no }}</code>
            <button type="button" class="btn btn-secondary btn-sm" @click="copyText(order.order_no)">{{ t('voucher.copy') }}</button>
          </div>
        </div>

        <div class="flex items-center justify-between rounded-xl border border-gray-200 p-4 dark:border-dark-600">
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('voucher.amountDue') }}</span>
          <span class="text-2xl font-bold text-gray-900 dark:text-white">{{ formatMoney(order.total_amount) }}</span>
        </div>

        <div class="space-y-3">
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('voucher.bankAccounts') }}</p>
          <div
            v-for="account in config.bankAccounts"
            :key="account.id"
            class="rounded-xl border border-gray-200 p-4 dark:border-dark-600"
          >
            <div class="flex flex-wrap items-start justify-between gap-2">
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">{{ account.bank_name }}</p>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ account.account_name }}</p>
                <p class="mt-2 font-mono text-base text-gray-900 dark:text-white">{{ account.account_number }}</p>
              </div>
              <button type="button" class="btn btn-secondary btn-sm" @click="copyBankAccount(account)">
                {{ t('voucher.copyAll') }}
              </button>
            </div>
          </div>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('voucher.expiresHint', { hours: config.orderTimeoutHours }) }}
        </p>

        <div class="flex gap-3">
          <button class="btn btn-secondary flex-1" @click="wizardStep = 'proof'">{{ t('voucher.uploadProof') }}</button>
          <button class="btn btn-primary flex-1" @click="wizardStep = 'proof'">{{ t('voucher.paidContinue') }}</button>
        </div>
      </div>
    </template>

    <!-- Step: Proof -->
    <template v-else-if="wizardStep === 'proof' && order">
      <div class="card p-6 space-y-5">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('voucher.proofTitle') }}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('voucher.proofDesc') }}</p>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('voucher.paymentRef') }}</label>
          <input
            v-model="paymentRef"
            type="text"
            class="input w-full"
            :placeholder="t('voucher.paymentRefPlaceholder')"
          />
        </div>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('voucher.paymentProof') }}</label>
          <label class="flex cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 px-4 py-8 transition-colors hover:border-primary-400 dark:border-dark-600 dark:hover:border-primary-500">
            <input type="file" class="hidden" accept="image/jpeg,image/png,application/pdf" @change="onProofSelected" />
            <Icon name="upload" size="lg" class="mb-2 text-gray-400" />
            <span class="text-sm text-gray-600 dark:text-gray-300">{{ t('voucher.uploadHint') }}</span>
            <span v-if="proofFileName" class="mt-2 text-xs font-medium text-primary-600 dark:text-primary-400">{{ proofFileName }}</span>
          </label>
        </div>

        <div v-if="config.bankAccounts.length > 1">
          <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('voucher.paidToBank') }}</label>
          <select v-model="selectedBankId" class="input w-full">
            <option v-for="account in config.bankAccounts" :key="account.id" :value="account.id">
              {{ account.bank_name }} · {{ account.account_number }}
            </option>
          </select>
        </div>

        <p v-if="proofError" class="text-xs text-red-600 dark:text-red-400">{{ proofError }}</p>

        <div class="flex gap-3">
          <button class="btn btn-secondary flex-1" @click="wizardStep = 'pay'">{{ t('common.back') }}</button>
          <button class="btn btn-primary flex-1" :disabled="submittingProof" @click="submitProof">
            <span v-if="submittingProof" class="flex items-center justify-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              {{ t('common.processing') }}
            </span>
            <span v-else>{{ t('voucher.submitProof') }}</span>
          </button>
        </div>
      </div>
    </template>

    <!-- Step: Waiting -->
    <template v-else-if="wizardStep === 'waiting' && order">
      <div class="card p-6 space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-xs text-gray-400 dark:text-gray-500">{{ t('voucher.orderNo') }}</p>
            <p class="font-mono font-semibold text-gray-900 dark:text-white">{{ order.order_no }}</p>
          </div>
          <span class="badge badge-info">{{ statusBadgeLabel }}</span>
        </div>

        <VoucherOrderTimeline :order="order" />

        <div class="rounded-xl bg-blue-50 p-4 text-sm text-blue-900 dark:bg-blue-900/20 dark:text-blue-100">
          {{ t('voucher.reviewSla', { hours: config.reviewSlaHours }) }}
        </div>

        <div v-if="order.status === 'rejected' && order.reject_reason" class="rounded-xl bg-red-50 p-4 text-sm text-red-900 dark:bg-red-900/20 dark:text-red-100">
          {{ order.reject_reason }}
        </div>

        <button class="btn btn-primary w-full" @click="resetFlow">{{ t('voucher.newOrder') }}</button>
      </div>
    </template>

    <!-- Step: Completed -->
    <template v-else-if="wizardStep === 'completed' && order">
      <div class="card p-6 space-y-5">
        <div class="text-center">
          <div class="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
            <Icon name="check" size="lg" class="text-green-600 dark:text-green-400" />
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('voucher.completedTitle') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('voucher.completedDesc') }}</p>
        </div>

        <VoucherOrderTimeline :order="order" />

        <div class="space-y-3">
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('voucher.yourPins') }}</p>
          <div
            v-for="(pin, idx) in (order.pins || [])"
            :key="idx"
            class="rounded-xl border border-gray-200 p-4 dark:border-dark-600"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <code class="font-mono text-base font-semibold text-gray-900 dark:text-white">{{ pin.pin_code }}</code>
              <button type="button" class="btn btn-secondary btn-sm" @click="copyText(pin.pin_code)">{{ t('voucher.copy') }}</button>
            </div>
            <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
              <span>{{ t('voucher.serial') }}: {{ pin.serial }}</span>
              <span>{{ t('voucher.expires') }}: {{ pin.expires_at }}</span>
            </div>
          </div>
        </div>

        <p class="text-xs text-amber-700 dark:text-amber-300">{{ t('voucher.pinSecurityHint') }}</p>
        <button class="btn btn-primary w-full" @click="resetFlow">{{ t('voucher.buyAgain') }}</button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import VoucherOrderTimeline from '@/components/voucher/VoucherOrderTimeline.vue'
import { formatPaymentAmount } from '@/components/payment/currency'
import { voucherAPI } from '@/api/voucher'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { VoucherCheckoutConfig, VoucherOrder, VoucherProduct, VoucherWizardStep } from '@/types/voucher'

const props = defineProps<{
  config: VoucherCheckoutConfig
  username: string
}>()

const { t, locale } = useI18n()

const wizardStep = ref<VoucherWizardStep>('select')
const selectedProduct = ref<VoucherProduct | null>(null)
const quantity = ref(1)
const order = ref<VoucherOrder | null>(null)
const paymentRef = ref('')
const proofFileName = ref('')
const proofFile = ref<File | null>(null)
const proofError = ref('')
const orderError = ref('')
const submittingProof = ref(false)
const creatingOrder = ref(false)
const selectedBankId = ref(props.config.bankAccounts[0]?.id ?? 1)

let pollTimer: ReturnType<typeof setInterval> | null = null

const stepItems = computed((): { key: VoucherWizardStep; label: string }[] => [
  { key: 'select', label: t('voucher.steps.select') },
  { key: 'confirm', label: t('voucher.steps.confirm') },
  { key: 'pay', label: t('voucher.steps.pay') },
  { key: 'proof', label: t('voucher.steps.proof') },
  { key: 'waiting', label: t('voucher.steps.waiting') },
  { key: 'completed', label: t('voucher.steps.completed') },
])

const maxQuantity = computed(() => {
  if (!selectedProduct.value) return 1
  return Math.min(props.config.maxQuantityPerOrder, selectedProduct.value.stock)
})

const subtotal = computed(() => {
  if (!selectedProduct.value) return 0
  return Math.round(selectedProduct.value.retail_price * quantity.value * 100) / 100
})

const feeAmount = computed(() => {
  if (props.config.feeRate <= 0) return 0
  return Math.ceil((subtotal.value * props.config.feeRate) / 100 * 100) / 100
})

const totalDue = computed(() => Math.round((subtotal.value + feeAmount.value) * 100) / 100)

const canProceedSelect = computed(() =>
  !!selectedProduct.value
  && selectedProduct.value.stock >= quantity.value
  && quantity.value >= 1,
)

const statusBadgeLabel = computed(() => {
  if (!order.value) return ''
  return t(`voucher.status.${order.value.status}`)
})

function formatMoney(value: number, withSymbol = true): string {
  if (!withSymbol) return value.toFixed(2)
  const localeCode = typeof locale === 'string' ? locale : (locale as { value?: string }).value
  return formatPaymentAmount(value, props.config.currency, localeCode)
}

function stepIndex(step: VoucherWizardStep): number {
  return stepItems.value.findIndex((s) => s.key === step)
}

function stepCircleClass(step: VoucherWizardStep): string {
  const current = stepIndex(wizardStep.value)
  const idx = stepIndex(step)
  if (idx < current) return 'bg-primary-500 text-white'
  if (idx === current) return 'bg-primary-100 text-primary-700 ring-2 ring-primary-500 dark:bg-primary-900/40 dark:text-primary-200'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}

function stepLabelClass(step: VoucherWizardStep): string {
  const current = stepIndex(wizardStep.value)
  const idx = stepIndex(step)
  if (idx <= current) return 'text-gray-900 dark:text-white'
  return 'text-gray-400 dark:text-gray-500'
}

function selectProduct(product: VoucherProduct) {
  if (product.stock <= 0) return
  selectedProduct.value = product
  quantity.value = 1
}

function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) return crypto.randomUUID()
  return `vc-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

async function createOrder() {
  if (!selectedProduct.value) return
  creatingOrder.value = true
  orderError.value = ''
  try {
    const res = await voucherAPI.createOrder({
      product_id: selectedProduct.value.id,
      quantity: quantity.value,
      idempotency_key: newIdempotencyKey(),
    })
    order.value = res.data.order
    wizardStep.value = 'pay'
  } catch (err: unknown) {
    orderError.value = extractApiErrorMessage(err)
  } finally {
    creatingOrder.value = false
  }
}

function onProofSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  proofFile.value = file
  proofFileName.value = file?.name || ''
  proofError.value = ''
}

async function submitProof() {
  if (!order.value) return
  if (!paymentRef.value.trim() && !proofFile.value) {
    proofError.value = t('voucher.proofRequired')
    return
  }
  submittingProof.value = true
  proofError.value = ''
  try {
    const form = new FormData()
    if (paymentRef.value.trim()) form.append('payment_ref', paymentRef.value.trim())
    if (props.config.bankAccounts.length > 1) form.append('bank_id', String(selectedBankId.value))
    if (proofFile.value) form.append('payment_proof', proofFile.value)
    const res = await voucherAPI.submitPaymentProof(order.value.id, form)
    order.value = res.data.order
    wizardStep.value = 'waiting'
    startPolling()
  } catch (err: unknown) {
    proofError.value = extractApiErrorMessage(err)
  } finally {
    submittingProof.value = false
  }
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function refreshOrderStatus() {
  if (!order.value) return
  try {
    const includePins = order.value.status === 'fulfilling' || order.value.status === 'completed'
    const res = await voucherAPI.getOrder(order.value.id, includePins)
    order.value = res.data.order
    if (order.value.status === 'completed') {
      stopPolling()
      wizardStep.value = 'completed'
    } else if (order.value.status === 'rejected' || order.value.status === 'expired') {
      stopPolling()
    }
  } catch {
    // ignore transient poll errors
  }
}

function startPolling() {
  stopPolling()
  void refreshOrderStatus()
  pollTimer = setInterval(() => { void refreshOrderStatus() }, 5000)
}

watch(wizardStep, (step) => {
  if (step === 'waiting' && order.value) startPolling()
  else if (step !== 'waiting') stopPolling()
})

onBeforeUnmount(() => stopPolling())

function resetFlow() {
  stopPolling()
  wizardStep.value = 'select'
  selectedProduct.value = null
  quantity.value = 1
  order.value = null
  paymentRef.value = ''
  proofFileName.value = ''
  proofFile.value = null
  proofError.value = ''
  orderError.value = ''
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    // ignore clipboard errors in unsupported contexts
  }
}

function copyBankAccount(account: { bank_name: string; account_name: string; account_number: string }) {
  const text = `${account.bank_name}\n${account.account_name}\n${account.account_number}`
  void copyText(text)
}
</script>
