package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListContractSeriesDetailsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListContractSeriesDetailsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContractSeriesDetailsLogic {
	return &ListContractSeriesDetailsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询系列生成合约谱系
func (l *ListContractSeriesDetailsLogic) ListContractSeriesDetails(in *option.ListContractSeriesDetailsReq) (*option.ListContractSeriesDetailsResp, error) {
	if in == nil || in.SeriesId <= 0 {
		return &option.ListContractSeriesDetailsResp{Base: helper.ErrResp(
			i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx),
		)}, nil
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListContractSeriesDetailsResp{Base: helper.ErrResp(
			i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx),
		)}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionContractSeriesDetailModel.FindPageBySeries(
		l.ctx, in.TenantId, in.SeriesId, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionContractSeriesDetail, 0, len(items))
	for _, item := range items {
		data = append(data, toContractSeriesDetailProto(item))
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	return &option.ListContractSeriesDetailsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: data, Total: total,
	}, nil
}
