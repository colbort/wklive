package adminlogic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRiskAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRiskAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRiskAccountsLogic {
	return &ListRiskAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询卖方风险账户
func (l *ListRiskAccountsLogic) ListRiskAccounts(in *option.ListRiskAccountsReq) (*option.ListRiskAccountsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionRiskAccountModel.FindPage(l.ctx, models.OptionRiskAccountPageFilter{
		TenantId: in.TenantId, UserId: in.UserId, AccountId: in.AccountId,
		SettleCoin: in.SettleCoin, Status: int64(in.Status),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionRiskAccount, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToRiskAccountProto(item))
	}
	return &option.ListRiskAccountsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID), Data: data,
	}, nil
}
