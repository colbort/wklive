package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

type AdminGetMarketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetMarketLogic {
	return &AdminGetMarketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个期权当前行情
func (l *AdminGetMarketLogic) AdminGetMarket(in *option.GetMarketReq) (*option.GetMarketResp, error) {
	item, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, in.TenantId, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetMarketResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.MarketNotFound, l.ctx))}, nil
		}
		return nil, err
	}

	return &option.GetMarketResp{Base: helper.OkResp(), Data: toMarketProto(item)}, nil
}
