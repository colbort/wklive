package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

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
func (l *RewardLogListLogic) RewardLogList(in *staking.AdminRewardLogListReq) (*staking.AdminRewardLogListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeRewardLogModel.FindPage(
		l.ctx, in.TenantId, cursor, limit, in.Uid, 0, in.ProductId, in.OrderNo,
		int64(in.RewardType), int64(in.RewardStatus), in.RewardTimesBegin, in.RewardTimesEnd,
	)
	if err != nil {
		return nil, err
	}

	resp := &staking.AdminRewardLogListResp{Page: helper.OkResp()}
	if len(items) == 0 {
		resp.Page = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeRewardLog, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, rewardLogToProto(item))
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
