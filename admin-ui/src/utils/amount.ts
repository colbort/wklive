export function centToAmount(value: unknown): number {
  const amount = Number(value || 0) / 100
  return Number.isFinite(amount) ? amount : 0
}

export function amountToCent(value: unknown): number {
  return Math.round(Number(value || 0) * 100)
}

export function formatCentAmount(value: unknown, digits = 2): string {
  return centToAmount(value).toFixed(digits)
}

export function formatCentFields<T extends Record<string, unknown>>(
  data: T | null | undefined,
  keys: Iterable<string>,
) {
  if (!data) return null
  const keySet = keys instanceof Set ? keys : new Set(keys)
  return Object.fromEntries(
    Object.entries(data).map(([key, value]) => [
      key,
      keySet.has(key) ? formatCentAmount(value) : value,
    ]),
  )
}
