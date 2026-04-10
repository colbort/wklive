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

type AppListContractsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListContractsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListContractsLogic {
	return &AppListContractsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取可交易期权合约列表
func (l *AppListContractsLogic) AppListContracts(in *option.AppListContractsReq) (*option.AppListContractsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	status := int64(in.Status)
	if status == 0 {
		status = int64(option.ContractStatus_CONTRACT_STATUS_TRADING)
	}
	items, total, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
		TenantId:         in.TenantId,
		UnderlyingSymbol: in.UnderlyingSymbol,
		OptionType:       int64(in.OptionType),
		Status:           status,
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

	return &option.AppListContractsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
	}, nil
}
