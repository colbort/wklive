package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInsuranceCoverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInsuranceCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInsuranceCoverLogic {
	return &GetInsuranceCoverLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetInsuranceCoverLogic) GetInsuranceCover(in *asset.GetInsuranceCoverReq) (*asset.GetInsuranceCoverResp, error) {
	liquidationNo := strings.TrimSpace(in.GetLiquidationNo())
	if in.GetTenantId() <= 0 || liquidationNo == "" {
		return nil, fmt.Errorf("invalid insurance cover query")
	}
	row, err := models.NewTAssetInsuranceCoverModel(l.svcCtx.DB, l.svcCtx.Config.CacheRedis).
		FindOneByTenantLiquidationNo(l.ctx, in.GetTenantId(), liquidationNo)
	if err != nil {
		return nil, err
	}
	return &asset.GetInsuranceCoverResp{
		Base: helper.OkResp(), PlatformAccountId: row.PlatformAccountId, Coin: row.Coin,
		LiquidationId: row.LiquidationId, LiquidationNo: row.LiquidationNo,
		RequestedAmount: row.RequestedAmount.String(), CoveredAmount: row.CoveredAmount.String(),
		RemainingAmount: row.RemainingAmount.String(), Status: row.Status, CreateTimes: row.CreateTimes,
	}, nil
}
