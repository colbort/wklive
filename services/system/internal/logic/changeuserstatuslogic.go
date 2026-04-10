package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeUserStatusLogic {
	return &ChangeUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangeUserStatusLogic) ChangeUserStatus(in *system.ChangeUserStatusReq) (*system.RespBase, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 1,
				Msg:  "用户不存在",
			},
		}, nil
	}

	if user.Status == in.Status {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 1,
				Msg:  "用户状态未改变",
			},
		}, nil
	}
	user.Status = in.Status
	if err = l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
