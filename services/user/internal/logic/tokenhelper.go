package logic

import (
	"time"
	"wklive/common/utils"
	"wklive/proto/common"
)

const refreshTokenExtendSeconds int64 = 7 * 24 * 3600

func buildTokenInfo(secret string, accessExpireSeconds int64, userId int64, username string, expand string) (*common.TokenInfo, error) {
	now := time.Now()
	accessTTL := time.Duration(accessExpireSeconds) * time.Second
	refreshTTL := time.Duration(accessExpireSeconds+refreshTokenExtendSeconds) * time.Second

	accessToken, err := utils.GenToken(secret, userId, username, expand, "", accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenToken(secret, userId, username, expand, "", refreshTTL)
	if err != nil {
		return nil, err
	}

	return &common.TokenInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpireTime:   now.Add(accessTTL).UnixMilli(),
	}, nil
}
