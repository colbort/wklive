package adminlogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSymbolConfigLogic {
	return &CreateSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSymbolConfigLogic) CreateSymbolConfig(in *liquidity.SaveSymbolConfigReq) (*liquidity.SymbolConfigResp, error) {
	row, err := buildSymbolConfig(l.ctx, l.svcCtx, in, nil)
	if err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.SymbolConfigModel.FindOneByTenantIdSymbolIdProductType(l.ctx, row.TenantId, row.SymbolId, row.ProductType); err == nil {
		return nil, fmt.Errorf("symbol liquidity config already exists")
	} else if err != models.ErrNotFound {
		return nil, err
	}
	result, err := l.svcCtx.SymbolConfigModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	row.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &liquidity.SymbolConfigResp{Base: helper.OkResp(), Data: helpers.SymbolConfigToProto(row)}, nil
}
