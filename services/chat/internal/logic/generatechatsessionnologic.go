package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateChatSessionNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateChatSessionNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateChatSessionNoLogic {
	return &GenerateChatSessionNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 生成会话编号
func (l *GenerateChatSessionNoLogic) GenerateChatSessionNo(in *chat.GenerateChatSessionNoReq) (*chat.GenerateChatSessionNoResp, error) {
	sessionNo, err := l.svcCtx.GenerateNo(l.ctx, "CS")
	if err != nil {
		return &chat.GenerateChatSessionNoResp{Base: helper.ErrResp(500, err.Error())}, nil
	} else {
		return &chat.GenerateChatSessionNoResp{Base: helper.OkResp(), SessionNo: sessionNo}, nil
	}
}
