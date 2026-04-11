package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginAccountListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarginAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginAccountListLogic {
	return &GetMarginAccountListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取保证金账户列表
func (l *GetMarginAccountListLogic) GetMarginAccountList(in *trade.GetMarginAccountListReq) (*trade.GetMarginAccountListResp, error) {
	list, err := l.svcCtx.ContractMarginAcctModel.FindList(l.ctx, models.ContractMarginAccountPageFilter{
		TenantId:    int64(in.TenantId),
		UserId:      int64(in.UserId),
		MarketType:  int64(in.MarketType),
		MarginAsset: in.MarginAsset,
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	resp := &trade.GetMarginAccountListResp{Base: helper.OkResp()}
	for _, item := range list {
		resp.List = append(resp.List, marginAccountToProto(item))
	}
	return resp, nil
}
