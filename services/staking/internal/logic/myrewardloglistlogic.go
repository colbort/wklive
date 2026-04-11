package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyRewardLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyRewardLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRewardLogListLogic {
	return &MyRewardLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的收益记录列表
func (l *MyRewardLogListLogic) MyRewardLogList(in *staking.AppMyRewardLogListReq) (*staking.AppMyRewardLogListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeRewardLogModel.FindPage(l.ctx, in.TenantId, cursor, limit, in.Uid, in.OrderId, 0, "", int64(in.RewardType), 0, 0, 0)
	if err != nil {
		return nil, err
	}

	resp := &staking.AppMyRewardLogListResp{Base: helper.OkResp()}
	if len(items) == 0 {
		resp.Base = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeRewardLog, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, rewardLogToProto(item))
	}
	resp.Base = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
