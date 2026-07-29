import { get, post } from '@/utils/request'
import type {
  BizTradeEvent,
  ContractLeverageConfig,
  ContractMarginSnapshot,
  ContractPosition,
  ContractPositionHistory,
  CreateSymbolReq,
  GetCancelLogListAdminReq,
  GetFillDetailAdminReq,
  GetFillListAdminReq,
  GetMarginSnapshotListAdminReq,
  GetOrderDetailAdminReq,
  GetOrderListAdminReq,
  GetPositionDetailAdminReq,
  GetPositionHistoryListAdminReq,
  GetPositionListAdminReq,
  GetRiskOrderCheckLogListReq,
  GetSymbolDetailAdminReq,
  GetSymbolLeverageConfigListReq,
  GetSymbolListAdminReq,
  GetTradeEventDetailReq,
  GetTradeEventListReq,
  GetUserLeverageConfigReq,
  GetUserSymbolLimitReq,
  GetUserTradeConfigReq,
  GetUserTradeLimitReq,
  RespBase,
  RetryTradeEventReq,
  RiskOrderCheckLog,
  RiskUserSymbolLimit,
  RiskUserTradeLimit,
  SetContractSymbolConfigReq,
  SetSecondsSymbolConfigReq,
  SetSpotSymbolConfigReq,
  SetSymbolLeverageConfigReq,
  SetUserLeverageConfigReq,
  SetUserSymbolLimitReq,
  SetUserTradeConfigReq,
  SetUserTradeLimitReq,
  TradeCancelLog,
  TradeFill,
  TradeOrder,
  TradeOrderDetailData,
  TradeSymbol,
  TradeSymbolDetailData,
  TradeSymbolDetailResp,
  TradeSymbolLeverageConfig,
  TradeUserConfig,
  ContractRiskLimitTier,
  SetContractRiskLimitTierReq,
  GetContractRiskLimitTierListReq,
  GetContractRiskLimitTierListResp,
  ContractFundingBatch,
  GetFundingBatchListReq,
  GetFundingBatchListResp,
  ContractFundingSettlement,
  GetFundingSettlementListReq,
  GetFundingSettlementListResp,
  ContractDeliveryBatch,
  GetDeliveryBatchListReq,
  GetDeliveryBatchListResp,
  ContractDeliverySettlement,
  GetDeliverySettlementListReq,
  GetDeliverySettlementListResp,
  ContractLiquidation,
  GetLiquidationListReq,
  GetLiquidationListResp,
  ContractAccountLiquidation,
  GetAccountLiquidationListReq,
  GetAccountLiquidationListResp,
  GetAccountLiquidationDetailReq,
  GetAccountLiquidationDetailResp,
  RetryAccountLiquidationReq,
  TradeSecondsPriceSnapshot,
  GetSecondsPriceSnapshotListReq,
  GetSecondsPriceSnapshotListResp,
  TradeAssetReservation,
  GetAssetReservationListReq,
  GetAssetReservationListResp,
  TradeSettlementInstruction,
  GetSettlementInstructionListReq,
  GetSettlementInstructionListResp,
  RetrySettlementInstructionReq,
  ContractReconciliationIssue,
  GetContractReconciliationIssueListReq,
  GetContractReconciliationIssueListResp,
  IgnoreContractReconciliationIssueReq,
  UpdateSymbolReq,
  InsuranceFundAccount,
  SetInsuranceFundAccountReq,
  GetInsuranceFundAccountListReq,
  GetInsuranceFundAccountListResp,
  TradeMarketSnapshot,
  GetMarketSnapshotListReq,
  GetMarketSnapshotListResp,
} from '@/services'

export function apiTradeListSymbols(
  params: GetSymbolListAdminReq,
): Promise<RespBase<TradeSymbol[]>> {
  return get<TradeSymbol[]>('/admin/trade/symbols', params)
}

export function apiTradeGetSymbol(params: GetSymbolDetailAdminReq): Promise<TradeSymbolDetailResp> {
  return get<TradeSymbolDetailData>('/admin/trade/symbols/detail', params)
}

export function apiTradeCreateSymbol(params: CreateSymbolReq): Promise<RespBase> {
  return post('/admin/trade/symbols', params)
}

export function apiTradeUpdateSymbol(params: UpdateSymbolReq): Promise<RespBase> {
  return post('/admin/trade/symbols/update', params)
}

export function apiTradeSetSpotConfig(params: SetSpotSymbolConfigReq): Promise<RespBase> {
  return post('/admin/trade/symbols/spot-config', params)
}

export function apiTradeSetContractConfig(params: SetContractSymbolConfigReq): Promise<RespBase> {
  return post('/admin/trade/symbols/contract-config', params)
}

export function apiTradeSetSecondsConfig(params: SetSecondsSymbolConfigReq): Promise<RespBase> {
  return post('/admin/trade/symbols/seconds-config', params)
}

export function apiTradeListSymbolLeverageConfigs(
  params: GetSymbolLeverageConfigListReq,
): Promise<RespBase<TradeSymbolLeverageConfig[]>> {
  return get<TradeSymbolLeverageConfig[]>('/admin/trade/symbols/leverage-configs', params)
}

export function apiTradeSetSymbolLeverageConfig(
  params: SetSymbolLeverageConfigReq,
): Promise<RespBase> {
  return post('/admin/trade/symbols/leverage-config', params)
}

export function apiTradeListOrders(params: GetOrderListAdminReq): Promise<RespBase<TradeOrder[]>> {
  return get<TradeOrder[]>('/admin/trade/orders', params)
}

export function apiTradeGetOrder(
  params: GetOrderDetailAdminReq,
): Promise<RespBase<TradeOrderDetailData>> {
  return get<TradeOrderDetailData>('/admin/trade/orders/detail', params)
}

export function apiTradeListFills(params: GetFillListAdminReq): Promise<RespBase<TradeFill[]>> {
  return get<TradeFill[]>('/admin/trade/fills', params)
}

export function apiTradeGetFill(params: GetFillDetailAdminReq): Promise<RespBase<TradeFill>> {
  return get<TradeFill>('/admin/trade/fills/detail', params)
}

export function apiTradeListPositions(
  params: GetPositionListAdminReq,
): Promise<RespBase<ContractPosition[]>> {
  return get<ContractPosition[]>('/admin/trade/positions', params)
}

export function apiTradeGetPosition(
  params: GetPositionDetailAdminReq,
): Promise<RespBase<ContractPosition>> {
  return get<ContractPosition>('/admin/trade/positions/detail', params)
}

export function apiTradeListPositionHistories(
  params: GetPositionHistoryListAdminReq,
): Promise<RespBase<ContractPositionHistory[]>> {
  return get<ContractPositionHistory[]>('/admin/trade/position-histories', params)
}

export function apiTradeListMarginSnapshots(
  params: GetMarginSnapshotListAdminReq,
): Promise<RespBase<ContractMarginSnapshot[]>> {
  return get<ContractMarginSnapshot[]>('/admin/trade/margin-snapshots', params)
}

export function apiTradeListCancelLogs(
  params: GetCancelLogListAdminReq,
): Promise<RespBase<TradeCancelLog[]>> {
  return get<TradeCancelLog[]>('/admin/trade/cancel-logs', params)
}

export function apiTradeGetUserTradeLimit(
  params: GetUserTradeLimitReq,
): Promise<RespBase<RiskUserTradeLimit>> {
  return get<RiskUserTradeLimit>('/admin/trade/user-trade-limit', params)
}

export function apiTradeSetUserTradeLimit(params: SetUserTradeLimitReq): Promise<RespBase> {
  return post('/admin/trade/user-trade-limit', params)
}

export function apiTradeGetUserSymbolLimit(
  params: GetUserSymbolLimitReq,
): Promise<RespBase<RiskUserSymbolLimit>> {
  return get<RiskUserSymbolLimit>('/admin/trade/user-symbol-limit', params)
}

export function apiTradeSetUserSymbolLimit(params: SetUserSymbolLimitReq): Promise<RespBase> {
  return post('/admin/trade/user-symbol-limit', params)
}

export function apiTradeGetUserTradeConfig(
  params: GetUserTradeConfigReq,
): Promise<RespBase<TradeUserConfig>> {
  return get<TradeUserConfig>('/admin/trade/user-trade-config', params)
}

export function apiTradeSetUserTradeConfig(params: SetUserTradeConfigReq): Promise<RespBase> {
  return post('/admin/trade/user-trade-config', params)
}

export function apiTradeListRiskLogs(
  params: GetRiskOrderCheckLogListReq,
): Promise<RespBase<RiskOrderCheckLog[]>> {
  return get<RiskOrderCheckLog[]>('/admin/trade/risk-order-check-logs', params)
}

export function apiTradeGetUserLeverageConfig(
  params: GetUserLeverageConfigReq,
): Promise<RespBase<ContractLeverageConfig>> {
  return get<ContractLeverageConfig>('/admin/trade/user-leverage-config', params)
}

export function apiTradeSetUserLeverageConfig(params: SetUserLeverageConfigReq): Promise<RespBase> {
  return post('/admin/trade/user-leverage-config', params)
}

export function apiTradeListEvents(
  params: GetTradeEventListReq,
): Promise<RespBase<BizTradeEvent[]>> {
  return get<BizTradeEvent[]>('/admin/trade/events', params)
}

export function apiTradeGetEvent(params: GetTradeEventDetailReq): Promise<RespBase<BizTradeEvent>> {
  return get<BizTradeEvent>('/admin/trade/events/detail', params)
}

export function apiTradeRetryEvent(params: RetryTradeEventReq): Promise<RespBase> {
  return post('/admin/trade/events/retry', params)
}

export function apiTradeListRiskTiers(
  params: GetContractRiskLimitTierListReq,
): Promise<GetContractRiskLimitTierListResp> {
  return get<ContractRiskLimitTier[]>('/admin/trade/risk-tiers', params)
}

export function apiTradeSetRiskTier(params: SetContractRiskLimitTierReq): Promise<RespBase> {
  return post('/admin/trade/risk-tiers', params)
}

export function apiTradeListFundingBatches(
  params: GetFundingBatchListReq,
): Promise<GetFundingBatchListResp> {
  return get<ContractFundingBatch[]>('/admin/trade/funding/batches', params)
}

export function apiTradeListFundingSettlements(
  params: GetFundingSettlementListReq,
): Promise<GetFundingSettlementListResp> {
  return get<ContractFundingSettlement[]>('/admin/trade/funding/settlements', params)
}

export function apiTradeListDeliveryBatches(
  params: GetDeliveryBatchListReq,
): Promise<GetDeliveryBatchListResp> {
  return get<ContractDeliveryBatch[]>('/admin/trade/delivery/batches', params)
}

export function apiTradeListDeliverySettlements(
  params: GetDeliverySettlementListReq,
): Promise<GetDeliverySettlementListResp> {
  return get<ContractDeliverySettlement[]>('/admin/trade/delivery/settlements', params)
}

export function apiTradeListLiquidations(
  params: GetLiquidationListReq,
): Promise<GetLiquidationListResp> {
  return get<ContractLiquidation[]>('/admin/trade/liquidations', params)
}

export function apiTradeListAccountLiquidations(
  params: GetAccountLiquidationListReq,
): Promise<GetAccountLiquidationListResp> {
  return get<ContractAccountLiquidation[]>('/admin/trade/account-liquidations', params)
}

export function apiTradeGetAccountLiquidationDetail(
  params: GetAccountLiquidationDetailReq,
): Promise<GetAccountLiquidationDetailResp> {
  return get<ContractAccountLiquidation>(
    '/admin/trade/account-liquidations/detail',
    params,
  ) as Promise<GetAccountLiquidationDetailResp>
}

export function apiTradeRetryAccountLiquidation(
  params: RetryAccountLiquidationReq,
): Promise<RespBase> {
  return post('/admin/trade/account-liquidations/retry', params)
}

export function apiTradeListSecondsPriceSnapshots(
  params: GetSecondsPriceSnapshotListReq,
): Promise<GetSecondsPriceSnapshotListResp> {
  return get<TradeSecondsPriceSnapshot[]>('/admin/trade/seconds/price-snapshots', params)
}

export function apiTradeListAssetReservations(
  params: GetAssetReservationListReq,
): Promise<GetAssetReservationListResp> {
  return get<TradeAssetReservation[]>('/admin/trade/operations/asset-reservations', params)
}

export function apiTradeListSettlementInstructions(
  params: GetSettlementInstructionListReq,
): Promise<GetSettlementInstructionListResp> {
  return get<TradeSettlementInstruction[]>(
    '/admin/trade/operations/settlement-instructions',
    params,
  )
}

export function apiTradeRetrySettlementInstruction(
  params: RetrySettlementInstructionReq,
): Promise<RespBase> {
  return post('/admin/trade/operations/settlement-instructions/retry', params)
}

export function apiTradeListContractReconciliationIssues(
  params: GetContractReconciliationIssueListReq,
): Promise<GetContractReconciliationIssueListResp> {
  return get<ContractReconciliationIssue[]>('/admin/trade/operations/reconciliation-issues', params)
}

export function apiTradeIgnoreContractReconciliationIssue(
  params: IgnoreContractReconciliationIssueReq,
): Promise<RespBase> {
  return post('/admin/trade/operations/reconciliation-issues/ignore', params)
}

export function apiTradeListInsuranceFundAccounts(
  params: GetInsuranceFundAccountListReq,
): Promise<GetInsuranceFundAccountListResp> {
  return get<InsuranceFundAccount[]>('/admin/trade/insurance-fund/accounts', params)
}
export function apiTradeSetInsuranceFundAccount(
  params: SetInsuranceFundAccountReq,
): Promise<RespBase> {
  return post('/admin/trade/insurance-fund/accounts', params)
}
export function apiTradeListMarketSnapshots(
  params: GetMarketSnapshotListReq,
): Promise<GetMarketSnapshotListResp> {
  return get<TradeMarketSnapshot[]>('/admin/trade/market-snapshots', params)
}
