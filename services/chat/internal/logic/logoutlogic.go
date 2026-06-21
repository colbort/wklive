package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 退出登录
func (l *LogoutLogic) Logout(in *chat.ChatAdminLogoutReq) (*chat.AdminCommonResp, error) {
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || userID <= 0 {
		return &chat.AdminCommonResp{Base: badBase("invalid login session")}, nil
	}
	if _, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, userID); err != nil {
		if err == models.ErrNotFound {
			return &chat.AdminCommonResp{Base: badBase("invalid login session")}, nil
		}
		return &chat.AdminCommonResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminCommonResp{Base: okBase()}, nil
}
