export const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '')

export const statusType = (s) =>
  ({ 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' })[s] || 'info'

export const quotaPercent = (used, limit) => {
  if (!limit || limit <= 0) return 0
  return Math.min(100, Math.round((used / limit) * 100))
}

export const quotaStatus = (used, limit) => {
  if (!limit || limit <= 0) return ''
  const p = quotaPercent(used, limit)
  if (p >= 100) return 'exception'
  if (p >= 80) return 'warning'
  return 'success'
}
