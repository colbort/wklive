package chat_ws

import (
	"fmt"
	"net/http"
	"strings"
	"wklive/proto/chat"

	"chat-api/internal/jwt"
	"chat-api/internal/logic/chat_ws"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	wsProtocolMerchantPrefix = "merchant."
	wsProtocolUserPrefix     = "user."
	wsProtocolNicknamePrefix = "nickname."
	wsProtocolAvatarPrefix   = "avatar."
	wsProtocol               = "wklive-chat"
)

var upgrader = websocket.Upgrader{
	Subprotocols: []string{wsProtocol},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwt.Verify(svcCtx.Config.Jwt.AccessSecret, jwt.TokenFromRequest(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		nickname := firstNonEmpty(claims.Nickname, fmt.Sprintf("user-%d", claims.UserId))
		if claims.IsGuest {
			logx.Infof("chat ws guest identity resolved by chatToken, merchantId=%d userId=%d nickname=%s", claims.MerchantId, claims.UserId, nickname)

		} else {
			logx.Infof("chat ws identity resolved by chatToken, merchantId=%d userId=%d nickname=%s", claims.MerchantId, claims.UserId, nickname)
		}
		sessionNo := strings.TrimSpace(claims.SessionNo)
		if sessionNo == "" {
			resp, err := svcCtx.ChatAppCli.GenerateChatSessionNo(r.Context(), &chat.GenerateChatSessionNoReq{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			sessionNo = resp.SessionNo
		}
		req := types.ChatWSMessagesReq{
			SessionNo:  sessionNo,
			MerchantId: claims.MerchantId,
			UserId:     claims.UserId,
			Nickname:   nickname,
			AvatarUrl:  claims.AvatarUrl,
			IsGuest:    claims.IsGuest,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("upgrade chat user ws failed, userId=%d merchantId=%d temporary=%t err=%v", req.UserId, req.MerchantId, req.IsGuest, err)
			return
		}
		l := chat_ws.NewMessagesLogic(r.Context(), svcCtx)
		l.Messages(conn, req)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}
