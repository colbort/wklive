package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBanksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBanksLogic {
	return &ListBanksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户银行卡列表
func (l *ListBanksLogic) ListBanks(in *user.ListBanksReq) (*user.ListBanksResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.ListBanksResp{
			Base: helper.GetErrResp(404, "用户不存在"),
		}, nil
	}
	items, total, err := l.svcCtx.UserBankModel.FindPage(l.ctx, tuser.TenantId, tuser.Id, in.Page.Cursor, in.Page.Limit)
	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	data := make([]*user.UserBank, 0)
	for _, item := range items {
		data = append(data, &user.UserBank{
			Id:              item.Id,
			TenantId:        item.TenantId,
			UserId:          item.UserId,
			BankName:        item.BankName,
			BankCode:        item.BankCode.String,
			AccountName:     item.AccountName,
			AccountNo:       item.AccountNo,
			MaskedAccountNo: "",
			BranchName:      item.BranchName.String,
			CountryCode:     item.CountryCode.String,
			IsDefault:       item.IsDefault == 1,
			Status:          user.BankStatus(item.Status),
			CreateTimes:     item.CreateTimes,
			UpdateTimes:     item.UpdateTimes,
		})
	}

	return &user.ListBanksResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		List: data,
	}, nil
}
