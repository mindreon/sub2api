import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { isAdminDistributionRoutePath } from '@/nav/adminDistributionNav'
import { isDistributionRoutePath } from '@/nav/distributionNav'
import type { NavWorkspace } from '@/nav/types'

export type NavWorkspaceScope = 'user' | 'admin'

const USER_PERSONAL_PREFIXES = [
  '/dashboard',
  '/keys',
  '/usage',
  '/available-channels',
  '/monitor',
  '/subscriptions',
  '/purchase',
  '/orders',
  '/redeem',
  '/affiliate',
  '/profile',
  '/custom/',
]

function isUserPersonalRoutePath(path: string): boolean {
  if (isDistributionRoutePath(path)) {
    return true
  }
  return USER_PERSONAL_PREFIXES.some((prefix) => path === prefix || path.startsWith(prefix))
}

export function useNavWorkspace(scope: NavWorkspaceScope = 'user') {
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const route = useRoute()
  const router = useRouter()

  const workspace = computed(() => (scope === 'admin' ? appStore.adminNavWorkspace : appStore.navWorkspace))
  const isDistributionWorkspace = computed(() => workspace.value === 'distribution')

  function setWorkspace(next: NavWorkspace): void {
    if (scope === 'admin') {
      appStore.setAdminNavWorkspace(next)
      return
    }
    appStore.setNavWorkspace(next)
  }

  function syncWorkspaceFromRoute(path = route.path): void {
    if (!isDistributionRoutePath(path)) {
      return
    }
    appStore.setNavWorkspace('distribution')
  }

  function syncAdminWorkspaceFromRoute(path = route.path): void {
    if (isAdminDistributionRoutePath(path)) {
      appStore.setAdminNavWorkspace('distribution')
      return
    }
    if (path.startsWith('/admin')) {
      appStore.setAdminNavWorkspace('consumer')
    }
  }

  function syncAllWorkspacesFromRoute(path = route.path): void {
    syncAdminWorkspaceFromRoute(path)
    if (isDistributionRoutePath(path)) {
      appStore.setNavWorkspace('distribution')
      return
    }
    if (isUserPersonalRoutePath(path)) {
      appStore.setNavWorkspace('consumer')
    }
  }

  async function switchWorkspace(next: NavWorkspace): Promise<void> {
    if (next === workspace.value) {
      return
    }

    setWorkspace(next)

    if (scope === 'admin') {
      if (next === 'distribution') {
        if (!isAdminDistributionRoutePath(route.path)) {
          await router.push('/admin/distribution/organizations')
        }
        return
      }

      if (isAdminDistributionRoutePath(route.path)) {
        await router.push('/admin/dashboard')
      }
      return
    }

    if (next === 'distribution') {
      if (!isDistributionRoutePath(route.path)) {
        await router.push('/distribution')
      }
      return
    }

    if (isDistributionRoutePath(route.path)) {
      await router.push(authStore.isAdmin ? '/keys' : '/dashboard')
    }
  }

  return {
    workspace,
    isDistributionWorkspace,
    setWorkspace,
    syncWorkspaceFromRoute,
    syncAdminWorkspaceFromRoute,
    syncAllWorkspacesFromRoute,
    switchWorkspace,
  }
}
