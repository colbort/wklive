package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarginAccountListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarginAccountListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginAccountListAdminLogic {
	return &GetMarginAccountListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取保证金账户列表
func (l *GetMarginAccountListAdminLogic) GetMarginAccountListAdmin(in *trade.GetMarginAccountListAdminReq) (*trade.GetMarginAccountListAdminResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	list, total, err := l.svcCtx.ContractMarginAcctModel.FindPage(l.ctx, models.ContractMarginAccountPageFilter{
		TenantId:    int64(in.TenantId),
		UserId:      int64(in.UserId),
		MarketType:  int64(in.MarketType),
		MarginAsset: in.MarginAsset,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(list) > 0 {
		lastID = int64(list[len(list)-1].Id)
	}
	resp := &trade.GetMarginAccountListAdminResp{Base: pageutil.Base(cursor, limit, len(list), total, lastID)}
	for _, item := range list {
		resp.List = append(resp.List, marginAccountToProto(item))
	}
	return resp, nil
}
