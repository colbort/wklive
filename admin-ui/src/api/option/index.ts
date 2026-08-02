import { get, post } from '@/utils/request'
import type {
  CreateContractReq,
  CreateSettlementPriceCorrectionReq,
  ForceCancelComboOrderReq,
  GetAccountReq,
  GetAdminComboOrderReq,
  GetBillReq,
  GetContractReq,
  GetExerciseReq,
  GetMarketReq,
  GetOrderReq,
  GetPositionReq,
  GetSettlementReq,
  GetTradeReq,
  ListAccountsReq,
  ListAdminComboOrdersReq,
  ListBillsReq,
  ListContractsReq,
  ListExercisesReq,
  ListMarketSnapshotsReq,
  ListOrdersReq,
  ListPositionsReq,
  ListSettlementsReq,
  ListSettlementPricesReq,
  ListTradesReq,
  OptionAccount,
  OptionAdminComboOrderDetail,
  OptionAdminCommonResp,
  OptionBill,
  OptionContractDetail,
  OptionComboOrderDetail,
  OptionExercise,
  OptionExerciseDetail,
  OptionForceCancelContractOrdersReq,
  OptionReleaseUserKillSwitchReq,
  OptionUserTradingControl,
  OptionTradingControlEvent,
  AdminGetUserTradingControlReq,
  ListTradingControlEventsReq,
  UpsertMMPConfigReq,
  ResetMMPConfigReq,
  ListMMPConfigsReq,
  OptionMMPConfig,
  CreatePortfolioRiskConfigReq,
  ReviewPortfolioRiskConfigReq,
  ListPortfolioRiskConfigsReq,
  OptionPortfolioRiskConfig,
  CreateInsuranceInventoryExitReq,
  ReviewInsuranceInventoryExitReq,
  ExecuteInsuranceInventoryExitReq,
  ListInsuranceInventoryExitsReq,
  OptionInsuranceInventoryExit,
  CreateTradeCorrectionReq,
  ReviewTradeCorrectionReq,
  ListTradeCorrectionsReq,
  OptionTradeCorrection,
  OptionRetryAssetInstructionReq,
  GetOperationsOverviewReq,
  ListAssetInstructionsReq,
  ListReconciliationIssuesReq,
  CreateTradingCalendarReq,
  ReviewTradingCalendarReq,
  ListTradingCalendarsReq,
  OptionTradingCalendar,
  HaltContractTradingReq,
  ResumeContractTradingReq,
  ListTradingHaltsReq,
  OptionTradingHalt,
  CreateCorporateActionReq,
  ReviewCorporateActionReq,
  ListCorporateActionsReq,
  ListCorporateActionPositionsReq,
  OptionCorporateAction,
  OptionCorporateActionPosition,
  CreateContractSeriesReq,
  ReviewContractSeriesReq,
  ListContractSeriesReq,
  ListContractSeriesDetailsReq,
  OptionContractSeries,
  OptionContractSeriesDetail,
  ReviewContractSeriesLaunchReq,
  OptionOperationsOverview,
  OptionAssetInstruction,
  OptionReconciliationIssue,
  OptionRetrySettlementInstructionReq,
  ListPhysicalDeliveryUnitsReq,
  RetryPhysicalDeliveryUnitReq,
  OptionPhysicalDeliveryUnit,
  OptionRetryTradeEventReq,
  OptionRetryLiquidationReq,
  OptionRetryExerciseReq,
  OptionRiskAccount,
  OptionLiquidation,
  ListOptionRiskAccountsReq,
  ListOptionLiquidationsReq,
  OptionMarket,
  OptionMarketSnapshot,
  OptionOrder,
  OptionOrderDetail,
  OptionPosition,
  OptionPositionDetail,
  OptionSettlement,
  OptionSettlementPrice,
  OptionSettlementDetail,
  OptionTrade,
  OptionTradeDetail,
  RespBase,
  UpdateContractReq,
  UpdateMarketReq,
  ReviewSettlementPriceReq,
} from '@/services'

export function apiOptionListContracts(
  params: ListContractsReq,
): Promise<RespBase<OptionContractDetail[]>> {
  return get<OptionContractDetail[]>('/admin/option/contracts', params)
}

export function apiOptionGetContract(
  params: GetContractReq,
): Promise<RespBase<OptionContractDetail>> {
  return get<OptionContractDetail>('/admin/option/contracts/detail', params)
}

export function apiOptionCreateContract(
  params: CreateContractReq,
): Promise<RespBase<{ id: number }>> {
  return post<{ id: number }>('/admin/option/contracts', params)
}

export function apiOptionUpdateContract(params: UpdateContractReq): Promise<OptionAdminCommonResp> {
  return post('/admin/option/contracts/update', params)
}

export function apiOptionForceCancelContractOrders(
  params: OptionForceCancelContractOrdersReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/contracts/force-cancel-orders', params)
}

export function apiOptionReleaseUserKillSwitch(
  params: OptionReleaseUserKillSwitchReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/trading-controls/release-kill-switch', params)
}

export function apiOptionGetUserTradingControl(
  params: AdminGetUserTradingControlReq,
): Promise<RespBase<OptionUserTradingControl>> {
  return get<OptionUserTradingControl>('/admin/option/trading-controls/detail', params)
}

export function apiOptionListTradingControlEvents(
  params: ListTradingControlEventsReq,
): Promise<RespBase<OptionTradingControlEvent[]>> {
  return get<OptionTradingControlEvent[]>('/admin/option/trading-controls/events', params)
}

export function apiOptionUpsertMMPConfig(
  params: UpsertMMPConfigReq,
): Promise<RespBase<OptionMMPConfig>> {
  return post<OptionMMPConfig>('/admin/option/mmp/config', params)
}

export function apiOptionResetMMPConfig(
  params: ResetMMPConfigReq,
): Promise<RespBase<OptionMMPConfig>> {
  return post<OptionMMPConfig>('/admin/option/mmp/reset', params)
}

export function apiOptionListMMPConfigs(
  params: ListMMPConfigsReq,
): Promise<RespBase<OptionMMPConfig[]>> {
  return get<OptionMMPConfig[]>('/admin/option/mmp/configs', params)
}

export function apiOptionCreatePortfolioRiskConfig(
  params: CreatePortfolioRiskConfigReq,
): Promise<RespBase<OptionPortfolioRiskConfig>> {
  return post<OptionPortfolioRiskConfig>('/admin/option/risk/portfolio-configs', params)
}

export function apiOptionReviewPortfolioRiskConfig(
  params: ReviewPortfolioRiskConfigReq,
): Promise<RespBase<OptionPortfolioRiskConfig>> {
  return post<OptionPortfolioRiskConfig>('/admin/option/risk/portfolio-configs/review', params)
}

export function apiOptionListPortfolioRiskConfigs(
  params: ListPortfolioRiskConfigsReq,
): Promise<RespBase<OptionPortfolioRiskConfig[]>> {
  return get<OptionPortfolioRiskConfig[]>('/admin/option/risk/portfolio-configs', params)
}

export function apiOptionCreateInsuranceInventoryExit(
  params: CreateInsuranceInventoryExitReq,
): Promise<RespBase<OptionInsuranceInventoryExit>> {
  return post<OptionInsuranceInventoryExit>('/admin/option/risk/insurance-inventory-exits', params)
}

export function apiOptionReviewInsuranceInventoryExit(
  params: ReviewInsuranceInventoryExitReq,
): Promise<RespBase<OptionInsuranceInventoryExit>> {
  return post<OptionInsuranceInventoryExit>(
    '/admin/option/risk/insurance-inventory-exits/review',
    params,
  )
}

export function apiOptionExecuteInsuranceInventoryExit(
  params: ExecuteInsuranceInventoryExitReq,
): Promise<RespBase<OptionInsuranceInventoryExit>> {
  return post<OptionInsuranceInventoryExit>(
    '/admin/option/risk/insurance-inventory-exits/execute',
    params,
  )
}

export function apiOptionListInsuranceInventoryExits(
  params: ListInsuranceInventoryExitsReq,
): Promise<RespBase<OptionInsuranceInventoryExit[]>> {
  return get<OptionInsuranceInventoryExit[]>('/admin/option/risk/insurance-inventory-exits', params)
}

export function apiOptionCreateTradeCorrection(
  params: CreateTradeCorrectionReq,
): Promise<RespBase<OptionTradeCorrection>> {
  return post<OptionTradeCorrection>('/admin/option/trade-corrections', params)
}

export function apiOptionReviewTradeCorrection(
  params: ReviewTradeCorrectionReq,
): Promise<RespBase<OptionTradeCorrection>> {
  return post<OptionTradeCorrection>('/admin/option/trade-corrections/review', params)
}

export function apiOptionListTradeCorrections(
  params: ListTradeCorrectionsReq,
): Promise<RespBase<OptionTradeCorrection[]>> {
  return get<OptionTradeCorrection[]>('/admin/option/trade-corrections', params)
}

export function apiOptionRetryAssetInstruction(
  params: OptionRetryAssetInstructionReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/recovery/asset-instructions/retry', params)
}

export function apiOptionGetOperationsOverview(
  params: GetOperationsOverviewReq,
): Promise<RespBase<OptionOperationsOverview>> {
  return get<OptionOperationsOverview>('/admin/option/operations/overview', params)
}

export function apiOptionListAssetInstructions(
  params: ListAssetInstructionsReq,
): Promise<RespBase<OptionAssetInstruction[]>> {
  return get<OptionAssetInstruction[]>('/admin/option/operations/asset-instructions', params)
}

export function apiOptionListReconciliationIssues(
  params: ListReconciliationIssuesReq,
): Promise<RespBase<OptionReconciliationIssue[]>> {
  return get<OptionReconciliationIssue[]>('/admin/option/operations/reconciliation-issues', params)
}

export function apiOptionCreateTradingCalendar(
  params: CreateTradingCalendarReq,
): Promise<RespBase<OptionTradingCalendar>> {
  return post<OptionTradingCalendar>('/admin/option/trading-calendars', params)
}

export function apiOptionReviewTradingCalendar(
  params: ReviewTradingCalendarReq,
): Promise<RespBase<OptionTradingCalendar>> {
  return post<OptionTradingCalendar>('/admin/option/trading-calendars/review', params)
}

export function apiOptionListTradingCalendars(
  params: ListTradingCalendarsReq,
): Promise<RespBase<OptionTradingCalendar[]>> {
  return get<OptionTradingCalendar[]>('/admin/option/trading-calendars', params)
}

export function apiOptionHaltContractTrading(
  params: HaltContractTradingReq,
): Promise<RespBase<OptionTradingHalt>> {
  return post<OptionTradingHalt>('/admin/option/trading-halts', params)
}

export function apiOptionResumeContractTrading(
  params: ResumeContractTradingReq,
): Promise<RespBase<OptionTradingHalt>> {
  return post<OptionTradingHalt>('/admin/option/trading-halts/resume', params)
}

export function apiOptionListTradingHalts(
  params: ListTradingHaltsReq,
): Promise<RespBase<OptionTradingHalt[]>> {
  return get<OptionTradingHalt[]>('/admin/option/trading-halts', params)
}

export function apiOptionCreateCorporateAction(
  params: CreateCorporateActionReq,
): Promise<RespBase<OptionCorporateAction>> {
  return post<OptionCorporateAction>('/admin/option/corporate-actions', params)
}

export function apiOptionReviewCorporateAction(
  params: ReviewCorporateActionReq,
): Promise<RespBase<OptionCorporateAction>> {
  return post<OptionCorporateAction>('/admin/option/corporate-actions/review', params)
}

export function apiOptionListCorporateActions(
  params: ListCorporateActionsReq,
): Promise<RespBase<OptionCorporateAction[]>> {
  return get<OptionCorporateAction[]>('/admin/option/corporate-actions', params)
}

export function apiOptionListCorporateActionPositions(
  params: ListCorporateActionPositionsReq,
): Promise<RespBase<OptionCorporateActionPosition[]>> {
  return get<OptionCorporateActionPosition[]>('/admin/option/corporate-actions/positions', params)
}

export function apiOptionCreateContractSeries(
  params: CreateContractSeriesReq,
): Promise<RespBase<OptionContractSeries>> {
  return post<OptionContractSeries>('/admin/option/contract-series', params)
}

export function apiOptionReviewContractSeries(
  params: ReviewContractSeriesReq,
): Promise<RespBase<OptionContractSeries>> {
  return post<OptionContractSeries>('/admin/option/contract-series/review', params)
}

export function apiOptionListContractSeries(
  params: ListContractSeriesReq,
): Promise<RespBase<OptionContractSeries[]>> {
  return get<OptionContractSeries[]>('/admin/option/contract-series', params)
}

export function apiOptionListContractSeriesDetails(
  params: ListContractSeriesDetailsReq,
): Promise<RespBase<OptionContractSeriesDetail[]>> {
  return get<OptionContractSeriesDetail[]>('/admin/option/contract-series/details', params)
}

export function apiOptionReviewContractSeriesLaunch(
  params: ReviewContractSeriesLaunchReq,
): Promise<RespBase<OptionContractSeries>> {
  return post<OptionContractSeries>('/admin/option/contract-series/launch-review', params)
}

export function apiOptionRetrySettlementInstruction(
  params: OptionRetrySettlementInstructionReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/settlements/retry-instruction', params)
}

export function apiOptionListPhysicalDeliveryUnits(
  params: ListPhysicalDeliveryUnitsReq,
): Promise<RespBase<OptionPhysicalDeliveryUnit[]>> {
  return get<OptionPhysicalDeliveryUnit[]>('/admin/option/physical-delivery/units', params)
}

export function apiOptionRetryPhysicalDeliveryUnit(
  params: RetryPhysicalDeliveryUnitReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/physical-delivery/units/retry', params)
}

export function apiOptionRetryTradeEvent(
  params: OptionRetryTradeEventReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/recovery/trade-events/retry', params)
}

export function apiOptionListRiskAccounts(
  params: ListOptionRiskAccountsReq,
): Promise<RespBase<OptionRiskAccount[]>> {
  return get<OptionRiskAccount[]>('/admin/option/risk/accounts', params)
}

export function apiOptionListLiquidations(
  params: ListOptionLiquidationsReq,
): Promise<RespBase<OptionLiquidation[]>> {
  return get<OptionLiquidation[]>('/admin/option/risk/liquidations', params)
}

export function apiOptionRetryLiquidation(
  params: OptionRetryLiquidationReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/risk/liquidations/retry', params)
}

export function apiOptionGetMarket(params: GetMarketReq): Promise<RespBase<OptionMarket>> {
  return get<OptionMarket>('/admin/option/market/detail', params)
}

export function apiOptionUpdateMarket(params: UpdateMarketReq): Promise<OptionAdminCommonResp> {
  return post('/admin/option/market/update', params)
}

export function apiOptionListMarketSnapshots(
  params: ListMarketSnapshotsReq,
): Promise<RespBase<OptionMarketSnapshot[]>> {
  return get<OptionMarketSnapshot[]>('/admin/option/market/snapshots', params)
}

export function apiOptionListOrders(params: ListOrdersReq): Promise<RespBase<OptionOrder[]>> {
  return get<OptionOrder[]>('/admin/option/orders', params)
}

export function apiOptionGetOrder(params: GetOrderReq): Promise<RespBase<OptionOrderDetail>> {
  return get<OptionOrderDetail>('/admin/option/orders/detail', params)
}

export function apiOptionListAdminComboOrders(
  params: ListAdminComboOrdersReq,
): Promise<RespBase<OptionComboOrderDetail[]>> {
  return get<OptionComboOrderDetail[]>('/admin/option/combo-orders', params)
}

export function apiOptionGetAdminComboOrder(
  params: GetAdminComboOrderReq,
): Promise<RespBase<OptionAdminComboOrderDetail>> {
  return get<OptionAdminComboOrderDetail>('/admin/option/combo-orders/detail', params)
}

export function apiOptionForceCancelComboOrder(
  params: ForceCancelComboOrderReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/combo-orders/force-cancel', params)
}

export function apiOptionListTrades(params: ListTradesReq): Promise<RespBase<OptionTrade[]>> {
  return get<OptionTrade[]>('/admin/option/trades', params)
}

export function apiOptionGetTrade(params: GetTradeReq): Promise<RespBase<OptionTradeDetail>> {
  return get<OptionTradeDetail>('/admin/option/trades/detail', params)
}

export function apiOptionListPositions(
  params: ListPositionsReq,
): Promise<RespBase<OptionPosition[]>> {
  return get<OptionPosition[]>('/admin/option/positions', params)
}

export function apiOptionGetPosition(
  params: GetPositionReq,
): Promise<RespBase<OptionPositionDetail>> {
  return get<OptionPositionDetail>('/admin/option/positions/detail', params)
}

export function apiOptionListExercises(
  params: ListExercisesReq,
): Promise<RespBase<OptionExercise[]>> {
  return get<OptionExercise[]>('/admin/option/exercises', params)
}

export function apiOptionGetExercise(
  params: GetExerciseReq,
): Promise<RespBase<OptionExerciseDetail>> {
  return get<OptionExerciseDetail>('/admin/option/exercises/detail', params)
}

export function apiOptionRetryExercise(
  params: OptionRetryExerciseReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/exercises/retry', params)
}

export function apiOptionListSettlements(
  params: ListSettlementsReq,
): Promise<RespBase<OptionSettlement[]>> {
  return get<OptionSettlement[]>('/admin/option/settlements', params)
}

export function apiOptionListSettlementPrices(
  params: ListSettlementPricesReq,
): Promise<RespBase<OptionSettlementPrice[]>> {
  return get<OptionSettlementPrice[]>('/admin/option/settlement-prices', params)
}

export function apiOptionCreateSettlementPriceCorrection(
  params: CreateSettlementPriceCorrectionReq,
): Promise<RespBase<OptionSettlementPrice>> {
  return post<OptionSettlementPrice>('/admin/option/settlement-prices/corrections', params)
}

export function apiOptionReviewSettlementPrice(
  params: ReviewSettlementPriceReq,
): Promise<RespBase<OptionSettlementPrice>> {
  return post<OptionSettlementPrice>('/admin/option/settlement-prices/review', params)
}

export function apiOptionGetSettlement(
  params: GetSettlementReq,
): Promise<RespBase<OptionSettlementDetail>> {
  return get<OptionSettlementDetail>('/admin/option/settlements/detail', params)
}

export function apiOptionListAccounts(params: ListAccountsReq): Promise<RespBase<OptionAccount[]>> {
  return get<OptionAccount[]>('/admin/option/accounts', params)
}

export function apiOptionGetAccount(params: GetAccountReq): Promise<RespBase<OptionAccount>> {
  return get<OptionAccount>('/admin/option/accounts/detail', params)
}

export function apiOptionListBills(params: ListBillsReq): Promise<RespBase<OptionBill[]>> {
  return get<OptionBill[]>('/admin/option/bills', params)
}

export function apiOptionGetBill(params: GetBillReq): Promise<RespBase<OptionBill>> {
  return get<OptionBill>('/admin/option/bills/detail', params)
}
