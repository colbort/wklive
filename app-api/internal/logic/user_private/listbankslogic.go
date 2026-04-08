// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBanksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBanksLogic {
	return &ListBanksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBanksLogic) ListBanks(req *types.ListBanksReq) (resp *types.ListBanksResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.ListBanks(l.ctx, &user.ListBanksReq{
		UserId: userId,
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.UserBank, len(result.List))
	for i, bank := range result.List {
		data[i] = types.UserBank{
			Id:              bank.Id,
			TenantId:        bank.TenantId,
			UserId:          bank.UserId,
			BankName:        bank.BankName,
			BankCode:        bank.BankCode,
			AccountName:     bank.AccountName,
			AccountNo:       bank.AccountNo,
			MaskedAccountNo: bank.MaskedAccountNo,
			BranchName:      bank.BranchName,
			CountryCode:     bank.CountryCode,
			IsDefault:       bank.IsDefault,
			Status:          int64(bank.Status.Number()),
			CreateTimes:     bank.CreateTimes,
			UpdateTimes:     bank.UpdateTimes,
		}
	}
	return &types.ListBanksResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}, nil
}
