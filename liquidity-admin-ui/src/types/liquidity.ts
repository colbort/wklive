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
