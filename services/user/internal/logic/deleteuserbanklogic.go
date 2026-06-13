package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserBankLogic {
	return &DeleteUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用户银行卡
func (l *DeleteUserBankLogic) DeleteUserBank(in *user.DeleteUserBankReq) (*user.AdminCommonResp, error) {
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if bank == nil {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(i18n.BankCardNotFound, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, bank.TenantId, i18n.NoPermissionDeleteThisBankCard); err != nil {
		return nil, err
	} else if base != nil {
		return &user.AdminCommonResp{
			Base: base,
		}, nil
	}

	err = l.svcCtx.UserBankModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
