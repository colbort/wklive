package logic

import (
	"context"
	"errors"

	pageutil "wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListSettlementsLogic {
	return &AdminListSettlementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询到期结算记录列表
func (l *AdminListSettlementsLogic) AdminListSettlements(in *option.ListSettlementsReq) (*option.ListSettlementsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionSettlementModel.FindPage(l.ctx, models.OptionSettlementPageFilter{
		TenantId:            in.TenantId,
		ContractId:          in.ContractId,
		Status:              int64(in.Status),
		SettlementTimeStart: pageutil.TimeRangeStart(in.SettlementTimeRange),
		SettlementTimeEnd:   pageutil.TimeRangeEnd(in.SettlementTimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionSettlementDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := buildSettlementDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		list = append(list, detail)
	}

	return &option.ListSettlementsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
