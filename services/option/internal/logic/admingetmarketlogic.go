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
			return &option.GetMarketResp{Base: helper.GetErrResp(404, "行情不存在")}, nil
		}
		return nil, err
	}

	return &option.GetMarketResp{Base: helper.OkResp(), Data: toMarketProto(item)}, nil
}
