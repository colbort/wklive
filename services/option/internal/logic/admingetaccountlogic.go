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

type AdminGetAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetAccountLogic {
	return &AdminGetAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个账户资产详情
func (l *AdminGetAccountLogic) AdminGetAccount(in *option.GetAccountReq) (*option.GetAccountResp, error) {
	items, _, err := l.svcCtx.OptionAccountModel.FindPage(l.ctx, models.OptionAccountPageFilter{
		TenantId:   in.TenantId,
		Uid:        in.Uid,
		AccountId:  in.AccountId,
		MarginCoin: in.MarginCoin,
	}, 0, 1)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if len(items) == 0 {
		return &option.GetAccountResp{Base: helper.GetErrResp(404, "账户资产不存在")}, nil
	}

	return &option.GetAccountResp{Base: helper.OkResp(), Data: toAccountProto(items[0])}, nil
}
