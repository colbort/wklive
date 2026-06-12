package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultBankLogic {
	return &SetDefaultBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置默认银行卡
func (l *SetDefaultBankLogic) SetDefaultBank(in *user.SetDefaultBankReq) (*user.AppCommonResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	// 获取银行卡信息
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if bank == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(i18n.BankCardNotFound, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}

	// 验证银行卡是否属于该用户
	if bank.UserId != userId {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(i18n.PermissionDeniedForBankCard, i18n.Translate(i18n.PermissionDeniedForBankCard, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()

	// 将其他所有卡设置为非默认
	// TODO: 自定义查询方法，找出该用户的所有银行卡并更新

	// 将当前卡设置为默认
	bank.IsDefault = int64(common.YesNo_YES_NO_YES)
	bank.UpdateTimes = now
	err = l.svcCtx.UserBankModel.Update(l.ctx, bank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 设置银行卡 %d 为默认卡成功", userId, in.Id)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
