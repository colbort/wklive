package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyRedeemLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMyRedeemLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyRedeemLogListLogic {
	return &MyRedeemLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取我的赎回记录列表
func (l *MyRedeemLogListLogic) MyRedeemLogList(in *staking.AppMyRedeemLogListReq) (*staking.AppMyRedeemLogListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeRedeemLogModel.FindPage(l.ctx, in.TenantId, cursor, limit, in.Uid, in.OrderId, 0, "", "", 0, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	resp := &staking.AppMyRedeemLogListResp{Base: helper.OkResp()}
	if len(items) == 0 {
		resp.Base = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeRedeemLog, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, redeemLogToProto(item))
	}
	resp.Base = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
