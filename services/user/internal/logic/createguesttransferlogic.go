package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	guestTransferKeyPrefix    = "guest:transfer:"
	defaultTransferTTLSeconds = int64(120)
	guestTransferPath         = "/profile"
)

type guestTransferPayload struct {
	TenantID     int64  `json:"tenantId"`
	UserID       int64  `json:"userId"`
	DeviceID     string `json:"deviceId"`
	Username     string `json:"username"`
	SourceOrigin string `json:"sourceOrigin"`
	TargetOrigin string `json:"targetOrigin"`
	CreatedAt    int64  `json:"createdAt"`
}

type CreateGuestTransferLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGuestTransferLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGuestTransferLogic {
	return &CreateGuestTransferLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateGuestTransfer creates a short-lived, one-time code for moving a guest session to another origin.
func (l *CreateGuestTransferLogic) CreateGuestTransfer(in *user.CreateGuestTransferReq) (*user.CreateGuestTransferResp, error) {
	sourceOrigin, ok := normalizeTransferOrigin(in.GetSourceOrigin())
	if !ok {
		return &user.CreateGuestTransferResp{Base: helper.OkResp()}, nil
	}

	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	guest, err := l.svcCtx.UserModel.FindOne(l.ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return createGuestTransferError(l.ctx), nil
		}
		return nil, err
	}
	if guest.IsGuest != 2 || guest.Deleted != 0 || guest.Status != 1 {
		return createGuestTransferError(l.ctx), nil
	}
	domain, err := l.svcCtx.SystemCli.ResolveTenantDomain(l.ctx, &system.ResolveTenantDomainReq{
		TenantId: guest.TenantId, SourceOrigin: sourceOrigin,
	})
	if err != nil {
		return nil, err
	}
	if domain.GetSourceStatus() != system.TenantDomainStatus_TENANT_DOMAIN_STATUS_RETIRED || domain.GetTargetOrigin() == "" {
		return &user.CreateGuestTransferResp{Base: helper.OkResp()}, nil
	}
	targetOrigin, ok := normalizeTransferOrigin(domain.GetTargetOrigin())
	if !ok || targetOrigin == sourceOrigin {
		return &user.CreateGuestTransferResp{Base: helper.OkResp()}, nil
	}

	code, err := randomGuestTransferCode()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	payload, err := json.Marshal(guestTransferPayload{
		TenantID: guest.TenantId, UserID: guest.Id, DeviceID: guest.DeviceId,
		Username: guest.Username, SourceOrigin: sourceOrigin, TargetOrigin: targetOrigin, CreatedAt: now.UnixMilli(),
	})
	if err != nil {
		return nil, err
	}
	ttl := l.svcCtx.Config.GuestTransfer.ExpireSeconds
	if ttl <= 0 {
		ttl = defaultTransferTTLSeconds
	}
	if err := l.svcCtx.Redis.SetexCtx(l.ctx, guestTransferRedisKey(code), string(payload), int(ttl)); err != nil {
		return nil, err
	}

	return &user.CreateGuestTransferResp{
		Base: helper.OkResp(),
		Data: &user.CreateGuestTransferData{
			Code: code, RedirectUrl: fmt.Sprintf("%s%s#code=%s", targetOrigin, guestTransferPath, url.QueryEscape(code)),
			ExpireAt: now.Add(time.Duration(ttl) * time.Second).UnixMilli(),
		},
	}, nil
}

func randomGuestTransferCode() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func guestTransferRedisKey(code string) string {
	hash := sha256.Sum256([]byte(code))
	return guestTransferKeyPrefix + hex.EncodeToString(hash[:])
}

func normalizeTransferOrigin(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	hostname := strings.ToLower(parsed.Hostname())
	ip := net.ParseIP(hostname)
	allowHTTP := hostname == "localhost" || (ip != nil && (ip.IsLoopback() || ip.IsPrivate()))
	if scheme != "https" && !(scheme == "http" && allowHTTP) {
		return "", false
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", false
	}
	return scheme + "://" + strings.ToLower(parsed.Host), true
}

func createGuestTransferError(ctx context.Context) *user.CreateGuestTransferResp {
	return &user.CreateGuestTransferResp{Base: helper.ErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, ctx))}
}
