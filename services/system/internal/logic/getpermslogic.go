package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermsLogic {
	return &GetPermsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPermsLogic) GetPerms(in *system.Empty) (*system.PermsResp, error) {
	// todo: add your logic here and delete this line

	return &system.PermsResp{}, nil
}
