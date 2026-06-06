export function compactParams<T extends object>(params: T) {
  const result: Record<string, unknown> = {}

  function append(prefix: string, value: unknown) {
    if (value === undefined || value === null || value === '') return

    if (Array.isArray(value)) {
      result[prefix] = value
      return
    }

    if (typeof value === 'object') {
      for (const [childKey, childValue] of Object.entries(value as Record<string, unknown>)) {
        append(`${prefix}.${childKey}`, childValue)
      }
      return
    }

    result[prefix] = value
  }

  for (const [key, value] of Object.entries(params as Record<string, unknown>)) {
    append(key, value)
  }

  return result
}

export function buildPath(
  template: string,
  params: Record<string, string | number | undefined | null>,
) {
  return Object.entries(params).reduce((path, [key, value]) => {
    return path.replace(`:${key}`, encodeURIComponent(String(value ?? '')))
  }, template)
}
