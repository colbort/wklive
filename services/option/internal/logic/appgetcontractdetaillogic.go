package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppGetContractDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppGetContractDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetContractDetailLogic {
	return &AppGetContractDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取期权合约详情
func (l *AppGetContractDetailLogic) AppGetContractDetail(in *option.AppGetContractDetailReq) (*option.AppGetContractDetailResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppGetContractDetailResp{}, nil
}
