package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
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
			return &option.GetTradeResp{Base: helper.GetErrResp(404, "成交不存在")}, nil
		}
		return nil, err
	}
	data, err := buildTradeDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetTradeResp{Base: helper.OkResp(), Data: data}, nil
}
