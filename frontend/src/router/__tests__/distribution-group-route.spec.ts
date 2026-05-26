import { describe, expect, it, vi } from 'vitest'

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: false,
  isAdmin: false,
  isSimpleMode: false,
}))

const appStore = vi.hoisted(() => ({
  siteName: 'Sub2API',
  backendModeEnabled: false,
  cachedPublicSettings: null as null | Record<string, unknown>,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    customMenuItems: [],
  }),
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn(),
    isLoading: { value: false },
  }),
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn(),
    cancelPendingPrefetch: vi.fn(),
    resetPrefetchState: vi.fn(),
  }),
}))

describe('distribution group routes', () => {
  it('redirects user distribution group URLs to the real distribution page', async () => {
    const { default: router } = await import('@/router')

    const redirects: Array<[string, string]> = [
      ['/distribution/group/organization-management', '/distribution#members'],
      ['/distribution/group/promotion-management', '/distribution#promotion-links'],
      ['/distribution/group/commission-settlement', '/distribution#wallet'],
      ['/distribution/group/risk-alerts', '/distribution#alert-events'],
    ]

    for (const [path, target] of redirects) {
      const route = router.getRoutes().find((record) => record.path === path)
      expect(route?.redirect).toBe(target)
    }
  })

  it('redirects admin distribution group URLs to the matching admin tab pages', async () => {
    const { default: router } = await import('@/router')

    const redirects: Array<[string, string]> = [
      ['/admin/distribution/group/organization-management', '/admin/distribution/organizations'],
      ['/admin/distribution/group/promotion-management', '/admin/distribution/promotion-links'],
      ['/admin/distribution/group/commission-settlement', '/admin/distribution/wallets'],
      ['/admin/distribution/group/risk-alerts', '/admin/distribution/alert-events'],
    ]

    for (const [path, target] of redirects) {
      const route = router.getRoutes().find((record) => record.path === path)
      expect(route?.redirect).toBe(target)
    }
  })
})
