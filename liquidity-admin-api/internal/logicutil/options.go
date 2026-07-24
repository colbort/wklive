package logicutil

import (
	"sort"

	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/liquidity"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func LiquidityOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		enumGroup("productType", "产品类型", common.ProductType_PRODUCT_TYPE_UNKNOWN.Descriptor()),
		enumGroup("contractType", "合约期限类型", common.ContractType_CONTRACT_TYPE_NOT_APPLICABLE.Descriptor()),
		enumGroup("tradeSide", "买卖方向", common.Side_SIDE_UNKNOWN.Descriptor()),
		enumGroup("yesNo", "是否", common.YesNo_YES_NO_UNKNOWN.Descriptor()),
		enumGroup("enabled", "启用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		enumGroup("providerType", "提供方类型", liquidity.ProviderType_PROVIDER_TYPE_UNKNOWN.Descriptor()),
		enumGroup("providerEnvironment", "提供方环境", liquidity.ProviderEnvironment_PROVIDER_ENVIRONMENT_UNKNOWN.Descriptor()),
		enumGroup("providerStatus", "提供方状态", liquidity.ProviderStatus_PROVIDER_STATUS_UNKNOWN.Descriptor()),
		enumGroup("healthStatus", "健康状态", liquidity.HealthStatus_HEALTH_STATUS_UNKNOWN.Descriptor()),
		enumGroup("liquidityMode", "流动性模式", liquidity.LiquidityMode_LIQUIDITY_MODE_UNKNOWN.Descriptor()),
		enumGroup("symbolLiquidityStatus", "策略状态", liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_UNKNOWN.Descriptor()),
		enumGroup("quoteCycleStatus", "报价周期状态", liquidity.QuoteCycleStatus_QUOTE_CYCLE_STATUS_UNKNOWN.Descriptor()),
		enumGroup("quoteOrderStatus", "报价单状态", liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNKNOWN.Descriptor()),
		enumGroup("externalOrderPurpose", "外部订单用途", liquidity.ExternalOrderPurpose_EXTERNAL_ORDER_PURPOSE_UNKNOWN.Descriptor()),
		enumGroup("externalOrderType", "外部订单类型", liquidity.ExternalOrderType_EXTERNAL_ORDER_TYPE_UNKNOWN.Descriptor()),
		enumGroup("externalTimeInForce", "外部订单有效方式", liquidity.ExternalTimeInForce_EXTERNAL_TIME_IN_FORCE_UNKNOWN.Descriptor()),
		enumGroup("externalOrderStatus", "外部订单状态", liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_UNKNOWN.Descriptor()),
		enumGroup("externalFillSettlementStatus", "外部成交结算状态", liquidity.ExternalFillSettlementStatus_EXTERNAL_FILL_SETTLEMENT_STATUS_UNKNOWN.Descriptor()),
		enumGroup("hedgeTriggerType", "对冲触发类型", liquidity.HedgeTriggerType_HEDGE_TRIGGER_TYPE_UNKNOWN.Descriptor()),
		enumGroup("hedgeStatus", "对冲状态", liquidity.HedgeStatus_HEDGE_STATUS_UNKNOWN.Descriptor()),
		enumGroup("inventorySource", "库存来源", liquidity.InventorySource_INVENTORY_SOURCE_UNKNOWN.Descriptor()),
		enumGroup("liquidityRiskLevel", "风险等级", liquidity.RiskLevel_RISK_LEVEL_UNKNOWN.Descriptor()),
		enumGroup("riskActionType", "风险动作", liquidity.RiskActionType_RISK_ACTION_TYPE_UNKNOWN.Descriptor()),
		enumGroup("riskEventStatus", "风险事件状态", liquidity.RiskEventStatus_RISK_EVENT_STATUS_UNKNOWN.Descriptor()),
		enumGroup("reconcileType", "对账类型", liquidity.ReconcileType_RECONCILE_TYPE_UNKNOWN.Descriptor()),
		enumGroup("reconcileStatus", "对账状态", liquidity.ReconcileStatus_RECONCILE_STATUS_UNKNOWN.Descriptor()),
		enumGroup("reconcileDifferenceType", "对账差异类型", liquidity.ReconcileDifferenceType_RECONCILE_DIFFERENCE_TYPE_UNKNOWN.Descriptor()),
		enumGroup("reconcileDifferenceStatus", "对账差异状态", liquidity.ReconcileDifferenceStatus_RECONCILE_DIFFERENCE_STATUS_UNKNOWN.Descriptor()),
	}
}

func enumGroup(key, label string, descriptor protoreflect.EnumDescriptor) types.OptionsGroup {
	values := make([]int, 0, descriptor.Values().Len())
	codes := make(map[int32]string, descriptor.Values().Len())
	for i := 0; i < descriptor.Values().Len(); i++ {
		value := descriptor.Values().Get(i)
		number := int32(value.Number())
		values = append(values, int(number))
		codes[number] = string(value.Name())
	}
	sort.Ints(values)
	items := make([]types.OptionsItem, 0, len(values))
	for _, raw := range values {
		value := int32(raw)
		items = append(items, types.OptionsItem{Value: value, Code: codes[value]})
	}
	return types.OptionsGroup{Key: key, Label: label, Options: items}
}
