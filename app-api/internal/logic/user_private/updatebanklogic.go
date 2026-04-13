// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBankLogic {
	return &UpdateBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateBankLogic) UpdateBank(req *types.UpdateBankReq) (resp *types.UpdateBankResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.UpdateBank(l.ctx, &user.UpdateBankReq{
		UserId:      userId,
		Id:          req.Id,
		BankName:    req.BankName,
		BankCode:    req.BankCode,
		AccountName: req.AccountName,
		AccountNo:   req.AccountNo,
		BranchName:  req.BranchName,
		CountryCode: req.CountryCode,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateBankResp{
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
			CreateTimes: result.Bank.CreateTimes,
			UpdateTimes: result.Bank.UpdateTimes,
		},
	}, nil
}
