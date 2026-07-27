package adminlogic

import (
	"context"
	"errors"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserLeverageConfigLogic {
	return &SetUserLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户杠杆配置
func (l *SetUserLeverageConfigLogic) SetUserLeverageConfig(in *trade.SetUserLeverageConfigReq) (*trade.CommonResp, error) {
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
	}

	now := utils.NowMillis()
	item, err := l.svcCtx.ContractLeverageCfgModel.FindOneByTenantIdUserIdSymbolIdMarginMode(l.ctx, in.TenantId, in.UserId, in.SymbolId, int64(in.MarginMode))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TContractLeverageConfig{
			TenantId:    in.TenantId,
			UserId:      in.UserId,
			SymbolId:    in.SymbolId,
			MarginMode:  int64(in.MarginMode),
			Enabled:     int64(common.Enable_ENABLE_ENABLED),
			CreateTimes: now,
		}
	}
	item.LongLeverage = in.LongLeverage
	item.ShortLeverage = in.ShortLeverage
	item.OperatorId = in.OperatorId
	item.Source = int64(in.Source)
	item.Enabled = helpers.EnableToModel(in.Enabled, item.Enabled)
	item.Remark = in.Remark
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.ContractLeverageCfgModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.ContractLeverageCfgModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
