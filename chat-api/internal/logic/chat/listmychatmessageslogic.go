// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"encoding/json"
	"wklive/proto/chat"
	"wklive/proto/common"

	"chat-api/internal/jwt"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatMessagesLogic {
	return &ListMyChatMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyChatMessagesLogic) ListMyChatMessages(req *types.ListChatMessagesReq) (resp *types.ListChatMessagesResp, err error) {
	claims, ok := jwt.ClaimsFromContext(l.ctx)
	if !ok || claims.SessionNo == "" {
		return &types.ListChatMessagesResp{
			RespBase: types.RespBase{Code: 200, Msg: "ok"},
			Data:     []types.ChatMessage{},
		}, nil
	}

	rpcResp, err := l.svcCtx.ChatAppCli.ListMyChatMessages(l.ctx, &chat.ListMyChatMessagesReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		SessionNo: claims.SessionNo,
	})
	if err != nil {
		return nil, err
	}
	return &types.ListChatMessagesResp{
		RespBase: respBaseToType(rpcResp.GetBase()),
		Data:     protoMessagesToTypes(claims.MerchantId, rpcResp.GetData()),
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
		Width:       item.GetWidth(),
		Height:      item.GetHeight(),
		Duration:    item.GetDuration(),
		Status:      int64(item.GetStatus()),
		Extra:       protoPayloadToJSONString(item),
		CreateTime:  item.GetCreateTimes(),
		UpdateTime:  item.GetUpdateTimes(),
	}
}

func protoPayloadToJSONString(item *chat.ChatMessage) string {
	if item == nil || item.GetPayload() == nil {
		return ""
	}
	bs, err := json.Marshal(item.GetPayload().AsMap())
	if err != nil {
		return ""
	}
	return string(bs)
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
