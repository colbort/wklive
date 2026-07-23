package adminlogic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSettlementLogic {
	return &GetSettlementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个到期结算记录详情
func (l *GetSettlementLogic) GetSettlement(in *option.GetSettlementReq) (*option.GetSettlementResp, error) {
	item, err := findSettlementByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.Id, in.SettlementNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetSettlementResp{Base: helper.ErrResp(i18n.SettlementRecordNotFound, i18n.Translate(i18n.SettlementRecordNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	data, err := buildSettlementDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetSettlementResp{Base: helper.OkResp(), Data: data}, nil
}
