import { authHttp, http } from "./http";
import { compactParams } from "./utils";
import type { RespBase } from "../types/api";
import type {
  OptionCancelOrderReq,
  OptionCancelComboOrderReq,
  ExerciseReq,
  SetExerciseInstructionReq,
  GetExerciseInstructionReq,
  ActivateKillSwitchReq,
  GetContractDetailReq,
  GetOrderBookReq,
  OptionGetOrderDetailReq,
  GetPositionDetailReq,
  ListAccountsReq,
  ListBillsReq,
  ListContractsReq,
  ListOptionChainReq,
  ListCurrentOrdersReq,
  ListExercisesReq,
  ListHistoryOrdersReq,
  ListPositionsReq,
  ListTradesReq,
  OptionPlaceOrderReq,
  OptionPlaceComboOrderReq,
  OptionGetComboOrderReq,
  OptionListComboOrdersReq,
  OptionAccount,
  OptionBill,
  OptionContractDetail,
  OptionChainRow,
  OptionExerciseDetail,
  OptionExerciseInstruction,
  OptionUserTradingControl,
  OptionMMPConfig,
  GetMMPConfigReq,
  OptionOrderDetail,
  OptionOrderBook,
  OptionComboOrderDetail,
  OptionPositionDetail,
  OptionTradeDetail,
} from "../types/option";

export function apiOptionListContracts(
  params: ListContractsReq,
): Promise<RespBase & { data: OptionContractDetail[] }> {
  return http
    .get("/option/contracts", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetContractDetail(
  params: GetContractDetailReq,
): Promise<RespBase & { data: OptionContractDetail }> {
  return http
    .get("/option/contracts/detail", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListOptionChain(params: ListOptionChainReq): Promise<
  RespBase & {
    data: OptionChainRow[];
    generatedAt: number;
    statisticsWindowStart: number;
  }
> {
  return http
    .get("/option/chain", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetOrderBook(
  params: GetOrderBookReq,
): Promise<RespBase & { data: OptionOrderBook }> {
  return http
    .get("/option/order-book", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionPlaceOrder(
  params: OptionPlaceOrderReq,
): Promise<RespBase & { data: { orderNo: string; orderId: number } }> {
  return authHttp
    .post("/option/orders", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionCancelOrder(
  params: OptionCancelOrderReq,
): Promise<RespBase> {
  return authHttp
    .post("/option/orders/cancel", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionPlaceComboOrder(
  params: OptionPlaceComboOrderReq,
): Promise<RespBase & { data: OptionComboOrderDetail }> {
  return authHttp
    .post("/option/combo-orders", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionCancelComboOrder(
  params: OptionCancelComboOrderReq,
): Promise<RespBase> {
  return authHttp
    .post("/option/combo-orders/cancel", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetComboOrder(
  params: OptionGetComboOrderReq,
): Promise<RespBase & { data: OptionComboOrderDetail }> {
  return authHttp
    .get("/option/combo-orders/detail", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListComboOrders(
  params: OptionListComboOrdersReq,
): Promise<RespBase & { data: OptionComboOrderDetail[] }> {
  return authHttp
    .get("/option/combo-orders", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetOrderDetail(
  params: OptionGetOrderDetailReq,
): Promise<RespBase & { data: OptionOrderDetail }> {
  return authHttp
    .get("/option/orders/detail", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListCurrentOrders(
  params: ListCurrentOrdersReq,
): Promise<RespBase & { data: OptionOrderDetail[] }> {
  return authHttp
    .get("/option/orders/current", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListHistoryOrders(
  params: ListHistoryOrdersReq,
): Promise<RespBase & { data: OptionOrderDetail[] }> {
  return authHttp
    .get("/option/orders/history", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListTrades(
  params: ListTradesReq,
): Promise<RespBase & { data: OptionTradeDetail[] }> {
  return authHttp
    .get("/option/trades", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListPositions(
  params: ListPositionsReq,
): Promise<RespBase & { data: OptionPositionDetail[] }> {
  return authHttp
    .get("/option/positions", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetPositionDetail(
  params: GetPositionDetailReq,
): Promise<RespBase & { data: OptionPositionDetail }> {
  return authHttp
    .get("/option/positions/detail", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionExercise(
  params: ExerciseReq,
): Promise<RespBase & { data: { exerciseNo: string; exerciseId: number } }> {
  return authHttp
    .post("/option/exercise", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionSetExerciseInstruction(
  params: SetExerciseInstructionReq,
): Promise<RespBase & { data: OptionExerciseInstruction }> {
  return authHttp
    .post("/option/exercise/instruction", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetExerciseInstruction(
  params: GetExerciseInstructionReq,
): Promise<RespBase & { data: OptionExerciseInstruction }> {
  return authHttp
    .get("/option/exercise/instruction", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetUserTradingControl(): Promise<
  RespBase & { data: OptionUserTradingControl }
> {
  return authHttp
    .get("/option/trading-control")
    .then((res: { data: any }) => res.data);
}

export function apiOptionGetMMPConfig(
  params: GetMMPConfigReq,
): Promise<RespBase & { data: OptionMMPConfig }> {
  return authHttp
    .get("/option/mmp/config", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionActivateKillSwitch(
  params: ActivateKillSwitchReq = {},
): Promise<RespBase & { data: OptionUserTradingControl }> {
  return authHttp
    .post("/option/kill-switch", params)
    .then((res: { data: any }) => res.data);
}

export function apiOptionListExercises(
  params: ListExercisesReq,
): Promise<RespBase & { data: OptionExerciseDetail[] }> {
  return authHttp
    .get("/option/exercises", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListAccounts(
  params: ListAccountsReq,
): Promise<RespBase & { data: OptionAccount[] }> {
  return authHttp
    .get("/option/accounts", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}

export function apiOptionListBills(
  params: ListBillsReq,
): Promise<RespBase & { data: OptionBill[] }> {
  return authHttp
    .get("/option/bills", { params: compactParams(params) })
    .then((res: { data: any }) => res.data);
}
