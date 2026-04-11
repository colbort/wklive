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

type AdminGetTradeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetTradeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetTradeLogic {
	return &AdminGetTradeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个成交记录详情
func (l *AdminGetTradeLogic) AdminGetTrade(in *option.GetTradeReq) (*option.GetTradeResp, error) {
	item, err := findTradeByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.Id, in.TradeNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetTradeResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.TradeNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	data, err := buildTradeDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetTradeResp{Base: helper.OkResp(), Data: data}, nil
}
