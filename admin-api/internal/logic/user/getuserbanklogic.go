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

type GetUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserBankLogic {
	return &GetUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserBankLogic) GetUserBank(req *types.GetUserBankReq) (resp *types.GetUserBankResp, err error) {
	result, err := l.svcCtx.UserCli.GetUserBank(l.ctx, &user.GetUserBankReq{
		TenantId: req.TenantId,
		Id:       req.Id,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserBankResp{
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
			CreateTime:  result.Bank.CreateTime,
			UpdateTime:  result.Bank.UpdateTime,
		},
	}, nil
}
