package applogic

import (
	"context"
	"errors"
	helpers "wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListLogic {
	return &GetPositionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取持仓列表
func (l *GetPositionListLogic) GetPositionList(in *trade.GetPositionListReq) (*trade.GetPositionListResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	data, err := l.svcCtx.ContractPositionModel.FindList(l.ctx, models.ContractPositionPageFilter{
		TenantId:     tenantId,
		UserId:       userId,
		SymbolId:     in.SymbolId,
		ContractType: int64(in.ContractType),
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	resp := &trade.GetPositionListResp{Base: helper.OkResp()}
	for _, item := range data {
		resp.Data = append(resp.Data, helpers.PositionToProto(item))
	}
	return resp, nil
}
