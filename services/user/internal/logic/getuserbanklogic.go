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

type GetUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserBankLogic {
	return &GetUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员获取用户银行卡详情
func (l *GetUserBankLogic) GetUserBank(in *user.GetUserBankReq) (*user.GetUserBankResp, error) {
	// 获取银行卡信息
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if bank == nil {
		return &user.GetUserBankResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}

	return &user.GetUserBankResp{
		Base: helper.OkResp(),
		Bank: toUserBankItemProto(bank),
	}, nil
}
