import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from '@/stores'
import { isAdminDistributionRoutePath } from '@/nav/adminDistributionNav'
import { isDistributionRoutePath } from '@/nav/distributionNav'

describe('useNavWorkspace helpers', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('detects distribution routes', () => {
    expect(isDistributionRoutePath('/distribution')).toBe(true)
    expect(isDistributionRoutePath('/distribution/overview')).toBe(true)
    expect(isDistributionRoutePath('/dashboard')).toBe(false)
  })

  it('persists workspace in the app store', () => {
    const appStore = useAppStore()
    appStore.setNavWorkspace('distribution')
    expect(appStore.navWorkspace).toBe('distribution')
    expect(localStorage.getItem('sub2api-nav-workspace')).toBe('distribution')
  })

  it('detects admin distribution routes', () => {
    expect(isAdminDistributionRoutePath('/admin/distribution')).toBe(true)
    expect(isAdminDistributionRoutePath('/admin/distribution/organizations')).toBe(true)
    expect(isAdminDistributionRoutePath('/admin/dashboard')).toBe(false)
  })

  it('persists admin workspace separately', () => {
    const appStore = useAppStore()
    appStore.setAdminNavWorkspace('distribution')
    expect(appStore.adminNavWorkspace).toBe('distribution')
    expect(localStorage.getItem('sub2api-admin-nav-workspace')).toBe('distribution')
  })
})
