import type { ItickTenantProduct } from '@/types/itick'

export type DesktopStat = {
  label: string
  value: string
  down?: boolean
}

export type DesktopProductRow = {
  key: string
  product: ItickTenantProduct
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
}
