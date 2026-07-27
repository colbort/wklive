package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/common"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type SymbolConfigUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSymbolConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolConfigUpdateLogic {
	return &SymbolConfigUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SymbolConfigUpdateLogic) SymbolConfigUpdate(req *types.SaveSymbolConfigReq) (*types.RespBase, error) {
	_, userID, err := logicutil.Identity(l.ctx)
	if err != nil {
		return nil, err
	}
	in := logicutil.Convert[pb.SaveSymbolConfigReq](req)
	in.Id = req.Id
	in.OperatorId = userID
	out, err := l.svcCtx.LiquidityCli.UpdateSymbolConfig(l.ctx, in)
	if err != nil {
		return nil, err
	}

	// The create form models a single default quote level. Keep that level's
	// executable quantity in sync when editing an existing single-level config.
	detail, err := l.svcCtx.LiquidityCli.GetSymbolConfigDetail(l.ctx, &pb.GetSymbolConfigDetailReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	if len(detail.Levels) == 1 {
		level := detail.Levels[0]
		_, err = l.svcCtx.LiquidityCli.SetStrategyLevels(l.ctx, &pb.SetStrategyLevelsReq{
			ConfigId:      req.Id,
			ConfigVersion: out.Data.Version,
			OperatorId:    userID,
			Levels: []*pb.StrategyLevelInput{{
				LevelNo:      level.LevelNo,
				BidSpreadBps: level.BidSpreadBps,
				AskSpreadBps: level.AskSpreadBps,
				BidQty:       req.MinQuoteQty,
				AskQty:       req.MinQuoteQty,
				Enabled:      common.Enable(level.Enabled),
			}},
		})
		if err != nil {
			return nil, err
		}
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}
