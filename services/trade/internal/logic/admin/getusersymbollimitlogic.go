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

type GetUserSymbolLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSymbolLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSymbolLimitLogic {
	return &GetUserSymbolLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户交易对限制
func (l *GetUserSymbolLimitLogic) GetUserSymbolLimit(in *trade.GetUserSymbolLimitReq) (*trade.GetUserSymbolLimitResp, error) {
	tenantID, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed || tenantID <= 0 {
		return &trade.GetUserSymbolLimitResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	item, err := l.svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolId(l.ctx, tenantID, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetUserSymbolLimitResp{Base: helper.OkResp()}
	if item != nil {
		resp.Data = helpers.RiskUserSymbolLimitToProto(item)
	}
	return resp, nil
}
