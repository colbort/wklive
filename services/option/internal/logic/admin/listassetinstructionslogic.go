package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAssetInstructionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAssetInstructionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAssetInstructionsLogic {
	return &ListAssetInstructionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资产指令，统一定位失败和人工处理项
func (l *ListAssetInstructionsLogic) ListAssetInstructions(in *option.ListAssetInstructionsReq) (*option.ListAssetInstructionsResp, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListAssetInstructionsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionAssetInstructionModel.FindPage(
		l.ctx, models.OptionAssetInstructionPageFilter{
			TenantId: in.TenantId, UserId: in.UserId, BizNo: in.BizNo,
			Status: int64(in.Status), ReconciliationStatus: int64(in.ReconciliationStatus),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionAssetInstruction, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToAssetInstructionProto(item))
	}
	return &option.ListAssetInstructionsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}
