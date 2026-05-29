import type { DistributionMember, DistributionMemberRole } from '@/api/admin/distribution'

export const DISTRIBUTION_PROMOTER_ROLES: DistributionMemberRole[] = ['agent', 'kol1', 'kol2']

export function isDistributionPromoterRole(roleType: string | null | undefined): boolean {
  const normalized = String(roleType || '').trim().toLowerCase()
  return DISTRIBUTION_PROMOTER_ROLES.includes(normalized as DistributionMemberRole)
}

export function isActiveDistributionMemberStatus(status: string | null | undefined): boolean {
  return String(status || '').trim().toLowerCase() === 'active'
}

export function filterMyActivePromoterMembers(
  members: DistributionMember[],
  userId: number,
): DistributionMember[] {
  if (userId <= 0) return []
  return members.filter(
    (member) =>
      member.user_id === userId &&
      isDistributionPromoterRole(member.role_type) &&
      isActiveDistributionMemberStatus(member.status),
  )
}

export function canManageDistributionMembersNav(
  canManageChannel: boolean,
  promoterMembers: DistributionMember[],
): boolean {
  if (canManageChannel) return true
  return promoterMembers.some((member) => member.role_type === 'agent' || member.role_type === 'kol1')
}

export function canAccessDistributionPromotionNav(
  canManageChannel: boolean,
  promoterMembers: DistributionMember[],
): boolean {
  return canManageChannel || promoterMembers.length > 0
}

export function formatDistributionMemberIdentity(
  member: Pick<DistributionMember, 'username' | 'user_email' | 'member_id'>,
  roleLabel?: string,
): string {
  const primary = member.username
    ? `${member.username} (${member.user_email})`
    : member.user_email || `#${member.member_id}`
  if (!roleLabel) return primary
  return `${primary} · ${roleLabel}`
}
