// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package security

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/reqenc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EncryptionSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEncryptionSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EncryptionSessionLogic {
	return &EncryptionSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EncryptionSessionLogic) EncryptionSession(req *types.CreateEncryptionSessionReq) (resp *types.EncryptionSessionResp, err error) {
	session, err := l.svcCtx.RequestEncryption.CreateSession(l.ctx, reqenc.CreateSessionRequest{
		Version:      req.Version,
		RSAKid:       req.RsaKid,
		EncryptedKey: req.EncryptedKey,
	})
	if err != nil {
		return nil, err
	}
	return &types.EncryptionSessionResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: types.EncryptionSessionData{
			KeyId:       session.KeyID,
			ExpiresAt:   session.ExpiresAt,
			RotateAfter: session.RotateAfter,
		},
	}, nil
}
