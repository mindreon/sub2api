import type { DistributionLevelConfig } from '@/api/admin/settings'

export type DistributionLevelSource = 'channel' | 'global'

export interface DistributionLevelSelectOption {
  code: string
  name: string
  commission_rate: number
  source: DistributionLevelSource
  label: string
}

export const DISTRIBUTION_LEVELS_MAX = 20

export function normalizeDistributionLevelConfigs(raw: unknown): DistributionLevelConfig[] | null {
  if (!Array.isArray(raw)) return null
  const out: DistributionLevelConfig[] = []
  for (const item of raw) {
    if (!item || typeof item !== 'object') return null
    const cfg = item as Partial<DistributionLevelConfig> & {
      commission_rate?: unknown
      sort_order?: unknown
      active?: unknown
    }
    const code = String(cfg.code || '').trim().toUpperCase()
    if (!code) continue
    const name = String(cfg.name || '').trim() || code
    out.push({
      code,
      name,
      commission_rate: Math.min(100, Math.max(0, Number(cfg.commission_rate) || 0)),
      active: cfg.active !== false,
      sort_order: Math.max(0, Math.floor(Number(cfg.sort_order) || 0)),
      note: String(cfg.note || '').trim(),
    })
  }
  return out
}

export function cloneDistributionLevels(levels: DistributionLevelConfig[]): DistributionLevelConfig[] {
  return levels.map((level) => ({ ...level }))
}

export function sortDistributionLevels(levels: DistributionLevelConfig[]): DistributionLevelConfig[] {
  return [...levels].sort((a, b) => {
    if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
    return a.code.localeCompare(b.code)
  })
}

export function suggestLevelCodeFromName(name: string): string {
  const trimmed = String(name || '').trim()
  if (!trimmed) return ''
  const ascii = trimmed
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-zA-Z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '')
    .toUpperCase()
  if (ascii) return ascii.slice(0, 32)
  return trimmed.slice(0, 8).toUpperCase()
}

export function validateDistributionLevelConfigs(levels: DistributionLevelConfig[]): string | null {
  if (levels.length > DISTRIBUTION_LEVELS_MAX) {
    return 'max'
  }
  const seen = new Set<string>()
  for (const level of levels) {
    const code = String(level.code || '').trim().toUpperCase()
    const name = String(level.name || '').trim()
    if (!code || !name) {
      return 'required'
    }
    if (seen.has(code)) {
      return 'duplicate'
    }
    seen.add(code)
    const rate = Number(level.commission_rate)
    if (!Number.isFinite(rate) || rate < 0 || rate > 100) {
      return 'rate'
    }
  }
  return null
}

export function formatLevelCommissionPercent(rate: number): string {
  const value = Math.min(100, Math.max(0, Number(rate) || 0))
  const rounded = Math.round(value * 10000) / 10000
  return Number.isInteger(rounded) ? String(rounded) : String(rounded)
}

export function levelCommissionRateToMemberRate(rate: number): number {
  const percent = Math.min(100, Math.max(0, Number(rate) || 0))
  return percent / 100
}

export function buildLevelSelectOptions(
  channelLevels: DistributionLevelConfig[],
  globalLevels: DistributionLevelConfig[],
  formatLabel: (level: DistributionLevelConfig, source: DistributionLevelSource) => string,
): DistributionLevelSelectOption[] {
  const options: DistributionLevelSelectOption[] = []
  const seen = new Set<string>()

  for (const level of sortDistributionLevels(channelLevels)) {
    if (!level.active) continue
    const code = level.code.toUpperCase()
    if (seen.has(code)) continue
    seen.add(code)
    options.push({
      code,
      name: level.name,
      commission_rate: level.commission_rate,
      source: 'channel',
      label: formatLabel(level, 'channel'),
    })
  }

  for (const level of sortDistributionLevels(globalLevels)) {
    if (!level.active) continue
    const code = level.code.toUpperCase()
    if (seen.has(code)) continue
    seen.add(code)
    options.push({
      code,
      name: level.name,
      commission_rate: level.commission_rate,
      source: 'global',
      label: formatLabel(level, 'global'),
    })
  }

  return options
}

export function parseDistributionLevelsFromConfig(raw: unknown): DistributionLevelConfig[] {
  return normalizeDistributionLevelConfigs(raw) ?? []
}
