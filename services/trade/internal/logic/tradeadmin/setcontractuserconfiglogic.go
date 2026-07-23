package tradeadminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetContractUserConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetContractUserConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractUserConfigLogic {
	return &SetContractUserConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户合约偏好配置
func (l *SetContractUserConfigLogic) SetContractUserConfig(in *trade.SetContractUserConfigReq) (*trade.AdminCommonResp, error) {
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.AdminCommonResp{Base: base}, nil
	}
	item, err := l.svcCtx.ContractUserConfigModel.FindOneByTenantIdUserIdSymbolId(l.ctx, in.TenantId, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	now := utils.NowMillis()
	if item == nil {
		item = &models.TContractUserConfig{TenantId: in.TenantId, UserId: in.UserId, SymbolId: in.SymbolId, CreateTimes: now}
	}
	item.PositionMode, item.MarginMode, item.DefaultLeverage = int64(in.PositionMode), int64(in.MarginMode), in.DefaultLeverage
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.ContractUserConfigModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.ContractUserConfigModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
