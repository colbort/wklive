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

type AddBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBankLogic {
	return &AddBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddBankLogic) AddBank(req *types.AddBankReq) (resp *types.AddBankResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.AddBank(l.ctx, &user.AddBankReq{
		UserId:      userId,
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

	return &types.AddBankResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Bank: types.UserBank{},
	}, nil
}
