package staking

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/staking"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStakingOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetStakingOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStakingOptionsLogic {
	return &GetStakingOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetStakingOptionsLogic) GetStakingOptions() (resp *types.GetStakingOptionsResp, err error) {
	return &types.GetStakingOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("yesNo", "是否", staking.YesNo_YES_NO_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("productStatus", "产品状态", staking.ProductStatus_PRODUCT_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("productType", "产品类型", staking.ProductType_PRODUCT_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("interestMode", "计息模式", staking.InterestMode_INTEREST_MODE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("rewardMode", "奖励模式", staking.RewardMode_REWARD_MODE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("orderStatus", "订单状态", staking.OrderStatus_ORDER_STATUS_UNKNOWN.Descriptor()),
		},
	}, nil
}
