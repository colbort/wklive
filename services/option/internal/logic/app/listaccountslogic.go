package applogic

import (
	"context"
	"errors"
	"wklive/services/option/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAccountsLogic {
	return &ListAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取账户资产列表
func (l *ListAccountsLogic) ListAccounts(in *option.UserListAccountsReq) (*option.UserListAccountsResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	items, _, err := l.svcCtx.OptionAccountModel.FindPage(l.ctx, models.OptionAccountPageFilter{
		TenantId:  tenantId,
		UserId:    userId,
		AccountId: in.AccountId,
	}, 0, 100)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	data := make([]*option.OptionAccount, 0, len(items))
	for _, item := range items {
		data = append(data, helpers.ToAccountProto(item))
	}

	return &option.UserListAccountsResp{Base: helper.OkResp(), Data: data}, nil
}
