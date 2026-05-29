export interface NavItem {
  path: string
  label: string
  icon?: unknown
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: NavItem[]
  expandOnly?: boolean
  featureFlag?: () => boolean | undefined
}

export type NavWorkspace = 'consumer' | 'distribution'
