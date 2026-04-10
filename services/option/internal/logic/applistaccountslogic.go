package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
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
	items, _, err := l.svcCtx.OptionAccountModel.FindPage(l.ctx, models.OptionAccountPageFilter{
		TenantId:  in.TenantId,
		Uid:       in.Uid,
		AccountId: in.AccountId,
	}, 0, 100)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	list := make([]*option.OptionAccount, 0, len(items))
	for _, item := range items {
		list = append(list, toAccountProto(item))
	}

	return &option.AppListAccountsResp{Base: helper.OkResp(), List: list}, nil
}
