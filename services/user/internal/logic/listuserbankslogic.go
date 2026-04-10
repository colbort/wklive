package logic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserBanksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserBanksLogic {
	return &ListUserBanksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员查询用户银行卡列表
func (l *ListUserBanksLogic) ListUserBanks(in *user.ListUserBanksReq) (*user.ListUserBanksResp, error) {
	banks, total, err := l.svcCtx.UserBankModel.FindPage(l.ctx, in.TenantId, in.UserId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(banks)) == in.Page.Limit {
		nextCursor = banks[len(banks)-1].Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(banks)) == in.Page.Limit

	bankList := make([]*user.UserBankListItem, 0, len(banks))
	for _, bank := range banks {
		if in.Status != 0 && bank.Status != int64(in.Status) {
			continue
		}
		u, err := l.svcCtx.UserModel.FindOne(l.ctx, bank.UserId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if u == nil {
			continue
		}
		if in.Keyword != "" {
			keyword := strings.ToLower(in.Keyword)
			if !strings.Contains(strings.ToLower(bank.BankName), keyword) &&
				!strings.Contains(strings.ToLower(bank.AccountName), keyword) &&
				!strings.Contains(strings.ToLower(bank.AccountNo), keyword) &&
				!strings.Contains(strings.ToLower(u.Username), keyword) &&
				!strings.Contains(strings.ToLower(u.UserNo), keyword) {
				continue
			}
		}

		bankList = append(bankList, &user.UserBankListItem{
			Id:              bank.Id,
			UserId:          bank.UserId,
			UserNo:          u.UserNo,
			Username:        u.Username,
			BankName:        bank.BankName,
			BankCode:        bank.BankCode.String,
			AccountName:     bank.AccountName,
			AccountNo:       bank.AccountNo,
			MaskedAccountNo: maskAccountNo(bank.AccountNo),
			BranchName:      bank.BranchName.String,
			CountryCode:     bank.CountryCode.String,
			IsDefault:       bank.IsDefault == 1,
			Status:          user.BankStatus(bank.Status),
			CreateTimes:     bank.CreateTimes,
		})
	}

	return &user.ListUserBanksResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		List: bankList,
	}, nil
}
