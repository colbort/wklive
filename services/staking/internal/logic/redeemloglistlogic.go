package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RedeemLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRedeemLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedeemLogListLogic {
	return &RedeemLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取赎回记录列表
func (l *RedeemLogListLogic) RedeemLogList(in *staking.AdminRedeemLogListReq) (*staking.AdminRedeemLogListResp, error) {
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeRedeemLogModel.FindPage(
		l.ctx,
		models.StakeRedeemLogPageFilter{
			TenantId:     in.TenantId,
			UserId:       in.UserId,
			ProductId:    in.ProductId,
			OrderNo:      in.OrderNo,
			RedeemNo:     in.RedeemNo,
			RedeemType:   int64(in.RedeemType),
			RedeemStatus: int64(in.RedeemStatus),
			RedeemBegin:  in.RedeemTimesBegin,
			RedeemEnd:    in.RedeemTimesEnd,
		},
		cursor,
		limit,
	)
	if err != nil {
		return nil, err
	}

	resp := &staking.AdminRedeemLogListResp{Page: helper.OkResp()}
	if len(items) == 0 {
		resp.Page = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeRedeemLog, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, redeemLogToProto(item))
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
