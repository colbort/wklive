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
	sessionNo := ""
	if in.IsGuest {
		sn, err := l.svcCtx.GenerateNo(l.ctx, "GCS")
		if err != nil {
			return nil, err
		}
		sessionNo = sn
	} else {
		session, err := l.svcCtx.ChatSessionModel.FindByUser(l.ctx, in.MerchantId, in.UserId)
		if err == nil {
			sessionNo = session.SessionNo
		} else {
			sn, err := l.svcCtx.GenerateNo(l.ctx, "CS")
			if err != nil {
				return nil, err
			}
			sessionNo = sn
		}
	}
	if sessionNo == "" {
		return &chat.GenerateChatSessionNoResp{Base: helper.ErrResp(500, "session no is empty")}, nil
	} else {
		return &chat.GenerateChatSessionNoResp{Base: helper.OkResp(), SessionNo: sessionNo}, nil
	}
}
