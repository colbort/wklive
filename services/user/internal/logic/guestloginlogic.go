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
	"wklive/proto/system"
	"wklive/proto/user"
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
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil {
		return nil, err
	}
	if tenant == nil || tenant.Data == nil {
		return &user.GuestLoginResp{
			Base: helper.GetErrResp(401, i18n.Translate(i18n.TenantNotFound, l.ctx)),
		}, nil
	}

	if in.DeviceId == "" && in.Fingerprint == "" {
		return &user.GuestLoginResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.PleaseSwitchDeviceToLogin, l.ctx)),
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
		matched, err = l.findMatchedGuestByFingerprint(tenant.Data.Id, in.Fingerprint, in.RegisterIp)
		if err != nil {
			return nil, err
		}
	}

	if matched != nil {
		now := utils.NowMillis()
		matched.LastLoginIp = sql.NullString{String: in.RegisterIp, Valid: in.RegisterIp != ""}
		matched.LastLoginTime = now
		matched.UpdateTimes = now
		if matched.DeviceId == "" {
			matched.DeviceId = fmt.Sprintf("%d", matched.Id)
		}
		_ = l.svcCtx.UserModel.Update(l.ctx, matched)
		if err := l.saveFingerprint(tenant.Data.Id, matched.Id, matched.DeviceId, in.Fingerprint, in.RegisterIp, now); err != nil {
			return nil, err
		}

		resp, err := l.buildGuestLoginResp(tenant.Data.Id, matched, false)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	err = l.checkGuestLimit(in.RegisterIp)
	if err != nil {
		return nil, err
	}

	userId := l.svcCtx.Node.Generate().Int64()
	deviceId := fmt.Sprintf("%d", userId)
	now := utils.NowMillis()
	guest := &models.TUser{
		Id:             userId,
		TenantId:       tenant.Data.Id,
		UserNo:         fmt.Sprintf("G%d", userId),
		Username:       fmt.Sprintf("Guest%d", userId),
		Nickname:       sql.NullString{String: fmt.Sprintf("Guest%d", userId), Valid: true},
		Avatar:         sql.NullString{},
		PasswordHash:   "",
		RegisterType:   int64(user.RegisterType_REGISTER_TYPE_GUEST),
		Status:         1,
		MemberLevel:    0,
		Language:       sql.NullString{},
		Timezone:       sql.NullString{},
		InviteCode:     sql.NullString{},
		Signature:      sql.NullString{},
		Source:         sql.NullString{String: "guest", Valid: true},
		ReferrerUserId: sql.NullInt64{},
		LastLoginIp:    sql.NullString{String: in.RegisterIp, Valid: in.RegisterIp != ""},
		LastLoginTime:  now,
		RegisterIp:     sql.NullString{String: in.RegisterIp, Valid: in.RegisterIp != ""},
		RegisterTime:   now,
		IsGuest:        2,
		IsRecharge:     0,
		DeviceId:       deviceId,
		Fingerprint:    sql.NullString{String: "", Valid: true},
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
	if err := l.saveFingerprint(tenant.Data.Id, guest.Id, guest.DeviceId, in.Fingerprint, in.RegisterIp, now); err != nil {
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
	str["tid"] = tenantId
	expand, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		guest.Id,
		guest.Username,
		string(expand),
		"",
		time.Duration(24*3600)*time.Second,
	)
	if err != nil {
		return nil, err
	}
	return &user.GuestLoginResp{
		Base: helper.OkResp(),
		Data: &user.GuestLogin{
			Token:    token,
			Uid:      guest.DeviceId,
			Username: guest.Username,
			IsNew:    isNew,
		},
	}, nil
}

func (l *GuestLoginLogic) findMatchedGuestByFingerprint(tenantId int64, fingerprint string, registerIp string) (*models.TUser, error) {
	current, ok := parseFingerprint(fingerprint, registerIp)
	if !ok {
		return nil, nil
	}

	matchKey := buildFingerprintMatchKey(current)

	var bestUserId int64
	bestScore := 0
	cursor := int64(0)
	limit := int64(500)
	for {
		// 查找得分最高的
		fingerprintCandidates, err := l.svcCtx.FingerprintModel.FindGuestFingerprintCandidates(l.ctx, tenantId, matchKey, cursor, limit)
		if err != nil {
			return nil, err
		}
		if len(fingerprintCandidates) == 0 {
			break
		}

		for _, candidate := range fingerprintCandidates {
			cursor = candidate.Id
			stored, ok := parseFingerprint(candidate.Fingerprint, candidate.SourceIp)
			if !ok {
				continue
			}
			score := scoreGuestFingerprint(current, stored)
			if score <= bestScore {
				continue
			}
			bestScore = score
			bestUserId = candidate.UserId
		}

		if int64(len(fingerprintCandidates)) < limit {
			break
		}
	}

	// 有 得分 高于 90，的 用户
	if bestScore >= 90 {
		best, err := l.svcCtx.UserModel.FindByTenantIdUserId(l.ctx, tenantId, bestUserId)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return best, nil
	}

	return nil, nil
}

func (l *GuestLoginLogic) saveFingerprint(tenantId int64, userId int64, deviceId string, fingerprint string, sourceIp string, now int64) error {
	if fingerprint == "" {
		return nil
	}
	parsed, ok := parseFingerprint(fingerprint, sourceIp)
	if !ok {
		return nil
	}
	matchKey := buildFingerprintMatchKey(parsed)

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

func parseFingerprint(raw string, ip string) (map[string]any, bool) {
	if raw == "" {
		return nil, false
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, false
	}
	out["ip"] = ip
	return out, true
}

func buildFingerprintMatchKey(fingerprint map[string]any) string {
	keys := []string{
		"platform",
		"browserName",
		"osName",
		"deviceType",
	}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		value, ok := fingerprint[key].(string)
		if ok && value != "" {
			values[key] = value
		}
	}
	if len(values) < 2 {
		return ""
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

func scoreGuestFingerprint(left, right map[string]any) int {
	score := 0
	score += scoreStringField(left, right, "userAgent", 14)
	score += scoreStringField(left, right, "platform", 7)
	score += scoreStringField(left, right, "timezone", 8)
	score += scoreStringField(left, right, "language", 3)
	score += scoreAnyField(left, right, "languages", 2)
	score += scoreStringField(left, right, "browserName", 5)
	score += scoreStringField(left, right, "browserMajorVersion", 3)
	score += scoreStringField(left, right, "osName", 10)
	score += scoreStringField(left, right, "deviceType", 6)
	score += scoreNumberField(left, right, "screenWidth", 4)
	score += scoreNumberField(left, right, "screenHeight", 4)
	score += scoreNumberField(left, right, "availWidth", 1)
	score += scoreNumberField(left, right, "availHeight", 1)
	score += scoreNumberField(left, right, "colorDepth", 2)
	score += scoreNumberField(left, right, "pixelRatio", 5)
	score += scoreNumberField(left, right, "hardwareConcurrency", 4)
	score += scoreNumberField(left, right, "deviceMemory", 3)
	score += scoreNumberField(left, right, "maxTouchPoints", 4)
	score += scoreBoolField(left, right, "cookieEnabled", 1)
	score += scoreBoolField(left, right, "localStorageSupported", 1)
	score += scoreBoolField(left, right, "sessionStorageSupported", 1)
	score += scoreBoolField(left, right, "indexedDBSupported", 1)
	score += scoreStringField(left, right, "ip", 10)
	return score
}

func scoreStringField(left, right map[string]any, field string, weight int) int {
	lv, lok := left[field].(string)
	rv, rok := right[field].(string)
	if !lok || !rok || lv == "" || rv == "" || lv != rv {
		return 0
	}
	return weight
}

func scoreNumberField(left, right map[string]any, field string, weight int) int {
	lv, lok := left[field].(float64)
	rv, rok := right[field].(float64)
	if !lok || !rok || lv != rv {
		return 0
	}
	return weight
}

func scoreBoolField(left, right map[string]any, field string, weight int) int {
	lv, lok := left[field].(bool)
	rv, rok := right[field].(bool)
	if !lok || !rok || lv != rv {
		return 0
	}
	return weight
}

func scoreAnyField(left, right map[string]any, field string, weight int) int {
	lv, lok := left[field]
	rv, rok := right[field]
	if !lok || !rok {
		return 0
	}
	lb, err := json.Marshal(lv)
	if err != nil {
		return 0
	}
	rb, err := json.Marshal(rv)
	if err != nil {
		return 0
	}
	if string(lb) != string(rb) {
		return 0
	}
	return weight
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
		return errors.New(i18n.Translate(i18n.RegistrationTooFrequentRetry, l.ctx))
	}

	return nil
}
