package adminlogic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpLogListLogic {
	return &OpLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OpLogListLogic) OpLogList(in *system.OpLogListReq) (*system.OpLogListResp, error) {
	tenantID, _ := utils.GetTenantIdFromMd(l.ctx)
	items, total, err := l.svcCtx.OpLogModel.FindPage(
		l.ctx,
		models.OpLogPageFilter{
			TenantId: tenantID,
			Username: in.Username,
			Method:   requestMethodToString(in.Method),
			Path:     in.Path,
		},
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*system.OpLogItem, 0, len(items))
	for _, item := range items {
		data = append(data, &system.OpLogItem{
			Id:          item.Id,
			TenantId:    item.TenantId,
			UserId:      item.UserId,
			Username:    item.Username,
			Module:      item.Module,
			Action:      item.Action,
			Method:      requestMethodToProto(item.Method),
			Path:        item.Path,
			Req:         item.Req.String,
			Resp:        item.Resp.String,
			Ip:          item.Ip,
			CostMs:      item.CostMs,
			CreateTimes: item.CreateTimes,
			UpdateTimes: item.UpdateTimes,
		})
	}

	return &system.OpLogListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
