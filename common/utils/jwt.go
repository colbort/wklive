package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
