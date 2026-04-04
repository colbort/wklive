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

type AddUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserBankLogic {
	return &AddUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddUserBankLogic) AddUserBank(req *types.AddUserBankReq) (resp *types.AddUserBankResp, err error) {
	result, err := l.svcCtx.UserCli.AddUserBank(l.ctx, &user.AddUserBankReq{
		TenantId:    req.TenantId,
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

	return &types.AddUserBankResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Bank: types.UserBank{
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
			CreateTimes:  result.Bank.CreateTimes,
			UpdateTimes:  result.Bank.UpdateTimes,
		},
	}, nil
}
