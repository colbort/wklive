package logicutil

import (
	"wklive/app-api/internal/types"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/payment"
	"wklive/proto/staking"
	"wklive/proto/trade"
	"wklive/proto/user"
)

func CoreOptions() []types.OptionsGroup {
	options := make([]types.OptionsGroup, 0)
	options = append(options, CommonOptions()...)
	options = append(options, UserOptions()...)
	options = append(options, AssetOptions()...)
	options = append(options, PaymentOptions()...)
	options = append(options, TradeOptions()...)
	options = append(options, OptionOptions()...)
	options = append(options, StakingOptions()...)
	options = append(options, ItickOptions()...)
	return options
}

func CommonOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("walletType", "钱包类型", common.WalletType_WALLET_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("chainCode", "链类型", common.ChainCode_CHAIN_CODE_UNKNOWN.Descriptor()),
		EnumGroup("yesNo", "是否", common.YesNo_YES_NO_UNKNOWN.Descriptor()),
		EnumGroup("visible", "显示状态", common.Switch_SWITCH_UNKNOWN.Descriptor()),
		EnumGroup("tradeSide", "买卖方向", common.Side_SIDE_UNKNOWN.Descriptor()),
		EnumGroup("status", "状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("commonStatus", "通用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("enableStatus", "启用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("assetStatus", "资产状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("bankStatus", "银行卡状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
	}
}

func UserOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("registerType", "注册方式", user.RegisterType_REGISTER_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("loginType", "登录方式", user.LoginType_LOGIN_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("userStatus", "用户状态", user.UserStatus_USER_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("riskLevel", "风控等级", user.RiskLevel_RISK_LEVEL_NORMAL.Descriptor()),
		EnumGroup("gender", "性别", user.Gender_GENDER_UNKNOWN.Descriptor()),
		EnumGroup("idType", "证件类型", user.IdType_ID_TYPE_NONE.Descriptor()),
		EnumGroup("kycLevel", "KYC等级", user.KycLevel_KYC_LEVEL_NONE.Descriptor()),
		EnumGroup("verifyStatus", "实名状态", user.VerifyStatus_VERIFY_STATUS_NONE.Descriptor()),
	}
}

func AssetOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("freezeStatus", "冻结状态", asset.FreezeStatus_FREEZE_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("lockStatus", "锁定状态", asset.LockStatus_LOCK_STATUS_UNKNOWN.Descriptor()),
	}
}

func PaymentOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("platformType", "平台类型", payment.PlatformType_PLATFORM_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("openStatus", "开通状态", payment.OpenStatus_OPEN_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("sceneType", "场景类型", payment.SceneType_SCENE_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("feeType", "费用类型", payment.FeeType_FEE_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("clientType", "客户端类型", payment.ClientType_CLIENT_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("payOrderStatus", "支付订单状态", payment.PayOrderStatus_PAY_ORDER_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("notifyProcessStatus", "回调处理状态", payment.NotifyProcessStatus_NOTIFY_PROCESS_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("signResult", "签名结果", payment.SignResult_SIGN_RESULT_UNKNOWN.Descriptor()),
	}
}

func TradeOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("marketType", "市场类型", trade.MarketType_MARKET_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("contractType", "合约类型", trade.ContractType_CONTRACT_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("symbolStatus", "交易对状态", trade.SymbolStatus_SYMBOL_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("orderType", "订单类型", trade.OrderType_ORDER_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("triggerKind", "触发用途", trade.TriggerKind_TRIGGER_KIND_NONE.Descriptor()),
		EnumGroup("timeInForce", "有效方式", trade.TimeInForce_TIME_IN_FORCE_UNKNOWN.Descriptor()),
		EnumGroup("orderStatus", "订单状态", trade.OrderStatus_ORDER_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("eventStatus", "事件状态", trade.EventStatus_EVENT_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("marginMode", "保证金模式", trade.MarginMode_MARGIN_MODE_UNKNOWN.Descriptor()),
		EnumGroup("positionMode", "持仓模式", trade.PositionMode_POSITION_MODE_UNKNOWN.Descriptor()),
	}
}

func OptionOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("optionType", "期权类型", option.OptionType_OPTION_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("exerciseStyle", "行权方式", option.ExerciseStyle_EXERCISE_STYLE_UNKNOWN.Descriptor()),
		EnumGroup("contractStatus", "合约状态", option.ContractStatus_CONTRACT_STATUS_UNKNOWN.Descriptor()),
	}
}

func StakingOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("productStatus", "产品状态", staking.ProductStatus_PRODUCT_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("productType", "产品类型", staking.ProductType_PRODUCT_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("interestMode", "计息模式", staking.InterestMode_INTEREST_MODE_UNKNOWN.Descriptor()),
		EnumGroup("rewardMode", "奖励模式", staking.RewardMode_REWARD_MODE_UNKNOWN.Descriptor()),
		EnumGroup("stakingOrderStatus", "质押订单状态", staking.OrderStatus_ORDER_STATUS_UNKNOWN.Descriptor()),
	}
}

func ItickOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("categoryType", "产品类型", itick.CategoryType_CATEGORY_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("klineType", "K线周期", itick.KlineType_KLINE_TYPE_UNKNOWN.Descriptor()),
	}
}
