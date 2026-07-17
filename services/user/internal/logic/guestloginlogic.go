package logic

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/proto/user"
	"wklive/services/user/internal/constant"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GuestLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGuestLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GuestLoginLogic {
	return &GuestLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 游客登了
func (l *GuestLoginLogic) GuestLogin(in *user.GuestLoginReq) (*user.GuestLoginResp, error) {
	registerIP, _ := utils.GetClientIPFromMd(l.ctx)
	tenantCode, err := utils.GetTenantCodeFromMd(l.ctx)
	if err != nil || tenantCode == "" {
		return &user.GuestLoginResp{
			Base: helper.ErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx)),
		}, nil
	}
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &tenantCode,
	})
	if err != nil {
		return nil, err
	}
	if tenant == nil || tenant.Data == nil {
		return &user.GuestLoginResp{
			Base: helper.ErrResp(i18n.TenantNotFound, i18n.Translate(i18n.TenantNotFound, l.ctx)),
		}, nil
	}

	if in.DeviceId == "" && in.Fingerprint == "" {
		return &user.GuestLoginResp{
			Base: helper.ErrResp(i18n.PleaseSwitchDeviceToLogin, i18n.Translate(i18n.PleaseSwitchDeviceToLogin, l.ctx)),
		}, nil
	}

	var matched *models.TUser
	if in.DeviceId != "" {
		matched, err = l.svcCtx.UserModel.FindGuestByDeviceId(l.ctx, tenant.Data.Id, in.DeviceId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
	}

	if matched == nil && in.Fingerprint != "" {
		matched, err = l.findMatchedGuestByFingerprint(tenant.Data.Id, in.Fingerprint)
		if err != nil {
			return nil, err
		}
	}

	if matched != nil {
		now := utils.NowMillis()
		matched.LastLoginIp = sql.NullString{String: registerIP, Valid: registerIP != ""}
		matched.LastLoginTime = now
		matched.UpdateTimes = now
		if matched.DeviceId == "" {
			matched.DeviceId = fmt.Sprintf("%d", matched.Id)
		}
		_ = l.svcCtx.UserModel.Update(l.ctx, matched)
		if err := l.saveFingerprint(tenant.Data.Id, matched.Id, matched.DeviceId, in.Fingerprint, registerIP, now); err != nil {
			return nil, err
		}

		resp, err := l.buildGuestLoginResp(tenant.Data.Id, matched, false)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	err = l.checkGuestLimit(registerIP)
	if err != nil {
		return nil, err
	}

	userNo := l.svcCtx.Node.Generate().Int64()
	inviteCode, err := l.svcCtx.GenerateInviteCode(l.ctx, tenant.Data.Id)
	if err != nil {
		return nil, err
	}
	nickname, err := constant.RandomNickname()
	if err != nil {
		return nil, err
	}
	deviceId := fmt.Sprintf("%d", userNo)
	now := utils.NowMillis()
	guest := &models.TUser{
		TenantId:       tenant.Data.Id,
		UserNo:         fmt.Sprintf("G%d", userNo),
		Username:       fmt.Sprintf("Guest%d", userNo),
		Nickname:       sql.NullString{String: nickname, Valid: true},
		Avatar:         sql.NullString{},
		PasswordHash:   "",
		RegisterType:   int64(user.RegisterType_REGISTER_TYPE_GUEST),
		Status:         1,
		MemberLevel:    0,
		Language:       sql.NullString{},
		Timezone:       sql.NullString{},
		InviteCode:     sql.NullString{String: inviteCode, Valid: true},
		Signature:      sql.NullString{},
		Source:         sql.NullString{String: "guest", Valid: true},
		ReferrerUserId: sql.NullInt64{},
		LastLoginIp:    sql.NullString{String: registerIP, Valid: registerIP != ""},
		LastLoginTime:  now,
		RegisterIp:     sql.NullString{String: registerIP, Valid: registerIP != ""},
		RegisterTime:   now,
		IsGuest:        2,
		IsRecharge:     int64(common.YesNo_YES_NO_NO),
		DeviceId:       deviceId,
		Fingerprint:    sql.NullString{String: in.Fingerprint, Valid: true},
		Remark:         sql.NullString{},
		Deleted:        0,
		CreateTimes:    now,
		UpdateTimes:    now,
	}
	result, err := l.svcCtx.UserModel.Insert(l.ctx, guest)
	if err != nil {
		return nil, err
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	guest.Id = insertId
	if err := l.saveFingerprint(tenant.Data.Id, guest.Id, guest.DeviceId, in.Fingerprint, registerIP, now); err != nil {
		return nil, err
	}

	resp, err := l.buildGuestLoginResp(tenant.Data.Id, guest, true)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (l *GuestLoginLogic) buildGuestLoginResp(tenantId int64, guest *models.TUser, isNew bool) (*user.GuestLoginResp, error) {
	str := make(map[string]any, 1)
	str["tenantId"] = tenantId
	expand, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}

	token, err := buildTokenInfo(
		l.svcCtx.Config.Jwt.AccessSecret,
		24*3600,
		guest.Id, guest.Username, string(expand),
	)
	if err != nil {
		return nil, err
	}
	return &user.GuestLoginResp{
		Base: helper.OkResp(),
		Data: &user.GuestLogin{
			Token:    token.AccessToken,
			UserId:   guest.DeviceId,
			Username: guest.Username,
			IsNew:    isNew,
		},
	}, nil
}

func (l *GuestLoginLogic) findMatchedGuestByFingerprint(tenantId int64, fingerprint string) (*models.TUser, error) {
	visitorId, ok := parseFingerprintVisitorID(fingerprint)
	if !ok {
		return nil, nil
	}

	matchKey := buildFingerprintMatchKey(visitorId)

	cursor := int64(0)
	limit := int64(500)
	for {
		fingerprintCandidates, err := l.svcCtx.FingerprintModel.FindGuestFingerprintCandidates(l.ctx, tenantId, matchKey, cursor, limit)
		if err != nil {
			return nil, err
		}
		if len(fingerprintCandidates) == 0 {
			break
		}

		for _, candidate := range fingerprintCandidates {
			cursor = candidate.Id
			storedVisitorId, ok := parseFingerprintVisitorID(candidate.Fingerprint)
			if !ok || storedVisitorId != visitorId {
				continue
			}

			matched, err := l.svcCtx.UserModel.FindByTenantIdUserId(l.ctx, tenantId, candidate.UserId)
			if err != nil {
				if errors.Is(err, models.ErrNotFound) {
					continue
				}
				return nil, err
			}
			return matched, nil
		}

		if int64(len(fingerprintCandidates)) < limit {
			break
		}
	}

	return nil, nil
}

func (l *GuestLoginLogic) saveFingerprint(tenantId int64, userId int64, deviceId string, fingerprint string, sourceIp string, now int64) error {
	if fingerprint == "" {
		return nil
	}
	visitorId, ok := parseFingerprintVisitorID(fingerprint)
	if !ok {
		return nil
	}
	matchKey := buildFingerprintMatchKey(visitorId)

	hash := sha256.Sum256([]byte(fingerprint))
	return l.svcCtx.FingerprintModel.UpsertSeen(l.ctx, &models.TUserFingerprint{
		TenantId:        tenantId,
		UserId:          userId,
		DeviceId:        deviceId,
		FingerprintHash: hex.EncodeToString(hash[:]),
		MatchKey:        matchKey,
		Fingerprint:     fingerprint,
		SourceIp:        sql.NullString{String: sourceIp, Valid: sourceIp != ""},
		FirstSeenTime:   now,
		LastSeenTime:    now,
		CreateTimes:     now,
		UpdateTimes:     now,
	})
}

func parseFingerprintVisitorID(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	var out struct {
		VisitorID string `json:"visitorId"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", false
	}
	return out.VisitorID, out.VisitorID != ""
}

func buildFingerprintMatchKey(visitorId string) string {
	hash := sha256.Sum256([]byte("fingerprintjs:" + visitorId))
	return hex.EncodeToString(hash[:])
}

func (l *GuestLoginLogic) checkGuestLimit(ip string) error {
	key := fmt.Sprintf("guest:ip:%s:count", ip)

	// 自增次数
	count, err := l.svcCtx.Redis.Incr(key)
	if err != nil {
		return err
	}

	// 第一次自增，设置过期 1 小时
	if count == 1 {
		err := l.svcCtx.Redis.Expire(key, int(time.Hour.Seconds()))
		if err != nil {
			return err
		}
	}

	// 限制 2 次
	if count > 2 {
		return i18n.StatusError(l.ctx, i18n.RegistrationTooFrequentRetry)
	}

	return nil
}
