import type { NavItem } from './types'

export interface DistributionNavIcons {
  ChartIcon: unknown
  UsersIcon: unknown
  GlobeIcon: unknown
  UserIcon: unknown
  CreditCardIcon: unknown
  OrderIcon: unknown
  PriceTagIcon: unknown
  BellIcon: unknown
}

export interface UserDistributionNavOptions {
  canManageChannel: boolean
  canAccessPromotionNav: boolean
  canManageMembersNav: boolean
  canAccessChannelFinanceNav: boolean
}

type TranslateFn = (key: string) => string

export function buildUserDistributionNavItems(
  t: TranslateFn,
  icons: DistributionNavIcons,
  options: UserDistributionNavOptions,
): NavItem[] {
  const { ChartIcon, UsersIcon, GlobeIcon, UserIcon, CreditCardIcon, OrderIcon, PriceTagIcon, BellIcon } = icons
  const {
    canManageChannel,
    canAccessPromotionNav,
    canManageMembersNav,
    canAccessChannelFinanceNav,
  } = options

  const items: NavItem[] = []

  if (canManageChannel || canManageMembersNav) {
    items.push({
      path: '/distribution/group/organization-management',
      label: t('distribution.groups.organizationManagement'),
      icon: UsersIcon,
      expandOnly: true,
      children: [
        {
          path: '/distribution/overview',
          label: t('distribution.overview.title'),
          icon: ChartIcon,
          featureFlag: () => canManageChannel,
        },
        {
          path: '/distribution#members',
          label: t('distribution.tabs.members'),
          icon: UsersIcon,
          featureFlag: () => canManageMembersNav,
        },
      ],
    })
  }

  if (canAccessPromotionNav) {
    items.push({
      path: '/distribution/group/promotion-management',
      label: t('distribution.groups.promotionManagement'),
      icon: GlobeIcon,
      expandOnly: true,
      children: [
        { path: '/distribution#promotion-links', label: t('distribution.tabs.promotionLinks'), icon: GlobeIcon },
        { path: '/distribution#attributions', label: t('distribution.tabs.attributions'), icon: UserIcon },
      ],
    })
  }

  const settlementChildren: NavItem[] = []
  if (canAccessChannelFinanceNav) {
    settlementChildren.push(
      { path: '/distribution#wallet', label: t('distribution.tabs.wallet'), icon: CreditCardIcon },
      { path: '/distribution#wallet-requests', label: t('distribution.tabs.walletRequests'), icon: OrderIcon },
      {
        path: '/distribution#wholesale-pricing',
        label: t('distribution.tabs.wholesalePricing'),
        icon: PriceTagIcon,
      },
    )
  }
  if (canAccessPromotionNav || canAccessChannelFinanceNav) {
    settlementChildren.push({
      path: '/distribution#commissions',
      label: t('distribution.tabs.commissions'),
      icon: ChartIcon,
    })
  }
  if (settlementChildren.length > 0) {
    items.push({
      path: '/distribution/group/commission-settlement',
      label: t('distribution.groups.commissionSettlement'),
      icon: CreditCardIcon,
      expandOnly: true,
      children: settlementChildren,
    })
  }

  if (canAccessChannelFinanceNav) {
    items.push({
      path: '/distribution#alert-events',
      label: t('distribution.tabs.alertEvents'),
      icon: BellIcon,
    })
  }

  return items
}

export function isDistributionRoutePath(path: string): boolean {
  return path === '/distribution' || path.startsWith('/distribution/')
}
