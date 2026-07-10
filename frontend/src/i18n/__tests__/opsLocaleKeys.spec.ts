import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'

function flattenKeys(obj: Record<string, any>, prefix = ''): string[] {
  const keys: string[] = []
  for (const [k, v] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${k}` : k
    if (typeof v === 'object' && v !== null && !Array.isArray(v)) {
      keys.push(...flattenKeys(v, fullKey))
    } else {
      keys.push(fullKey)
    }
  }
  return keys
}

describe('ops locale key completeness', () => {
  const requiredKeys = [
    'admin.ops.result',
    'admin.ops.timeRange.custom',
    'admin.ops.customTimeRange.startTime',
    'admin.ops.customTimeRange.endTime',
  ]

  for (const key of requiredKeys) {
    it(`en locale has ${key}`, () => {
      const enKeys = flattenKeys(en)
      expect(enKeys).toContain(key)
    })
  }
})

describe('groups locale key completeness', () => {
  it('en locale has admin.groups.failedToSave', () => {
    const enKeys = flattenKeys(en)
    expect(enKeys).toContain('admin.groups.failedToSave')
  })
})

describe('gas merge locale key completeness', () => {
  const requiredKeys = [
    'nav.workspaceSwitcherLabel',
    'nav.workspaceAdminDistribution',
    'redeem.balanceAddedDistribution',
    'payment.tabVoucher',
    'voucher.account',
    'distribution.title',
    'distribution.overview.title',
    'distributionLevels.add',
    'admin.distribution.title',
    'admin.distribution.actions.createOrganization',
    'admin.settings.features.distribution.title',
    'admin.settings.voucher.title',
    'admin.voucher.replenishmentTitle',
  ]

  for (const [locale, messages] of Object.entries({ en, zh })) {
    it(`${locale} locale keeps gas merge namespaces`, () => {
      const keys = flattenKeys(messages)
      for (const key of requiredKeys) {
        expect(keys, `${locale} missing ${key}`).toContain(key)
      }
    })
  }
})
