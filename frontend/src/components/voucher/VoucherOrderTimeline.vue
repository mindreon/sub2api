<template>
  <ol class="space-y-4">
    <li v-for="item in timeline" :key="item.status" class="flex gap-3">
      <div class="relative flex flex-col items-center">
        <div
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[10px] font-bold"
          :class="item.done ? 'bg-primary-500 text-white' : item.active ? 'bg-primary-100 text-primary-700 ring-2 ring-primary-400 dark:bg-primary-900/40 dark:text-primary-200' : 'bg-gray-100 text-gray-400 dark:bg-dark-700'"
        >
          <Icon v-if="item.done" name="check" size="xs" />
          <span v-else>{{ item.index }}</span>
        </div>
        <div v-if="!item.last" class="mt-1 w-px flex-1 bg-gray-200 dark:bg-dark-600" />
      </div>
      <div class="pb-4">
        <p class="text-sm font-medium" :class="item.done || item.active ? 'text-gray-900 dark:text-white' : 'text-gray-400'">
          {{ t(`voucher.status.${item.status}`) }}
        </p>
        <p v-if="item.at" class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ formatTime(item.at) }}</p>
        <p v-else-if="item.pendingHint" class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">{{ item.pendingHint }}</p>
      </div>
    </li>
  </ol>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { VoucherOrder, VoucherOrderStatus } from '@/types/voucher'

const props = defineProps<{
  order: VoucherOrder
}>()

const { t } = useI18n()

const STATUS_FLOW: VoucherOrderStatus[] = [
  'pending_payment',
  'payment_submitted',
  'payment_verified',
  'fulfilling',
  'completed',
]

const statusRank: Record<VoucherOrderStatus, number> = {
  pending_payment: 0,
  payment_submitted: 1,
  payment_verified: 2,
  fulfilling: 3,
  completed: 4,
  rejected: -1,
  expired: -1,
}

const timeline = computed(() => {
  const current = statusRank[props.order.status]
  return STATUS_FLOW.map((status, index) => {
    const done = current > index || (current === index && status === 'completed')
    const active = current === index && status !== 'completed'
    let at = ''
    if (status === 'pending_payment') at = props.order.created_at
    if (status === 'payment_submitted' && current >= 1) at = props.order.created_at
    if (status === 'completed' && props.order.status === 'completed') at = props.order.completed_at || props.order.created_at
    return {
      status,
      index: index + 1,
      done: done || props.order.status === 'completed',
      active,
      at: done || active ? at : '',
      pendingHint: !done && !active && status === 'payment_verified' ? t('voucher.timelinePendingReview') : '',
      last: index === STATUS_FLOW.length - 1,
    }
  })
})

function formatTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}
</script>
