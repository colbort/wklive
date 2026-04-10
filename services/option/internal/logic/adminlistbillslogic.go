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

type AdminListBillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListBillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListBillsLogic {
	return &AdminListBillsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资金流水列表
func (l *AdminListBillsLogic) AdminListBills(in *option.ListBillsReq) (*option.ListBillsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionBillModel.FindPage(l.ctx, models.OptionBillPageFilter{
		TenantId:        in.TenantId,
		Uid:             in.Uid,
		AccountId:       in.AccountId,
		BizNo:           in.BizNo,
		RefType:         int64(in.RefType),
		CreateTimeStart: pageutil.TimeRangeStart(in.CreateTimeRange),
		CreateTimeEnd:   pageutil.TimeRangeEnd(in.CreateTimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionBill, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		list = append(list, toBillProto(item))
	}

	return &option.ListBillsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
