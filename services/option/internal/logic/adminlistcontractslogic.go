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

type AdminListContractsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListContractsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListContractsLogic {
	return &AdminListContractsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询期权合约列表
func (l *AdminListContractsLogic) AdminListContracts(in *option.ListContractsReq) (*option.ListContractsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
		TenantId:         in.TenantId,
		ContractCode:     in.ContractCode,
		UnderlyingSymbol: in.UnderlyingSymbol,
		OptionType:       int64(in.OptionType),
		Status:           int64(in.Status),
		ListTimeStart:    pageutil.TimeRangeStart(in.ListTimeRange),
		ListTimeEnd:      pageutil.TimeRangeEnd(in.ListTimeRange),
		ExpireTimeStart:  pageutil.TimeRangeStart(in.ExpireTimeRange),
		ExpireTimeEnd:    pageutil.TimeRangeEnd(in.ExpireTimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionContractDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := buildContractDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		list = append(list, detail)
	}

	return &option.ListContractsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
