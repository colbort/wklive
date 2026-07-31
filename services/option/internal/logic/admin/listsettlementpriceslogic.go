package adminlogic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSettlementPricesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSettlementPricesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettlementPricesLogic {
	return &ListSettlementPricesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询到期结算价版本
func (l *ListSettlementPricesLogic) ListSettlementPrices(in *option.ListSettlementPricesReq) (*option.ListSettlementPricesResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionSettlementPriceModel.FindPage(
		l.ctx,
		models.OptionSettlementPricePageFilter{
			TenantId: in.TenantId, ContractId: in.ContractId, Status: int64(in.Status),
		},
		cursor, limit,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	data := make([]*option.OptionSettlementPrice, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToSettlementPriceProto(item))
	}
	return &option.ListSettlementPricesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
