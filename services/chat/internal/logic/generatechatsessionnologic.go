package logic

import (
	"context"
	"fmt"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

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
	for attempt := 0; attempt < sessionNoInsertAttempts; attempt++ {
		sessionNo := nextNo("CS")
		_, err := l.svcCtx.ChatSessionModel.FindOneBySessionNo(l.ctx, sessionNo)
		if err == models.ErrNotFound {
			return &chat.GenerateChatSessionNoResp{Base: okBase(), SessionNo: sessionNo}, nil
		}
		if err != nil {
			return &chat.GenerateChatSessionNoResp{Base: errorBase(err)}, nil
		}
	}

	return &chat.GenerateChatSessionNoResp{Base: errorBase(fmt.Errorf("failed to generate unique session_no"))}, nil
}
