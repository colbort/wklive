// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankLogic {
	return &UpdateUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBankLogic) UpdateUserBank(req *types.UpdateUserBankReq) (resp *types.UpdateUserBankResp, err error) {
	result, err := l.svcCtx.UserCli.UpdateUserBank(l.ctx, &user.UpdateUserBankReq{
		TenantId:    req.TenantId,
		Id:          req.Id,
		UserId:      req.UserId,
		BankName:    req.BankName,
		BankCode:    req.BankCode,
		AccountName: req.AccountName,
		AccountNo:   req.AccountNo,
		BranchName:  req.BranchName,
		CountryCode: req.CountryCode,
		IsDefault:   req.IsDefault,
		Status:      user.BankStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateUserBankResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Bank: types.UserBankItem{
			Id:          result.Bank.Id,
			TenantId:    result.Bank.TenantId,
			UserId:      result.Bank.UserId,
			BankName:    result.Bank.BankName,
			BankCode:    result.Bank.BankCode,
			AccountName: result.Bank.AccountName,
			AccountNo:   result.Bank.AccountNo,
			BranchName:  result.Bank.BranchName,
			CountryCode: result.Bank.CountryCode,
			IsDefault:   result.Bank.IsDefault,
			Status:      int64(result.Bank.Status),
			CreateTimes: result.Bank.CreateTimes,
			UpdateTimes: result.Bank.UpdateTimes,
		},
	}, nil
}
