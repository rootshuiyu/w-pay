const TOKEN_KEY = 'nbpay_token'
const USER_KEY = 'nbpay_user'
const LEGACY_TOKEN = 'wpay_token'
const LEGACY_USER = 'wpay_user'

function migrateLegacyKeys() {
  if (!localStorage.getItem(TOKEN_KEY) && localStorage.getItem(LEGACY_TOKEN)) {
    localStorage.setItem(TOKEN_KEY, localStorage.getItem(LEGACY_TOKEN))
    localStorage.removeItem(LEGACY_TOKEN)
  }
  if (!localStorage.getItem(USER_KEY) && localStorage.getItem(LEGACY_USER)) {
    localStorage.setItem(USER_KEY, localStorage.getItem(LEGACY_USER))
    localStorage.removeItem(LEGACY_USER)
  }
}

export function getToken() {
  migrateLegacyKeys()
  return localStorage.getItem(TOKEN_KEY)
}

export function getUser() {
  migrateLegacyKeys()
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || '{}')
  } catch {
    return {}
  }
}

export function setAuth(token, user) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
  localStorage.removeItem(LEGACY_TOKEN)
  localStorage.removeItem(LEGACY_USER)
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  localStorage.removeItem(LEGACY_TOKEN)
  localStorage.removeItem(LEGACY_USER)
}

migrateLegacyKeys()
