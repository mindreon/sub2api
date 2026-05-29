import { describe, expect, it } from 'vitest'

import {
  buildLevelSelectOptions,
  normalizeDistributionLevelConfigs,
  suggestLevelCodeFromName,
  validateDistributionLevelConfigs,
} from '../distributionLevels'

describe('distributionLevels utils', () => {
  it('normalizes level configs', () => {
    const result = normalizeDistributionLevelConfigs([
      { code: ' vip ', name: '', commission_rate: 120, active: true, sort_order: 2, note: ' core ' },
      { code: '  ' },
    ])

    expect(result).toEqual([
      {
        code: 'VIP',
        name: 'VIP',
        commission_rate: 100,
        active: true,
        sort_order: 2,
        note: 'core',
      },
    ])
  })

  it('validates duplicate codes', () => {
    const error = validateDistributionLevelConfigs([
      { code: 'GOLD', name: 'Gold', commission_rate: 10, active: true, sort_order: 0, note: '' },
      { code: 'gold', name: 'Gold 2', commission_rate: 8, active: true, sort_order: 1, note: '' },
    ])
    expect(error).toBe('duplicate')
  })

  it('suggests code from name', () => {
    expect(suggestLevelCodeFromName('金牌代理')).toBeTruthy()
    expect(suggestLevelCodeFromName('Gold Agent')).toBe('GOLD_AGENT')
  })

  it('builds select options with channel precedence', () => {
    const options = buildLevelSelectOptions(
      [{ code: 'VIP', name: 'VIP', commission_rate: 15, active: true, sort_order: 0, note: '' }],
      [{ code: 'VIP', name: 'Global VIP', commission_rate: 10, active: true, sort_order: 0, note: '' }],
      (level, source) => `${level.code}-${source}`,
    )

    expect(options).toHaveLength(1)
    expect(options[0]?.label).toBe('VIP-channel')
  })
})
