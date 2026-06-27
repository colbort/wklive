// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"
	"strings"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"wklive/proto/chat"
	"wklive/proto/common"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatMessagesLogic {
	return &PageChatMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatMessagesLogic) PageChatMessages(req *types.PageChatMessagesReq) (resp *types.PageChatMessagesResp, err error) {
	sessionNo := strings.TrimSpace(req.SessionNo)
	if l.svcCtx.ChatAdminCli != nil {
		transientSession, err := l.svcCtx.ChatAdminCli.GetTransientChatSession(l.ctx, &chat.GetTransientChatSessionReq{
			MerchantId: req.MerchantId,
			SessionNo:  sessionNo,
		})
		if err == nil && transientSession.GetBase().GetCode() == 200 {
			rpcResp, err := l.svcCtx.ChatAdminCli.ListTransientChatMessages(l.ctx, &chat.ListTransientChatMessagesReq{
				MerchantId: req.MerchantId,
				SessionNo:  sessionNo,
				SenderType: chat.ChatSenderType(req.SenderType),
				Page: &common.PageReq{
					Cursor: req.Cursor,
					Limit:  req.Limit,
				},
			})
			if err != nil {
				return nil, err
			}
			return &types.PageChatMessagesResp{
				RespBase: respBaseToType(rpcResp.GetBase()),
				Data:     protoMessagesToTypes(req.MerchantId, rpcResp.GetData()),
			}, nil
		}
	}
	rpcResp, err := l.svcCtx.ChatAdminCli.PageChatMessages(l.ctx, &chat.PageChatMessagesReq{
		SessionNo:  sessionNo,
		SenderType: chat.ChatSenderType(req.SenderType),
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
	})
	if err != nil {
		return nil, err
	}
	return &types.PageChatMessagesResp{
		RespBase: respBaseToType(rpcResp.GetBase()),
		Data:     protoMessagesToTypes(req.MerchantId, rpcResp.GetData()),
	}, nil
}

func respBaseToType(base *common.RespBase) types.RespBase {
	if base == nil {
		return types.RespBase{}
	}
	return types.RespBase{
		Code:       base.GetCode(),
		Msg:        base.GetMsg(),
		Total:      base.GetTotal(),
		HasNext:    base.GetHasNext(),
		HasPrev:    base.GetHasPrev(),
		NextCursor: base.GetNextCursor(),
		PrevCursor: base.GetPrevCursor(),
	}
}

func protoMessagesToTypes(merchantId int64, list []*chat.ChatMessage) []types.ChatMessage {
	if len(list) == 0 {
		return []types.ChatMessage{}
	}
	resp := make([]types.ChatMessage, 0, len(list))
	for _, item := range list {
		resp = append(resp, protoMessageToType(merchantId, item))
	}
	return resp
}

func protoMessageToType(merchantId int64, item *chat.ChatMessage) types.ChatMessage {
	if item == nil {
		return types.ChatMessage{}
	}
	return types.ChatMessage{
		MessageNo:   item.GetMessageNo(),
		SessionNo:   item.GetSessionNo(),
		MerchantId:  merchantId,
		Sender:      protoMessageSenderToType(item.GetSender()),
		Receiver:    protoMessageSenderToType(item.GetReceiver()),
		MessageType: int64(item.GetMessageType()),
		Content:     item.GetContent(),
		Url:         item.GetUrl(),
		FileName:    item.GetFileName(),
		FileSize:    item.GetFileSize(),
		MimeType:    item.GetMimeType(),
		Width:       int64(item.GetWidth()),
		Height:      int64(item.GetHeight()),
		Duration:    int64(item.GetDuration()),
		Status:      int64(item.GetStatus()),
		Extra:       item.GetExtra(),
		CreateTime:  item.GetCreateTime(),
		UpdateTime:  item.GetUpdateTime(),
	}
}

func protoMessageSenderToType(item *chat.ChatMessageUser) types.ChatMessageSender {
	if item == nil {
		return types.ChatMessageSender{}
	}
	return types.ChatMessageSender{
		Id:         item.GetId(),
		SenderType: int64(item.GetType()),
		Nickname:   item.GetNickname(),
		AvatarUrl:  item.GetAvatarUrl(),
	}
}
