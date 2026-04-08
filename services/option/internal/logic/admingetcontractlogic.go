package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetContractLogic {
	return &AdminGetContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个期权合约详情
func (l *AdminGetContractLogic) AdminGetContract(in *option.GetContractReq) (*option.GetContractResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetContractResp{}, nil
}
