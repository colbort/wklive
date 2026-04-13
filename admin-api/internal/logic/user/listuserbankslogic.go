// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserBanksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserBanksLogic {
	return &ListUserBanksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserBanksLogic) ListUserBanks(req *types.ListUserBanksReq) (resp *types.ListUserBanksResp, err error) {
	result, err := l.svcCtx.UserCli.ListUserBanks(l.ctx, &user.ListUserBanksReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId: req.TenantId,
		UserId:   req.UserId,
		Keyword:  req.Keyword,
		Status:   user.BankStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.UserBankItem, len(result.List))
	for i, item := range result.List {
		data[i] = types.UserBankItem{
			Id:          item.Id,
			TenantId:    item.TenantId,
			UserId:      item.UserId,
			BankName:    item.BankName,
			BankCode:    item.BankCode,
			AccountName: item.AccountName,
			AccountNo:   item.AccountNo,
			BranchName:  item.BranchName,
			CountryCode: item.CountryCode,
			IsDefault:   item.IsDefault,
			Status:      int64(item.Status),
			CreateTimes: item.CreateTimes,
			UpdateTimes: item.UpdateTimes,
		}
	}

	return &types.ListUserBanksResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}
