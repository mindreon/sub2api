import { describe, expect, it } from 'vitest'

import { isSidebarGroupExpanded, toggleSidebarGroup } from '../sidebarGroupExpand'

describe('sidebarGroupExpand', () => {
  it('allows collapsing an active group', () => {
    const expanded = new Set<string>()
    const collapsed = new Set<string>()

    expect(isSidebarGroupExpanded('/admin/distribution/group/organization-management', expanded, collapsed, true)).toBe(true)

    const toggled = toggleSidebarGroup(
      '/admin/distribution/group/organization-management',
      expanded,
      collapsed,
      true,
    )

    expect(
      isSidebarGroupExpanded(
        '/admin/distribution/group/organization-management',
        toggled.expandedGroups,
        toggled.collapsedGroups,
        true,
      ),
    ).toBe(false)
  })

  it('re-expands a group after toggle open', () => {
    const collapsed = new Set(['/admin/distribution/group/organization-management'])
    const expanded = new Set<string>()

    const toggled = toggleSidebarGroup(
      '/admin/distribution/group/organization-management',
      expanded,
      collapsed,
      false,
    )

    expect(
      isSidebarGroupExpanded(
        '/admin/distribution/group/organization-management',
        toggled.expandedGroups,
        toggled.collapsedGroups,
        true,
      ),
    ).toBe(true)
  })
})
