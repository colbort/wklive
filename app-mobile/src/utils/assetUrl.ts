export function resolveSystemAssetUrl(assetUrl: string | undefined, url?: string) {
  const value = url?.trim()
  if (!value) return ''
  if (/^(https?:)?\/\//i.test(value) || /^(data|blob):/i.test(value)) return value

  const path = value.replace(/^\.\//, '').replace(/^\/+/, '')
  const base = assetUrl?.replace(/\/+$/, '') || ''
  if (!base) return `/${path}`

  return `${base}/${path}`
}
