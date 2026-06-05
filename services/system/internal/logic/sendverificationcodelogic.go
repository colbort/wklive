package logic

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	uc "wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送验证码
func (l *SendVerificationCodeLogic) SendVerificationCode(in *system.SendVerificationCodeReq) (*system.RespBase, error) {
	target := strings.TrimSpace(in.Email)
	if in.Channel == system.VerificationCodeChannel_VERIFICATION_CODE_CHANNEL_PHONE {
		target = strings.TrimSpace(in.Phone)
	}
	if target == "" || in.Channel == system.VerificationCodeChannel_VERIFICATION_CODE_CHANNEL_UNKNOWN {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx)),
		}, nil
	}
	if in.Scene == system.VerificationCodeScene_VERIFICATION_CODE_SCENE_UNKNOWN {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, l.ctx)),
		}, nil
	}

	code, err := verificationCode(6)
	if err != nil {
		return nil, err
	}

	provider, sendErr := l.send(in, target, code)
	status := system.VerificationCodeStatus_VERIFICATION_CODE_STATUS_SUCCESS
	errMsg := ""
	if sendErr != nil {
		status = system.VerificationCodeStatus_VERIFICATION_CODE_STATUS_FAILED
		errMsg = sendErr.Error()
	}

	if err = l.insertRecord(in.TenantId, in.Channel, target, in.Scene, code, status, provider, errMsg); err != nil {
		return nil, err
	}

	if sendErr != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InternalServerError, sendErr.Error()),
		}, nil
	}

	return &system.RespBase{Base: helper.OkResp()}, nil
}

func (l *SendVerificationCodeLogic) send(in *system.SendVerificationCodeReq, target, code string) (string, error) {
	switch in.Channel {
	case system.VerificationCodeChannel_VERIFICATION_CODE_CHANNEL_EMAIL:
		return l.sendEmail(in.TenantId, target, verificationSceneTemplateValue(in.Scene), code)
	case system.VerificationCodeChannel_VERIFICATION_CODE_CHANNEL_PHONE:
		return l.sendPhone(in.TenantId, target, verificationSceneTemplateValue(in.Scene), code)
	default:
		return "", fmt.Errorf("unsupported verification code channel")
	}
}

func (l *SendVerificationCodeLogic) sendEmail(tenantId int64, email, scene, code string) (string, error) {
	config, err := l.emailConfig(tenantId)
	if err != nil {
		return "smtp", err
	}
	if !config.Enabled {
		return "smtp", fmt.Errorf("email verification code is disabled")
	}
	if config.SmtpHost == "" || config.SmtpPort == 0 || config.FromEmail == "" {
		return "smtp", fmt.Errorf("email config is incomplete")
	}

	subject := renderVerificationTemplate(config.SubjectTemplate, "", code, scene)
	if subject == "" {
		subject = "Verification code"
	}
	body := renderVerificationTemplate(config.BodyTemplate, "", code, scene)
	if body == "" {
		body = fmt.Sprintf("Your verification code is %s", code)
	}

	from := config.FromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + email,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="UTF-8"`,
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort)
	var auth smtp.Auth
	if config.Username != "" || config.Password != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)
	}
	return "smtp", smtp.SendMail(addr, auth, config.FromEmail, []string{email}, []byte(msg))
}

func (l *SendVerificationCodeLogic) sendPhone(tenantId int64, phone, scene, code string) (string, error) {
	config, err := l.phoneConfig(tenantId)
	if err != nil {
		return "", err
	}
	if !config.Enabled {
		return config.Provider, fmt.Errorf("phone verification code is disabled")
	}
	if config.Endpoint == "" {
		return config.Provider, fmt.Errorf("phone config endpoint is empty")
	}

	method := strings.ToUpper(strings.TrimSpace(config.Method))
	if method == "" {
		method = http.MethodPost
	}
	body := renderVerificationTemplate(config.BodyTemplate, phone, code, scene)
	if body == "" {
		payload := map[string]string{"phone": phone, "code": code, "scene": scene}
		b, _ := json.Marshal(payload)
		body = string(b)
	}

	req, err := http.NewRequestWithContext(l.ctx, method, config.Endpoint, bytes.NewBufferString(body))
	if err != nil {
		return config.Provider, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range parseHeaders(config.HeadersJson) {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return config.Provider, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return config.Provider, fmt.Errorf("phone provider returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return config.Provider, nil
}

func (l *SendVerificationCodeLogic) emailConfig(tenantId int64) (*system.EmailConfig, error) {
	value, err := l.configValue(tenantId, system.SysConfigType_EMAIL_CONFIG)
	if err != nil {
		return nil, err
	}
	var config system.EmailConfig
	if err = json.Unmarshal([]byte(value), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (l *SendVerificationCodeLogic) phoneConfig(tenantId int64) (*system.PhoneConfig, error) {
	value, err := l.configValue(tenantId, system.SysConfigType_PHONE_CONFIG)
	if err != nil {
		return nil, err
	}
	var config system.PhoneConfig
	if err = json.Unmarshal([]byte(value), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (l *SendVerificationCodeLogic) configValue(tenantId int64, key system.SysConfigType) (string, error) {
	config, err := l.svcCtx.ConfigModel.FindOneByTenantIdConfigKey(l.ctx, tenantId, sql.NullString{String: key.String(), Valid: true})
	if err == models.ErrNotFound && tenantId != 0 {
		config, err = l.svcCtx.ConfigModel.FindOneByTenantIdConfigKey(l.ctx, 0, sql.NullString{String: key.String(), Valid: true})
	}
	if err != nil {
		if err == models.ErrNotFound {
			return "", fmt.Errorf("%s is not configured", key.String())
		}
		return "", err
	}
	return config.ConfigValue.String, nil
}

func (l *SendVerificationCodeLogic) insertRecord(
	tenantId int64,
	channel system.VerificationCodeChannel,
	target string,
	scene system.VerificationCodeScene,
	code string,
	status system.VerificationCodeStatus,
	provider string,
	errMsg string,
) error {
	now := uc.NowMillis()
	_, err := l.svcCtx.VerificationCodeRecordModel.Insert(l.ctx, &models.SysVerificationCodeRecord{
		TenantId:     tenantId,
		Channel:      int64(channel),
		Target:       target,
		Scene:        int64(scene),
		Code:         code,
		Status:       int64(status),
		Provider:     sql.NullString{String: provider, Valid: provider != ""},
		ErrorMessage: sql.NullString{String: errMsg, Valid: errMsg != ""},
		CreateTimes:  now,
		UpdateTimes:  now,
	})
	return err
}

func verificationCode(length int) (string, error) {
	var builder strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + n.Int64()))
	}
	return builder.String(), nil
}

func renderVerificationTemplate(template, phone, code, scene string) string {
	replacer := strings.NewReplacer(
		"{{phone}}", phone,
		"{{code}}", code,
		"{{scene}}", scene,
	)
	return replacer.Replace(template)
}

func verificationSceneTemplateValue(scene system.VerificationCodeScene) string {
	switch scene {
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_REGISTER:
		return "register"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_LOGIN:
		return "login"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_RESET_PASSWORD:
		return "reset_password"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_BIND_EMAIL:
		return "bind_email"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_BIND_PHONE:
		return "bind_phone"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_CHANGE_PASSWORD:
		return "change_password"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_WITHDRAW:
		return "withdraw"
	case system.VerificationCodeScene_VERIFICATION_CODE_SCENE_TEST:
		return "test"
	default:
		return "unknown"
	}
}

func parseHeaders(headersJSON string) map[string]string {
	headers := make(map[string]string)
	if strings.TrimSpace(headersJSON) == "" {
		return headers
	}
	_ = json.Unmarshal([]byte(headersJSON), &headers)
	return headers
}
