package adminlogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/common"
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
	if _, err := l.svcCtx.SymbolConfigModel.FindOneBySymbolIdProductType(l.ctx, row.SymbolId, row.ProductType); err == nil {
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
	defaultLevel := []models.LiquidityStrategyLevelInput{{
		LevelNo:      1,
		BidSpreadBps: row.BaseSpreadBps,
		AskSpreadBps: row.BaseSpreadBps,
		BidQty:       row.MinQuoteQty,
		AskQty:       row.MinQuoteQty,
		Enabled:      int64(common.Switch_SWITCH_ON),
	}}
	if err := l.svcCtx.StrategyLevelModel.Replace(l.ctx, row.Id, defaultLevel, row.CreateTimes); err != nil {
		// 主配置和默认档位应作为一个完整的创建结果；档位失败时补偿删除主配置。
		if deleteErr := l.svcCtx.SymbolConfigModel.Delete(l.ctx, row.Id); deleteErr != nil {
			l.Errorf("rollback symbol config failed configId=%d err=%v", row.Id, deleteErr)
		}
		return nil, err
	}
	return &liquidity.SymbolConfigResp{Base: helper.OkResp(), Data: helpers.SymbolConfigToProto(row)}, nil
}
