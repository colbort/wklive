package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RewardLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRewardLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RewardLogListLogic {
	return &RewardLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取收益记录列表
func (l *RewardLogListLogic) RewardLogList(in *staking.RewardLogListReq) (*staking.RewardLogListResp, error) {
	tenantId, base, err := helpers.AdminTenantReadScopeResp(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if base != nil {
		return &staking.RewardLogListResp{Page: base}, nil
	}
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeRewardLogModel.FindPage(
		l.ctx,
		models.StakeRewardLogPageFilter{
			TenantId:     tenantId,
			UserId:       in.UserId,
			ProductId:    in.ProductId,
			OrderNo:      in.OrderNo,
			RewardType:   int64(in.RewardType),
			RewardStatus: int64(in.RewardStatus),
			RewardBegin:  in.RewardTimesBegin,
			RewardEnd:    in.RewardTimesEnd,
		},
		cursor,
		limit,
	)
	if err != nil {
		return nil, err
	}

	resp := &staking.RewardLogListResp{Page: helper.OkResp()}
	if len(items) == 0 {
		resp.Page = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeRewardLog, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, helpers.RewardLogToProto(item))
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
