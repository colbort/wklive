package applogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTradingControlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTradingControlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradingControlLogic {
	return &GetUserTradingControlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询当前用户 kill switch
func (l *GetUserTradingControlLogic) GetUserTradingControl(in *option.GetUserTradingControlReq) (*option.GetUserTradingControlResp, error) {
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.OptionUserTradingControlModel.FindOneByTenantIdUserId(
		l.ctx, tenantID, userID,
	)
	if errors.Is(err, models.ErrNotFound) {
		return &option.GetUserTradingControlResp{
			Base: helper.OkResp(),
			Data: &option.OptionUserTradingControl{
				TenantId: tenantID, UserId: userID, KillSwitch: common.YesNo_YES_NO_NO,
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetUserTradingControlResp{
		Base: helper.OkResp(), Data: optionUserTradingControl(item),
	}, nil
}
