package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerificationCodeRecordListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerificationCodeRecordListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerificationCodeRecordListLogic {
	return &VerificationCodeRecordListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证码发送记录
func (l *VerificationCodeRecordListLogic) VerificationCodeRecordList(in *system.VerificationCodeRecordListReq) (*system.VerificationCodeRecordListResp, error) {
	limit := int64(20)
	cursor := int64(0)
	if in.Page != nil {
		limit = in.Page.Limit
		cursor = in.Page.Cursor
	}

	rows, total, err := l.svcCtx.VerificationCodeRecordModel.FindPage(
		l.ctx,
		in.TenantId,
		int64(in.Channel),
		in.Target,
		int64(in.Scene),
		int64(in.Status),
		cursor,
		limit,
	)
	if err != nil {
		return nil, err
	}

	data := make([]*system.VerificationCodeRecordItem, 0, len(rows))
	for _, row := range rows {
		data = append(data, verificationCodeRecordItem(row))
	}

	lastID := int64(0)
	if len(rows) > 0 {
		lastID = rows[len(rows)-1].Id
	}

	return &system.VerificationCodeRecordListResp{
		Base: pageutil.Base(cursor, limit, len(rows), total, lastID),
		Data: data,
	}, nil
}
