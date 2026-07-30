package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/market/kline"
	"wklive/services/market/internal/pkg/utils"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncProductKlineHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncProductKlineHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductKlineHistoryLogic {
	return &SyncProductKlineHistoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SyncProductKlineHistoryLogic) SyncProductKlineHistory(in *market.SyncProductKlineHistoryReq) (*market.SyncProductKlineHistoryResp, error) {
	if strings.TrimSpace(l.svcCtx.Config.Itick.ApiUrl) == "" {
		return &market.SyncProductKlineHistoryResp{Base: helper.ErrResp(i18n.ApiURLRequired, i18n.Translate(i18n.ApiURLRequired, l.ctx))}, nil
	}
	if strings.TrimSpace(l.svcCtx.Config.Itick.Token) == "" {
		return &market.SyncProductKlineHistoryResp{Base: helper.ErrResp(i18n.ApiTokenRequired, i18n.Translate(i18n.ApiTokenRequired, l.ctx))}, nil
	}
	interval := utils.KlineTypeToInterval(in.KType)
	if interval == "" || strings.TrimSpace(in.CategoryCode) == "" || strings.TrimSpace(in.Market) == "" || strings.TrimSpace(in.Symbol) == "" {
		return &market.SyncProductKlineHistoryResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	worker := kline.NewSyncKlinesWorker(l.ctx, l.svcCtx, nil, "", "")
	result, err := worker.FetchProductHistory(kline.KlineJob{ApiUrl: l.svcCtx.Config.Itick.ApiUrl, Token: l.svcCtx.Config.Itick.Token,
		Category: in.CategoryCode, Market: in.Market, Symbol: in.Symbol, KType: int32(in.KType)}, interval, in.EndTs)
	if err != nil {
		return nil, err
	}
	return &market.SyncProductKlineHistoryResp{Base: helper.OkResp(), SyncedCount: int64(result.NewCount)}, nil
}
