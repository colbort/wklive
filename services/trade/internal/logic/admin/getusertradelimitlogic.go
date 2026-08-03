package adminlogic

import (
	"context"
	"errors"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTradeLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTradeLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradeLimitLogic {
	return &GetUserTradeLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户交易限制
func (l *GetUserTradeLimitLogic) GetUserTradeLimit(in *trade.GetUserTradeLimitReq) (*trade.GetUserTradeLimitResp, error) {
	tenantID, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed || tenantID <= 0 {
		return &trade.GetUserTradeLimitResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	item, err := l.svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdProductTypeContractType(l.ctx, tenantID, in.UserId, int64(in.ProductType), int64(in.ContractType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetUserTradeLimitResp{Base: helper.OkResp()}
	if item != nil {
		resp.Data = helpers.RiskUserTradeLimitToProto(item)
	}
	return resp, nil
}
