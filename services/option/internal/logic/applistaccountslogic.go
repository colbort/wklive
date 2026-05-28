package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListAccountsLogic {
	return &AppListAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取账户资产列表
func (l *AppListAccountsLogic) AppListAccounts(in *option.AppListAccountsReq) (*option.AppListAccountsResp, error) {
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
		data = append(data, toAccountProto(item))
	}

	return &option.AppListAccountsResp{Base: helper.OkResp(), Data: data}, nil
}
