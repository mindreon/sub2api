<template>
  <BaseDialog :show="show" :title="title" width="normal" @close="emit('close')">
    <form class="space-y-4" @submit.prevent="emit('submit')">
      <div class="grid gap-4 sm:grid-cols-2">
        <label v-if="showChannelOrgField" class="block">
          <span class="input-label">{{ t(fieldKey('channelOrgId')) }}</span>
          <input
            v-model.number="memberForm.channel_org_id"
            type="number"
            min="1"
            class="input mt-1"
            required
            @change="emit('channel-org-change')"
          />
        </label>
        <label class="block">
          <span class="input-label">{{ t(fieldKey('userId')) }}</span>
          <div class="relative mt-1">
            <input
              v-model="memberUserLookup.keyword"
              type="text"
              class="input pr-8"
              :placeholder="t(userSearchPlaceholderKey)"
              @input="emit('member-user-input')"
              @focus="emit('member-user-focus')"
            />
            <button
              v-if="memberForm.user_id > 0"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              @click="emit('clear-member-user')"
            >
              <Icon name="x" size="sm" :stroke-width="2" />
            </button>
            <div
              v-if="memberUserLookup.open && (memberUserLookup.results.length > 0 || memberUserLookup.keyword)"
              class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
            >
              <div v-if="memberUserLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
              <div v-else-if="memberUserLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
              <button
                v-for="user in memberUserLookup.results"
                :key="user.id"
                type="button"
                class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                @click="emit('select-member-user', user)"
              >
                <div class="font-medium text-gray-900 dark:text-white">{{ user.username || user.email }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ user.email }} · #{{ user.id }}</div>
              </button>
            </div>
          </div>
        </label>
        <label class="block">
          <span class="input-label">{{ t(`${namespace}.${roleFieldKey}`) }}</span>
          <select v-model="memberForm.role_type" class="input mt-1">
            <option v-for="role in roleOptions" :key="role" :value="role">{{ t(`${namespace}.roles.${role}`) }}</option>
          </select>
        </label>
        <label v-if="!hideParentFieldForAgent || memberForm.role_type !== 'agent'" class="block">
          <span class="input-label">{{ t(parentMemberLabelKey) }}</span>
          <div class="relative mt-1">
            <input
              v-model="parentMemberLookup.keyword"
              type="text"
              class="input pr-8"
              :placeholder="t(parentSearchPlaceholderKey)"
              :disabled="disableParentLookup"
              :required="parentFieldRequiredForNonAgent && memberForm.role_type !== 'agent'"
              @input="emit('parent-member-input')"
              @focus="emit('parent-member-focus')"
            />
            <button
              v-if="memberForm.parent_member_id"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              @click="emit('clear-parent-member')"
            >
              <Icon name="x" size="sm" :stroke-width="2" />
            </button>
            <div
              v-if="parentMemberLookup.open && (parentMemberLookup.results.length > 0 || parentMemberLookup.keyword)"
              class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
            >
              <div v-if="parentMemberLookup.loading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
              <div v-else-if="parentMemberLookup.results.length === 0" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.noOptionsFound') }}</div>
              <button
                v-for="member in parentMemberLookup.results"
                :key="member.member_id"
                type="button"
                class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                @click="emit('select-parent-member', member)"
              >
                <div class="font-medium text-gray-900 dark:text-white">{{ member.username || member.user_email }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  {{ member.user_email }} · #{{ member.member_id }}<template v-if="member.role_type"> · {{ t(`${namespace}.roles.${member.role_type}`) }}</template>
                </div>
              </button>
            </div>
          </div>
        </label>
        <label class="block">
          <span class="input-label">{{ t(fieldKey('levelCode')) }}</span>
          <input v-model.trim="memberForm.level_code" class="input mt-1" :placeholder="levelCodePlaceholderKey ? t(levelCodePlaceholderKey) : undefined" />
          <p v-if="levelCodeDescriptionKey" class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t(levelCodeDescriptionKey) }}
          </p>
        </label>
        <label class="block">
          <span class="input-label">{{ t(fieldKey('commissionRate')) }}</span>
          <input v-model.number="memberForm.commission_rate" type="number" min="0" max="1" step="0.0001" class="input mt-1" required />
        </label>
        <label class="block">
          <span class="input-label">{{ t(fieldKey('status')) }}</span>
          <select v-model="memberForm.status" class="input mt-1">
            <option value="active">{{ t(`${namespace}.statuses.active`) }}</option>
            <option value="inactive">{{ t(`${namespace}.statuses.inactive`) }}</option>
            <option value="disabled">{{ t(`${namespace}.statuses.disabled`) }}</option>
          </select>
        </label>
      </div>
      <div class="flex justify-end gap-2">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? t('common.saving') : t('common.create') }}</button>
      </div>
    </form>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUser } from '@/types'
import type { DistributionMember, DistributionMemberRole } from '@/api/admin/distribution'

type DialogNamespace = 'distribution' | 'admin.distribution'

type LookupState<T> = {
  keyword: string
  loading: boolean
  open: boolean
  results: T[]
  selected: T | null
}

type MemberFormModel = {
  channel_org_id?: number
  user_id: number
  role_type: DistributionMemberRole
  parent_member_id?: number | null
  level_code?: string
  commission_rate: number
  status?: 'active' | 'inactive' | 'disabled'
}

const props = withDefaults(defineProps<{
  show: boolean
  title: string
  saving: boolean
  namespace: DialogNamespace
  userSearchPlaceholderKey: string
  parentSearchPlaceholderKey: string
  roleFieldKey: string
  parentMemberLabelKey?: string
  levelCodePlaceholderKey?: string
  levelCodeDescriptionKey?: string
  showChannelOrgField?: boolean
  hideParentFieldForAgent?: boolean
  parentFieldRequiredForNonAgent?: boolean
  disableParentLookup?: boolean
  memberForm: MemberFormModel
  roleOptions: DistributionMemberRole[]
  memberUserLookup: LookupState<AdminUser>
  parentMemberLookup: LookupState<DistributionMember>
}>(), {
  parentMemberLabelKey: undefined,
  levelCodePlaceholderKey: undefined,
  levelCodeDescriptionKey: undefined,
  showChannelOrgField: false,
  hideParentFieldForAgent: false,
  parentFieldRequiredForNonAgent: false,
  disableParentLookup: false,
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit'): void
  (e: 'channel-org-change'): void
  (e: 'member-user-input'): void
  (e: 'member-user-focus'): void
  (e: 'clear-member-user'): void
  (e: 'select-member-user', user: AdminUser): void
  (e: 'parent-member-input'): void
  (e: 'parent-member-focus'): void
  (e: 'clear-parent-member'): void
  (e: 'select-parent-member', member: DistributionMember): void
}>()

const { t } = useI18n()

function fieldKey(name: string) {
  return `${props.namespace}.fields.${name}`
}
</script>
