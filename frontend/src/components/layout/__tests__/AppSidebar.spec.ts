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

describe('AppSidebar distribution navigation', () => {
  it('renders distribution management as grouped submenu sections', () => {
    expect(componentSource).toContain("path: '/admin/distribution'")
    expect(componentSource).toContain("expandOnly: true")
    expect(componentSource).toContain("t('admin.distribution.groups.organizationManagement')")
    expect(componentSource).toContain("t('admin.distribution.groups.promotionManagement')")
    expect(componentSource).toContain("t('admin.distribution.groups.commissionSettlement')")
    expect(componentSource).toContain("t('admin.distribution.groups.riskAlerts')")
    expect(componentSource).toContain("path: '/admin/distribution/organizations'")
    expect(componentSource).toContain("path: '/admin/distribution/members'")
    expect(componentSource).toContain("path: '/admin/distribution/promotion-links'")
    expect(componentSource).toContain("path: '/admin/distribution/wallets'")
    expect(componentSource).toContain("path: '/admin/distribution/alert-events'")
    expect(componentSource).toContain("path: '/admin/distribution/wallet-requests'")
    expect(componentSource).toContain("path: '/admin/distribution/wallet-transactions'")
    expect(componentSource).toContain("path: '/admin/distribution/attributions'")
    expect(componentSource).toContain("path: '/admin/distribution/commissions'")
  })

  it('renders user distribution navigation with a dedicated channel overview entry', () => {
    expect(componentSource).toContain("path: '/distribution'")
    expect(componentSource).toContain("expandOnly: true")
    expect(componentSource).toContain("t('distribution.groups.organizationManagement')")
    expect(componentSource).toContain("t('distribution.groups.promotionManagement')")
    expect(componentSource).toContain("t('distribution.groups.commissionSettlement')")
    expect(componentSource).toContain("t('distribution.groups.riskAlerts')")
    expect(componentSource).toContain("path: '/distribution/overview'")
    expect(componentSource).toContain("t('distribution.overview.title')")
    expect(componentSource).toContain("path: '/distribution#members'")
    expect(componentSource).toContain("path: '/distribution#wallet'")
    expect(componentSource).toContain("path: '/distribution#commissions'")
  })
})
