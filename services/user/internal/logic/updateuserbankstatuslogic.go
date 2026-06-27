package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBankStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBankStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankStatusLogic {
	return &UpdateUserBankStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户银行卡状态
func (l *UpdateUserBankStatusLogic) UpdateUserBankStatus(in *user.UpdateUserBankStatusReq) (*user.AdminCommonResp, error) {
	// 获取银行卡信息
	userBank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userBank == nil {
		return &user.AdminCommonResp{
			Base: helper.ErrResp(i18n.BankCardNotFound, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, userBank.TenantId, i18n.NoPermissionOperateThisBankCard); err != nil {
		return nil, err
	} else if base != nil {
		return &user.AdminCommonResp{
			Base: base,
		}, nil
	}

	// 更新银行卡状态
	if in.Enabled != 0 {
		userBank.Enabled = int64(in.Enabled)
	}
	userBank.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.UserBankModel.Update(l.ctx, userBank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新银行卡 %d 状态为 %d", in.Id, in.Enabled)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
