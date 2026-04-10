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

type AdminListAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListAccountsLogic {
	return &AdminListAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询账户资产列表
func (l *AdminListAccountsLogic) AdminListAccounts(in *option.ListAccountsReq) (*option.ListAccountsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionAccountModel.FindPage(l.ctx, models.OptionAccountPageFilter{
		TenantId:  in.TenantId,
		Uid:       in.Uid,
		AccountId: in.AccountId,
		Status:    int64(in.Status),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionAccount, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		list = append(list, toAccountProto(item))
	}

	return &option.ListAccountsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
