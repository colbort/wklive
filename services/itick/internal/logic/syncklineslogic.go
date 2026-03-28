package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncKlinesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncKlinesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncKlinesLogic {
	return &SyncKlinesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步K线数据
func (l *SyncKlinesLogic) SyncKlines(in *itick.SyncKlinesReq) (*itick.SyncKlinesResp, error) {
	// todo: add your logic here and delete this line

	return &itick.SyncKlinesResp{}, nil
}
