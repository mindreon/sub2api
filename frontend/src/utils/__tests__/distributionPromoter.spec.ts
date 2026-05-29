import { describe, expect, it } from 'vitest'

import type { DistributionMember } from '@/api/admin/distribution'
import {
  canAccessDistributionPromotionNav,
  canManageDistributionMembersNav,
  filterMyActivePromoterMembers,
  isDistributionPromoterRole,
} from '@/utils/distributionPromoter'

function member(partial: Partial<DistributionMember> & Pick<DistributionMember, 'member_id' | 'user_id' | 'role_type'>): DistributionMember {
  return {
    channel_org_id: 88,
    level_code: '',
    commission_rate: 0,
    status: 'active',
    user_email: 'a@example.com',
    username: 'alice',
    created_at: '2026-05-24T00:00:00Z',
    updated_at: '2026-05-24T00:00:00Z',
    ...partial,
  }
}

describe('distributionPromoter utils', () => {
  it('detects promoter roles', () => {
    expect(isDistributionPromoterRole('agent')).toBe(true)
    expect(isDistributionPromoterRole('kol1')).toBe(true)
    expect(isDistributionPromoterRole('manager')).toBe(false)
  })

  it('filters active promoter memberships for the current user', () => {
    const items = [
      member({ member_id: 1, user_id: 7, role_type: 'kol1' }),
      member({ member_id: 2, user_id: 8, role_type: 'kol1' }),
      member({ member_id: 3, user_id: 7, role_type: 'manager' as DistributionMember['role_type'] }),
      member({ member_id: 4, user_id: 7, role_type: 'agent', status: 'inactive' }),
    ]

    expect(filterMyActivePromoterMembers(items, 7)).toEqual([items[0]])
  })

  it('derives navigation access from channel role and promoter memberships', () => {
    const promoters = [member({ member_id: 1, user_id: 7, role_type: 'kol2' })]

    expect(canAccessDistributionPromotionNav(true, [])).toBe(true)
    expect(canAccessDistributionPromotionNav(false, promoters)).toBe(true)
    expect(canAccessDistributionPromotionNav(false, [])).toBe(false)

    expect(canManageDistributionMembersNav(true, [])).toBe(true)
    expect(canManageDistributionMembersNav(false, promoters)).toBe(false)
    expect(
      canManageDistributionMembersNav(false, [member({ member_id: 2, user_id: 7, role_type: 'agent' })]),
    ).toBe(true)
  })
})
