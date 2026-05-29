<template>
  <AdminDistributionLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('admin.distribution.title') }}</h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-dark-400">{{ t('admin.distribution.description') }}</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-md" :disabled="loading" :title="t('common.refresh')" @click="loadActiveTab">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button v-if="activeTab === 'organizations'" type="button" class="btn btn-primary btn-md" @click="openOrganizationDialog()">
            <Icon name="plus" size="md" />
            {{ t('admin.distribution.actions.createOrganization') }}
          </button>
          <button v-if="activeTab === 'members'" type="button" class="btn btn-primary btn-md" @click="openMemberDialog">
            <Icon name="plus" size="md" />
            {{ t('admin.distribution.actions.createMember') }}
          </button>
          <button v-if="activeTab === 'promotion-links'" type="button" class="btn btn-primary btn-md" @click="openLinkDialog">
            <Icon name="plus" size="md" />
            {{ t('admin.distribution.actions.createPromotionLink') }}
          </button>
        </div>
      </div>

      <TablePageLayout>
        <template #filters>
          <div class="flex flex-wrap items-center gap-3">
            <div
              v-if="activeTab !== 'organizations'"
              class="relative w-full sm:w-72"
            >
              <input
                v-model="filterChannelOrgLookup.keyword"
                type="text"
                class="input w-full pr-8"
                :placeholder="t('admin.distribution.fields.channelOrgIdPlaceholder')"
                @input="scheduleFilterChannelOrgLookup"
                @focus="filterChannelOrgLookup.open = true"
                @keyup.enter="reloadFromFirstPage"
              />
              <button
                v-if="filters.channel_org_id"
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                @click="clearFilterChannelOrgSelection(); reloadFromFirstPage()"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
              <div
                v-if="filterChannelOrgLookup.open && (filterChannelOrgLookup.results.length > 0 || filterChannelOrgLookup.keyword)"
                class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
              >
                <div v-if="filterChannelOrgLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
                <div v-else-if="filterChannelOrgLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
                <button
                  v-for="organization in filterChannelOrgLookup.results"
                  :key="organization.id"
                  type="button"
                  class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="selectFilterChannelOrg(organization)"
                >
                  <div class="font-medium text-gray-900 dark:text-white">{{ organization.name }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    #{{ organization.id }} · {{ organizationTypeLabel(organization.type) }}
                  </div>
                </button>
              </div>
            </div>
            <div
              v-if="usesUserFilter(activeTab)"
              class="relative w-full sm:w-72"
            >
              <input
                v-model="filterUserLookup.keyword"
                type="text"
                class="input w-full pr-8"
                :placeholder="t('admin.usage.searchUserPlaceholder')"
                @input="scheduleUserLookup(filterUserLookup, clearFilterUserSelection)"
                @focus="filterUserLookup.open = true"
              />
              <button
                v-if="filters.user_id"
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                @click="clearFilterUserSelection(); reloadFromFirstPage()"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
              <div
                v-if="filterUserLookup.open && (filterUserLookup.results.length > 0 || filterUserLookup.keyword)"
                class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
              >
                <div v-if="filterUserLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
                <div v-else-if="filterUserLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
                <button
                  v-for="user in filterUserLookup.results"
                  :key="user.id"
                  type="button"
                  class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="selectFilterUser(user)"
                >
                  <div class="font-medium text-gray-900 dark:text-white">{{ user.username || user.email }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ user.email }} · #{{ user.id }}</div>
                </button>
              </div>
            </div>
            <select
              v-if="activeTab === 'members' || activeTab === 'promotion-links'"
              v-model="filters.role_type"
              class="input w-full sm:w-44"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allRoles') }}</option>
              <option v-for="role in roleOptions" :key="role" :value="role">{{ roleLabel(role) }}</option>
            </select>
            <select
              v-if="activeTab === 'wallet-requests'"
              v-model="filters.request_type"
              class="input w-full sm:w-44"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allRequestTypes') }}</option>
              <option v-for="item in walletRequestTypeOptions" :key="item" :value="item">{{ walletRequestTypeLabel(item) }}</option>
            </select>
            <select
              v-if="activeTab === 'wallet-requests'"
              v-model="filters.request_status"
              class="input w-full sm:w-44"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allRequestStatuses') }}</option>
              <option v-for="item in walletRequestStatusOptions" :key="item" :value="item">{{ walletRequestStatusLabel(item) }}</option>
            </select>
            <select
              v-if="activeTab === 'alert-events'"
              v-model="filters.alert_type"
              class="input w-full sm:w-48"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allAlertTypes') }}</option>
              <option v-for="item in alertTypeOptions" :key="item" :value="item">{{ alertTypeLabel(item) }}</option>
            </select>
            <select
              v-if="activeTab === 'alert-events'"
              v-model="filters.alert_status"
              class="input w-full sm:w-44"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allAlertStatuses') }}</option>
              <option v-for="item in alertStatusOptions" :key="item" :value="item">{{ alertStatusLabel(item) }}</option>
            </select>
            <select
              v-if="activeTab === 'alert-events'"
              v-model="filters.alert_severity"
              class="input w-full sm:w-44"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allSeverities') }}</option>
              <option v-for="item in alertSeverityOptions" :key="item" :value="item">{{ alertSeverityLabel(item) }}</option>
            </select>
            <select
              v-if="activeTab === 'wallet-transactions'"
              v-model="filters.transaction_type"
              class="input w-full sm:w-52"
              @change="reloadFromFirstPage"
            >
              <option value="">{{ t('admin.distribution.filters.allTransactionTypes') }}</option>
              <option v-for="item in walletTransactionTypeOptions" :key="item" :value="item">{{ transactionTypeLabel(item) }}</option>
            </select>
          </div>
        </template>

        <template #table>
          <DataTable :columns="columns" :data="rows" :loading="loading">
            <template #cell-id="{ row }">
              <span class="font-mono text-sm">#{{ row.id }}</span>
            </template>
            <template #cell-member_id="{ row }">
              <span class="font-mono text-sm">#{{ row.member_id }}</span>
            </template>
            <template #cell-user="{ row }">
              <div class="space-y-0.5">
                <div class="font-medium text-gray-900 dark:text-white">#{{ row.user_id }} {{ row.user_email || '-' }}</div>
                <div class="text-xs text-gray-500 dark:text-dark-400">{{ row.username || '-' }}</div>
              </div>
            </template>
            <template #cell-name="{ row }">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
            </template>
            <template #cell-type="{ row }">
              <StatusPill :label="organizationTypeLabel(row.type)" tone="blue" />
            </template>
            <template #cell-role_type="{ row }">
              <StatusPill :label="roleLabel(row.role_type)" tone="blue" />
            </template>
            <template #cell-status="{ row }">
              <StatusPill :label="statusLabel(row.status)" :tone="statusTone(row.status)" />
            </template>
            <template #cell-channel_org_id="{ row }">
              <span class="font-mono text-sm">#{{ row.channel_org_id }}</span>
            </template>
            <template #cell-organization_name="{ row }">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.organization_name || '-' }}</div>
            </template>
            <template #cell-organization_type="{ row }">
              <StatusPill :label="organizationTypeLabel(row.organization_type)" tone="blue" />
            </template>
            <template #cell-prepaid_balance="{ row }">
              ${{ formatAmount(row.prepaid_balance) }}
            </template>
            <template #cell-commission_reserved="{ row }">
              ${{ formatAmount(row.commission_reserved) }}
            </template>
            <template #cell-total_recharged="{ row }">
              ${{ formatAmount(row.total_recharged) }}
            </template>
            <template #cell-total_consumed="{ row }">
              ${{ formatAmount(row.total_consumed) }}
            </template>
            <template #cell-warning_threshold="{ row }">
              ${{ formatAmount(row.warning_threshold) }}
            </template>
            <template #cell-updated_at="{ row }">
              {{ formatDateTime(row.updated_at) }}
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
            <template #cell-parent_member_id="{ row }">
              <span class="font-mono text-sm">{{ row.parent_member_id ? `#${row.parent_member_id}` : '-' }}</span>
            </template>
            <template #cell-level_code="{ row }">
              <span v-if="row.role_type === 'agent'" class="font-mono text-sm">{{ row.level_code || '-' }}</span>
              <span v-else class="text-sm text-gray-400">-</span>
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
            <template #cell-prepaid_balance_before="{ row }">
              ${{ formatAmount(row.prepaid_balance_before) }}
            </template>
            <template #cell-prepaid_balance_after="{ row }">
              ${{ formatAmount(row.prepaid_balance_after) }}
            </template>
            <template #cell-commission_reserved_after="{ row }">
              ${{ formatAmount(row.commission_reserved_after) }}
            </template>
            <template #cell-base_amount="{ row }">
              ${{ formatAmount(row.base_amount) }}
            </template>
            <template #cell-frozen_until="{ row }">
              {{ formatDateTime(row.frozen_until) }}
            </template>
            <template #cell-actions="{ row }">
              <div class="flex flex-wrap justify-end gap-2">
                <button
                  v-if="activeTab === 'organizations'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openOrganizationDialog(row)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  v-if="activeTab === 'wallets'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openWalletDialog(row)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  v-if="activeTab === 'wallets'"
                  type="button"
                  class="btn btn-primary btn-sm"
                  @click="openRechargeDialog(row)"
                >
                  {{ t('admin.distribution.actions.rechargeWallet') }}
                </button>
                <button
                  v-if="activeTab === 'wallets'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openRefundDialog(row)"
                >
                  {{ t('admin.distribution.actions.refundWallet') }}
                </button>
                <button
                  v-if="activeTab === 'wallet-requests' && row.status === 'pending'"
                  type="button"
                  class="btn btn-primary btn-sm"
                  @click="reviewWalletRequest(row, 'approve')"
                >
                  {{ t('admin.distribution.actions.approveWalletRequest') }}
                </button>
                <button
                  v-if="activeTab === 'wallet-requests' && row.status === 'pending'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="reviewWalletRequest(row, 'reject')"
                >
                  {{ t('admin.distribution.actions.rejectWalletRequest') }}
                </button>
                <button
                  v-if="activeTab === 'attributions'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openAttributionAuditDialog(row)"
                >
                  {{ t('admin.distribution.actions.viewAttributionAudits') }}
                </button>
                <button
                  v-if="activeTab === 'attributions'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openAttributionDialog(row)"
                >
                  {{ t('admin.distribution.actions.adjustAttribution') }}
                </button>
                <button
                  v-if="activeTab === 'commissions' && row.status !== 'settled'"
                  type="button"
                  class="btn btn-primary btn-sm"
                  @click="openSettleDialog(row)"
                >
                  {{ t('admin.distribution.actions.settleCommission') }}
                </button>
                <button
                  v-if="activeTab === 'commissions' && row.status !== 'reversed'"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="reverseCommission(row)"
                >
                  {{ t('admin.distribution.actions.reverseCommission') }}
                </button>
              </div>
            </template>
              <template #cell-created_at="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
              <template #cell-bound_at="{ row }">
                {{ formatDateTime(row.bound_at) }}
              </template>
              <template #cell-code="{ row }">
                <span class="font-mono text-sm">{{ row.code || '-' }}</span>
              </template>
              <template #cell-target_type="{ row }">
                <StatusPill :label="targetTypeLabel(row.target_type)" tone="blue" />
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

    <BaseDialog :show="organizationDialog" :title="selectedOrganization ? t('admin.distribution.dialogs.organizationEditTitle') : t('admin.distribution.dialogs.organizationTitle')" width="wide" @close="organizationDialog = false">
      <form class="space-y-4" @submit.prevent="submitOrganization">
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.name') }}</span>
          <input v-model.trim="organizationForm.name" class="input mt-1" required />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.type') }}</span>
          <select v-model="organizationForm.type" class="input mt-1">
            <option value="reseller">{{ organizationTypeLabel('reseller') }}</option>
            <option value="platform">{{ organizationTypeLabel('platform') }}</option>
            <option value="oem">{{ organizationTypeLabel('oem') }}</option>
          </select>
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.ownerUser') }}</span>
          <div class="relative mt-1">
            <input
              v-model="ownerUserLookup.keyword"
              type="text"
              class="input pr-8"
              :placeholder="t('admin.usage.searchUserPlaceholder')"
              @input="scheduleUserLookup(ownerUserLookup, clearOwnerUserSelection)"
              @focus="ownerUserLookup.open = true"
            />
            <button
              v-if="organizationForm.owner_user_id"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              @click="clearOwnerUserSelection"
            >
              <Icon name="x" size="sm" :stroke-width="2" />
            </button>
            <div
              v-if="ownerUserLookup.open && (ownerUserLookup.results.length > 0 || ownerUserLookup.keyword)"
              class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
            >
              <div v-if="ownerUserLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
              <div v-else-if="ownerUserLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
              <button
                v-for="user in ownerUserLookup.results"
                :key="user.id"
                type="button"
                class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                @click="selectOwnerUser(user)"
              >
                <div class="font-medium text-gray-900 dark:text-white">{{ user.username || user.email }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ user.email }} · #{{ user.id }}</div>
              </button>
            </div>
          </div>
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.wholesaleDiscountRate') }}</span>
          <input v-model.number="organizationConfigForm.wholesale_discount_rate" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.refundFeeRate') }}</span>
          <input v-model.number="organizationConfigForm.refund_fee_rate" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.commissionUpperRatio') }}</span>
          <input v-model.number="organizationConfigForm.commission_upper_ratio" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.managementRewardCap') }}</span>
          <input v-model.number="organizationConfigForm.management_reward_cap" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.directCommissionRateCap') }}</span>
          <input v-model.number="organizationConfigForm.direct_commission_rate_cap" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.channelCommissionRate') }}</span>
          <input v-model.number="organizationConfigForm.channel_commission_rate" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.firstRechargeMinAmount') }}</span>
          <input v-model.number="organizationConfigForm.first_recharge_min_amount" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.rechargeMinAmount') }}</span>
          <input v-model.number="organizationConfigForm.recharge_min_amount" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.consumptionLimit') }}</span>
          <input v-model.number="organizationConfigForm.consumption_limit" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.consumptionWarningThreshold') }}</span>
          <input v-model.number="organizationConfigForm.consumption_warning_threshold" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.rechargeLeadTimeDays') }}</span>
          <input v-model.number="organizationConfigForm.recharge_lead_time_days" type="number" min="0" step="1" class="input mt-1" />
        </label>
        <label class="block sm:col-span-2">
          <span class="input-label">{{ t('admin.distribution.fields.rechargeDeadlineNote') }}</span>
          <textarea v-model.trim="organizationConfigForm.recharge_deadline_note" class="input mt-1 min-h-[120px]" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.maxAgentCount') }}</span>
          <input v-model.number="organizationConfigForm.max_agent_count" type="number" min="0" step="1" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.maxKolCount') }}</span>
          <input v-model.number="organizationConfigForm.max_kol_count" type="number" min="0" step="1" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.maxManagerCount') }}</span>
          <input v-model.number="organizationConfigForm.max_manager_count" type="number" min="0" step="1" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.userCommissionTotalCap') }}</span>
          <input v-model.number="organizationConfigForm.user_commission_total_cap" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.memberCommissionTotalCap') }}</span>
          <input v-model.number="organizationConfigForm.member_commission_total_cap" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.teamRewardRate') }}</span>
          <input v-model.number="organizationConfigForm.team_reward_rate" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.teamRewardThreshold') }}</span>
          <input v-model.number="organizationConfigForm.team_reward_threshold" type="number" min="0" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.freezeHours') }}</span>
          <input v-model.number="organizationConfigForm.freeze_hours" type="number" min="0" max="720" step="1" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.kol2Rate') }}</span>
          <input v-model.number="organizationConfigForm.kol2_rate" type="number" min="0" max="1" step="0.0001" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.commissionSettlementMethod') }}</span>
          <select v-model="organizationConfigForm.commission_settlement_method" class="input mt-1">
            <option value="balance">{{ t('admin.distribution.settlementMethods.balance') }}</option>
            <option value="auto">{{ t('admin.distribution.settlementMethods.auto') }}</option>
            <option value="manual">{{ t('admin.distribution.settlementMethods.manual') }}</option>
            <option value="offline">{{ t('admin.distribution.settlementMethods.offline') }}</option>
          </select>
        </label>
        <div v-if="organizationForm.type === 'platform'" class="sm:col-span-2 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-900/40 dark:text-dark-300">
          <p>{{ t('admin.distribution.fields.levelsPlatformHint') }}</p>
          <RouterLink to="/admin/settings" class="mt-1 inline-block text-primary-600 hover:underline dark:text-primary-400">
            {{ t('admin.distribution.fields.levelsPlatformLink') }}
          </RouterLink>
        </div>
        <div v-else class="sm:col-span-2">
          <DistributionLevelsEditor
            ref="organizationLevelsEditorRef"
            v-model="organizationConfigForm.distribution_levels"
            :title="t('admin.distribution.fields.levelsTitle')"
            :description="t('admin.distribution.fields.levelsDesc')"
          />
        </div>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.logoUrl') }}</span>
          <input v-model.trim="organizationBrandForm.logo_url" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.primaryColor') }}</span>
          <input v-model.trim="organizationBrandForm.primary_color" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.domain') }}</span>
          <input v-model.trim="organizationBrandForm.domain" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.apiDomain') }}</span>
          <input v-model.trim="organizationBrandForm.api_domain" class="input mt-1" />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="organizationDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : selectedOrganization ? t('common.save') : t('common.create') }}</button>
        </div>
      </form>
    </BaseDialog>

    <DistributionMemberFormDialog
      :show="memberDialog"
      :title="t('admin.distribution.dialogs.memberTitle')"
      :saving="saving"
      namespace="admin.distribution"
      role-field-key="fields.role"
      parent-member-label-key="admin.distribution.fields.parentMemberId"
      user-search-placeholder-key="admin.usage.searchUserPlaceholder"
      parent-search-placeholder-key="admin.distribution.fields.referrerMemberPlaceholder"
      channel-org-search-placeholder-key="admin.distribution.fields.channelOrgIdPlaceholder"
      level-code-description-key="admin.distribution.fields.levelCodeDesc"
      :show-channel-org-field="true"
      :channel-org-lookup="channelOrgLookup"
      :disable-parent-lookup="memberForm.channel_org_id <= 0"
      :level-options="memberLevelOptions"
      :member-form="memberForm"
      :role-options="roleOptions"
      :member-user-lookup="memberUserLookup"
      :parent-member-lookup="parentMemberLookup"
      @close="memberDialog = false"
      @submit="submitMember"
      @channel-org-change="handleMemberChannelOrgChange"
      @channel-org-input="scheduleChannelOrgLookup"
      @channel-org-focus="channelOrgLookup.open = true"
      @clear-channel-org="clearChannelOrgSelection"
      @select-channel-org="selectChannelOrg"
      @member-user-input="scheduleUserLookup(memberUserLookup, clearMemberUserSelection)"
      @member-user-focus="memberUserLookup.open = true"
      @clear-member-user="clearMemberUserSelection"
      @select-member-user="selectMemberUser"
      @parent-member-input="scheduleMemberLookup(parentMemberLookup, clearParentMemberSelection, memberForm.channel_org_id)"
      @parent-member-focus="parentMemberLookup.open = true"
      @clear-parent-member="clearParentMemberSelection"
      @select-parent-member="selectParentMember"
    />

    <BaseDialog :show="linkDialog" :title="t('admin.distribution.dialogs.linkTitle')" width="normal" @close="linkDialog = false">
      <form class="space-y-4" @submit.prevent="submitLink">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block">
            <span class="input-label">{{ t('admin.distribution.fields.promotionMember') }}</span>
            <div class="relative mt-1">
              <input
                v-model="linkMemberLookup.keyword"
                type="text"
                class="input pr-8"
                :placeholder="t('admin.distribution.fields.referrerMemberPlaceholder')"
                @input="scheduleMemberLookup(linkMemberLookup, clearLinkMemberSelection, filters.channel_org_id || undefined)"
                @focus="linkMemberLookup.open = true"
              />
              <button
                v-if="linkForm.member_id > 0"
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                @click="clearLinkMemberSelection"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
              <div
                v-if="linkMemberLookup.open && (linkMemberLookup.results.length > 0 || linkMemberLookup.keyword)"
                class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
              >
                <div v-if="linkMemberLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
                <div v-else-if="linkMemberLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
                <button
                  v-for="member in linkMemberLookup.results"
                  :key="member.member_id"
                  type="button"
                  class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="selectLinkMember(member)"
                >
                  <div class="font-medium text-gray-900 dark:text-white">{{ member.username || member.user_email }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ member.user_email }} · {{ roleLabel(member.role_type) }}</div>
                </button>
              </div>
            </div>
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.distribution.fields.code') }}</span>
            <input v-model.trim="linkForm.code" class="input mt-1" :placeholder="t('admin.distribution.fields.codePlaceholder')" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.distribution.fields.targetType') }}</span>
            <select v-model="linkForm.target_type" class="input mt-1">
              <option value="registration">{{ targetTypeLabel('registration') }}</option>
              <option value="oauth">{{ targetTypeLabel('oauth') }}</option>
              <option value="manual">{{ targetTypeLabel('manual') }}</option>
            </select>
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.distribution.fields.status') }}</span>
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

    <BaseDialog :show="walletDialog" :title="t('admin.distribution.dialogs.walletTitle')" width="normal" @close="walletDialog = false">
      <form class="space-y-4" @submit.prevent="submitWallet">
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.warningThreshold') }}</span>
          <input v-model.number="walletForm.warning_threshold" type="number" min="0" step="0.0001" class="input mt-1" required />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="walletDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="rechargeDialog" :title="t('admin.distribution.dialogs.rechargeTitle')" width="normal" @close="rechargeDialog = false">
      <form class="space-y-4" @submit.prevent="submitRecharge">
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.rechargeAmount') }}</span>
          <input v-model.number="rechargeForm.amount" type="number" min="0.0001" step="0.0001" class="input mt-1" required />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.referenceNo') }}</span>
          <input v-model.trim="rechargeForm.reference_no" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.note') }}</span>
          <textarea v-model.trim="rechargeForm.note" class="input mt-1 min-h-[120px]" />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="rechargeDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('admin.distribution.actions.rechargeWallet') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="refundDialog" :title="t('admin.distribution.dialogs.refundTitle')" width="normal" @close="refundDialog = false">
      <form class="space-y-4" @submit.prevent="submitRefund">
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.refundAmount') }}</span>
          <input v-model.number="refundForm.amount" type="number" min="0.0001" step="0.0001" class="input mt-1" required />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.referenceNo') }}</span>
          <input v-model.trim="refundForm.reference_no" class="input mt-1" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.note') }}</span>
          <textarea v-model.trim="refundForm.note" class="input mt-1 min-h-[120px]" />
        </label>
        <p class="text-xs text-gray-500 dark:text-dark-400">
          {{ t('admin.distribution.fields.refundHint') }}
        </p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="refundDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('admin.distribution.actions.refundWallet') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="settleDialog" :title="t('admin.distribution.dialogs.settleTitle')" width="normal" @close="settleDialog = false">
      <form class="space-y-4" @submit.prevent="submitSettle">
        <label class="block">
          <span class="input-label">{{ t('admin.distribution.fields.settlementMethod') }}</span>
          <select v-model="settleForm.settlement_method" class="input mt-1">
            <option value="manual">{{ t('admin.distribution.settlementMethods.manual') }}</option>
            <option value="offline">{{ t('admin.distribution.settlementMethods.offline') }}</option>
            <option value="balance">{{ t('admin.distribution.settlementMethods.balance') }}</option>
            <option value="auto">{{ t('admin.distribution.settlementMethods.auto') }}</option>
          </select>
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="settleDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('admin.distribution.actions.settleCommission') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="attributionDialog" :title="t('admin.distribution.dialogs.attributionTitle')" width="normal" @close="attributionDialog = false">
      <form class="space-y-4" @submit.prevent="submitAttribution">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block sm:col-span-2">
            <span class="input-label">{{ t('admin.distribution.fields.channelOrgId') }}</span>
            <div class="relative mt-1">
              <input
                v-model="attributionChannelOrgLookup.keyword"
                type="text"
                class="input pr-8"
                :placeholder="t('admin.distribution.fields.channelOrgIdPlaceholder')"
                required
                @input="scheduleAttributionChannelOrgLookup"
                @focus="attributionChannelOrgLookup.open = true"
              />
              <button
                v-if="attributionForm.channel_org_id > 0"
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                @click="clearAttributionChannelOrgSelection"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
              <div
                v-if="attributionChannelOrgLookup.open && (attributionChannelOrgLookup.results.length > 0 || attributionChannelOrgLookup.keyword)"
                class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
              >
                <div v-if="attributionChannelOrgLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
                <div v-else-if="attributionChannelOrgLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
                <button
                  v-for="organization in attributionChannelOrgLookup.results"
                  :key="organization.id"
                  type="button"
                  class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="selectAttributionChannelOrg(organization)"
                >
                  <div class="font-medium text-gray-900 dark:text-white">{{ organization.name }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    #{{ organization.id }} · {{ organizationTypeLabel(organization.type) }}
                  </div>
                </button>
              </div>
            </div>
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.distribution.fields.referrerMember') }}</span>
            <div class="relative mt-1">
              <input
                v-model="referrerMemberLookup.keyword"
                type="text"
                class="input pr-8"
                :placeholder="t('admin.distribution.fields.referrerMemberPlaceholder')"
                :disabled="attributionForm.channel_org_id <= 0"
                @input="scheduleMemberLookup(referrerMemberLookup, clearReferrerMemberSelection, attributionForm.channel_org_id)"
                @focus="referrerMemberLookup.open = true"
              />
              <button
                v-if="attributionForm.referrer_member_id"
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                @click="clearReferrerMemberSelection"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
              <div
                v-if="referrerMemberLookup.open && (referrerMemberLookup.results.length > 0 || referrerMemberLookup.keyword)"
                class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
              >
                <div v-if="referrerMemberLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
                <div v-else-if="referrerMemberLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
                <button
                  v-for="member in referrerMemberLookup.results"
                  :key="member.member_id"
                  type="button"
                  class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="selectReferrerMember(member)"
                >
                  <div class="font-medium text-gray-900 dark:text-white">{{ member.username || member.user_email }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ member.user_email }} · {{ roleLabel(member.role_type) }}</div>
                </button>
              </div>
            </div>
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.distribution.fields.promotionLink') }}</span>
            <div class="relative mt-1">
              <input
                v-model="attributionPromotionLinkLookup.keyword"
                type="text"
                class="input pr-8"
                :placeholder="t('admin.distribution.fields.promotionLinkPlaceholder')"
                :disabled="attributionForm.channel_org_id <= 0"
                @input="scheduleAttributionPromotionLinkLookup"
                @focus="attributionPromotionLinkLookup.open = true"
              />
              <button
                v-if="attributionForm.promotion_link_id"
                type="button"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                @click="clearAttributionPromotionLinkSelection"
              >
                <Icon name="x" size="sm" :stroke-width="2" />
              </button>
              <div
                v-if="attributionPromotionLinkLookup.open && (attributionPromotionLinkLookup.results.length > 0 || attributionPromotionLinkLookup.keyword)"
                class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
              >
                <div v-if="attributionPromotionLinkLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
                <div v-else-if="attributionPromotionLinkLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
                <button
                  v-for="link in attributionPromotionLinkLookup.results"
                  :key="link.id"
                  type="button"
                  class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  @click="selectAttributionPromotionLink(link)"
                >
                  <div class="font-medium text-gray-900 dark:text-white">{{ link.code }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    {{ link.username || link.user_email }} · {{ roleLabel(link.role_type) }} · {{ targetTypeLabel(link.target_type) }}
                  </div>
                </button>
              </div>
            </div>
          </label>
          <label class="block sm:col-span-2">
            <span class="input-label">{{ t('admin.distribution.fields.note') }}</span>
            <textarea v-model.trim="attributionForm.note" class="input mt-1 min-h-[120px]" />
          </label>
        </div>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="attributionDialog = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="attributionAuditDialog" :title="t('admin.distribution.dialogs.attributionAuditTitle')" width="extra-wide" @close="attributionAuditDialog = false">
      <div class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <div class="text-sm text-gray-500 dark:text-dark-400">
            #{{ selectedAttributionAuditUser?.user_id || '-' }} {{ selectedAttributionAuditUser?.user_email || '-' }}
          </div>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loadingAttributionAudits" @click="reloadAttributionAudits">
            {{ t('common.refresh') }}
          </button>
        </div>
        <div v-if="loadingAttributionAudits" class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="attributionAudits.length === 0" class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.distribution.messages.noAttributionAudits') }}
        </div>
        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.createdAt') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.operator') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.previousChannelOrgId') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.newChannelOrgId') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.previousReferrerMemberId') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.newReferrerMemberId') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.previousPromotionLinkId') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.newPromotionLinkId') }}</th>
                <th class="px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.distribution.columns.note') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr v-for="item in attributionAudits" :key="item.id">
                <td class="px-3 py-2 text-sm text-gray-700 dark:text-dark-300">{{ formatDateTime(item.created_at) }}</td>
                <td class="px-3 py-2 text-sm text-gray-700 dark:text-dark-300">
                  <div class="font-medium text-gray-900 dark:text-white">{{ item.operator_user_email || '-' }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">#{{ item.operator_user_id || '-' }} {{ item.operator_username || '' }}</div>
                </td>
                <td class="px-3 py-2 font-mono text-sm text-gray-700 dark:text-dark-300">{{ item.previous_channel_org_id ? `#${item.previous_channel_org_id}` : '-' }}</td>
                <td class="px-3 py-2 font-mono text-sm text-gray-700 dark:text-dark-300">#{{ item.new_channel_org_id }}</td>
                <td class="px-3 py-2 font-mono text-sm text-gray-700 dark:text-dark-300">{{ item.previous_referrer_member_id ? `#${item.previous_referrer_member_id}` : '-' }}</td>
                <td class="px-3 py-2 font-mono text-sm text-gray-700 dark:text-dark-300">{{ item.new_referrer_member_id ? `#${item.new_referrer_member_id}` : '-' }}</td>
                <td class="px-3 py-2 font-mono text-sm text-gray-700 dark:text-dark-300">{{ item.previous_promotion_link_id ? `#${item.previous_promotion_link_id}` : '-' }}</td>
                <td class="px-3 py-2 font-mono text-sm text-gray-700 dark:text-dark-300">{{ item.new_promotion_link_id ? `#${item.new_promotion_link_id}` : '-' }}</td>
                <td class="px-3 py-2 text-sm text-gray-700 dark:text-dark-300">{{ item.note || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </BaseDialog>
  </AdminDistributionLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute } from 'vue-router'
import { adminAPI, usersAPI } from '@/api/admin'
import AdminDistributionLayout from '@/components/layout/AdminDistributionLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DistributionMemberFormDialog from '@/components/distribution/DistributionMemberFormDialog.vue'
import DistributionLevelsEditor from '@/components/distribution/DistributionLevelsEditor.vue'
import type { DistributionLevelConfig } from '@/api/admin/settings'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import type { AdminUser } from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import {
  buildLevelSelectOptions,
  formatLevelCommissionPercent,
  normalizeDistributionLevelConfigs,
  parseDistributionLevelsFromConfig,
} from '@/utils/distributionLevels'
import { formatDateTime as formatDisplayDateTime } from '@/utils/format'
import {
  createDistributionMember,
  createDistributionOrganization,
  createDistributionPromotionLink,
  listDistributionAttributions,
  listDistributionAttributionAudits,
  listDistributionAlertEvents,
  listDistributionCommissions,
  listDistributionMembers,
  listDistributionOrganizations,
  listDistributionPromotionLinks,
  listDistributionWalletRequests,
  listDistributionWalletTransactions,
  rechargeDistributionWallet,
  reviewDistributionWalletRequest,
  refundDistributionWallet,
  listDistributionWallets,
  reverseDistributionCommission,
  settleDistributionCommission,
  updateDistributionAttribution,
  updateDistributionOrganization,
  updateDistributionWalletWarningThreshold,
  type DistributionAttributionAudit,
  type DistributionAttribution,
  type CreateDistributionMemberRequest,
  type CreateDistributionOrganizationRequest,
  type CreateDistributionPromotionLinkRequest,
  type DistributionAlertEvent,
  type DistributionCommission,
  type DistributionMember,
  type DistributionPromotionLink,
  type DistributionOrganization,
  type DistributionWallet,
  type DistributionWalletRequest,
  type DistributionListParams,
  type DistributionMemberRole,
  type DistributionOrganizationType,
} from '@/api/admin/distribution'

type AdminDistributionTab = 'organizations' | 'members' | 'promotion-links' | 'wallets' | 'alert-events' | 'wallet-requests' | 'wallet-transactions' | 'attributions' | 'commissions'
type TableRow = Record<string, any>
const VALID_TABS: AdminDistributionTab[] = ['organizations', 'members', 'promotion-links', 'wallets', 'alert-events', 'wallet-requests', 'wallet-transactions', 'attributions', 'commissions']

type LookupState<T> = {
  keyword: string
  loading: boolean
  open: boolean
  results: T[]
  selected: T | null
  timer: ReturnType<typeof setTimeout> | null
}

function createLookupState<T>(): LookupState<T> {
  return reactive({
    keyword: '',
    loading: false,
    open: false,
    results: [] as T[],
    selected: null as T | null,
    timer: null as ReturnType<typeof setTimeout> | null,
  }) as LookupState<T>
}

function formatUserLookupLabel(user: Pick<AdminUser, 'username' | 'email'>) {
  return user.username ? `${user.username} (${user.email})` : user.email
}

function formatMemberLookupLabel(member: Pick<DistributionMember, 'username' | 'user_email' | 'role_type'>) {
  const primary = member.username ? `${member.username} (${member.user_email})` : member.user_email
  return `${primary} · ${roleLabel(member.role_type)}`
}

function formatPromotionLinkLookupLabel(link: Pick<DistributionPromotionLink, 'code' | 'username' | 'user_email' | 'target_type' | 'role_type'>) {
  const owner = link.username ? `${link.username} (${link.user_email})` : link.user_email
  return `${link.code} · ${owner} · ${targetTypeLabel(link.target_type)}`
}

function formatOrganizationLookupLabel(org: Pick<DistributionOrganization, 'id' | 'name' | 'type'>) {
  return `${org.name} · #${org.id} · ${org.type}`
}

function usesUserFilter(tab: AdminDistributionTab) {
  return tab === 'members' || tab === 'promotion-links' || tab === 'attributions' || tab === 'commissions'
}

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const rows = ref<TableRow[]>([])
const totals = reactive({ organizations: 0, members: 0, promotionLinks: 0, wallets: 0, alertEvents: 0, walletRequests: 0, walletTransactions: 0, attributions: 0, commissions: 0 })
const walletDialog = ref(false)
const rechargeDialog = ref(false)
const refundDialog = ref(false)
const settleDialog = ref(false)
const attributionDialog = ref(false)
const attributionAuditDialog = ref(false)
const selectedWallet = ref<DistributionWallet | null>(null)
const selectedCommission = ref<DistributionCommission | null>(null)
const selectedAttribution = ref<DistributionAttribution | null>(null)
const selectedAttributionAuditUser = ref<TableRow | null>(null)
const attributionAudits = ref<DistributionAttributionAudit[]>([])
const loadingAttributionAudits = ref(false)
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const filters = reactive<{ channel_org_id?: number | null; user_id?: number | null; role_type: string; transaction_type: string; request_type: string; request_status: string; alert_type: string; alert_status: string; alert_severity: string }>({ channel_org_id: null, user_id: null, role_type: '', transaction_type: '', request_type: '', request_status: '', alert_type: '', alert_status: '', alert_severity: '' })
const organizationDialog = ref(false)
const memberDialog = ref(false)
const linkDialog = ref(false)
const selectedOrganization = ref<DistributionOrganization | null>(null)
const ownerUserLookup = createLookupState<AdminUser>()
const filterUserLookup = createLookupState<AdminUser>()
const filterChannelOrgLookup = createLookupState<DistributionOrganization>()
const memberUserLookup = createLookupState<AdminUser>()
const parentMemberLookup = createLookupState<DistributionMember>()
const channelOrgLookup = createLookupState<DistributionOrganization>()
const attributionChannelOrgLookup = createLookupState<DistributionOrganization>()
const linkMemberLookup = createLookupState<DistributionMember>()
const attributionPromotionLinkLookup = createLookupState<DistributionPromotionLink>()
const referrerMemberLookup = createLookupState<DistributionMember>()

const roleOptions: DistributionMemberRole[] = ['manager', 'agent', 'kol1', 'kol2']

const memberLevelOptions = computed(() =>
  buildLevelSelectOptions(
    organizationLevelsById.value[memberForm.channel_org_id] ?? [],
    globalDistributionLevels.value,
    (level, source) =>
      t('distributionLevels.optionLabel', {
        name: level.name,
        code: level.code,
        rate: formatLevelCommissionPercent(level.commission_rate),
        source: t(`distributionLevels.source.${source}`),
      }),
  ),
)

function cacheOrganizationLevels(org: Pick<DistributionOrganization, 'id' | 'config'>) {
  organizationLevelsById.value[org.id] = parseDistributionLevelsFromConfig(org.config?.distribution_levels)
}

async function loadGlobalDistributionLevels() {
  try {
    const settings = await adminAPI.settings.getSettings()
    globalDistributionLevels.value =
      normalizeDistributionLevelConfigs(settings.distribution_global_levels ?? []) ?? []
  } catch {
    globalDistributionLevels.value = []
  }
}

async function ensureOrganizationLevels(channelOrgId: number) {
  if (channelOrgId <= 0 || organizationLevelsById.value[channelOrgId]) return
  try {
    const result = await listDistributionOrganizations({ page: 1, page_size: 200 })
    for (const org of result.items || []) {
      cacheOrganizationLevels(org)
    }
  } catch {
    organizationLevelsById.value[channelOrgId] = []
  }
}

function handleMemberChannelOrgChange() {
  clearParentMemberSelection()
  memberForm.level_code = ''
  memberForm.commission_rate = 0
  void ensureOrganizationLevels(memberForm.channel_org_id)
}

onMounted(() => {
  void loadGlobalDistributionLevels()
})
const walletRequestTypeOptions = ['recharge', 'refund']
const walletRequestStatusOptions = ['pending', 'approved', 'rejected']
const walletTransactionTypeOptions = ['recharge', 'refund', 'consume', 'commission_reserve', 'commission_release', 'commission_settle', 'commission_deduct', 'commission_refund']
const alertTypeOptions = ['low_balance', 'balance_exhausted', 'consumption_warning', 'consumption_exhausted']
const alertStatusOptions = ['active', 'resolved']
const alertSeverityOptions = ['warning', 'critical']
const activeTab = computed<AdminDistributionTab>(() => {
  const tabParam = Array.isArray(route.params.tab) ? route.params.tab[0] : route.params.tab
  return VALID_TABS.includes(tabParam as AdminDistributionTab) ? (tabParam as AdminDistributionTab) : 'organizations'
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

async function searchAdminUsers(state: LookupState<AdminUser>) {
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
    void searchAdminUsers(state)
  }, 250)
}

async function searchDistributionOrganizationsForLookup(state: LookupState<DistributionOrganization>) {
  const keyword = state.keyword.trim()
  if (!keyword) {
    state.results = []
    return
  }
  state.loading = true
  try {
    const response = await listDistributionOrganizations({
      page: 1,
      page_size: 10,
      q: keyword,
    })
    state.results = response.items || []
  } catch {
    state.results = []
  } finally {
    state.loading = false
  }
}

function scheduleOrganizationLookup(
  state: LookupState<DistributionOrganization>,
  clearSelection: () => void,
) {
  const keyword = state.keyword
  if (state.selected && keyword !== formatOrganizationLookupLabel(state.selected)) {
    clearSelection()
    state.keyword = keyword
    state.open = true
  }
  if (state.timer) clearTimeout(state.timer)
  state.timer = setTimeout(() => {
    void searchDistributionOrganizationsForLookup(state)
  }, 250)
}

function scheduleChannelOrgLookup() {
  scheduleOrganizationLookup(channelOrgLookup, clearChannelOrgSelection)
}

function scheduleFilterChannelOrgLookup() {
  scheduleOrganizationLookup(filterChannelOrgLookup, clearFilterChannelOrgSelection)
}

function scheduleAttributionChannelOrgLookup() {
  scheduleOrganizationLookup(attributionChannelOrgLookup, clearAttributionChannelOrgSelection)
}

function selectChannelOrg(organization: DistributionOrganization) {
  channelOrgLookup.selected = organization
  channelOrgLookup.keyword = formatOrganizationLookupLabel(organization)
  channelOrgLookup.open = false
  memberForm.channel_org_id = organization.id
  cacheOrganizationLevels(organization)
  handleMemberChannelOrgChange()
}

function clearChannelOrgSelection() {
  resetLookupState(channelOrgLookup)
  memberForm.channel_org_id = 0
  memberForm.level_code = ''
  memberForm.commission_rate = 0
}

async function primeOrganizationLookupState(
  state: LookupState<DistributionOrganization>,
  channelOrgId: number,
  onResolved?: (organization: DistributionOrganization | null) => void,
) {
  resetLookupState(state)
  if (channelOrgId <= 0) {
    onResolved?.(null)
    return
  }
  try {
    const response = await listDistributionOrganizations({
      page: 1,
      page_size: 1,
      channel_org_id: channelOrgId,
    })
    const organization = response.items?.[0]
    if (organization) {
      state.selected = organization
      state.keyword = formatOrganizationLookupLabel(organization)
      cacheOrganizationLevels(organization)
      onResolved?.(organization)
      return
    }
  } catch {
    // fall through to keyword fallback
  }
  state.keyword = `#${channelOrgId}`
  onResolved?.(null)
}

async function primeChannelOrgLookup(channelOrgId: number) {
  await primeOrganizationLookupState(channelOrgLookup, channelOrgId, (organization) => {
    memberForm.channel_org_id = organization?.id || channelOrgId
  })
}

async function primeMemberLookupState(
  state: LookupState<DistributionMember>,
  memberId: number | null | undefined,
  channelOrgId?: number,
) {
  resetLookupState(state)
  if (!memberId || memberId <= 0) return
  try {
    const response = await listDistributionMembers({
      page: 1,
      page_size: 20,
      channel_org_id: channelOrgId || undefined,
      q: String(memberId),
    })
    const member = response.items?.find((item) => item.member_id === memberId) || response.items?.[0]
    if (member) {
      state.selected = member
      state.keyword = formatMemberLookupLabel(member)
      return
    }
  } catch {
    // fall through
  }
  state.keyword = `#${memberId}`
}

async function primePromotionLinkLookupState(
  state: LookupState<DistributionPromotionLink>,
  linkId: number | null | undefined,
  channelOrgId?: number,
) {
  resetLookupState(state)
  if (!linkId || linkId <= 0) return
  try {
    const response = await listDistributionPromotionLinks({
      page: 1,
      page_size: 20,
      channel_org_id: channelOrgId || undefined,
      q: String(linkId),
    })
    const link = response.items?.find((item) => item.id === linkId) || response.items?.[0]
    if (link) {
      state.selected = link
      state.keyword = formatPromotionLinkLookupLabel(link)
      return
    }
  } catch {
    // fall through
  }
  state.keyword = `#${linkId}`
}

async function searchDistributionMembersForLookup(state: LookupState<DistributionMember>, channelOrgID?: number) {
  const keyword = state.keyword.trim()
  if (!keyword) {
    state.results = []
    return
  }
  state.loading = true
  try {
    const response = await listDistributionMembers({
      page: 1,
      page_size: 10,
      channel_org_id: channelOrgID || undefined,
      q: keyword,
    })
    state.results = response.items
  } catch {
    state.results = []
  } finally {
    state.loading = false
  }
}

function scheduleMemberLookup(state: LookupState<DistributionMember>, clearSelection: () => void, channelOrgID?: number) {
  const keyword = state.keyword
  if (state.selected && keyword !== formatMemberLookupLabel(state.selected)) {
    clearSelection()
    state.keyword = keyword
    state.open = true
  }
  if (state.timer) clearTimeout(state.timer)
  state.timer = setTimeout(() => {
    void searchDistributionMembersForLookup(state, channelOrgID)
  }, 250)
}

async function primeOwnerUserLookup(userID: number | null | undefined) {
  resetLookupState(ownerUserLookup)
  if (!userID) return
  try {
    const user = await usersAPI.getById(userID)
    ownerUserLookup.selected = user
    ownerUserLookup.keyword = formatUserLookupLabel(user)
  } catch {
    ownerUserLookup.keyword = `#${userID}`
  }
}

const organizationForm = reactive<CreateDistributionOrganizationRequest>({
  type: 'reseller',
  name: '',
  owner_user_id: null,
})

const organizationConfigForm = reactive({
  wholesale_discount_rate: 0.5,
  refund_fee_rate: 0,
  commission_upper_ratio: 0.35,
  management_reward_cap: 0,
  direct_commission_rate_cap: 0,
  channel_commission_rate: 0,
  first_recharge_min_amount: 0,
  recharge_min_amount: 0,
  consumption_limit: 0,
  consumption_warning_threshold: 0,
  recharge_lead_time_days: 0,
  recharge_deadline_note: '',
  max_agent_count: 0,
  max_kol_count: 0,
  max_manager_count: 0,
  user_commission_total_cap: 0,
  member_commission_total_cap: 0,
  team_reward_rate: 0,
  team_reward_threshold: 0,
  freeze_hours: 168,
  kol2_rate: 0.05,
  commission_settlement_method: 'balance',
  distribution_levels: [] as DistributionLevelConfig[],
})

const organizationLevelsEditorRef = ref<InstanceType<typeof DistributionLevelsEditor> | null>(null)
const globalDistributionLevels = ref<DistributionLevelConfig[]>([])
const organizationLevelsById = ref<Record<number, DistributionLevelConfig[]>>({})

const organizationBrandForm = reactive({
  logo_url: '',
  primary_color: '',
  domain: '',
  api_domain: '',
})

const memberForm = reactive<CreateDistributionMemberRequest>({
  channel_org_id: 0,
  user_id: 0,
  role_type: 'agent',
  parent_member_id: null,
  level_code: '',
  commission_rate: 0,
  status: 'active',
})

const linkForm = reactive<CreateDistributionPromotionLinkRequest>({
  member_id: 0,
  code: '',
  target_type: 'registration',
  status: 'active',
})

const walletForm = reactive({
  warning_threshold: 0,
})

const rechargeForm = reactive({
  amount: 0,
  reference_no: '',
  note: '',
})

const refundForm = reactive({
  amount: 0,
  reference_no: '',
  note: '',
})

const settleForm = reactive({
  settlement_method: 'manual' as 'balance' | 'auto' | 'manual' | 'offline',
})

const attributionForm = reactive({
  channel_org_id: 0,
  referrer_member_id: null as number | null,
  promotion_link_id: null as number | null,
  note: '',
})

const columns = computed<Column[]>(() => {
  if (activeTab.value === 'organizations') {
    return [
      { key: 'id', label: t('admin.distribution.columns.id') },
      { key: 'name', label: t('admin.distribution.columns.name') },
      { key: 'type', label: t('admin.distribution.columns.type') },
      { key: 'owner_user_id', label: t('admin.distribution.columns.ownerUserId') },
      { key: 'status', label: t('admin.distribution.columns.status') },
      { key: 'created_at', label: t('admin.distribution.columns.createdAt') },
      { key: 'actions', label: t('common.actions') },
    ]
  }
  if (activeTab.value === 'members') {
    return [
      { key: 'member_id', label: t('admin.distribution.columns.memberId') },
      { key: 'user', label: t('admin.distribution.columns.user') },
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'role_type', label: t('admin.distribution.columns.role') },
      { key: 'level_code', label: t('admin.distribution.columns.levelCode') },
      { key: 'parent_member_id', label: t('admin.distribution.columns.parentMemberId') },
      { key: 'rate', label: t('admin.distribution.columns.commissionRate') },
      { key: 'status', label: t('admin.distribution.columns.status') },
      { key: 'created_at', label: t('admin.distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'promotion-links') {
    return [
      { key: 'id', label: t('admin.distribution.columns.id') },
      { key: 'code', label: t('admin.distribution.columns.code') },
      { key: 'user', label: t('admin.distribution.columns.user') },
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'member_id', label: t('admin.distribution.columns.memberId') },
      { key: 'role_type', label: t('admin.distribution.columns.role') },
      { key: 'target_type', label: t('admin.distribution.columns.targetType') },
      { key: 'status', label: t('admin.distribution.columns.status') },
      { key: 'created_at', label: t('admin.distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'wallets') {
    return [
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'organization_name', label: t('admin.distribution.columns.organizationName') },
      { key: 'organization_type', label: t('admin.distribution.columns.organizationType') },
      { key: 'prepaid_balance', label: t('admin.distribution.columns.prepaidBalance') },
      { key: 'commission_reserved', label: t('admin.distribution.columns.commissionReserved') },
      { key: 'total_recharged', label: t('admin.distribution.columns.totalRecharged') },
      { key: 'total_consumed', label: t('admin.distribution.columns.totalConsumed') },
      { key: 'warning_threshold', label: t('admin.distribution.columns.warningThreshold') },
      { key: 'status', label: t('admin.distribution.columns.status') },
      { key: 'updated_at', label: t('admin.distribution.columns.updatedAt') },
      { key: 'actions', label: t('common.actions') },
    ]
  }
  if (activeTab.value === 'wallet-transactions') {
    return [
      { key: 'id', label: t('admin.distribution.columns.id') },
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'organization_name', label: t('admin.distribution.columns.organizationName') },
      { key: 'transaction_type', label: t('admin.distribution.columns.transactionType') },
      { key: 'amount', label: t('admin.distribution.columns.amount') },
      { key: 'prepaid_balance_before', label: t('admin.distribution.columns.balanceBefore') },
      { key: 'prepaid_balance_after', label: t('admin.distribution.columns.balanceAfter') },
      { key: 'commission_reserved_after', label: t('admin.distribution.columns.reservedAfter') },
      { key: 'reference_no', label: t('admin.distribution.columns.referenceNo') },
      { key: 'note', label: t('admin.distribution.columns.note') },
      { key: 'created_at', label: t('admin.distribution.columns.createdAt') },
    ]
  }
  if (activeTab.value === 'alert-events') {
    return [
      { key: 'id', label: t('admin.distribution.columns.id') },
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'organization_name', label: t('admin.distribution.columns.organizationName') },
      { key: 'alert_type', label: t('admin.distribution.columns.alertType') },
      { key: 'severity', label: t('admin.distribution.columns.severity') },
      { key: 'status', label: t('admin.distribution.columns.status') },
      { key: 'details_summary', label: t('admin.distribution.columns.summary') },
      { key: 'triggered_at', label: t('admin.distribution.columns.triggeredAt') },
      { key: 'resolved_at', label: t('admin.distribution.columns.resolvedAt') },
      { key: 'last_observed_at', label: t('admin.distribution.columns.lastObservedAt') },
    ]
  }
  if (activeTab.value === 'wallet-requests') {
    return [
      { key: 'id', label: t('admin.distribution.columns.id') },
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'organization_name', label: t('admin.distribution.columns.organizationName') },
      { key: 'request_type', label: t('admin.distribution.columns.requestType') },
      { key: 'amount', label: t('admin.distribution.columns.amount') },
      { key: 'status', label: t('admin.distribution.columns.status') },
      { key: 'reference_no', label: t('admin.distribution.columns.referenceNo') },
      { key: 'note', label: t('admin.distribution.columns.note') },
      { key: 'created_at', label: t('admin.distribution.columns.createdAt') },
      { key: 'actions', label: t('common.actions') },
    ]
  }
  if (activeTab.value === 'attributions') {
    return [
      { key: 'user', label: t('admin.distribution.columns.user') },
      { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
      { key: 'referrer_member_id', label: t('admin.distribution.columns.referrerMemberId') },
      { key: 'promotion_link_id', label: t('admin.distribution.columns.promotionLinkId') },
      { key: 'bound_source', label: t('admin.distribution.columns.boundSource') },
      { key: 'bound_at', label: t('admin.distribution.columns.boundAt') },
      { key: 'actions', label: t('common.actions') },
    ]
  }
  return [
    { key: 'id', label: t('admin.distribution.columns.id') },
    { key: 'user', label: t('admin.distribution.columns.user') },
    { key: 'channel_org_id', label: t('admin.distribution.columns.channelOrgId') },
    { key: 'member_id', label: t('admin.distribution.columns.memberId') },
    { key: 'commission_type', label: t('admin.distribution.columns.commissionType') },
    { key: 'base_amount', label: t('admin.distribution.columns.baseAmount') },
    { key: 'rate', label: t('admin.distribution.columns.commissionRate') },
    { key: 'amount', label: t('admin.distribution.columns.amount') },
    { key: 'status', label: t('admin.distribution.columns.status') },
    { key: 'settlement_method', label: t('admin.distribution.columns.settlementMethod') },
    { key: 'frozen_until', label: t('admin.distribution.columns.frozenUntil') },
    { key: 'actions', label: t('common.actions') },
  ]
})

function buildParams(): DistributionListParams {
  return {
    page: pagination.page,
    page_size: pagination.page_size,
    channel_org_id: filters.channel_org_id || undefined,
    user_id: usesUserFilter(activeTab.value) ? filters.user_id || undefined : undefined,
    role_type: activeTab.value === 'members' || activeTab.value === 'promotion-links' ? filters.role_type || undefined : undefined,
    alert_type: activeTab.value === 'alert-events' ? filters.alert_type || undefined : undefined,
    severity: activeTab.value === 'alert-events' ? filters.alert_severity || undefined : undefined,
    request_type: activeTab.value === 'wallet-requests' ? filters.request_type || undefined : undefined,
    status: activeTab.value === 'wallet-requests' ? filters.request_status || undefined : activeTab.value === 'alert-events' ? filters.alert_status || undefined : undefined,
    transaction_type: activeTab.value === 'wallet-transactions' ? filters.transaction_type || undefined : undefined,
  }
}

async function loadActiveTab() {
  loading.value = true
  try {
    const params = buildParams()
    const result =
      activeTab.value === 'organizations'
        ? await listDistributionOrganizations({ page: pagination.page, page_size: pagination.page_size })
        : activeTab.value === 'members'
          ? await listDistributionMembers(params)
          : activeTab.value === 'promotion-links'
            ? await listDistributionPromotionLinks(params)
            : activeTab.value === 'wallets'
              ? await listDistributionWallets(params)
              : activeTab.value === 'alert-events'
                ? await listDistributionAlertEvents(params)
              : activeTab.value === 'wallet-requests'
                ? await listDistributionWalletRequests(params)
              : activeTab.value === 'wallet-transactions'
                ? await listDistributionWalletTransactions(params)
            : activeTab.value === 'attributions'
              ? await listDistributionAttributions(params)
              : await listDistributionCommissions(params)

    rows.value = result.items || []
    if (activeTab.value === 'organizations') {
      for (const item of rows.value) {
        cacheOrganizationLevels(item as DistributionOrganization)
      }
    }
    pagination.total = result.total || 0
    const totalKey = activeTab.value === 'promotion-links' ? 'promotionLinks' : activeTab.value === 'alert-events' ? 'alertEvents' : activeTab.value === 'wallet-requests' ? 'walletRequests' : activeTab.value === 'wallet-transactions' ? 'walletTransactions' : activeTab.value
    if (totalKey in totals) {
      totals[totalKey as keyof typeof totals] = result.total || 0
    }
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

function reloadFromFirstPage() {
  pagination.page = 1
  void loadActiveTab()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadActiveTab()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadActiveTab()
}

function selectFilterUser(user: AdminUser) {
  filterUserLookup.selected = user
  filterUserLookup.keyword = formatUserLookupLabel(user)
  filterUserLookup.open = false
  filters.user_id = user.id
  reloadFromFirstPage()
}

function clearFilterUserSelection() {
  resetLookupState(filterUserLookup)
  filters.user_id = null
}

function selectFilterChannelOrg(organization: DistributionOrganization) {
  filterChannelOrgLookup.selected = organization
  filterChannelOrgLookup.keyword = formatOrganizationLookupLabel(organization)
  filterChannelOrgLookup.open = false
  filters.channel_org_id = organization.id
  cacheOrganizationLevels(organization)
  reloadFromFirstPage()
}

function clearFilterChannelOrgSelection() {
  resetLookupState(filterChannelOrgLookup)
  filters.channel_org_id = null
}

function selectAttributionChannelOrg(organization: DistributionOrganization) {
  attributionChannelOrgLookup.selected = organization
  attributionChannelOrgLookup.keyword = formatOrganizationLookupLabel(organization)
  attributionChannelOrgLookup.open = false
  attributionForm.channel_org_id = organization.id
  cacheOrganizationLevels(organization)
  clearReferrerMemberSelection()
  clearAttributionPromotionLinkSelection()
}

function clearAttributionChannelOrgSelection() {
  resetLookupState(attributionChannelOrgLookup)
  attributionForm.channel_org_id = 0
  clearReferrerMemberSelection()
  clearAttributionPromotionLinkSelection()
}

async function searchPromotionLinksForLookup(
  state: LookupState<DistributionPromotionLink>,
  channelOrgID?: number,
) {
  const keyword = state.keyword.trim()
  if (!keyword) {
    state.results = []
    return
  }
  state.loading = true
  try {
    const response = await listDistributionPromotionLinks({
      page: 1,
      page_size: 10,
      channel_org_id: channelOrgID || undefined,
      q: keyword,
    })
    state.results = response.items || []
  } catch {
    state.results = []
  } finally {
    state.loading = false
  }
}

function scheduleAttributionPromotionLinkLookup() {
  const keyword = attributionPromotionLinkLookup.keyword
  if (
    attributionPromotionLinkLookup.selected &&
    keyword !== formatPromotionLinkLookupLabel(attributionPromotionLinkLookup.selected)
  ) {
    clearAttributionPromotionLinkSelection()
    attributionPromotionLinkLookup.keyword = keyword
    attributionPromotionLinkLookup.open = true
  }
  if (attributionPromotionLinkLookup.timer) clearTimeout(attributionPromotionLinkLookup.timer)
  attributionPromotionLinkLookup.timer = setTimeout(() => {
    void searchPromotionLinksForLookup(attributionPromotionLinkLookup, attributionForm.channel_org_id)
  }, 250)
}

function selectAttributionPromotionLink(link: DistributionPromotionLink) {
  attributionPromotionLinkLookup.selected = link
  attributionPromotionLinkLookup.keyword = formatPromotionLinkLookupLabel(link)
  attributionPromotionLinkLookup.open = false
  attributionForm.promotion_link_id = link.id
}

function clearAttributionPromotionLinkSelection() {
  resetLookupState(attributionPromotionLinkLookup)
  attributionForm.promotion_link_id = null
}

function selectOwnerUser(user: AdminUser) {
  ownerUserLookup.selected = user
  ownerUserLookup.keyword = formatUserLookupLabel(user)
  ownerUserLookup.open = false
  organizationForm.owner_user_id = user.id
}

function clearOwnerUserSelection() {
  resetLookupState(ownerUserLookup)
  organizationForm.owner_user_id = null
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

function selectLinkMember(member: DistributionMember) {
  linkMemberLookup.selected = member
  linkMemberLookup.keyword = formatMemberLookupLabel(member)
  linkMemberLookup.open = false
  linkForm.member_id = member.member_id
}

function clearLinkMemberSelection() {
  resetLookupState(linkMemberLookup)
  linkForm.member_id = 0
}

function selectReferrerMember(member: DistributionMember) {
  referrerMemberLookup.selected = member
  referrerMemberLookup.keyword = formatMemberLookupLabel(member)
  referrerMemberLookup.open = false
  attributionForm.referrer_member_id = member.member_id
}

function clearReferrerMemberSelection() {
  resetLookupState(referrerMemberLookup)
  attributionForm.referrer_member_id = null
}

function openOrganizationDialog(row?: DistributionOrganization) {
  selectedOrganization.value = row || null
  organizationForm.type = row?.type || 'reseller'
  organizationForm.name = row?.name || ''
  organizationForm.owner_user_id = row?.owner_user_id ?? null
  organizationConfigForm.wholesale_discount_rate = Number(row?.config?.wholesale_discount_rate ?? 0.5)
  organizationConfigForm.refund_fee_rate = Number(row?.config?.refund_fee_rate ?? 0)
  organizationConfigForm.commission_upper_ratio = Number(row?.config?.commission_upper_ratio ?? 0.35)
  organizationConfigForm.management_reward_cap = Number(row?.config?.management_reward_cap ?? 0)
  organizationConfigForm.direct_commission_rate_cap = Number(row?.config?.direct_commission_rate_cap ?? 0)
  organizationConfigForm.channel_commission_rate = Number(row?.config?.channel_commission_rate ?? 0)
  organizationConfigForm.first_recharge_min_amount = Number(row?.config?.first_recharge_min_amount ?? 0)
  organizationConfigForm.recharge_min_amount = Number(row?.config?.recharge_min_amount ?? 0)
  organizationConfigForm.consumption_limit = Number(row?.config?.consumption_limit ?? 0)
  organizationConfigForm.consumption_warning_threshold = Number(row?.config?.consumption_warning_threshold ?? 0)
  organizationConfigForm.recharge_lead_time_days = Number(row?.config?.recharge_lead_time_days ?? 0)
  organizationConfigForm.recharge_deadline_note = String(row?.config?.recharge_deadline_note ?? '')
  organizationConfigForm.max_agent_count = Number(row?.config?.max_agent_count ?? 0)
  organizationConfigForm.max_kol_count = Number(row?.config?.max_kol_count ?? 0)
  organizationConfigForm.max_manager_count = Number(row?.config?.max_manager_count ?? 0)
  organizationConfigForm.user_commission_total_cap = Number(row?.config?.user_commission_total_cap ?? 0)
  organizationConfigForm.member_commission_total_cap = Number(row?.config?.member_commission_total_cap ?? 0)
  organizationConfigForm.team_reward_rate = Number(row?.config?.team_reward_rate ?? 0)
  organizationConfigForm.team_reward_threshold = Number(row?.config?.team_reward_threshold ?? 0)
  organizationConfigForm.freeze_hours = Number(row?.config?.freeze_hours ?? 168)
  organizationConfigForm.kol2_rate = Number(row?.config?.kol2_rate ?? 0.05)
  organizationConfigForm.commission_settlement_method = String(row?.config?.commission_settlement_method ?? 'balance')
  organizationConfigForm.distribution_levels = parseDistributionLevelsFromConfig(row?.config?.distribution_levels)
  if (row?.id) {
    cacheOrganizationLevels(row)
  }
  organizationBrandForm.logo_url = String(row?.brand_config?.logo_url ?? '')
  organizationBrandForm.primary_color = String(row?.brand_config?.primary_color ?? row?.brand_config?.theme_color ?? '')
  organizationBrandForm.domain = String(row?.brand_config?.domain ?? '')
  organizationBrandForm.api_domain = String(row?.brand_config?.api_domain ?? '')
  organizationDialog.value = true
  void primeOwnerUserLookup(organizationForm.owner_user_id)
}

function openMemberDialog() {
  memberForm.channel_org_id = 0
  memberForm.user_id = 0
  memberForm.role_type = 'agent'
  memberForm.parent_member_id = null
  memberForm.level_code = ''
  memberForm.commission_rate = 0
  memberForm.status = 'active'
  resetLookupState(memberUserLookup)
  resetLookupState(parentMemberLookup)
  resetLookupState(channelOrgLookup)
  const presetChannelOrgId = filters.channel_org_id || 0
  if (presetChannelOrgId > 0) {
    memberForm.channel_org_id = presetChannelOrgId
    void primeChannelOrgLookup(presetChannelOrgId)
  }
  void ensureOrganizationLevels(memberForm.channel_org_id)
  memberDialog.value = true
}

function openLinkDialog() {
  linkForm.member_id = 0
  linkForm.code = ''
  linkForm.target_type = 'registration'
  linkForm.status = 'active'
  resetLookupState(linkMemberLookup)
  linkDialog.value = true
}

function openWalletDialog(row: DistributionWallet) {
  selectedWallet.value = row
  walletForm.warning_threshold = row.warning_threshold || 0
  walletDialog.value = true
}

function openRechargeDialog(row: DistributionWallet) {
  selectedWallet.value = row
  rechargeForm.amount = 0
  rechargeForm.reference_no = ''
  rechargeForm.note = ''
  rechargeDialog.value = true
}

function openRefundDialog(row: DistributionWallet) {
  selectedWallet.value = row
  refundForm.amount = 0
  refundForm.reference_no = ''
  refundForm.note = ''
  refundDialog.value = true
}

function openSettleDialog(row: DistributionCommission) {
  selectedCommission.value = row
  settleForm.settlement_method = 'manual'
  settleDialog.value = true
}

function openAttributionDialog(row: DistributionAttribution) {
  selectedAttribution.value = row
  attributionForm.channel_org_id = row.channel_org_id
  attributionForm.referrer_member_id = row.referrer_member_id ?? null
  attributionForm.promotion_link_id = row.promotion_link_id ?? null
  attributionForm.note = ''
  void primeOrganizationLookupState(attributionChannelOrgLookup, row.channel_org_id, (organization) => {
    attributionForm.channel_org_id = organization?.id || row.channel_org_id
  })
  void primeMemberLookupState(referrerMemberLookup, row.referrer_member_id, row.channel_org_id)
  void primePromotionLinkLookupState(attributionPromotionLinkLookup, row.promotion_link_id, row.channel_org_id)
  attributionDialog.value = true
}

async function loadAttributionAudits(userID: number) {
  loadingAttributionAudits.value = true
  try {
    const result = await listDistributionAttributionAudits({
      user_id: userID,
      page: 1,
      page_size: 20,
    })
    attributionAudits.value = result.items || []
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.loadFailed')))
  } finally {
    loadingAttributionAudits.value = false
  }
}

function openAttributionAuditDialog(row: TableRow) {
  selectedAttributionAuditUser.value = row
  attributionAudits.value = []
  attributionAuditDialog.value = true
  void loadAttributionAudits(Number(row.user_id || 0))
}

function reloadAttributionAudits() {
  const userID = Number(selectedAttributionAuditUser.value?.user_id || 0)
  if (userID <= 0) return
  void loadAttributionAudits(userID)
}

async function submitOrganization() {
  saving.value = true
  try {
    let distributionLevels: DistributionLevelConfig[] = []
    if (organizationForm.type !== 'platform') {
      if (!organizationLevelsEditorRef.value?.validate()) {
        appStore.showError(t('admin.distribution.errors.levelsValidationError'))
        return
      }
      distributionLevels =
        normalizeDistributionLevelConfigs(organizationConfigForm.distribution_levels) ?? []
    }
    const payload = {
      type: organizationForm.type,
      name: organizationForm.name,
      owner_user_id: organizationForm.owner_user_id || undefined,
      config: {
        wholesale_discount_rate: organizationConfigForm.wholesale_discount_rate,
        refund_fee_rate: organizationConfigForm.refund_fee_rate,
        commission_upper_ratio: organizationConfigForm.commission_upper_ratio,
        management_reward_cap: organizationConfigForm.management_reward_cap,
        direct_commission_rate_cap: organizationConfigForm.direct_commission_rate_cap,
        channel_commission_rate: organizationConfigForm.channel_commission_rate,
        first_recharge_min_amount: organizationConfigForm.first_recharge_min_amount,
        recharge_min_amount: organizationConfigForm.recharge_min_amount,
        consumption_limit: organizationConfigForm.consumption_limit,
        consumption_warning_threshold: organizationConfigForm.consumption_warning_threshold,
        recharge_lead_time_days: organizationConfigForm.recharge_lead_time_days,
        recharge_deadline_note: organizationConfigForm.recharge_deadline_note,
        max_agent_count: organizationConfigForm.max_agent_count,
        max_kol_count: organizationConfigForm.max_kol_count,
        max_manager_count: organizationConfigForm.max_manager_count,
        user_commission_total_cap: organizationConfigForm.user_commission_total_cap,
        member_commission_total_cap: organizationConfigForm.member_commission_total_cap,
        team_reward_rate: organizationConfigForm.team_reward_rate,
        team_reward_threshold: organizationConfigForm.team_reward_threshold,
        freeze_hours: organizationConfigForm.freeze_hours,
        kol2_rate: organizationConfigForm.kol2_rate,
        commission_settlement_method: organizationConfigForm.commission_settlement_method,
        distribution_levels: distributionLevels,
      },
      brand_config: {
        logo_url: organizationBrandForm.logo_url || '',
        primary_color: organizationBrandForm.primary_color || '',
        domain: organizationBrandForm.domain || '',
        api_domain: organizationBrandForm.api_domain || '',
      },
    }
    if (selectedOrganization.value) {
      const updated = await updateDistributionOrganization(selectedOrganization.value.id, payload)
      cacheOrganizationLevels(updated)
      appStore.showSuccess(t('admin.distribution.messages.organizationUpdated'))
    } else {
      const created = await createDistributionOrganization(payload)
      cacheOrganizationLevels(created)
      appStore.showSuccess(t('admin.distribution.messages.organizationCreated'))
    }
    organizationDialog.value = false
    selectedOrganization.value = null
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.createFailed')))
  } finally {
    saving.value = false
  }
}

async function submitMember() {
  if (memberForm.channel_org_id <= 0) {
    appStore.showError(t('admin.distribution.messages.channelOrgRequired'))
    return
  }
  if (memberForm.user_id <= 0) {
    appStore.showError(t('admin.distribution.fields.userId'))
    return
  }
  if (memberForm.role_type === 'agent' && !memberForm.level_code) {
    appStore.showError(t('distributionLevels.selectPlaceholder'))
    return
  }
  saving.value = true
  try {
    await createDistributionMember({
      channel_org_id: memberForm.channel_org_id,
      user_id: memberForm.user_id,
      role_type: memberForm.role_type,
      parent_member_id: memberForm.parent_member_id || undefined,
      level_code: memberForm.role_type === 'agent' ? memberForm.level_code || undefined : undefined,
      commission_rate: memberForm.commission_rate,
      status: memberForm.status,
    })
    appStore.showSuccess(t('admin.distribution.messages.memberCreated'))
    memberDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.createFailed')))
  } finally {
    saving.value = false
  }
}

async function submitLink() {
  if (linkForm.member_id <= 0) {
    appStore.showError(t('admin.distribution.fields.memberId'))
    return
  }
  saving.value = true
  try {
    await createDistributionPromotionLink({
      member_id: linkForm.member_id,
      code: linkForm.code || undefined,
      target_type: linkForm.target_type,
      status: linkForm.status,
    })
    appStore.showSuccess(t('admin.distribution.messages.linkCreated'))
    linkDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.createFailed')))
  } finally {
    saving.value = false
  }
}

async function submitWallet() {
  if (!selectedWallet.value) return
  saving.value = true
  try {
    await updateDistributionWalletWarningThreshold(selectedWallet.value.channel_org_id, walletForm.warning_threshold)
    appStore.showSuccess(t('admin.distribution.messages.walletUpdated'))
    walletDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

async function submitRecharge() {
  if (!selectedWallet.value) return
  saving.value = true
  try {
    await rechargeDistributionWallet(selectedWallet.value.channel_org_id, {
      amount: rechargeForm.amount,
      reference_no: rechargeForm.reference_no || undefined,
      note: rechargeForm.note || undefined,
    })
    appStore.showSuccess(t('admin.distribution.messages.walletRecharged'))
    rechargeDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

async function submitRefund() {
  if (!selectedWallet.value) return
  saving.value = true
  try {
    const result = await refundDistributionWallet(selectedWallet.value.channel_org_id, {
      amount: refundForm.amount,
      reference_no: refundForm.reference_no || undefined,
      note: refundForm.note || undefined,
    })
    appStore.showSuccess(
      t('admin.distribution.messages.walletRefunded', {
        netAmount: formatAmount(result.net_amount),
        feeAmount: formatAmount(result.fee_amount),
      }),
    )
    refundDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

async function reviewWalletRequest(row: DistributionWalletRequest | TableRow, action: 'approve' | 'reject') {
  saving.value = true
  try {
    await reviewDistributionWalletRequest(Number(row.id), {
      action,
      review_note: '',
    })
    appStore.showSuccess(t('admin.distribution.messages.walletRequestReviewed'))
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

async function submitSettle() {
  if (!selectedCommission.value) return
  saving.value = true
  try {
    await settleDistributionCommission(selectedCommission.value.id, settleForm.settlement_method)
    appStore.showSuccess(t('admin.distribution.messages.commissionSettled'))
    settleDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

async function submitAttribution() {
  if (!selectedAttribution.value) return
  if (attributionForm.channel_org_id <= 0) {
    appStore.showError(t('admin.distribution.messages.channelOrgRequired'))
    return
  }
  saving.value = true
  try {
    await updateDistributionAttribution(selectedAttribution.value.user_id, {
      channel_org_id: attributionForm.channel_org_id,
      referrer_member_id: attributionForm.referrer_member_id || undefined,
      promotion_link_id: attributionForm.promotion_link_id || undefined,
      note: attributionForm.note || undefined,
    })
    appStore.showSuccess(t('admin.distribution.messages.attributionUpdated'))
    attributionDialog.value = false
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

async function reverseCommission(row: DistributionCommission) {
  saving.value = true
  try {
    await reverseDistributionCommission(row.id)
    appStore.showSuccess(t('admin.distribution.messages.commissionReversed'))
    await loadActiveTab()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.distribution.errors', t('admin.distribution.errors.updateFailed')))
  } finally {
    saving.value = false
  }
}

function organizationTypeLabel(type: DistributionOrganizationType | string) {
  return t(`admin.distribution.organizationTypes.${type}`, type)
}

function roleLabel(role: DistributionMemberRole | string) {
  return t(`admin.distribution.roles.${role}`, role)
}

function targetTypeLabel(targetType: string) {
  return t(`admin.distribution.targetTypes.${targetType}`, targetType)
}

function commissionTypeLabel(commissionType: string) {
  return t(`admin.distribution.commissionTypes.${commissionType}`, commissionType || '-')
}

function transactionTypeLabel(transactionType: string) {
  return t(`distribution.transactionTypes.${transactionType}`, transactionType || '-')
}

function walletRequestTypeLabel(requestType: string) {
  return t(`admin.distribution.requestTypes.${requestType}`, requestType || '-')
}

function walletRequestStatusLabel(status: string) {
  return t(`admin.distribution.requestStatuses.${status}`, status || '-')
}

function alertTypeLabel(alertType: string) {
  return t(`admin.distribution.alertTypes.${alertType}`, alertType || '-')
}

function alertSeverityLabel(severity: string) {
  return t(`admin.distribution.severities.${severity}`, severity || '-')
}

function alertStatusLabel(status: string) {
  return t(`admin.distribution.alertStatuses.${status}`, status || '-')
}

function settlementMethodLabel(settlementMethod: string) {
  return t(`admin.distribution.settlementMethods.${settlementMethod}`, settlementMethod || '-')
}

function statusLabel(status: string) {
  if (status === 'active' || status === 'resolved') {
    return status === 'resolved' ? alertStatusLabel(status) : t(`admin.distribution.statuses.${status}`, status || '-')
  }
  if (status === 'warning' || status === 'critical' || status === 'info') {
    return alertSeverityLabel(status)
  }
  if (status === 'active' || status === 'resolved') {
    return alertStatusLabel(status)
  }
  return t(`admin.distribution.statuses.${status}`, status || '-')
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

function formatDateTime(value: string | null | undefined) {
  return value ? formatDisplayDateTime(value) : '-'
}

function alertSummary(row: DistributionAlertEvent | TableRow) {
  const details = row.details && typeof row.details === 'object' ? row.details as Record<string, unknown> : {}
  const note = String(details.recharge_deadline_note || '')
  switch (row.alert_type) {
    case 'low_balance':
      return t('admin.distribution.alertSummaries.lowBalance', {
        balance: formatAmount(Number(details.prepaid_balance || 0)),
        threshold: formatAmount(Number(details.warning_threshold || 0)),
        note,
      })
    case 'balance_exhausted':
      return t('admin.distribution.alertSummaries.balanceExhausted', {
        available: formatAmount(Number(details.available_balance || 0)),
        note,
      })
    case 'consumption_warning':
      return t('admin.distribution.alertSummaries.consumptionWarning', {
        remaining: formatAmount(Number(details.remaining_consumption || 0)),
        limit: formatAmount(Number(details.consumption_limit || 0)),
        note,
      })
    case 'consumption_exhausted':
      return t('admin.distribution.alertSummaries.consumptionExhausted', {
        consumed: formatAmount(Number(details.total_consumed || 0)),
        limit: formatAmount(Number(details.consumption_limit || 0)),
        note,
      })
    default:
      return note || '-'
  }
}

watch(activeTab, () => {
  rows.value = []
  pagination.page = 1
  pagination.total = 0
  if (!usesUserFilter(activeTab.value)) {
    clearFilterUserSelection()
  }
  if (activeTab.value !== 'members' && activeTab.value !== 'promotion-links') filters.role_type = ''
  if (activeTab.value !== 'alert-events') {
    filters.alert_type = ''
    filters.alert_status = ''
    filters.alert_severity = ''
  }
  if (activeTab.value !== 'wallet-requests') {
    filters.request_type = ''
    filters.request_status = ''
  }
  if (activeTab.value !== 'wallet-transactions') filters.transaction_type = ''
  void loadActiveTab()
}, { immediate: true })

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
