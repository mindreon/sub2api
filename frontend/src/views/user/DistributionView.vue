<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('distribution.title') }}</h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.description') }}</p>
        </div>
        <div class="flex gap-2">
          <button
            v-if="activeTab === 'commissions' && overview?.can_manage_channel && selectedCommissionIds.length > 0"
            type="button"
            class="btn btn-primary btn-md"
            @click="openSettleDialog()"
          >
            <Icon name="check" size="md" />
            {{ t('distribution.actions.batchSettleCommission', { count: selectedCommissionIds.length }) }}
          </button>
          <button
            v-if="activeTab === 'wallet-requests' && overview?.can_manage_channel"
            type="button"
            class="btn btn-primary btn-md"
            @click="openWalletRequestDialog('recharge')"
          >
            <Icon name="plus" size="md" />
            {{ t('distribution.actions.requestWalletRecharge') }}
          </button>
          <button
            v-if="activeTab === 'wallet-requests' && overview?.can_manage_channel"
            type="button"
            class="btn btn-secondary btn-md"
            @click="openWalletRequestDialog('refund')"
          >
            <Icon name="plus" size="md" />
            {{ t('distribution.actions.requestWalletRefund') }}
          </button>
          <button
            v-if="activeTab === 'members'"
            type="button"
            class="btn btn-primary btn-md"
            @click="openMemberDialog"
          >
            <Icon name="plus" size="md" />
            {{ t('distribution.actions.createMember') }}
          </button>
          <button
            v-if="activeTab === 'promotion-links'"
            type="button"
            class="btn btn-primary btn-md"
            @click="openLinkDialog"
          >
            <Icon name="plus" size="md" />
            {{ t('distribution.actions.createPromotionLink') }}
          </button>
          <button type="button" class="btn btn-secondary btn-md" :disabled="loading" :title="t('common.refresh')" @click="loadCurrentTab">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div v-if="notAttributed" class="rounded-lg border border-gray-200 bg-white p-8 text-center dark:border-dark-700 dark:bg-dark-900">
        <Icon name="users" size="xl" class="mx-auto text-gray-400 dark:text-dark-500" />
        <h2 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">{{ t('distribution.empty.title') }}</h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">{{ t('distribution.empty.description') }}</p>
      </div>

      <template v-else>
        <div class="min-w-0">
          <TablePageLayout>
          <template #filters>
            <select v-if="activeTab === 'members'" v-model="roleType" class="input w-full sm:w-44" @change="reloadFromFirstPage">
              <option value="">{{ t('distribution.filters.allRoles') }}</option>
              <option v-for="role in roleOptions" :key="role" :value="role">{{ roleLabel(role) }}</option>
            </select>
            <div v-else-if="activeTab === 'wholesale-pricing'" class="flex w-full flex-col gap-2 sm:flex-row sm:items-center">
              <input
                v-model.trim="wholesalePricingQuery"
                class="input w-full sm:w-72"
                :placeholder="t('distribution.filters.pricingSearchPlaceholder')"
                @keyup.enter="reloadFromFirstPage"
                @change="reloadFromFirstPage"
              />
              <span class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('distribution.messages.wholesaleDiscountRate', { rate: formatRate(wholesalePricingDiscountRate) }) }}
              </span>
            </div>
            <select v-else-if="activeTab === 'wallet'" v-model="walletTransactionType" class="input w-full sm:w-52" @change="reloadFromFirstPage">
              <option value="">{{ t('distribution.filters.allTransactionTypes') }}</option>
              <option v-for="item in walletTransactionTypeOptions" :key="item" :value="item">{{ transactionTypeLabel(item) }}</option>
            </select>
            <div v-else-if="activeTab === 'alert-events'" class="flex w-full flex-col gap-2 sm:flex-row sm:items-center">
              <select v-model="alertType" class="input w-full sm:w-48" @change="reloadFromFirstPage">
                <option value="">{{ t('distribution.filters.allAlertTypes') }}</option>
                <option v-for="item in alertTypeOptions" :key="item" :value="item">{{ alertTypeLabel(item) }}</option>
              </select>
              <select v-model="alertStatus" class="input w-full sm:w-44" @change="reloadFromFirstPage">
                <option value="">{{ t('distribution.filters.allAlertStatuses') }}</option>
                <option v-for="item in alertStatusOptions" :key="item" :value="item">{{ alertStatusLabel(item) }}</option>
              </select>
              <select v-model="alertSeverity" class="input w-full sm:w-44" @change="reloadFromFirstPage">
                <option value="">{{ t('distribution.filters.allSeverities') }}</option>
                <option v-for="item in alertSeverityOptions" :key="item" :value="item">{{ alertSeverityLabel(item) }}</option>
              </select>
            </div>
            <div v-else-if="activeTab === 'wallet-requests'" class="flex w-full flex-col gap-2 sm:flex-row sm:items-center">
              <select v-model="walletRequestType" class="input w-full sm:w-44" @change="reloadFromFirstPage">
                <option value="">{{ t('distribution.filters.allRequestTypes') }}</option>
                <option v-for="item in walletRequestTypeOptions" :key="item" :value="item">{{ walletRequestTypeLabel(item) }}</option>
              </select>
              <select v-model="walletRequestStatus" class="input w-full sm:w-44" @change="reloadFromFirstPage">
                <option value="">{{ t('distribution.filters.allRequestStatuses') }}</option>
                <option v-for="item in walletRequestStatusOptions" :key="item" :value="item">{{ walletRequestStatusLabel(item) }}</option>
              </select>
            </div>
          </template>

          <template #table>
            <DataTable :columns="columns" :data="rows" :loading="loading" row-key="id">
              <template #header-select>
                <div v-if="activeTab === 'commissions'" class="flex justify-center">
                  <input
                    type="checkbox"
                    class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    :checked="allVisibleSelectableCommissionsSelected"
                    @click.stop
                    @change="toggleSelectAllVisibleCommissions($event)"
                  />
                </div>
              </template>
              <template #cell-select="{ row }">
                <div v-if="activeTab === 'commissions'" class="flex justify-center">
                  <input
                    type="checkbox"
                    class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500 disabled:cursor-not-allowed disabled:opacity-50"
                    :checked="selectedCommissionIds.includes(row.id)"
                    :disabled="!canSettleCommission(row)"
                    @click.stop
                    @change="toggleSelectCommission(row.id, $event)"
                  />
                </div>
              </template>
              <template #cell-member_id="{ row }">
                <span class="font-mono text-sm">#{{ row.member_id }}</span>
              </template>
              <template #cell-id="{ row }">
                <span class="font-mono text-sm">#{{ row.id }}</span>
              </template>
              <template #cell-user="{ row }">
                <div class="space-y-0.5">
                  <div class="font-medium text-gray-900 dark:text-white">#{{ row.user_id }} {{ row.user_email || '-' }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">{{ row.username || '-' }}</div>
                </div>
              </template>
              <template #cell-channel_org_id="{ row }">
                <span class="font-mono text-sm">#{{ row.channel_org_id }}</span>
              </template>
              <template #cell-role_type="{ row }">
                <StatusPill :label="roleLabel(row.role_type)" tone="blue" />
              </template>
              <template #cell-status="{ row }">
                <StatusPill :label="statusLabel(row.status)" :tone="statusTone(row.status)" />
              </template>
              <template #cell-parent_member_id="{ row }">
                <span class="font-mono text-sm">{{ row.parent_member_id ? `#${row.parent_member_id}` : '-' }}</span>
              </template>
              <template #cell-level_code="{ row }">
                <span class="font-mono text-sm">{{ row.level_code || '-' }}</span>
              </template>
              <template #cell-referrer_member_id="{ row }">
                <span class="font-mono text-sm">{{ row.referrer_member_id ? `#${row.referrer_member_id}` : '-' }}</span>
              </template>
              <template #cell-promotion_link_id="{ row }">
                <span class="font-mono text-sm">{{ row.promotion_link_id ? `#${row.promotion_link_id}` : '-' }}</span>
              </template>
              <template #cell-rate="{ row }">
                {{ formatRate(row.rate ?? row.commission_rate) }}
              </template>
              <template #cell-commission_type="{ row }">
                <StatusPill :label="commissionTypeLabel(row.commission_type)" tone="blue" />
              </template>
              <template #cell-settlement_method="{ row }">
                <StatusPill :label="settlementMethodLabel(row.settlement_method)" tone="gray" />
              </template>
              <template #cell-base_amount="{ row }">
                ${{ formatAmount(row.base_amount) }}
              </template>
              <template #cell-amount="{ row }">
                <span class="font-medium text-gray-900 dark:text-white">${{ formatAmount(row.amount) }}</span>
              </template>
              <template #cell-transaction_type="{ row }">
                <StatusPill :label="transactionTypeLabel(row.transaction_type)" tone="blue" />
              </template>
              <template #cell-alert_type="{ row }">
                <StatusPill :label="alertTypeLabel(row.alert_type)" tone="blue" />
              </template>
              <template #cell-severity="{ row }">
                <StatusPill :label="alertSeverityLabel(row.severity)" :tone="row.severity === 'critical' ? 'red' : row.severity === 'warning' ? 'amber' : 'gray'" />
              </template>
              <template #cell-details_summary="{ row }">
                <div class="max-w-xl text-sm text-gray-700 dark:text-dark-300">{{ alertSummary(row) }}</div>
              </template>
              <template #cell-request_type="{ row }">
                <StatusPill :label="walletRequestTypeLabel(row.request_type)" tone="blue" />
              </template>
              <template #cell-prepaid_balance_before="{ row }">
                ${{ formatAmount(row.prepaid_balance_before) }}
              </template>
              <template #cell-prepaid_balance_after="{ row }">
                ${{ formatAmount(row.prepaid_balance_after) }}
              </template>
              <template #cell-commission_reserved_after="{ row }">
                ${{ formatAmount(row.commission_reserved_after) }}
              </template>
              <template #cell-created_at="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
              <template #cell-triggered_at="{ row }">
                {{ formatDateTime(row.triggered_at) }}
              </template>
              <template #cell-resolved_at="{ row }">
                {{ formatDateTime(row.resolved_at) }}
              </template>
              <template #cell-last_observed_at="{ row }">
                {{ formatDateTime(row.last_observed_at) }}
              </template>
              <template #cell-bound_at="{ row }">
                {{ formatDateTime(row.bound_at) }}
              </template>
              <template #cell-frozen_until="{ row }">
                {{ formatDateTime(row.frozen_until) }}
              </template>
              <template #cell-code="{ row }">
                <div class="space-y-1">
                  <div class="font-mono text-sm">{{ row.code || '-' }}</div>
                  <div v-if="activeTab === 'promotion-links' && buildPromotionLink(row)" class="flex items-center gap-2">
                    <button
                      type="button"
                      class="max-w-[320px] truncate text-left text-xs text-primary-600 hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
                      :title="buildPromotionLink(row)"
                      @click="copyPromotionLink(row)"
                    >
                      {{ buildPromotionLink(row) }}
                    </button>
                    <button
                      type="button"
                      class="text-xs text-gray-500 hover:text-gray-700 dark:text-dark-400 dark:hover:text-dark-200"
                      :title="t('common.copy')"
                      @click="copyPromotionLink(row)"
                    >
                      {{ t('common.copy') }}
                    </button>
                  </div>
                </div>
              </template>
              <template #cell-provider="{ row }">
                <span class="font-mono text-xs">{{ row.provider || '-' }}</span>
              </template>
              <template #cell-billing_mode="{ row }">
                <span class="font-mono text-xs">{{ row.billing_mode || '-' }}</span>
              </template>
              <template #cell-official_pricing="{ row }">
                <div class="space-y-1 text-xs">
                  <div v-if="hasWholesalePricingValues(row, 'official')" class="space-y-1">
                    <div v-if="row.official_input_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.input') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.official_input_price) }}</span>
                    </div>
                    <div v-if="row.official_output_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.output') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.official_output_price) }}</span>
                    </div>
                    <div v-if="row.official_cache_write_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.cacheWrite') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.official_cache_write_price) }}</span>
                    </div>
                    <div v-if="row.official_cache_read_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.cacheRead') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.official_cache_read_price) }}</span>
                    </div>
                    <div v-if="row.official_image_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.image') }}</span>
                      <span class="font-mono">{{ formatImagePrice(row.official_image_price) }}</span>
                    </div>
                  </div>
                  <span v-else class="text-gray-400 dark:text-dark-500">-</span>
                </div>
              </template>
              <template #cell-wholesale_pricing="{ row }">
                <div class="space-y-1 text-xs">
                  <div v-if="hasWholesalePricingValues(row, 'wholesale')" class="space-y-1">
                    <div v-if="row.wholesale_input_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.input') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.wholesale_input_price) }}</span>
                    </div>
                    <div v-if="row.wholesale_output_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.output') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.wholesale_output_price) }}</span>
                    </div>
                    <div v-if="row.wholesale_cache_write_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.cacheWrite') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.wholesale_cache_write_price) }}</span>
                    </div>
                    <div v-if="row.wholesale_cache_read_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.cacheRead') }}</span>
                      <span class="font-mono">{{ formatPerMillionPrice(row.wholesale_cache_read_price) }}</span>
                    </div>
                    <div v-if="row.wholesale_image_price > 0" class="flex justify-between gap-3">
                      <span>{{ t('distribution.pricing.image') }}</span>
                      <span class="font-mono">{{ formatImagePrice(row.wholesale_image_price) }}</span>
                    </div>
                  </div>
                  <span v-else class="text-gray-400 dark:text-dark-500">-</span>
                </div>
              </template>
              <template #cell-target_type="{ row }">
                <StatusPill :label="targetTypeLabel(row.target_type)" tone="blue" />
              </template>
              <template #cell-actions="{ row }">
                <div class="flex justify-end">
                  <button
                    v-if="activeTab === 'commissions' && overview?.can_manage_channel && row.status !== 'settled' && row.status !== 'cancelled' && row.status !== 'reversed'"
                    type="button"
                    class="btn btn-primary btn-sm"
                    @click="openSettleDialog(row)"
                  >
                    {{ t('distribution.actions.settleCommission') }}
                  </button>
                </div>
              </template>
            </DataTable>
          </template>

        <template #pagination>
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
              :page-size="pagination.page_size"
              :total="pagination.total"
              @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </template>
          </TablePageLayout>
        </div>
      </template>
    </div>

    <DistributionMemberFormDialog
      :show="memberDialog"
      :title="t('distribution.dialogs.memberTitle')"
      :saving="saving"
      namespace="distribution"
      role-field-key="fields.roleType"
      parent-member-label-key="distribution.columns.parentMemberId"
      user-search-placeholder-key="admin.usage.searchUserPlaceholder"
      parent-search-placeholder-key="distribution.fields.parentMemberIdPlaceholder"
      level-code-placeholder-key="distribution.fields.levelCodePlaceholder"
      :hide-parent-field-for-agent="true"
      :parent-field-required-for-non-agent="true"
      :member-form="memberForm"
      :role-options="roleOptions"
      :member-user-lookup="memberUserLookup"
      :parent-member-lookup="parentMemberLookup"
      @close="memberDialog = false"
      @submit="submitMember"
      @member-user-input="scheduleUserLookup(memberUserLookup, clearMemberUserSelection)"
      @member-user-focus="memberUserLookup.open = true"
      @clear-member-user="clearMemberUserSelection"
      @select-member-user="selectMemberUser"
      @parent-member-input="scheduleParentMemberLookup(parentMemberLookup, clearParentMemberSelection)"
      @parent-member-focus="parentMemberLookup.open = true"
      @clear-parent-member="clearParentMemberSelection"
      @select-parent-member="selectParentMember"
    />

    <BaseDialog :show="linkDialog" :title="t('distribution.dialogs.linkTitle')" width="normal" @close="linkDialog = false">
      <form class="space-y-4" @submit.prevent="submitLink">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.memberId') }}</span>
            <input v-model.number="linkForm.member_id" type="number" min="1" class="input mt-1" required />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.code') }}</span>
            <input v-model.trim="linkForm.code" class="input mt-1" :placeholder="t('distribution.fields.codePlaceholder')" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.targetType') }}</span>
            <select v-model="linkForm.target_type" class="input mt-1">
              <option value="registration">{{ targetTypeLabel('registration') }}</option>
              <option value="oauth">{{ targetTypeLabel('oauth') }}</option>
              <option value="manual">{{ targetTypeLabel('manual') }}</option>
            </select>
          </label>
          <label class="block">
            <span class="input-label">{{ t('distribution.fields.status') }}</span>
            <select v-model="linkForm.status" class="input mt-1">
              <option value="active">{{ statusLabel('active') }}</option>
              <option value="inactive">{{ statusLabel('inactive') }}</option>
              <option value="disabled">{{ statusLabel('disabled') }}</option>
            </select>
          </label>
        </div>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="linkDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('common.create') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="walletRequestDialog" :title="walletRequestDialogTitle" width="normal" @close="walletRequestDialog = false">
      <form class="space-y-4" @submit.prevent="submitWalletRequest">
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.amount') }}</span>
          <input v-model.number="walletRequestForm.amount" type="number" min="0.0001" step="0.0001" class="input mt-1" required />
        </label>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.referenceNo') }}</span>
          <input v-model.trim="walletRequestForm.reference_no" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.note') }}</span>
          <textarea v-model.trim="walletRequestForm.note" class="input mt-1 min-h-[120px]" />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="walletRequestDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('common.create') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="settleDialog" :title="t('distribution.dialogs.settleTitle')" width="normal" @close="settleDialog = false">
      <form class="space-y-4" @submit.prevent="submitSettle">
        <p v-if="settleTargetIds.length > 1" class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('distribution.messages.selectedCommissionCount', { count: settleTargetIds.length }) }}
        </p>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.settlementMethod') }}</span>
          <select v-model="settleForm.settlement_method" class="input mt-1">
            <option value="manual">{{ t('distribution.settlementMethods.manual') }}</option>
            <option value="offline">{{ t('distribution.settlementMethods.offline') }}</option>
          </select>
        </label>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.referenceNo') }}</span>
          <input v-model.trim="settleForm.settlement_reference_no" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('distribution.fields.note') }}</span>
          <textarea v-model.trim="settleForm.settlement_note" class="input mt-1 min-h-[120px]" />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="settleDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('distribution.actions.settleCommission') }}</button>
        </div>
      </form>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DistributionMemberFormDialog from '@/components/distribution/DistributionMemberFormDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { usersAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminUser } from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDateTime as formatDisplayDateTime } from '@/utils/format'
import { formatScaled } from '@/utils/pricing'
import {
  createMyDistributionMember,
  getDistributionOverview,
  createMyDistributionPromotionLink,
  listMyDistributionAttributions,
  listMyDistributionAlertEvents,
  listMyDistributionCommissions,
  listMyDistributionMembers,
  listMyDistributionPromotionLinks,
  listMyDistributionWalletRequests,
  listMyDistributionWholesalePricing,
  listMyDistributionWalletTransactions,
  settleMyDistributionCommission,
  submitMyDistributionWalletRequest,
  type DistributionOverview,
  type DistributionWholesalePricingItem,
} from '@/api/distribution'
import type { DistributionCommission, DistributionMember, DistributionMemberRole } from '@/api/admin/distribution'

type DistributionTab = 'wallet' | 'alert-events' | 'wallet-requests' | 'members' | 'promotion-links' | 'attributions' | 'commissions' | 'wholesale-pricing'
type LookupState<T> = {
  keyword: string
  loading: boolean
  open: boolean
  results: T[]
  selected: T | null
  timer: ReturnType<typeof setTimeout> | null
}

type TableRow = Record<string, any>

function createLookupState<T>(): LookupState<T> {
  return reactive({
    keyword: '',
    loading: false,
    open: false,
    results: [],
    selected: null,
    timer: null,
  }) as LookupState<T>
}

function formatUserLookupLabel(user: Pick<AdminUser, 'username' | 'email'>) {
  return user.username ? `${user.username} (${user.email})` : user.email
}

function formatMemberLookupLabel(member: Pick<DistributionMember, 'username' | 'user_email' | 'member_id'>) {
  const primary = member.username ? `${member.username} (${member.user_email})` : member.user_email
  return `${primary} · #${member.member_id}`
}

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const overview = ref<DistributionOverview | null>(null)
const notAttributed = ref(false)
const activeTab = ref<DistributionTab>('wallet')
const loading = ref(false)
const saving = ref(false)
const rows = ref<TableRow[]>([])
const totals = reactive({ wallet: 0, alertEvents: 0, members: 0, promotionLinks: 0, attributions: 0, commissions: 0, wholesalePricing: 0 })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const roleType = ref('')
const walletTransactionType = ref('')
const alertType = ref('')
const alertSeverity = ref('')
const alertStatus = ref('')
const walletRequestType = ref('')
const walletRequestStatus = ref('')
const wholesalePricingQuery = ref('')
const wholesalePricingDiscountRate = ref(0.5)
const memberDialog = ref(false)
const linkDialog = ref(false)
const walletRequestDialog = ref(false)
const settleDialog = ref(false)
const selectedCommission = ref<DistributionCommission | null>(null)
const selectedCommissionIds = ref<number[]>([])
const hasInitializedDistributionPage = ref(false)
const memberUserLookup = createLookupState<AdminUser>()
const parentMemberLookup = createLookupState<DistributionMember>()
const memberForm = reactive({
  user_id: 0,
  role_type: 'kol1' as 'agent' | 'kol1' | 'kol2',
  parent_member_id: null as number | null,
  level_code: '',
  commission_rate: 0,
  status: 'active' as 'active' | 'inactive' | 'disabled',
})
const linkForm = reactive({
  member_id: 0,
  code: '',
  target_type: 'registration' as 'registration' | 'oauth' | 'manual',
  status: 'active' as 'active' | 'inactive' | 'disabled',
})
const walletRequestForm = reactive({
  request_type: 'recharge' as 'recharge' | 'refund',
  amount: 0,
  reference_no: '',
  note: '',
})
const settleForm = reactive({
  settlement_method: 'manual' as 'manual' | 'offline',
  settlement_reference_no: '',
  settlement_note: '',
})
const roleOptions: DistributionMemberRole[] = ['manager', 'agent', 'kol1', 'kol2']
const walletRequestTypeOptions = ['recharge', 'refund']
const walletRequestStatusOptions = ['pending', 'approved', 'rejected']
const walletTransactionTypeOptions = ['recharge', 'refund', 'consume', 'commission_reserve', 'commission_release', 'commission_settle', 'commission_deduct', 'commission_refund']
const alertTypeOptions = ['low_balance', 'balance_exhausted', 'consumption_warning', 'consumption_exhausted']
const alertStatusOptions = ['active', 'resolved']
const alertSeverityOptions = ['warning', 'critical']
const summary = computed(() => overview.value?.summary ?? null)
const distributionTabValues = new Set<DistributionTab>([
  'wallet',
  'alert-events',
  'wallet-requests',
  'members',
  'promotion-links',
  'attributions',
  'commissions',
  'wholesale-pricing',
])
const visibleSelectableCommissionIds = computed(() =>
  activeTab.value === 'commissions'
    ? rows.value
        .filter((row): row is DistributionCommission => canSettleCommission(row))
        .map((row) => row.id)
    : [],
)
const allVisibleSelectableCommissionsSelected = computed(
  () =>
    visibleSelectableCommissionIds.value.length > 0 &&
    visibleSelectableCommissionIds.value.every((commissionId) => selectedCommissionIds.value.includes(commissionId)),
)
const settleTargetIds = computed(() => {
  if (selectedCommission.value) {
    return [selectedCommission.value.id]
  }
  return selectedCommissionIds.value
})

function resetLookupState<T>(state: LookupState<T>) {
  state.keyword = ''
  state.loading = false
  state.open = false
  state.results = []
  state.selected = null
  if (state.timer) {
    clearTimeout(state.timer)
    state.timer = null
  }
}

async function searchMemberUsers(state: LookupState<AdminUser>) {
  const keyword = state.keyword.trim()
  if (!keyword) {
    state.results = []
    return
  }
  state.loading = true
  try {
    const response = await usersAPI.list(1, 10, { search: keyword })
    state.results = response.items
  } catch {
    state.results = []
  } finally {
    state.loading = false
  }
}

function scheduleUserLookup(state: LookupState<AdminUser>, clearSelection: () => void) {
  const keyword = state.keyword
  if (state.selected && keyword !== formatUserLookupLabel(state.selected)) {
    clearSelection()
    state.keyword = keyword
    state.open = true
  }
  if (state.timer) clearTimeout(state.timer)
  state.timer = setTimeout(() => {
    void searchMemberUsers(state)
  }, 250)
}

async function searchParentMembers(state: LookupState<DistributionMember>) {
  const keyword = state.keyword.trim()
  if (!keyword) {
    state.results = []
    return
  }
  state.loading = true
  try {
    const response = await listMyDistributionMembers({
      page: 1,
      page_size: 10,
      q: keyword,
    })
    state.results = response.items
  } catch {
    state.results = []
  } finally {
    state.loading = false
  }
}

function scheduleParentMemberLookup(state: LookupState<DistributionMember>, clearSelection: () => void) {
  const keyword = state.keyword
  if (state.selected && keyword !== formatMemberLookupLabel(state.selected)) {
    clearSelection()
    state.keyword = keyword
    state.open = true
  }
  if (state.timer) clearTimeout(state.timer)
  state.timer = setTimeout(() => {
    void searchParentMembers(state)
  }, 250)
}

function tabFromHash(hash: string): DistributionTab {
  const value = hash.replace(/^#/, '')
  if (value && distributionTabValues.has(value as DistributionTab)) {
    return value as DistributionTab
  }
  return 'wallet'
}

const walletRequestDialogTitle = computed(() =>
  walletRequestForm.request_type === 'refund'
    ? t('distribution.dialogs.walletRefundRequestTitle')
    : t('distribution.dialogs.walletRechargeRequestTitle'),
)
const columns = computed<Column[]>(() => {
  if (activeTab.value === 'wallet') {
    return [
      { key: 'transaction_type', label: t('distribution.columns.transactionType') },
      { key: 'amount', label: t('distribution.columns.amount') },
      { key: 'prepaid_balance_before', label: t('distribution.columns.balanceBefore') },
      { key: 'prepaid_balance_after', label: t('distribution.columns.balanceAfter') },
      { key: 'commission_reserved_after', label: t('distribution.columns.reservedAfter') },
      { key: 'reference_no', label: t('distribution.columns.referenceNo') },
      { key: 'note', label: t('distribution.columns.note') },
      { key: 'created_at', label: t('distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'members') {
    return [
      { key: 'member_id', label: t('distribution.columns.memberId') },
      { key: 'user', label: t('distribution.columns.user') },
      { key: 'role_type', label: t('distribution.columns.role') },
      { key: 'level_code', label: t('distribution.columns.levelCode') },
      { key: 'parent_member_id', label: t('distribution.columns.parentMemberId') },
      { key: 'rate', label: t('distribution.columns.commissionRate') },
      { key: 'status', label: t('distribution.columns.status') },
      { key: 'created_at', label: t('distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'promotion-links') {
    return [
      { key: 'id', label: t('distribution.columns.id') },
      { key: 'code', label: t('distribution.columns.code') },
      { key: 'user', label: t('distribution.columns.user') },
      { key: 'member_id', label: t('distribution.columns.memberId') },
      { key: 'role_type', label: t('distribution.columns.role') },
      { key: 'target_type', label: t('distribution.columns.targetType') },
      { key: 'status', label: t('distribution.columns.status') },
      { key: 'created_at', label: t('distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'attributions') {
    return [
      { key: 'user', label: t('distribution.columns.user') },
      { key: 'referrer_member_id', label: t('distribution.columns.referrerMemberId') },
      { key: 'promotion_link_id', label: t('distribution.columns.promotionLinkId') },
      { key: 'bound_source', label: t('distribution.columns.boundSource') },
      { key: 'bound_at', label: t('distribution.columns.boundAt') },
    ]
  }
  if (activeTab.value === 'wholesale-pricing') {
    return [
      { key: 'model', label: t('distribution.columns.model') },
      { key: 'provider', label: t('distribution.columns.provider') },
      { key: 'billing_mode', label: t('distribution.columns.billingMode') },
      { key: 'official_pricing', label: t('distribution.columns.officialPricing') },
      { key: 'wholesale_pricing', label: t('distribution.columns.wholesalePricing') },
    ]
  }
  if (activeTab.value === 'wallet-requests') {
    return [
      { key: 'request_type', label: t('distribution.columns.requestType') },
      { key: 'amount', label: t('distribution.columns.amount') },
      { key: 'status', label: t('distribution.columns.status') },
      { key: 'reference_no', label: t('distribution.columns.referenceNo') },
      { key: 'note', label: t('distribution.columns.note') },
      { key: 'created_at', label: t('distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'alert-events') {
    return [
      { key: 'alert_type', label: t('distribution.columns.alertType') },
      { key: 'severity', label: t('distribution.columns.severity') },
      { key: 'status', label: t('distribution.columns.status') },
      { key: 'details_summary', label: t('distribution.columns.summary') },
      { key: 'triggered_at', label: t('distribution.columns.triggeredAt') },
      { key: 'resolved_at', label: t('distribution.columns.resolvedAt') },
      { key: 'last_observed_at', label: t('distribution.columns.lastObservedAt') },
    ]
  }
  return [
    { key: 'select', label: '', class: 'w-12' },
    { key: 'id', label: t('distribution.columns.id') },
    { key: 'user', label: t('distribution.columns.user') },
    { key: 'member_id', label: t('distribution.columns.memberId') },
    { key: 'commission_type', label: t('distribution.columns.commissionType') },
    { key: 'base_amount', label: t('distribution.columns.baseAmount') },
    { key: 'rate', label: t('distribution.columns.commissionRate') },
    { key: 'amount', label: t('distribution.columns.amount') },
    { key: 'status', label: t('distribution.columns.status') },
    { key: 'settlement_method', label: t('distribution.columns.settlementMethod') },
    { key: 'frozen_until', label: t('distribution.columns.frozenUntil') },
    { key: 'actions', label: t('common.actions') },
  ]
})

async function loadOverview() {
  try {
    overview.value = await getDistributionOverview()
    notAttributed.value = false
  } catch (error: any) {
    if (error?.reason === 'DISTRIBUTION_ATTRIBUTION_NOT_FOUND' || error?.status === 404) {
      notAttributed.value = true
      return
    }
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.loadFailed')))
  }
}

async function loadCurrentTab() {
  if (notAttributed.value) return
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      role_type: activeTab.value === 'members' ? roleType.value || undefined : undefined,
      alert_type: activeTab.value === 'alert-events' ? alertType.value || undefined : undefined,
      severity: activeTab.value === 'alert-events' ? alertSeverity.value || undefined : undefined,
      request_type: activeTab.value === 'wallet-requests' ? walletRequestType.value || undefined : undefined,
      status: activeTab.value === 'wallet-requests' ? walletRequestStatus.value || undefined : activeTab.value === 'alert-events' ? alertStatus.value || undefined : undefined,
      transaction_type: activeTab.value === 'wallet' ? walletTransactionType.value || undefined : undefined,
      q: activeTab.value === 'wholesale-pricing' ? wholesalePricingQuery.value || undefined : undefined,
    }
    const result =
      activeTab.value === 'wallet'
        ? await listMyDistributionWalletTransactions(params)
        : activeTab.value === 'alert-events'
          ? await listMyDistributionAlertEvents(params)
        : activeTab.value === 'wallet-requests'
          ? await listMyDistributionWalletRequests(params)
        : activeTab.value === 'wholesale-pricing'
          ? await listMyDistributionWholesalePricing(params)
        : activeTab.value === 'members'
        ? await listMyDistributionMembers(params)
        : activeTab.value === 'promotion-links'
          ? await listMyDistributionPromotionLinks(params)
        : activeTab.value === 'attributions'
          ? await listMyDistributionAttributions(params)
          : await listMyDistributionCommissions(params)

    rows.value = result.items || []
    if ('discount_rate' in result && typeof result.discount_rate === 'number') {
      wholesalePricingDiscountRate.value = result.discount_rate
    }
    if (activeTab.value === 'commissions') {
      selectedCommissionIds.value = selectedCommissionIds.value.filter((commissionId) => visibleSelectableCommissionIds.value.includes(commissionId))
    } else {
      selectedCommissionIds.value = []
    }
    pagination.total = result.total || 0
    const totalKey = activeTab.value === 'promotion-links' ? 'promotionLinks' : activeTab.value === 'wholesale-pricing' ? 'wholesalePricing' : activeTab.value === 'alert-events' ? 'alertEvents' : activeTab.value === 'wallet-requests' ? 'wallet' : activeTab.value
    if (totalKey in totals) {
      totals[totalKey as keyof typeof totals] = result.total || 0
    }
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

function openMemberDialog() {
  memberForm.user_id = 0
  memberForm.role_type = 'kol1'
  memberForm.parent_member_id = null
  memberForm.level_code = ''
  memberForm.commission_rate = 0
  memberForm.status = 'active'
  resetLookupState(memberUserLookup)
  resetLookupState(parentMemberLookup)
  memberDialog.value = true
}

function selectMemberUser(user: AdminUser) {
  memberUserLookup.selected = user
  memberUserLookup.keyword = formatUserLookupLabel(user)
  memberUserLookup.open = false
  memberForm.user_id = user.id
}

function clearMemberUserSelection() {
  resetLookupState(memberUserLookup)
  memberForm.user_id = 0
}

function selectParentMember(member: DistributionMember) {
  parentMemberLookup.selected = member
  parentMemberLookup.keyword = formatMemberLookupLabel(member)
  parentMemberLookup.open = false
  memberForm.parent_member_id = member.member_id
}

function clearParentMemberSelection() {
  resetLookupState(parentMemberLookup)
  memberForm.parent_member_id = null
}

function openLinkDialog() {
  linkForm.member_id = 0
  linkForm.code = ''
  linkForm.target_type = 'registration'
  linkForm.status = 'active'
  linkDialog.value = true
}

function openWalletRequestDialog(requestType: 'recharge' | 'refund') {
  walletRequestForm.request_type = requestType
  walletRequestForm.amount = 0
  walletRequestForm.reference_no = ''
  walletRequestForm.note = ''
  walletRequestDialog.value = true
}

function openSettleDialog(row?: DistributionCommission) {
  selectedCommission.value = row || null
  settleForm.settlement_method = 'manual'
  settleForm.settlement_reference_no = row?.settlement_reference_no || ''
  settleForm.settlement_note = row?.settlement_note || ''
  settleDialog.value = true
}

function canSettleCommission(row: DistributionCommission | TableRow) {
  return row.status !== 'settled' && row.status !== 'cancelled' && row.status !== 'reversed'
}

function toggleSelectAllVisibleCommissions(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  selectedCommissionIds.value = checked ? [...visibleSelectableCommissionIds.value] : []
}

function toggleSelectCommission(commissionId: number, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  if (checked) {
    if (!selectedCommissionIds.value.includes(commissionId)) {
      selectedCommissionIds.value = [...selectedCommissionIds.value, commissionId]
    }
    return
  }
  selectedCommissionIds.value = selectedCommissionIds.value.filter((id) => id !== commissionId)
}

function reloadFromFirstPage() {
  pagination.page = 1
  void loadCurrentTab()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadCurrentTab()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadCurrentTab()
}

async function submitLink() {
  saving.value = true
  try {
    await createMyDistributionPromotionLink({
      member_id: linkForm.member_id,
      code: linkForm.code || undefined,
      target_type: linkForm.target_type,
      status: linkForm.status,
    })
    appStore.showSuccess(t('distribution.messages.linkCreated'))
    linkDialog.value = false
    await loadCurrentTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.createFailed')))
  } finally {
    saving.value = false
  }
}

async function submitMember() {
  saving.value = true
  try {
    await createMyDistributionMember({
      user_id: memberForm.user_id,
      role_type: memberForm.role_type,
      parent_member_id: memberForm.role_type === 'agent' ? undefined : memberForm.parent_member_id || undefined,
      level_code: memberForm.level_code || undefined,
      commission_rate: memberForm.commission_rate,
      status: memberForm.status,
    })
    appStore.showSuccess(t('distribution.messages.memberCreated'))
    memberDialog.value = false
    if (activeTab.value !== 'members') {
      activeTab.value = 'members'
    }
    await loadCurrentTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.createFailed')))
  } finally {
    saving.value = false
  }
}

async function submitWalletRequest() {
  saving.value = true
  try {
    await submitMyDistributionWalletRequest({
      request_type: walletRequestForm.request_type,
      amount: walletRequestForm.amount,
      reference_no: walletRequestForm.reference_no || undefined,
      note: walletRequestForm.note || undefined,
    })
    appStore.showSuccess(t('distribution.messages.walletRequestSubmitted'))
    walletRequestDialog.value = false
    if (activeTab.value !== 'wallet-requests') {
      activeTab.value = 'wallet-requests'
    }
    await loadCurrentTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.createFailed')))
  } finally {
    saving.value = false
  }
}

async function submitSettle() {
  const targetIds = settleTargetIds.value
  if (targetIds.length === 0) return
  saving.value = true
  let settledCount = 0
  try {
    for (const commissionId of targetIds) {
      await settleMyDistributionCommission(commissionId, {
        settlement_method: settleForm.settlement_method,
        settlement_reference_no: settleForm.settlement_reference_no || undefined,
        settlement_note: settleForm.settlement_note || undefined,
      })
      settledCount += 1
    }
    appStore.showSuccess(
      settledCount > 1 ? t('distribution.messages.commissionsSettled', { count: settledCount }) : t('distribution.messages.commissionSettled'),
    )
    settleDialog.value = false
    selectedCommission.value = null
    selectedCommissionIds.value = []
    await loadOverview()
    await loadCurrentTab()
  } catch (error) {
    if (settledCount > 0) {
      appStore.showSuccess(t('distribution.messages.commissionsSettledPartial', { count: settledCount }))
    }
    appStore.showError(extractI18nErrorMessage(error, t, 'distribution.errors', t('distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

function roleLabel(role: DistributionMemberRole | string) {
  return t(`distribution.roles.${role}`, role)
}

function transactionTypeLabel(transactionType: string) {
  return t(`distribution.transactionTypes.${transactionType}`, transactionType)
}

function walletRequestTypeLabel(requestType: string) {
  return t(`distribution.requestTypes.${requestType}`, requestType)
}

function walletRequestStatusLabel(status: string) {
  return t(`distribution.requestStatuses.${status}`, status)
}

function alertTypeLabel(alertType: string) {
  return t(`distribution.alertTypes.${alertType}`, alertType)
}

function alertSeverityLabel(severity: string) {
  return t(`distribution.severities.${severity}`, severity)
}

function alertStatusLabel(status: string) {
  return t(`distribution.alertStatuses.${status}`, status)
}

function targetTypeLabel(targetType: string) {
  return t(`distribution.targetTypes.${targetType}`, targetType)
}

function commissionTypeLabel(commissionType: string) {
  return t(`distribution.commissionTypes.${commissionType}`, commissionType || '-')
}

function settlementMethodLabel(settlementMethod: string) {
  return t(`distribution.settlementMethods.${settlementMethod}`, settlementMethod || '-')
}

function statusLabel(status: string) {
  if (status === 'resolved') return alertStatusLabel(status)
  return t(`distribution.statuses.${status}`, status || '-')
}

function statusTone(status: string) {
  if (status === 'active' || status === 'available' || status === 'settled') return 'green'
  if (status === 'frozen' || status === 'inactive') return 'amber'
  return 'gray'
}

function formatRate(value: number | null | undefined) {
  return `${(Number(value || 0) * 100).toFixed(2)}%`
}

function formatAmount(value: number | null | undefined) {
  return Number(value || 0).toFixed(4)
}

function alertSummary(row: TableRow) {
  const details = row.details && typeof row.details === 'object' ? row.details as Record<string, unknown> : {}
  const note = String(details.recharge_deadline_note || '')
  switch (row.alert_type) {
    case 'low_balance':
      return t('distribution.alertSummaries.lowBalance', {
        balance: formatAmount(Number(details.prepaid_balance || 0)),
        threshold: formatAmount(Number(details.warning_threshold || 0)),
        note,
      })
    case 'balance_exhausted':
      return t('distribution.alertSummaries.balanceExhausted', {
        available: formatAmount(Number(details.available_balance || 0)),
        note,
      })
    case 'consumption_warning':
      return t('distribution.alertSummaries.consumptionWarning', {
        remaining: formatAmount(Number(details.remaining_consumption || 0)),
        limit: formatAmount(Number(details.consumption_limit || 0)),
        note,
      })
    case 'consumption_exhausted':
      return t('distribution.alertSummaries.consumptionExhausted', {
        consumed: formatAmount(Number(details.total_consumed || 0)),
        limit: formatAmount(Number(details.consumption_limit || 0)),
        note,
      })
    default:
      return note || '-'
  }
}

function formatPerMillionPrice(value: number | null | undefined) {
  return value && value > 0 ? `${formatScaled(value, 1_000_000)} / 1M` : '-'
}

function formatImagePrice(value: number | null | undefined) {
  return value && value > 0 ? `${formatScaled(value, 1)} / image` : '-'
}

function formatDateTime(value: string | null | undefined) {
  return value ? formatDisplayDateTime(value) : '-'
}

function buildPromotionLink(row: TableRow) {
  const code = String(row?.code || '').trim()
  if (!code) return ''

  const configuredDomain = String(summary.value?.organization?.brand_config?.domain ?? '').trim()
  const baseOrigin = resolvePromotionLinkBaseOrigin(configuredDomain)
  if (!baseOrigin) return ''

  return `${baseOrigin}/register?promo_code=${encodeURIComponent(code)}`
}

async function copyPromotionLink(row: TableRow) {
  const promotionLink = buildPromotionLink(row)
  if (!promotionLink || typeof navigator === 'undefined' || !navigator.clipboard) return
  await navigator.clipboard.writeText(promotionLink)
  appStore.showSuccess(t('common.copied'))
}

function resolvePromotionLinkBaseOrigin(domain: string) {
  const value = domain.trim()
  if (value) {
    if (/^https?:\/\//i.test(value)) {
      return value.replace(/\/+$/, '')
    }
    if (typeof window !== 'undefined' && window.location?.protocol) {
      return `${window.location.protocol}//${value.replace(/\/+$/, '')}`
    }
    return `https://${value.replace(/\/+$/, '')}`
  }
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin.replace(/\/+$/, '')
  }
  return ''
}

function hasWholesalePricingValues(row: DistributionWholesalePricingItem, prefix: 'official' | 'wholesale') {
  return row[`${prefix}_input_price`] > 0 ||
    row[`${prefix}_output_price`] > 0 ||
    row[`${prefix}_cache_write_price`] > 0 ||
    row[`${prefix}_cache_read_price`] > 0 ||
    row[`${prefix}_image_price`] > 0
}

watch(activeTab, () => {
  if (!hasInitializedDistributionPage.value) return
  rows.value = []
  pagination.page = 1
  pagination.total = 0
  selectedCommissionIds.value = []
  selectedCommission.value = null
  if (activeTab.value !== 'wallet') walletTransactionType.value = ''
  if (activeTab.value !== 'alert-events') {
    alertType.value = ''
    alertSeverity.value = ''
    alertStatus.value = ''
  }
  if (activeTab.value !== 'wallet-requests') {
    walletRequestType.value = ''
    walletRequestStatus.value = ''
  }
  if (activeTab.value !== 'wholesale-pricing') wholesalePricingQuery.value = ''
  if (activeTab.value !== 'members') roleType.value = ''
  void loadCurrentTab()
})

watch(
  () => route.hash,
  () => {
    const nextTab = tabFromHash(route.hash)
    if (activeTab.value !== nextTab) {
      activeTab.value = nextTab
    }
  },
  { immediate: true }
)

watch(
  () => memberForm.role_type,
  (roleType) => {
    if (roleType === 'agent') {
      resetLookupState(parentMemberLookup)
      memberForm.parent_member_id = null
    }
  },
)

onMounted(async () => {
  await loadOverview()
  await loadCurrentTab()
  hasInitializedDistributionPage.value = true
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
