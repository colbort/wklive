import type { MarketTenantCategory } from '@/types/market'
import type {
  TradeSymbol,
  TradeSymbolContract,
  TradeSymbolLeverageConfig,
  TradeSymbolSeconds,
  TradeSymbolSpot,
} from '@/types/trade'

export const PRODUCT_TYPE_SPOT = 1
export const PRODUCT_TYPE_DERIVATIVE = 2
export const PRODUCT_TYPE_SECONDS = 3

export const CONTRACT_TYPE_PERPETUAL = 1
export const CONTRACT_TYPE_DELIVERY = 2
export const CONTRACT_VALUE_TYPE_LINEAR = 1
export const CONTRACT_VALUE_TYPE_INVERSE = 2

export type TradeMarketMode =
  | 'spot'
  | 'seconds'
  | 'delivery-linear'
  | 'delivery-inverse'
  | 'perpetual-linear'
  | 'perpetual-inverse'

export type TradeExperience = 'spot' | 'contract' | 'seconds'

export type TradeSymbolDetail = {
  symbol: TradeSymbol | null
  spot: TradeSymbolSpot | null
  contract: TradeSymbolContract | null
  leverageConfigs: TradeSymbolLeverageConfig[]
  secondsConfigs: TradeSymbolSeconds[]
}

export type TradeCategoryCode =
  | 'crypto'
  | 'forex'
  | 'stock'
  | 'future'
  | 'indices'
  | 'fund'
  | 'unknown'

export type TradeCategoryConfig = {
  code: TradeCategoryCode
  preferredModes: TradeMarketMode[]
  showDepthPreview: boolean
  showPremarket: boolean
  showProductSubtitle: boolean
}

const SPOT_FIRST_MODES: TradeMarketMode[] = [
  'spot',
  'seconds',
  'perpetual-linear',
  'perpetual-inverse',
  'delivery-linear',
  'delivery-inverse',
]

const CONTRACT_FIRST_MODES: TradeMarketMode[] = [
  'delivery-linear',
  'delivery-inverse',
  'perpetual-linear',
  'perpetual-inverse',
  'spot',
  'seconds',
]

const CATEGORY_CONFIGS: Record<TradeCategoryCode, TradeCategoryConfig> = {
  crypto: {
    code: 'crypto',
    preferredModes: SPOT_FIRST_MODES,
    showDepthPreview: true,
    showPremarket: false,
    showProductSubtitle: false,
  },
  forex: {
    code: 'forex',
    preferredModes: SPOT_FIRST_MODES,
    showDepthPreview: false,
    showPremarket: false,
    showProductSubtitle: false,
  },
  stock: {
    code: 'stock',
    preferredModes: SPOT_FIRST_MODES,
    showDepthPreview: false,
    showPremarket: true,
    showProductSubtitle: true,
  },
  future: {
    code: 'future',
    preferredModes: CONTRACT_FIRST_MODES,
    showDepthPreview: true,
    showPremarket: false,
    showProductSubtitle: false,
  },
  indices: {
    code: 'indices',
    preferredModes: SPOT_FIRST_MODES,
    showDepthPreview: false,
    showPremarket: false,
    showProductSubtitle: false,
  },
  fund: {
    code: 'fund',
    preferredModes: SPOT_FIRST_MODES,
    showDepthPreview: false,
    showPremarket: false,
    showProductSubtitle: false,
  },
  unknown: {
    code: 'unknown',
    preferredModes: SPOT_FIRST_MODES,
    showDepthPreview: false,
    showPremarket: false,
    showProductSubtitle: false,
  },
}

export function normalizeTradeCategoryCode(categoryCode?: string): TradeCategoryCode {
  const code = String(categoryCode || '').trim().toLowerCase()
  return Object.prototype.hasOwnProperty.call(CATEGORY_CONFIGS, code)
    ? (code as TradeCategoryCode)
    : 'unknown'
}

export function getTradeCategoryConfig(
  category?: Pick<MarketTenantCategory, 'categoryCode'> | null,
): TradeCategoryConfig {
  return CATEGORY_CONFIGS[normalizeTradeCategoryCode(category?.categoryCode)]
}

export function tradeExperienceForProductType(productType?: number): TradeExperience {
  if (productType === PRODUCT_TYPE_DERIVATIVE) return 'contract'
  if (productType === PRODUCT_TYPE_SECONDS) return 'seconds'
  return 'spot'
}

export function tradeExperienceForMode(mode: TradeMarketMode): TradeExperience {
  if (mode === 'seconds') return 'seconds'
  if (mode === 'spot') return 'spot'
  return 'contract'
}

export function resolveTradeExperience(
  symbol: Pick<TradeSymbol, 'productType'> | null | undefined,
  mode: TradeMarketMode,
): TradeExperience {
  return symbol ? tradeExperienceForProductType(symbol.productType) : tradeExperienceForMode(mode)
}

export function matchesTradeMarketMode(
  symbol: Pick<TradeSymbol, 'productType' | 'contractType' | 'contractValueType'>,
  mode: TradeMarketMode,
) {
  if (mode === 'spot') return symbol.productType === PRODUCT_TYPE_SPOT
  if (mode === 'seconds') return symbol.productType === PRODUCT_TYPE_SECONDS
  if (symbol.productType !== PRODUCT_TYPE_DERIVATIVE) return false

  const contractType = mode.startsWith('delivery')
    ? CONTRACT_TYPE_DELIVERY
    : CONTRACT_TYPE_PERPETUAL
  const contractValueType = mode.endsWith('linear')
    ? CONTRACT_VALUE_TYPE_LINEAR
    : CONTRACT_VALUE_TYPE_INVERSE

  return symbol.contractType === contractType && symbol.contractValueType === contractValueType
}
