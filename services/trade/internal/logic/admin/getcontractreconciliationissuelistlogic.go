package adminlogic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetContractReconciliationIssueListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContractReconciliationIssueListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractReconciliationIssueListLogic {
	return &GetContractReconciliationIssueListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询及人工忽略跨服务对账差异；恢复态只能由自动对账证据关闭
func (l *GetContractReconciliationIssueListLogic) GetContractReconciliationIssueList(in *trade.GetContractReconciliationIssueListReq) (*trade.GetContractReconciliationIssueListResp, error) {
	cursor, limit := pageutil.Input(in.GetPage())
	rows, total, err := l.svcCtx.ContractReconcileIssueModel.FindPage(l.ctx, models.ContractReconciliationIssuePageFilter{
		TenantId:  helpers.AdminTenantID(l.ctx, in.GetTenantId()),
		Status:    int64(in.GetStatus()),
		CheckType: in.GetCheckType(),
		BizNo:     in.GetBizNo(),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	resp := &trade.GetContractReconciliationIssueListResp{}
	lastID := int64(0)
	for _, row := range rows {
		resp.Data = append(resp.Data, mapper.ContractReconciliationIssueProto(row))
		lastID = row.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(rows), total, lastID)
	return resp, nil
}
