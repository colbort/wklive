package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListContractSeriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListContractSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContractSeriesLogic {
	return &ListContractSeriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询系列版本
func (l *ListContractSeriesLogic) ListContractSeries(in *option.ListContractSeriesReq) (*option.ListContractSeriesResp, error) {
	if in == nil {
		return &option.ListContractSeriesResp{Base: helper.ErrResp(
			i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx),
		)}, nil
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListContractSeriesResp{Base: helper.ErrResp(
			i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx),
		)}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionContractSeriesModel.FindPage(
		l.ctx, models.OptionContractSeriesPageFilter{
			TenantId: in.TenantId, SeriesCode: in.SeriesCode, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionContractSeries, 0, len(items))
	for _, item := range items {
		expiries, findErr := l.svcCtx.OptionContractSeriesExpiryModel.FindBySeries(l.ctx, item.TenantId, item.Id)
		if findErr != nil {
			return nil, findErr
		}
		bands, findErr := l.svcCtx.OptionContractSeriesStrikeBandModel.FindBySeries(l.ctx, item.TenantId, item.Id)
		if findErr != nil {
			return nil, findErr
		}
		data = append(data, toContractSeriesProto(item, expiries, bands))
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	return &option.ListContractSeriesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: data, Total: total,
	}, nil
}
