package logic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/pageutil"
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
	items, total, err := l.svcCtx.UserBankModel.FindPage(l.ctx, in.TenantId, in.UserId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*user.UserBankItem, 0)
	for _, bank := range items {
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

		data = append(data, toUserBankItemProto(bank))
	}

	return &user.ListUserBanksResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		List: data,
	}, nil
}
