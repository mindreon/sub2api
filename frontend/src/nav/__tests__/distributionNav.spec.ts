import { describe, expect, it } from 'vitest'

import { buildUserDistributionNavItems } from '@/nav/distributionNav'

const icons = {
  ChartIcon: 'chart',
  UsersIcon: 'users',
  GlobeIcon: 'globe',
  UserIcon: 'user',
  CreditCardIcon: 'credit',
  OrderIcon: 'order',
  PriceTagIcon: 'price',
  BellIcon: 'bell',
}

const t = (key: string) => key

function flattenPaths(items: ReturnType<typeof buildUserDistributionNavItems>): string[] {
  const paths: string[] = []
  for (const item of items) {
    paths.push(item.path)
    for (const child of item.children || []) {
      paths.push(child.path)
    }
  }
  return paths
}

describe('buildUserDistributionNavItems', () => {
  it('shows promotion management for channel managers and promoters', () => {
    const managerPaths = flattenPaths(
      buildUserDistributionNavItems(t, icons, {
        canManageChannel: true,
        canAccessPromotionNav: true,
        canManageMembersNav: true,
        canAccessChannelFinanceNav: true,
      }),
    )
    expect(managerPaths).toContain('/distribution#promotion-links')
    expect(managerPaths).toContain('/distribution#wallet')

    const promoterPaths = flattenPaths(
      buildUserDistributionNavItems(t, icons, {
        canManageChannel: false,
        canAccessPromotionNav: true,
        canManageMembersNav: false,
        canAccessChannelFinanceNav: false,
      }),
    )
    expect(promoterPaths).toContain('/distribution#promotion-links')
    expect(promoterPaths).toContain('/distribution#commissions')
    expect(promoterPaths).not.toContain('/distribution#wallet')
    expect(promoterPaths).not.toContain('/distribution#members')
  })

  it('hides promotion management when the user has no promoter access', () => {
    const paths = flattenPaths(
      buildUserDistributionNavItems(t, icons, {
        canManageChannel: false,
        canAccessPromotionNav: false,
        canManageMembersNav: false,
        canAccessChannelFinanceNav: false,
      }),
    )
    expect(paths).not.toContain('/distribution#promotion-links')
    expect(paths).not.toContain('/distribution#attributions')
  })
})
