export interface PageQuery {
  keyword?: string;
  status?: number | string;
  cursor?: number;
  limit?: number;
  count?: number;
  [key: string]: unknown;
}

export interface ListResponse<T = Record<string, unknown>> {
  data: T[];
  total?: number;
  hasNext?: boolean;
  hasPrev?: boolean;
  nextCursor?: number;
  prevCursor?: number;
}

export interface OptionItem {
  value: number;
  code: string;
}

export interface OptionGroup {
  key: string;
  label: string;
  options: OptionItem[];
}

export interface Provider {
  id: number;
  providerCode: string;
  providerName: string;
  providerType: number;
  venueCode: string;
  environment: number;
  status: number;
  lastHealthStatus: number;
  lastHealthAt: number;
  credentialConfigured: boolean;
  version: number;
}

export interface ProvisionInternalProviderRequest {
  symbolId: number;
  providerCode: string;
  providerName: string;
  baseAmount: string;
  quoteAmount: string;
  remark?: string;
}

export interface ConfigSymbolOption {
  symbolId: number;
  symbol: string;
  displaySymbol: string;
  productType: number;
  contractType: number;
  walletType: number;
  contractValueType: number;
  categoryType: number;
  referencePriceSource: string;
}

export interface ConfigProviderOption {
  providerId: number;
  providerCode: string;
  providerName: string;
  providerType: number;
  tradeUserId: number;
  status: number;
}

export interface ConfigTradingUserOption {
  tradeUserId: number;
  username: string;
}

export interface ConfigOptions {
  symbols: ConfigSymbolOption[];
  providers: ConfigProviderOption[];
  tradingUsers: ConfigTradingUserOption[];
}

export interface SymbolConfig {
  id: number;
  symbolId: number;
  symbol: string;
  productType: number;
  contractType: number;
  liquidityMode: number;
  referencePriceSource: string;
  baseSpreadBps: string;
  maxSpreadBps: string;
  refreshIntervalMs: number;
  status: number;
  version: number;
}

export interface StrategyLevel {
  id: number;
  configId: number;
  levelNo: number;
  bidSpreadBps: string;
  askSpreadBps: string;
  bidQty: string;
  askQty: string;
  enabled: number;
}

export interface SymbolConfigDetail extends SymbolConfig {
  internalProviderId: number;
  externalProviderId: number;
  externalSymbol: string;
  referencePriceKind: string;
  quoteValidityMs: number;
  quoteTtlMs: number;
  repriceThresholdBps: string;
  maxPriceDeviationBps: string;
  priceTick: string;
  qtyStep: string;
  minQuoteQty: string;
  maxQuoteQty: string;
  maxQuoteNotional: string;
  targetBaseInventory: string;
  minBaseInventory: string;
  maxBaseInventory: string;
  maxNetExposure: string;
  maxDailyNotional: string;
  inventorySkewBps: string;
  hedgeThreshold: string;
  hedgeRatio: string;
  selfTradePrevention: number;
  pauseReason: string;
  createTimes: number;
  updateTimes: number;
}

export interface SymbolConfigDetailResponse {
  data: SymbolConfigDetail;
  levels: StrategyLevel[];
}
