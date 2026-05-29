import type { NavItem } from './types'

export interface AdminDistributionNavIcons {
  ChannelIcon: unknown
  UsersIcon: unknown
  GlobeIcon: unknown
  UserIcon: unknown
  CreditCardIcon: unknown
  OrderIcon: unknown
  OrderListIcon: unknown
  ChartIcon: unknown
  BellIcon: unknown
  SettingsIcon: unknown
}

type TranslateFn = (key: string) => string

export function isAdminDistributionRoutePath(path: string): boolean {
  return path === '/admin/distribution' || path.startsWith('/admin/distribution/')
}

export function buildAdminDistributionNavItems(t: TranslateFn, icons: AdminDistributionNavIcons): NavItem[] {
  const {
    ChannelIcon,
    UsersIcon,
    GlobeIcon,
    UserIcon,
    CreditCardIcon,
    OrderIcon,
    OrderListIcon,
    ChartIcon,
    BellIcon,
    SettingsIcon,
  } = icons

  return [
    {
      path: '/admin/distribution/global-settings',
      label: t('admin.distribution.tabs.globalSettings'),
      icon: SettingsIcon,
      hideInSimpleMode: true,
    },
    {
      path: '/admin/distribution/group/organization-management',
      label: t('admin.distribution.groups.organizationManagement'),
      icon: UsersIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      children: [
        { path: '/admin/distribution/organizations', label: t('admin.distribution.tabs.organizations'), icon: ChannelIcon },
        { path: '/admin/distribution/members', label: t('admin.distribution.tabs.members'), icon: UsersIcon },
      ],
    },
    {
      path: '/admin/distribution/group/promotion-management',
      label: t('admin.distribution.groups.promotionManagement'),
      icon: GlobeIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      children: [
        { path: '/admin/distribution/promotion-links', label: t('admin.distribution.tabs.promotionLinks'), icon: GlobeIcon },
        { path: '/admin/distribution/attributions', label: t('admin.distribution.tabs.attributions'), icon: UserIcon },
      ],
    },
    {
      path: '/admin/distribution/group/commission-settlement',
      label: t('admin.distribution.groups.commissionSettlement'),
      icon: CreditCardIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      children: [
        { path: '/admin/distribution/wallets', label: t('admin.distribution.tabs.wallets'), icon: CreditCardIcon },
        { path: '/admin/distribution/wallet-requests', label: t('admin.distribution.tabs.walletRequests'), icon: OrderIcon },
        { path: '/admin/distribution/wallet-transactions', label: t('admin.distribution.tabs.walletTransactions'), icon: OrderListIcon },
        { path: '/admin/distribution/commissions', label: t('admin.distribution.tabs.commissions'), icon: ChartIcon },
      ],
    },
    {
      path: '/admin/distribution/alert-events',
      label: t('admin.distribution.tabs.alertEvents'),
      icon: BellIcon,
      hideInSimpleMode: true,
    },
  ]
}
