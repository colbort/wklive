package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"
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
			Base: helper.GetErrResp(404, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}
	if userBank.TenantId != in.TenantId {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionOperateThisBankCard, l.ctx)),
		}, nil
	}

	// 更新银行卡状态
	if in.Status != 0 {
		userBank.Status = int64(in.Status)
	}
	userBank.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserBankModel.Update(l.ctx, userBank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新银行卡 %d 状态为 %d", in.Id, in.Status)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
