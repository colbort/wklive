package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	pbpayment "wklive/proto/payment"

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
			logicutil.EnumGroup("platformType", "平台类型", pbpayment.PlatformType_PLATFORM_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("status", "状态", pbpayment.CommonStatus_COMMON_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("openStatus", "开通状态", pbpayment.OpenStatus_OPEN_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("sceneType", "场景类型", pbpayment.SceneType_SCENE_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("feeType", "费用类型", pbpayment.FeeType_FEE_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("clientType", "客户端类型", pbpayment.ClientType_CLIENT_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("payOrderStatus", "支付订单状态", pbpayment.PayOrderStatus_PAY_ORDER_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("notifyProcessStatus", "回调处理状态", pbpayment.NotifyProcessStatus_NOTIFY_PROCESS_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("signResult", "签名结果", pbpayment.SignResult_SIGN_RESULT_UNKNOWN.Descriptor()),
		},
	}, nil
}
