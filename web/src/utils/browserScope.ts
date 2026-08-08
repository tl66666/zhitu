const STORAGE_KEY = 'zhitu-browser-token-v1'
const TOKEN_PATTERN = /^[0-9a-f]{64}$/

let memoryToken = ''

const createToken = (): string => {
  const bytes = new Uint8Array(32)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

// getBrowserToken returns a stable, high-entropy identifier scoped to this
// browser profile. It is sent only to authenticated business API endpoints.
export const getBrowserToken = (): string => {
  if (memoryToken) return memoryToken

  try {
    const existing = localStorage.getItem(STORAGE_KEY)?.toLowerCase() || ''
    if (TOKEN_PATTERN.test(existing)) {
      memoryToken = existing
      return memoryToken
    }
  } catch {
    // Storage may be unavailable in hardened/private browser environments.
  }

  memoryToken = createToken()
  try {
    localStorage.setItem(STORAGE_KEY, memoryToken)
  } catch {
    // Keep the token stable for the current page lifetime.
  }
  return memoryToken
}

export const isBrowserScopedUrl = (url: string | undefined): boolean =>
  !!url && url.includes('/api/v1/')
