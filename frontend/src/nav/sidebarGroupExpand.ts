export function isSidebarGroupExpanded(
  itemPath: string,
  expandedGroups: Set<string>,
  collapsedGroups: Set<string>,
  groupActive: boolean,
): boolean {
  if (collapsedGroups.has(itemPath)) {
    return false
  }
  return expandedGroups.has(itemPath) || groupActive
}

export function toggleSidebarGroup(
  itemPath: string,
  expandedGroups: Set<string>,
  collapsedGroups: Set<string>,
  currentlyExpanded: boolean,
): { expandedGroups: Set<string>, collapsedGroups: Set<string> } {
  const nextExpanded = new Set(expandedGroups)
  const nextCollapsed = new Set(collapsedGroups)

  if (currentlyExpanded) {
    nextCollapsed.add(itemPath)
    nextExpanded.delete(itemPath)
  } else {
    nextCollapsed.delete(itemPath)
    nextExpanded.add(itemPath)
  }

  return { expandedGroups: nextExpanded, collapsedGroups: nextCollapsed }
}
