package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListPositionsLogic {
	return &AdminListPositionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询持仓列表
func (l *AdminListPositionsLogic) AdminListPositions(in *option.ListPositionsReq) (*option.ListPositionsResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListPositionsResp{}, nil
}
