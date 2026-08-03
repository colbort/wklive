package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationListLogic {
	return &OperationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取资金操作列表
func (l *OperationListLogic) OperationList(in *staking.OperationListReq) (*staking.OperationListResp, error) {
	tenantID, base, err := helpers.AdminTenantReadScopeResp(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	if base != nil {
		return &staking.OperationListResp{Page: base}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	items, total, err := l.svcCtx.StakeOperationModel.FindAdminPage(l.ctx, models.StakeOperationPageFilter{
		TenantId: tenantID, UserId: in.GetUserId(), OperationType: in.GetOperationType(), Status: in.GetStatus(),
		OperationNo: strings.TrimSpace(in.GetOperationNo()), OrderNo: strings.TrimSpace(in.GetOrderNo()),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	resp := &staking.OperationListResp{Page: helper.OkResp(), Data: make([]*staking.StakeOperation, 0, len(items))}
	lastID := int64(0)
	for _, item := range items {
		resp.Data = append(resp.Data, helpers.OperationToProto(item))
		lastID = item.Id
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, lastID)
	return resp, nil
}
