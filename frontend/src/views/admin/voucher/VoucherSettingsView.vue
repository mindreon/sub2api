<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.voucher.settingsPageTitle') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.voucher.description') }}</p>
      </div>
      <VoucherSettingsCard @saved="onSaved" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import VoucherSettingsCard from '@/components/admin/VoucherSettingsCard.vue'
import { useAdminSettingsStore } from '@/stores/adminSettings'

const { t } = useI18n()
const adminSettingsStore = useAdminSettingsStore()

function onSaved() {
  void adminSettingsStore.fetchVoucherNav(true)
}
</script>
