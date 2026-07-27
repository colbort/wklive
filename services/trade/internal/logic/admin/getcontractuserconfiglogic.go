package adminlogic

import (
	"context"
	"errors"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetContractUserConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContractUserConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractUserConfigLogic {
	return &GetContractUserConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户合约偏好配置
func (l *GetContractUserConfigLogic) GetContractUserConfig(in *trade.GetContractUserConfigReq) (*trade.GetContractUserConfigResp, error) {
	item, err := l.svcCtx.ContractUserConfigModel.FindOneByTenantIdUserIdSymbolId(l.ctx, in.TenantId, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	return &trade.GetContractUserConfigResp{Base: helper.OkResp(), Data: helpers.ContractUserConfigToProto(item)}, nil
}
