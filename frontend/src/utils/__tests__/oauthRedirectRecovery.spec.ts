import { describe, expect, it } from 'vitest'
import {
  buildOAuthCallbackRecoveryTarget,
  buildOAuthCallbackRecoveryTargetFromLocation,
  readOAuthTokenResponseFromHash,
  recoverOAuthCallbackHashFromLoginRedirect,
} from '@/utils/oauthRedirectRecovery'

describe('oauthRedirectRecovery', () => {
  it('routes token hash on dashboard to oauth callback', () => {
    const target = buildOAuthCallbackRecoveryTarget(
      '/dashboard',
      '',
      '#access_token=abc&redirect=%2Fdashboard'
    )

    expect(target).toBe('/auth/oauth/callback#access_token=abc&redirect=%2Fdashboard')
  })

  it('ignores oauth callback routes', () => {
    const target = buildOAuthCallbackRecoveryTarget(
      '/auth/oauth/callback',
      '',
      '#access_token=abc'
    )

    expect(target).toBeNull()
  })

  it('does not recover back into callback when leaving callback page', () => {
    const target = buildOAuthCallbackRecoveryTargetFromLocation(
      '/auth/oauth/callback',
      '/dashboard',
      '',
      '#access_token=abc&redirect=%2Fdashboard'
    )

    expect(target).toBeNull()
  })

  it('recovers token embedded in login redirect query', () => {
    const hash = recoverOAuthCallbackHashFromLoginRedirect(
      '/dashboard#access_token=abc&redirect=%2Fdashboard'
    )

    expect(hash).toBe('access_token=abc&redirect=%2Fdashboard')
  })

  it('reads token payload from callback hash', () => {
    const token = readOAuthTokenResponseFromHash(
      '#access_token=abc&refresh_token=rt&expires_in=3600&redirect=%2Fdashboard'
    )

    expect(token).toEqual({
      access_token: 'abc',
      refresh_token: 'rt',
      expires_in: 3600
    })
  })

})
