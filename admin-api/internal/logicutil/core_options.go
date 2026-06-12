package logicutil

import (
	"wklive/admin-api/internal/types"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/payment"
	"wklive/proto/staking"
	"wklive/proto/system"
	"wklive/proto/trade"
	"wklive/proto/user"
)

func CoreOptions() []types.OptionsGroup {
	options := make([]types.OptionsGroup, 0)
	options = append(options, CommonOptions()...)
	options = append(options, SystemOptions()...)
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
		EnumGroup("enabled", "启用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("commonStatus", "通用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("enableStatus", "启用状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("assetStatus", "资产状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
		EnumGroup("bankStatus", "银行卡状态", common.Enable_ENABLE_UNKNOWN.Descriptor()),
	}
}

func SystemOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("sysConfigType", "系统配置类型", system.SysConfigType_UNKNOWN.Descriptor()),
		EnumGroup("menuType", "菜单类型", system.MenuType_MENU_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("method", "请求方法", system.RequestMethod_REQUEST_METHOD_UNKNOWN.Descriptor()),
		EnumGroup("jobStatus", "任务状态", system.JobStatus_JOB_STATUS_DISABLED.Descriptor()),
		EnumGroup("verificationCodeChannel", "验证码渠道", system.VerificationCodeChannel_VERIFICATION_CODE_CHANNEL_UNKNOWN.Descriptor()),
		EnumGroup("verificationCodeScene", "验证码业务场景", system.VerificationCodeScene_VERIFICATION_CODE_SCENE_UNKNOWN.Descriptor()),
		EnumGroup("verificationCodeStatus", "验证码发送状态", system.VerificationCodeStatus_VERIFICATION_CODE_STATUS_UNKNOWN.Descriptor()),
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
		EnumGroup("bizType", "业务类型", asset.BizType_BIZ_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("assetSceneType", "业务场景", asset.SceneType_SCENE_TYPE_UNKNOWN.Descriptor()),
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
		EnumGroup("rechargeType", "充值类型", payment.RechargeType_RECHARGE_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("clientType", "客户端类型", payment.ClientType_CLIENT_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("payOrderStatus", "支付订单状态", payment.PayOrderStatus_PAY_ORDER_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("notifyProcessStatus", "回调处理状态", payment.NotifyProcessStatus_NOTIFY_PROCESS_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("signResult", "签名结果", payment.SignResult_SIGN_RESULT_UNKNOWN.Descriptor()),
		EnumGroup("cryptoRechargeAddressSource", "链上充值地址来源", payment.CryptoRechargeAddressSource_CRYPTO_RECHARGE_ADDRESS_SOURCE_UNKNOWN.Descriptor()),
		EnumGroup("cryptoRechargeAddressType", "链上充值地址类型", payment.CryptoRechargeAddressType_CRYPTO_RECHARGE_ADDRESS_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("cryptoRechargeAddressStatus", "链上充值地址状态", payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_UNKNOWN.Descriptor()),
		EnumGroup("cryptoRechargeTxStatus", "链上充值交易状态", payment.CryptoRechargeTxStatus_CRYPTO_RECHARGE_TX_STATUS_UNKNOWN.Descriptor()),
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
		Group(
			"leverageValue",
			"杠杆倍数",
			Option(1, "LEVERAGE_VALUE_1X"),
			Option(2, "LEVERAGE_VALUE_2X"),
			Option(5, "LEVERAGE_VALUE_5X"),
			Option(10, "LEVERAGE_VALUE_10X"),
			Option(20, "LEVERAGE_VALUE_20X"),
			Option(50, "LEVERAGE_VALUE_50X"),
			Option(75, "LEVERAGE_VALUE_75X"),
			Option(100, "LEVERAGE_VALUE_100X"),
			Option(125, "LEVERAGE_VALUE_125X"),
		),
	}
}

func OptionOptions() []types.OptionsGroup {
	return []types.OptionsGroup{
		EnumGroup("optionType", "期权类型", option.OptionType_OPTION_TYPE_UNKNOWN.Descriptor()),
		EnumGroup("exerciseStyle", "行权方式", option.ExerciseStyle_EXERCISE_STYLE_UNKNOWN.Descriptor()),
		EnumGroup("settlementType", "结算方式", option.SettlementType_SETTLEMENT_TYPE_UNKNOWN.Descriptor()),
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
