package utils

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/logx"
)

type Claims struct {
	Uid      int64  `json:"uid"`
	Username string `json:"username"`
	PermsVer int64  `json:"permsVer"`
	jwt.RegisteredClaims
}

func GenToken(secret string, uid int64, username string, permsVer int64, issuser string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Uid:      uid,
		Username: username,
		PermsVer: permsVer,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuser,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)), // 防止时钟偏差
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func GetUidFromCtx(ctx context.Context) (int64, error) {
	var uid int64
	jsonUid, ok := ctx.Value("uid").(json.Number)
	if !ok {
		logx.WithContext(ctx).Errorf("GetUidFromCtx err : not found uid in context")
		return 0, errors.New("uid not found in context")

	}
	if int64Uid, err := jsonUid.Int64(); err == nil {
		uid = int64Uid
	} else {
		logx.WithContext(ctx).Errorf("GetUidFromCtx err : %+v", err)
		return 0, err
	}
	return uid, nil
}
