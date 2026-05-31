<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.voucher.replenishmentTitle') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.voucher.replenishmentDescription') }}</p>
      </div>
      <VoucherReplenishmentCard />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import VoucherReplenishmentCard from '@/components/admin/VoucherReplenishmentCard.vue'

const { t } = useI18n()
</script>
