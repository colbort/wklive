// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package security

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EncryptionConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEncryptionConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EncryptionConfigLogic {
	return &EncryptionConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EncryptionConfigLogic) EncryptionConfig() (resp *types.EncryptionConfigResp, err error) {
	config := l.svcCtx.RequestEncryption.ConfigData()
	return &types.EncryptionConfigResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: types.EncryptionConfigData{
			Version:             config.Version,
			Mode:                string(config.Mode),
			Enabled:             config.Enabled,
			Required:            config.Required,
			RsaKid:              config.RSAKid,
			PublicKey:           config.PublicKey,
			KeyAlgorithm:        config.KeyAlgorithm,
			ContentAlgorithm:    config.ContentAlgorithm,
			SessionTtlSeconds:   config.SessionTTLSeconds,
			RotateBeforeSeconds: config.RotateBeforeSeconds,
			ServerTime:          config.ServerTime,
			ProtectedPrefixes:   config.ProtectedPrefixes,
		},
	}, nil
}
