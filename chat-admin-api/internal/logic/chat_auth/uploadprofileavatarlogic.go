// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_auth

import (
	"context"
	"errors"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadProfileAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadProfileAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadProfileAvatarLogic {
	return &UploadProfileAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadProfileAvatarLogic) UploadProfileAvatar() (resp *types.ChatAdminProfileResp, err error) {
	// The multipart handler owns file validation/storage and then calls
	// UpdateProfileLogic. This generated no-argument logic must never report a
	// zero-value success if it is invoked directly.
	return nil, errors.New("profile avatar upload must use the multipart HTTP handler")
}
