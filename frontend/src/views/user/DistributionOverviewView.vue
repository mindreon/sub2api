<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('distribution.overview.title') }}</h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.overview.description') }}</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-md" @click="router.push('/distribution')">
            <Icon name="chevronLeft" size="md" />
            {{ t('distribution.actions.viewDistributionCenter') }}
          </button>
          <button type="button" class="btn btn-secondary btn-md" @click="openSettingsDialog">
            <Icon name="cog" size="md" />
            {{ t('distribution.actions.manageChannel') }}
          </button>
          <button type="button" class="btn btn-secondary btn-md" :disabled="loading" :title="t('common.refresh')" @click="reloadOverview">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ t('distribution.stats.channelOrgId') }}</div>
          <div class="mt-2 font-mono text-2xl font-semibold text-gray-900 dark:text-white">#{{ overview?.channel_org_id || '-' }}</div>
        </div>
        <div v-for="stat in stats" :key="stat.key" class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ stat.label }}</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ stat.value }}</div>
        </div>
      </div>

      <div v-if="channelWarnings.length" class="grid gap-3">
        <div
          v-for="item in channelWarnings"
          :key="item.key"
          class="rounded-lg border px-4 py-3"
          :class="item.tone === 'amber'
            ? 'border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-100'
            : 'border-red-200 bg-red-50 text-red-900 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-100'"
        >
          <div class="text-sm font-semibold">{{ item.title }}</div>
          <div class="mt-1 text-sm opacity-90">{{ item.description }}</div>
        </div>
      </div>

      <TablePageLayout>
        <template #filters>
          <div class="flex w-full flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <DateRangePicker
              :start-date="analyticsStartDate"
              :end-date="analyticsEndDate"
              @update:startDate="analyticsStartDate = $event"
              @update:endDate="analyticsEndDate = $event"
              @change="reloadAnalytics"
            />
            <div class="flex flex-wrap items-center gap-2">
              <select v-model="analyticsGranularity" class="input w-full sm:w-36" @change="reloadAnalytics">
                <option value="day">{{ t('distribution.analytics.granularity.day') }}</option>
                <option value="week">{{ t('distribution.analytics.granularity.week') }}</option>
                <option value="month">{{ t('distribution.analytics.granularity.month') }}</option>
              </select>
              <select v-model.number="analyticsRankingLimit" class="input w-full sm:w-36" @change="reloadAnalytics">
                <option :value="5">{{ t('distribution.analytics.topRanking', { count: 5 }) }}</option>
                <option :value="10">{{ t('distribution.analytics.topRanking', { count: 10 }) }}</option>
                <option :value="20">{{ t('distribution.analytics.topRanking', { count: 20 }) }}</option>
              </select>
            </div>
          </div>
        </template>

        <template #table>
          <div class="space-y-6 p-5">
            <section class="space-y-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('distribution.analytics.channelTitle') }}</h2>
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.channelDescription') }}</p>
                </div>
              </div>

              <div v-if="analytics?.channel" class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
                <div v-for="item in analyticsChannelCards" :key="item.key" class="rounded-lg border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                  <div class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ item.label }}</div>
                  <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ item.value }}</div>
                </div>
              </div>
              <div v-else class="rounded-lg border border-dashed border-gray-200 bg-gray-50/70 p-6 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800/60 dark:text-dark-400">
                {{ t('distribution.analytics.channelUnavailable') }}
              </div>

              <DistributionAnalyticsTrendChart
                v-if="analytics?.channel"
                :trend-data="analytics.channel.trend"
                :loading="loading"
              />

              <div class="rounded-lg border border-gray-200 dark:border-dark-700">
                <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-700">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('distribution.analytics.memberRankingTitle') }}</h3>
                </div>
                <div v-if="analytics?.channel?.member_ranking?.length" class="overflow-x-auto">
                  <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
                    <thead class="bg-gray-50 dark:bg-dark-800">
                      <tr>
                        <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.columns.memberId') }}</th>
                        <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.columns.user') }}</th>
                        <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.columns.role') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.registeredUsers') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.rechargeAmount') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.consumptionAmount') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.commissionAmount') }}</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
                      <tr v-for="item in analytics.channel.member_ranking" :key="item.member_id">
                        <td class="px-4 py-3 font-mono text-sm text-gray-700 dark:text-dark-300">#{{ item.member_id }}</td>
                        <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-300">
                          <div class="font-medium text-gray-900 dark:text-white">#{{ item.user_id }} {{ item.user_email || '-' }}</div>
                          <div class="text-xs text-gray-500 dark:text-dark-400">{{ item.username || '-' }}</div>
                        </td>
                        <td class="px-4 py-3"><StatusPill :label="roleLabel(item.role_type)" tone="blue" /></td>
                        <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">{{ item.registered_users }}</td>
                        <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">${{ formatAmount(item.recharge_amount) }}</td>
                        <td class="px-4 py-3 text-right text-sm font-medium text-gray-900 dark:text-white">${{ formatAmount(item.consumption_amount) }}</td>
                        <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">${{ formatAmount(item.commission_amount) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div v-else class="p-6 text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.emptyRanking') }}</div>
              </div>
            </section>

            <section class="space-y-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('distribution.analytics.personalTitle') }}</h2>
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.personalDescription') }}</p>
                </div>
                <div v-if="analytics?.personal?.role_types?.length" class="flex flex-wrap gap-2">
                  <StatusPill v-for="role in analytics.personal.role_types" :key="role" :label="roleLabel(role)" tone="blue" />
                </div>
              </div>

              <div v-if="analytics?.personal" class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
                <div v-for="item in analyticsPersonalCards" :key="item.key" class="rounded-lg border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                  <div class="text-xs font-medium uppercase text-gray-500 dark:text-dark-400">{{ item.label }}</div>
                  <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ item.value }}</div>
                </div>
              </div>
              <div v-else class="rounded-lg border border-dashed border-gray-200 bg-gray-50/70 p-6 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800/60 dark:text-dark-400">
                {{ t('distribution.analytics.personalUnavailable') }}
              </div>

              <div class="rounded-lg border border-gray-200 dark:border-dark-700">
                <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-700">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('distribution.analytics.childRankingTitle') }}</h3>
                </div>
                <div v-if="analytics?.personal?.child_member_ranking?.length" class="overflow-x-auto">
                  <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
                    <thead class="bg-gray-50 dark:bg-dark-800">
                      <tr>
                        <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.columns.memberId') }}</th>
                        <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.columns.user') }}</th>
                        <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.columns.role') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.registeredUsers') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.consumptionAmount') }}</th>
                        <th class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.metrics.commissionAmount') }}</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
                      <tr v-for="item in analytics.personal.child_member_ranking" :key="item.member_id">
                        <td class="px-4 py-3 font-mono text-sm text-gray-700 dark:text-dark-300">#{{ item.member_id }}</td>
                        <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-300">
                          <div class="font-medium text-gray-900 dark:text-white">#{{ item.user_id }} {{ item.user_email || '-' }}</div>
                          <div class="text-xs text-gray-500 dark:text-dark-400">{{ item.username || '-' }}</div>
                        </td>
                        <td class="px-4 py-3"><StatusPill :label="roleLabel(item.role_type)" tone="blue" /></td>
                        <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">{{ item.registered_users }}</td>
                        <td class="px-4 py-3 text-right text-sm font-medium text-gray-900 dark:text-white">${{ formatAmount(item.consumption_amount) }}</td>
                        <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">${{ formatAmount(item.commission_amount) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div v-else class="p-6 text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.analytics.emptyChildRanking') }}</div>
              </div>
            </section>
          </div>
        </template>
      </TablePageLayout>
    </div>

    <BaseDialog :show="settingsDialog" :title="t('distribution.dialogs.settingsTitle')" width="normal" @close="settingsDialog = false">
      <form class="space-y-4" @submit.prevent="submitSettings">
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.channelName') }}</span>
          <input v-model.trim="settingsForm.name" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.commissionSettlementMethod') }}</span>
          <select v-model="settingsForm.commission_settlement_method" class="input mt-1">
            <option value="balance">{{ t('distribution.settlementMethods.balance') }}</option>
            <option value="auto">{{ t('distribution.settlementMethods.auto') }}</option>
            <option value="manual">{{ t('distribution.settlementMethods.manual') }}</option>
            <option value="offline">{{ t('distribution.settlementMethods.offline') }}</option>
          </select>
        </label>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.levelsJson') }}</span>
          <textarea v-model="settingsForm.distribution_levels_json" class="input mt-1 min-h-[180px] font-mono text-xs" spellcheck="false" />
        </label>
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.wholesaleDiscountRate') }}</span>
            <input :value="formatRate(settingsForm.wholesale_discount_rate)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.refundFeeRate') }}</span>
            <input :value="formatRate(settingsForm.refund_fee_rate)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.firstRechargeMinAmount') }}</span>
            <input :value="formatMoney(settingsForm.first_recharge_min_amount)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.rechargeMinAmount') }}</span>
            <input :value="formatMoney(settingsForm.recharge_min_amount)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.warningThreshold') }}</span>
            <input :value="formatMoney(settingsForm.warning_threshold)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.consumptionLimit') }}</span>
            <input :value="formatMoney(settingsForm.consumption_limit)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.consumptionWarningThreshold') }}</span>
            <input :value="formatMoney(settingsForm.consumption_warning_threshold)" class="input mt-1" disabled />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.rechargeLeadTimeDays') }}</span>
            <input :value="String(settingsForm.recharge_lead_time_days)" class="input mt-1" disabled />
          </label>
          <label class="block sm:col-span-2">
            <span class="input-label">{{ t('distribution.fields.rechargeDeadlineNote') }}</span>
            <textarea :value="settingsForm.recharge_deadline_note" class="input mt-1 min-h-[120px]" disabled />
          </label>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.logoUrl') }}</span>
            <input v-model.trim="settingsForm.logo_url" class="input mt-1" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.primaryColor') }}</span>
            <input v-model.trim="settingsForm.primary_color" class="input mt-1" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.domain') }}</span>
            <input v-model.trim="settingsForm.domain" class="input mt-1" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.apiDomain') }}</span>
            <input v-model.trim="settingsForm.api_domain" class="input mt-1" />
          </label>
        </div>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="settingsDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </form>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import DistributionAnalyticsTrendChart from '@/components/charts/DistributionAnalyticsTrendChart.vue'
import Icon from '@/components/icons/Icon.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import {
  getDistributionOverview,
  getMyDistributionAnalytics,
  updateMyDistributionOrganization,
  type DistributionOverview,
  type MyDistributionAnalyticsResponse,
} from '@/api/distribution'
import type { DistributionMemberRole } from '@/api/admin/distribution'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const overview = ref<DistributionOverview | null>(null)
const analytics = ref<MyDistributionAnalyticsResponse | null>(null)
const loading = ref(false)
const saving = ref(false)
const settingsDialog = ref(false)
const analyticsStartDate = ref(buildRelativeDateString(6))
const analyticsEndDate = ref(buildRelativeDateString(0))
const analyticsGranularity = ref<'day' | 'week' | 'month'>('day')
const analyticsRankingLimit = ref(10)
const settingsForm = reactive({
  name: '',
  commission_settlement_method: 'manual',
  distribution_levels_json: '[]',
  wholesale_discount_rate: 0.5,
  refund_fee_rate: 0,
  first_recharge_min_amount: 0,
  recharge_min_amount: 0,
  warning_threshold: 0,
  consumption_limit: 0,
  consumption_warning_threshold: 0,
  recharge_lead_time_days: 0,
  recharge_deadline_note: '',
  logo_url: '',
  primary_color: '',
  domain: '',
  api_domain: '',
})

const summary = computed(() => overview.value?.summary ?? null)

const stats = computed(() => [
  { key: 'walletBalance', label: t('distribution.stats.walletBalance'), value: formatMoney(summary.value?.wallet?.prepaid_balance ?? 0) },
  { key: 'walletReserved', label: t('distribution.stats.walletReserved'), value: formatMoney(summary.value?.wallet?.commission_reserved ?? 0) },
  { key: 'walletRecharged', label: t('distribution.stats.totalRecharged'), value: formatMoney(summary.value?.wallet?.total_recharged ?? 0) },
  { key: 'walletConsumed', label: t('distribution.stats.totalConsumed'), value: formatMoney(summary.value?.wallet?.total_consumed ?? 0) },
  { key: 'members', label: t('distribution.stats.members'), value: String(summary.value?.member_count ?? 0) },
  { key: 'promotionLinks', label: t('distribution.stats.promotionLinks'), value: String(summary.value?.promotion_link_count ?? 0) },
])

const analyticsChannelCards = computed(() => {
  const channelSummary = analytics.value?.channel?.summary
  if (!channelSummary) return []
  return [
    { key: 'registeredUsers', label: t('distribution.analytics.metrics.registeredUsers'), value: String(channelSummary.registered_users ?? 0) },
    { key: 'rechargeAmount', label: t('distribution.analytics.metrics.rechargeAmount'), value: formatMoney(channelSummary.recharge_amount) },
    { key: 'consumptionAmount', label: t('distribution.analytics.metrics.consumptionAmount'), value: formatMoney(channelSummary.consumption_amount) },
    { key: 'commissionAmount', label: t('distribution.analytics.metrics.commissionAmount'), value: formatMoney(channelSummary.commission_amount) },
    { key: 'settledCommissionAmount', label: t('distribution.analytics.metrics.settledCommissionAmount'), value: formatMoney(channelSummary.settled_commission_amount) },
    { key: 'commissionExpenseRatio', label: t('distribution.analytics.metrics.commissionExpenseRatio'), value: formatRate(channelSummary.commission_expense_ratio) },
    { key: 'commissionUpperRatio', label: t('distribution.analytics.metrics.commissionUpperRatio'), value: formatRate(channelSummary.commission_upper_ratio) },
  ]
})

const analyticsPersonalCards = computed(() => {
  const personalSummary = analytics.value?.personal?.summary
  if (!personalSummary) return []
  return [
    { key: 'registeredUsers', label: t('distribution.analytics.metrics.registeredUsers'), value: String(personalSummary.registered_users ?? 0) },
    { key: 'rechargeAmount', label: t('distribution.analytics.metrics.rechargeAmount'), value: formatMoney(personalSummary.recharge_amount) },
    { key: 'consumptionAmount', label: t('distribution.analytics.metrics.consumptionAmount'), value: formatMoney(personalSummary.consumption_amount) },
    { key: 'commissionAmount', label: t('distribution.analytics.metrics.commissionAmount'), value: formatMoney(personalSummary.commission_amount) },
    { key: 'settledCommissionAmount', label: t('distribution.analytics.metrics.settledCommissionAmount'), value: formatMoney(personalSummary.settled_commission_amount) },
  ]
})

const channelWarnings = computed(() => {
  const wallet = summary.value?.wallet
  const organization = summary.value?.organization
  if (!wallet || !organization) return []

  const warnings: Array<{ key: string; tone: 'amber' | 'red'; title: string; description: string }> = []
  const availableBalance = Number(wallet.prepaid_balance ?? 0) - Number(wallet.commission_reserved ?? 0)
  const warningThreshold = Number(wallet.warning_threshold ?? 0)
  const consumptionLimit = Number(organization.config?.consumption_limit ?? 0)
  const consumptionWarningThreshold = Number(organization.config?.consumption_warning_threshold ?? 0)
  const remainingConsumption = consumptionLimit - Number(wallet.total_consumed ?? 0)

  if (wallet.status && wallet.status !== 'active') {
    let description = t('distribution.warnings.suspendedGenericDescription', {
      status: statusLabel(wallet.status),
      note: String(organization.config?.recharge_deadline_note ?? ''),
    })
    if (consumptionLimit > 0 && remainingConsumption <= 0) {
      description = t('distribution.warnings.suspendedByConsumptionDescription', {
        limit: formatMoney(consumptionLimit),
        consumed: formatMoney(wallet.total_consumed),
        note: String(organization.config?.recharge_deadline_note ?? ''),
      })
    } else if (availableBalance <= 0) {
      description = t('distribution.warnings.suspendedByBalanceDescription', {
        balance: formatMoney(wallet.prepaid_balance),
        reserved: formatMoney(wallet.commission_reserved),
        available: formatMoney(availableBalance),
        note: String(organization.config?.recharge_deadline_note ?? ''),
      })
    }
    warnings.push({
      key: 'service-suspended',
      tone: 'red',
      title: t('distribution.warnings.suspendedTitle'),
      description,
    })
  }

  if (warningThreshold > 0 && Number(wallet.prepaid_balance ?? 0) <= warningThreshold) {
    warnings.push({
      key: 'low-balance',
      tone: 'amber',
      title: t('distribution.warnings.lowBalanceTitle'),
      description: t('distribution.warnings.lowBalanceDescription', {
        balance: formatMoney(wallet.prepaid_balance),
        threshold: formatMoney(warningThreshold),
        leadDays: Number(organization.config?.recharge_lead_time_days ?? 0),
        note: String(organization.config?.recharge_deadline_note ?? ''),
      }),
    })
  }
  if (consumptionLimit > 0 && consumptionWarningThreshold > 0 && remainingConsumption <= consumptionWarningThreshold) {
    warnings.push({
      key: 'consumption-limit',
      tone: remainingConsumption <= 0 ? 'red' : 'amber',
      title: t('distribution.warnings.consumptionTitle'),
      description: t('distribution.warnings.consumptionDescription', {
        remaining: formatMoney(remainingConsumption),
        limit: formatMoney(consumptionLimit),
        threshold: formatMoney(consumptionWarningThreshold),
        note: String(organization.config?.recharge_deadline_note ?? ''),
      }),
    })
  }

  return warnings
})

async function loadOverview() {
  const result = await getDistributionOverview()
  if (!result.can_manage_channel) {
    router.replace('/distribution')
    return false
  }
  overview.value = result
  return true
}

async function loadAnalytics() {
  analytics.value = await getMyDistributionAnalytics({
    start_date: analyticsStartDate.value,
    end_date: analyticsEndDate.value,
    granularity: analyticsGranularity.value,
    limit: analyticsRankingLimit.value,
  })
}

async function reloadOverview() {
  loading.value = true
  try {
    const canView = await loadOverview()
    if (!canView) return
    await loadAnalytics()
  } catch (error: any) {
    if (error?.reason === 'DISTRIBUTION_ATTRIBUTION_NOT_FOUND' || error?.status === 404) {
      router.replace('/distribution')
      return
    }
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

function reloadAnalytics() {
  void reloadOverview()
}

function openSettingsDialog() {
  const organization = summary.value?.organization
  const wallet = summary.value?.wallet
  settingsForm.name = organization?.name || ''
  settingsForm.commission_settlement_method = String(organization?.config?.commission_settlement_method ?? 'manual')
  settingsForm.distribution_levels_json = JSON.stringify(organization?.config?.distribution_levels ?? [], null, 2)
  settingsForm.wholesale_discount_rate = Number(organization?.config?.wholesale_discount_rate ?? 0.5)
  settingsForm.refund_fee_rate = Number(organization?.config?.refund_fee_rate ?? 0)
  settingsForm.first_recharge_min_amount = Number(organization?.config?.first_recharge_min_amount ?? 0)
  settingsForm.recharge_min_amount = Number(organization?.config?.recharge_min_amount ?? 0)
  settingsForm.warning_threshold = Number(wallet?.warning_threshold ?? 0)
  settingsForm.consumption_limit = Number(organization?.config?.consumption_limit ?? 0)
  settingsForm.consumption_warning_threshold = Number(organization?.config?.consumption_warning_threshold ?? 0)
  settingsForm.recharge_lead_time_days = Number(organization?.config?.recharge_lead_time_days ?? 0)
  settingsForm.recharge_deadline_note = String(organization?.config?.recharge_deadline_note ?? '')
  settingsForm.logo_url = String(organization?.brand_config?.logo_url ?? '')
  settingsForm.primary_color = String(organization?.brand_config?.primary_color ?? organization?.brand_config?.theme_color ?? '')
  settingsForm.domain = String(organization?.brand_config?.domain ?? '')
  settingsForm.api_domain = String(organization?.brand_config?.api_domain ?? '')
  settingsDialog.value = true
}

async function submitSettings() {
  saving.value = true
  try {
    let distributionLevels: Array<Record<string, unknown>> = []
    try {
      const parsed = JSON.parse(settingsForm.distribution_levels_json || '[]')
      distributionLevels = Array.isArray(parsed) ? parsed : []
    } catch {
      appStore.showError(t('distribution.errors.levelsFormatError'))
      return
    }
    await updateMyDistributionOrganization({
      name: settingsForm.name || undefined,
      config: {
        commission_settlement_method: settingsForm.commission_settlement_method,
        distribution_levels: distributionLevels,
      },
      brand_config: {
        logo_url: settingsForm.logo_url || '',
        primary_color: settingsForm.primary_color || '',
        domain: settingsForm.domain || '',
        api_domain: settingsForm.api_domain || '',
      },
    })
    appStore.showSuccess(t('distribution.messages.settingsUpdated'))
    settingsDialog.value = false
    await reloadOverview()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

function roleLabel(role: DistributionMemberRole | string) {
  return t(`distribution.roles.${role}`, role)
}

function statusLabel(status: string) {
  if (status === 'resolved') return t(`distribution.alertStatuses.${status}`, status)
  return t(`distribution.statuses.${status}`, status || '-')
}

function formatRate(value: number | null | undefined) {
  return `${(Number(value || 0) * 100).toFixed(2)}%`
}

function formatAmount(value: number | null | undefined) {
  return Number(value || 0).toFixed(4)
}

function formatMoney(value: number | null | undefined) {
  return formatAmount(value)
}

function buildRelativeDateString(daysAgo: number) {
  const date = new Date()
  date.setDate(date.getDate() - daysAgo)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

onMounted(() => {
  void reloadOverview()
})

const StatusPill = defineComponent({
  props: {
    label: { type: String, required: true },
    tone: { type: String, default: 'gray' },
  },
  setup(props) {
    const toneClass = computed(() => {
      if (props.tone === 'green') return 'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-500/10 dark:text-emerald-300'
      if (props.tone === 'amber') return 'bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-500/10 dark:text-amber-300'
      if (props.tone === 'blue') return 'bg-sky-50 text-sky-700 ring-sky-600/20 dark:bg-sky-500/10 dark:text-sky-300'
      return 'bg-gray-50 text-gray-700 ring-gray-600/20 dark:bg-dark-700 dark:text-dark-300'
    })
    return () => h('span', { class: ['inline-flex rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset', toneClass.value] }, props.label)
  },
})
</script>
