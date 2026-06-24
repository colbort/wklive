package chat_ws

import (
	"fmt"
	"net/http"
	"strings"

	"chat-api/internal/jwt"
	"chat-api/internal/logic/chat_ws"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	wsProtocolMerchantPrefix = "merchant."
	wsProtocolUserPrefix     = "user."
	wsProtocolNicknamePrefix = "nickname."
	wsProtocolAvatarPrefix   = "avatar."
)

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwt.Verify(svcCtx.Config.Jwt.AccessSecret, jwt.TokenFromRequest(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		nickname := firstNonEmpty(claims.Nickname, fmt.Sprintf("user-%d", claims.UserId))
		if claims.IsGuest {
			logx.Infof("chat ws guest identity resolved by chatToken, merchantId=%d userId=%d sessionNo=%s nickname=%s", claims.MerchantId, claims.UserId, nickname)

		} else {
			logx.Infof("chat ws identity resolved by chatToken, merchantId=%d userId=%d sessionNo=%s nickname=%s", claims.MerchantId, claims.UserId, nickname)
		}
		req := types.ChatWSMessagesReq{
			SessionNo:  claims.SessionNo,
			MerchantId: claims.MerchantId,
			UserId:     claims.UserId,
			Nickname:   nickname,
			AvatarUrl:  claims.AvatarUrl,
			IsGuest:    claims.IsGuest,
		}
		l := chat_ws.NewMessagesLogic(r.Context(), svcCtx)
		l.Messages(w, r, req)
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
