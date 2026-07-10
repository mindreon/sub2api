import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar workspace navigation', () => {
  it('uses workspace switching for user distribution navigation', () => {
    expect(componentSource).toContain('showUserWorkspaceSwitcher')
    expect(componentSource).toContain('activeUserNavItems')
    expect(componentSource).toContain('distributionNavItems')
    expect(componentSource).not.toContain('function buildUserDistributionNavItem')
  })

  it('uses workspace switching for admin platform and distribution navigation', () => {
    expect(componentSource).toContain('showAdminWorkspaceSwitcher')
    expect(componentSource).toContain('activeAdminNavItems')
    expect(componentSource).toContain('adminDistributionNavItems')
    expect(componentSource).toContain('adminPlatformNavItems')
    expect(componentSource).toContain("adminNavWorkspace === 'distribution'")
    expect(componentSource).not.toContain("path: '/admin/distribution/organizations'")
  })

  it('supports personal distribution workspace for admins with channel access', () => {
    expect(componentSource).toContain('showAdminPersonalWorkspaceSwitcher')
    expect(componentSource).toContain('hasPersonalDistributionAccess')
    expect(componentSource).toContain('activePersonalNavItems')
  })
})

describe('distribution navigation builders', () => {
  const distributionNavPath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../nav/distributionNav.ts')
  const adminDistributionNavPath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../nav/adminDistributionNav.ts')
  const distributionNavSource = readFileSync(distributionNavPath, 'utf8')
  const adminDistributionNavSource = readFileSync(adminDistributionNavPath, 'utf8')

  it('renders user distribution groups as top-level sidebar entries', () => {
    expect(distributionNavSource).not.toContain("t('nav.distribution')")
    expect(distributionNavSource).toContain("path: '/distribution/group/organization-management'")
    expect(distributionNavSource).toContain("t('distribution.groups.organizationManagement')")
    expect(distributionNavSource).toContain("path: '/distribution/overview'")
    expect(distributionNavSource).toContain("path: '/distribution#members'")
    expect(distributionNavSource).toContain("path: '/distribution#wallet'")
    expect(distributionNavSource).toContain("path: '/distribution#commissions'")
    expect(distributionNavSource).toContain("path: '/distribution#alert-events'")
    expect(distributionNavSource).not.toContain("path: '/distribution/group/risk-alerts'")
  })

  it('renders admin distribution groups as top-level sidebar entries', () => {
    expect(adminDistributionNavSource).not.toContain("t('nav.distributionManagement')")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/group/organization-management'")
    expect(adminDistributionNavSource).toContain("t('admin.distribution.groups.organizationManagement')")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/global-settings'")
    expect(adminDistributionNavSource).toContain("t('admin.distribution.tabs.globalSettings')")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/organizations'")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/members'")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/promotion-links'")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/wallets'")
    expect(adminDistributionNavSource).toContain("path: '/admin/distribution/alert-events'")
    expect(adminDistributionNavSource).not.toContain("path: '/admin/distribution/group/risk-alerts'")
  })
})
