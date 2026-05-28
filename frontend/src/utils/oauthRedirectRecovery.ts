import type { OAuthTokenResponse } from '@/api/auth'

const OAUTH_CALLBACK_PATHS = ['/auth/oauth/callback', '/auth/callback']

export function safeDecodeURIComponent(value: string): string {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

export function repeatedlyDecode(value: string, maxRound = 3): string {
  let result = value
  for (let i = 0; i < maxRound; i += 1) {
    const decoded = safeDecodeURIComponent(result)
    if (decoded === result) {
      break
    }
    result = decoded
  }
  return result
}

export function isOAuthCallbackPath(path: string): boolean {
  return OAUTH_CALLBACK_PATHS.some(
    (callbackPath) => path === callbackPath || path.startsWith(`${callbackPath}/`)
  )
}

export function parseOAuthFragmentParams(hash: string): URLSearchParams {
  const raw = hash.startsWith('#') ? hash.slice(1) : hash
  return new URLSearchParams(raw)
}

function readOAuthAccessTokenFromHash(hash: string): string {
  return parseOAuthFragmentParams(hash).get('access_token')?.trim() || ''
}

export function sanitizeOAuthRedirectPath(path: string | null | undefined): string {
  if (!path) return '/dashboard'
  if (!path.startsWith('/')) return '/dashboard'
  if (path.startsWith('//')) return '/dashboard'
  if (path.includes('://')) return '/dashboard'
  if (path.includes('\n') || path.includes('\r')) return '/dashboard'
  return path
}

export function readOAuthTokenResponseFromHash(hash: string): OAuthTokenResponse | null {
  const params = parseOAuthFragmentParams(hash)
  const accessToken = params.get('access_token')?.trim() || ''
  if (!accessToken) {
    return null
  }

  const response: OAuthTokenResponse = { access_token: accessToken }
  const refreshToken = params.get('refresh_token')?.trim() || ''
  if (refreshToken) {
    response.refresh_token = refreshToken
  }
  const expiresIn = Number.parseInt(params.get('expires_in')?.trim() || '', 10)
  if (Number.isFinite(expiresIn) && expiresIn > 0) {
    response.expires_in = expiresIn
  }
  const tokenType = params.get('token_type')?.trim() || ''
  if (tokenType) {
    response.token_type = tokenType
  }
  return response
}

export function readOAuthRedirectTargetFromHash(hash: string): string {
  const redirect = repeatedlyDecode(parseOAuthFragmentParams(hash).get('redirect') || '', 4)
  return sanitizeOAuthRedirectPath(redirect || '/dashboard')
}

export function recoverOAuthCallbackHashFromLoginRedirect(
  redirectParam: unknown
): string | null {
  if (typeof redirectParam !== 'string' || !redirectParam.trim()) {
    return null
  }

  const normalized = repeatedlyDecode(redirectParam.trim(), 4)
  const hashIndex = normalized.indexOf('#')
  if (hashIndex < 0) {
    return null
  }

  const hashPayload = normalized.slice(hashIndex + 1)
  const params = new URLSearchParams(hashPayload)
  const accessToken = params.get('access_token')?.trim() || ''
  if (!accessToken) {
    return null
  }

  const redirectInHash = repeatedlyDecode(params.get('redirect') || '', 4)
  if (
    redirectInHash.startsWith('/') &&
    !redirectInHash.startsWith('//') &&
    !redirectInHash.includes('://')
  ) {
    params.set('redirect', redirectInHash)
  } else {
    params.set('redirect', '/dashboard')
  }
  return params.toString()
}

/**
 * When OAuth tokens are delivered in the URL hash to a non-callback page
 * (for example /dashboard#access_token=...), route to the dedicated callback page.
 */
export function buildOAuthCallbackRecoveryTarget(
  path: string,
  search: string,
  hash: string
): string | null {
  const accessToken = readOAuthAccessTokenFromHash(hash)
  if (!accessToken || isOAuthCallbackPath(path)) {
    return null
  }

  const query = search.startsWith('?') ? search : search ? `?${search}` : ''
  const fragment = hash.startsWith('#') ? hash : hash ? `#${hash}` : ''
  return `/auth/oauth/callback${query}${fragment}`
}

export function buildOAuthCallbackRecoveryTargetFromLocation(
  currentPath: string,
  targetPath: string,
  search: string,
  hash: string
): string | null {
  if (isOAuthCallbackPath(currentPath)) {
    return null
  }

  return buildOAuthCallbackRecoveryTarget(targetPath, search, hash)
}
