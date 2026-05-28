package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
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
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.ContractMarginAcctModel.FindPage(l.ctx, models.ContractMarginAccountPageFilter{
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		MarketType:  int64(in.MarketType),
		MarginAsset: in.MarginAsset,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetMarginAccountListAdminResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, marginAccountToProto(item))
	}
	return resp, nil
}
