package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminCreateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateContractLogic {
	return &AdminCreateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建期权合约
func (l *AdminCreateContractLogic) AdminCreateContract(in *option.CreateContractReq) (*option.CreateContractResp, error) {
	// todo: add your logic here and delete this line

	return &option.CreateContractResp{}, nil
}
