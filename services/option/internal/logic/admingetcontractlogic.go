package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
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
	item, err := findContractByCodeOrID(l.ctx, l.svcCtx, in.TenantId, in.Id, in.ContractCode)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetContractResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	data, err := buildContractDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetContractResp{Base: helper.OkResp(), Data: data}, nil
}
