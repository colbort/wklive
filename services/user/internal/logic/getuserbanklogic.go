package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/common"
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
			Base: &common.RespBase{
				Code: 404,
				Msg:  "银行卡不存在",
			},
		}, nil
	}

	bankProto := &user.UserBank{
		Id:          bank.Id,
		TenantId:    bank.TenantId,
		UserId:      bank.UserId,
		BankName:    bank.BankName,
		BankCode:    bank.BankCode.String,
		AccountName: bank.AccountName,
		AccountNo:   bank.AccountNo,
		BranchName:  bank.BranchName.String,
		CountryCode: bank.CountryCode.String,
		IsDefault:   bank.IsDefault == 1,
		Status:      user.BankStatus(bank.Status),
		CreateTimes: bank.CreateTimes,
		UpdateTimes: bank.UpdateTimes,
	}

	return &user.GetUserBankResp{
		Base: helper.OkResp(),
		Bank: bankProto,
	}, nil
}