package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

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
	item, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppGetContractDetailResp{Base: helper.GetErrResp(404, "合约不存在")}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && item.TenantId != in.TenantId {
		return &option.AppGetContractDetailResp{Base: helper.GetErrResp(404, "合约不存在")}, nil
	}
	data, err := buildContractDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.AppGetContractDetailResp{Base: helper.OkResp(), Data: data}, nil
}
