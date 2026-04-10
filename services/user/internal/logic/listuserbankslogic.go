package logic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

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
	// TODO: 实现复杂查询逻辑
	// 需要支持多个过滤条件：keyword, status
	// 使用 FindPage 或自定义查询方法

	bankList := []*user.UserBankListItem{}

	return &user.ListUserBanksResp{
		Base: &common.RespBase{
			Code:  200,
			Msg:   "OK",
			Total: int64(len(bankList)),
		},
		List: bankList,
	}, nil
}