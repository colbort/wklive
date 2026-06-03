import { t } from '@/i18n'

type MarketCategoryLike = {
  categoryCode?: string
  categoryName?: string
} | null | undefined

const CATEGORY_LABEL_KEYS: Record<string, string> = {
  forex: 'market.categoryForex',
  crypto: 'market.categoryCrypto',
  stock: 'market.categoryStock',
  future: 'market.categoryFuture',
  indices: 'market.categoryIndices',
  fund: 'market.categoryFund',
}

export function marketCategoryCodeLabel(categoryCode?: string, fallback = '') {
  const code = String(categoryCode || '').trim().toLowerCase()
  const key = CATEGORY_LABEL_KEYS[code]
  return key ? t(key) : fallback
}

export function marketCategoryLabel(category: MarketCategoryLike) {
  return marketCategoryCodeLabel(category?.categoryCode, category?.categoryName || '')
}
