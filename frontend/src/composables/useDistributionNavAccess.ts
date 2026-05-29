import { computed, ref } from 'vue'
import { getDistributionOverview, listMyDistributionMembers } from '@/api/distribution'
import type { DistributionMember } from '@/api/admin/distribution'
import {
  canAccessDistributionPromotionNav,
  canManageDistributionMembersNav,
  filterMyActivePromoterMembers,
} from '@/utils/distributionPromoter'

const hasDistributionAccess = ref(false)
const hasPersonalDistributionAccess = ref(false)
const canManageDistributionChannel = ref(false)
const myPromoterMembers = ref<DistributionMember[]>([])
let syncPromise: Promise<void> | null = null

export function useDistributionNavAccess() {
  const hasPromoterAccess = computed(() => myPromoterMembers.value.length > 0)
  const canAccessPromotionNav = computed(() =>
    canAccessDistributionPromotionNav(canManageDistributionChannel.value, myPromoterMembers.value),
  )
  const canManageMembersNav = computed(() =>
    canManageDistributionMembersNav(canManageDistributionChannel.value, myPromoterMembers.value),
  )
  const canAccessChannelFinanceNav = computed(() => canManageDistributionChannel.value)

  async function syncDistributionNavAccess(isAuthenticated: boolean, isAdminUser: boolean): Promise<void> {
    if (!isAuthenticated) {
      hasDistributionAccess.value = false
      hasPersonalDistributionAccess.value = false
      canManageDistributionChannel.value = false
      myPromoterMembers.value = []
      return
    }

    if (syncPromise) {
      await syncPromise
      return
    }

    syncPromise = (async () => {
      try {
        const overview = await getDistributionOverview()
        const hasChannel = overview.channel_org_id > 0
        canManageDistributionChannel.value = hasChannel && overview.can_manage_channel

        if (!hasChannel) {
          hasDistributionAccess.value = false
          hasPersonalDistributionAccess.value = false
          myPromoterMembers.value = []
          return
        }

        try {
          const members = await listMyDistributionMembers({ page: 1, page_size: 100 })
          myPromoterMembers.value = filterMyActivePromoterMembers(members.items || [], overview.user_id)
        } catch {
          myPromoterMembers.value = []
        }

        if (isAdminUser) {
          hasPersonalDistributionAccess.value = true
          hasDistributionAccess.value = false
        } else {
          hasDistributionAccess.value = true
          hasPersonalDistributionAccess.value = false
        }
      } catch {
        hasDistributionAccess.value = false
        hasPersonalDistributionAccess.value = false
        canManageDistributionChannel.value = false
        myPromoterMembers.value = []
      } finally {
        syncPromise = null
      }
    })()

    await syncPromise
  }

  return {
    hasDistributionAccess,
    hasPersonalDistributionAccess,
    canManageDistributionChannel,
    myPromoterMembers,
    hasPromoterAccess,
    canAccessPromotionNav,
    canManageMembersNav,
    canAccessChannelFinanceNav,
    syncDistributionNavAccess,
  }
}
