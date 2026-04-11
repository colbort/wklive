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
	// 获取银行卡信息
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if bank == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}

	// 验证银行卡是否属于该用户
	if bank.UserId != in.UserId {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.PermissionDeniedForBankCard, l.ctx)),
		}, nil
	}

	now := time.Now().UnixMilli()

	// 将其他所有卡设置为非默认
	// TODO: 自定义查询方法，找出该用户的所有银行卡并更新

	// 将当前卡设置为默认
	bank.IsDefault = 1
	bank.UpdateTimes = now
	err = l.svcCtx.UserBankModel.Update(l.ctx, bank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 设置银行卡 %d 为默认卡成功", in.UserId, in.Id)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
