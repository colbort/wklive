package jwt

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	wsProtocolTokenPrefix = "token."
)

type contextKey string

const claimsContextKey contextKey = "chat-token-claims"

type Claims struct {
	MerchantId int64  `json:"merchantId"`
	UserId     int64  `json:"userId"`
	SessionNo  string `json:"sessionNo,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	AvatarUrl  string `json:"avatarUrl,omitempty"`
	IsGuest    bool   `json:"isGuest,omitempty"`
	ExpireAt   int64  `json:"expireAt"`
	jwt.RegisteredClaims
}

func Sign(secret string, claims Claims) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", fmt.Errorf("chat token secret is required")
	}
	if claims.MerchantId <= 0 {
		return "", fmt.Errorf("merchantId is required")
	}
	if claims.UserId <= 0 {
		return "", fmt.Errorf("userId is required")
	}
	if claims.ExpireAt <= time.Now().UnixMilli() {
		return "", fmt.Errorf("expireAt is invalid")
	}
	now := time.Now()
	if claims.RegisteredClaims.IssuedAt == nil {
		claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	}
	if claims.RegisteredClaims.NotBefore == nil {
		claims.RegisteredClaims.NotBefore = jwt.NewNumericDate(now.Add(-5 * time.Second))
	}
	if claims.RegisteredClaims.ExpiresAt == nil {
		claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.UnixMilli(claims.ExpireAt))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Verify(secret string, tokenString string) (Claims, error) {
	var claims Claims
	secret = strings.TrimSpace(secret)
	tokenString = strings.TrimSpace(tokenString)
	if secret == "" {
		return claims, fmt.Errorf("chat token secret is required")
	}
	if tokenString == "" {
		return claims, fmt.Errorf("invalid chat token")
	}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !token.Valid {
		return Claims{}, fmt.Errorf("invalid chat token")
	}
	if claims.MerchantId <= 0 || claims.UserId <= 0 {
		return claims, fmt.Errorf("invalid chat token identity")
	}
	if claims.ExpireAt <= time.Now().UnixMilli() {
		return claims, fmt.Errorf("chat token expired")
	}
	return claims, nil
}

func ContextWithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	return claims, ok
}

func TokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if token := strings.TrimSpace(r.URL.Query().Get("chatToken")); token != "" {
		return token
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	if token := strings.TrimSpace(r.Header.Get("x-chat-token")); token != "" {
		return token
	}
	return tokenFromProtocol(r.Header.Get("Sec-WebSocket-Protocol"))
}

func ProtocolToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	return wsProtocolTokenPrefix + base64.RawURLEncoding.EncodeToString([]byte(token))
}

func tokenFromProtocol(protocols string) string {
	for _, part := range strings.Split(protocols, ",") {
		value := strings.TrimSpace(part)
		if !strings.HasPrefix(value, wsProtocolTokenPrefix) {
			continue
		}
		raw := strings.TrimPrefix(value, wsProtocolTokenPrefix)
		decoded, err := base64.RawURLEncoding.DecodeString(raw)
		if err != nil {
			return ""
		}
		return string(decoded)
	}
	return ""
}
