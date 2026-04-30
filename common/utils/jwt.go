package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
	Expand   string `json:"expand"`
	jwt.RegisteredClaims
}

func GenToken(secret string, userId int64, username string, expand string, issuser string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserId:   userId,
		Username: username,
		Expand:   expand,
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

func GetUserIdFromCtx(ctx context.Context) (int64, error) {
	v := ctx.Value(CtxKeyUid)
	if v == nil {
		return 0, errors.New("userId not found in context")
	}

	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case json.Number:
		return val.Int64()
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, fmt.Errorf("invalid userId type: %T, value=%#v", v, v)
	}
}

func GetUsernameFromCtx(ctx context.Context) (string, error) {
	username, ok := ctx.Value(CtxKeyUsername).(string)
	if !ok {
		return "", fmt.Errorf("%s not found in context", CtxKeyUsername)
	}
	return username, nil
}

func GetTenantIdFromCtx(ctx context.Context) (int64, error) {
	tenantId, ok := ctx.Value(CtxKeyTenantId).(int64)
	if !ok {
		return 0, fmt.Errorf("%s not found in context", CtxKeyTenantId)
	}
	return tenantId, nil
}

func GetTenantCodeFromCtx(ctx context.Context) (string, error) {
	tenantCode, ok := ctx.Value(CtxKeyTenantCode).(string)
	if !ok {
		return "", fmt.Errorf("%s not found in context", CtxKeyTenantCode)
	}
	return tenantCode, nil
}
