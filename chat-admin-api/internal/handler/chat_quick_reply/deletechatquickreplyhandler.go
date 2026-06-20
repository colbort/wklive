// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_quick_reply

import (
	"net/http"

	"chat-admin-api/internal/logic/chat_quick_reply"
	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteChatQuickReplyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteChatQuickReplyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat_quick_reply.NewDeleteChatQuickReplyLogic(r.Context(), svcCtx)
		resp, err := l.DeleteChatQuickReply(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
