import Decimal from 'decimal.js'

/** Asset RPC values are decimal natural units, not integer payment cents. */
export function formatAssetAmount(value: unknown): string {
  try {
    return new Decimal(String(value ?? 0)).toFixed()
  } catch {
    return '0'
  }
}

export function formatAssetFields<T extends Record<string, unknown>>(
  data: T | null | undefined,
  keys: Iterable<string>,
) {
  if (!data) return null
  const keySet = keys instanceof Set ? keys : new Set(keys)
  return Object.fromEntries(
    Object.entries(data).map(([key, value]) => [
      key,
      keySet.has(key) ? formatAssetAmount(value) : value,
    ]),
  )
}
