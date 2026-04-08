package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUpdateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateContractLogic {
	return &AdminUpdateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权合约
func (l *AdminUpdateContractLogic) AdminUpdateContract(in *option.UpdateContractReq) (*option.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &option.AdminCommonResp{}, nil
}
