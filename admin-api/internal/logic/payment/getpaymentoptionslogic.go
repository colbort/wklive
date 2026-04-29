package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaymentOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentOptionsLogic {
	return &GetPaymentOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaymentOptionsLogic) GetPaymentOptions() (resp *types.GetPaymentOptionsResp, err error) {
	return &types.GetPaymentOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("platformType", "平台类型", payment.PlatformType_PLATFORM_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("status", "状态", payment.CommonStatus_COMMON_STATUS_UNKNOWN.Descriptor()),
			logicutil.Group("visible", "显示状态",
				logicutil.Option(0, "YES_NO_NO"),
				logicutil.Option(1, "YES_NO_YES"),
			),
			logicutil.Group("yesNo", "是否",
				logicutil.Option(0, "YES_NO_NO"),
				logicutil.Option(1, "YES_NO_YES"),
			),
			logicutil.EnumGroup("openStatus", "开通状态", payment.OpenStatus_OPEN_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("sceneType", "场景类型", payment.SceneType_SCENE_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("feeType", "费用类型", payment.FeeType_FEE_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("clientType", "客户端类型", payment.ClientType_CLIENT_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("payOrderStatus", "支付订单状态", payment.PayOrderStatus_PAY_ORDER_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("notifyProcessStatus", "回调处理状态", payment.NotifyProcessStatus_NOTIFY_PROCESS_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("signResult", "签名结果", payment.SignResult_SIGN_RESULT_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("cryptoRechargeAddressSource", "链上充值地址来源", payment.CryptoRechargeAddressSource_CRYPTO_RECHARGE_ADDRESS_SOURCE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("cryptoRechargeAddressType", "链上充值地址类型", payment.CryptoRechargeAddressType_CRYPTO_RECHARGE_ADDRESS_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("cryptoRechargeAddressStatus", "链上充值地址状态", payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("cryptoRechargeTxStatus", "链上充值交易状态", payment.CryptoRechargeTxStatus_CRYPTO_RECHARGE_TX_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("chainCode", "链类型", common.ChainCode_CHAIN_CODE_UNKNOWN.Descriptor()),
		},
	}, nil
}
