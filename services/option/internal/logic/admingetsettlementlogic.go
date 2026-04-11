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

type AdminGetSettlementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetSettlementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetSettlementLogic {
	return &AdminGetSettlementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个到期结算记录详情
func (l *AdminGetSettlementLogic) AdminGetSettlement(in *option.GetSettlementReq) (*option.GetSettlementResp, error) {
	item, err := findSettlementByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.Id, in.SettlementNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetSettlementResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.SettlementRecordNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	data, err := buildSettlementDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetSettlementResp{Base: helper.OkResp(), Data: data}, nil
}
