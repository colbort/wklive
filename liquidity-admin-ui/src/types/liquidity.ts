export interface PageQuery {
  keyword?: string;
  status?: number | string;
  cursor?: number;
  limit?: number;
  [key: string]: unknown;
}

export interface ListResponse<T = Record<string, unknown>> {
  data: T[];
  page?: { total?: number; nextCursor?: number; hasMore?: boolean };
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
