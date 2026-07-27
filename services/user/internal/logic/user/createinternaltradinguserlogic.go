package userlogic

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type CreateInternalTradingUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateInternalTradingUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateInternalTradingUserLogic {
	return &CreateInternalTradingUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateInternalTradingUserLogic) CreateInternalTradingUser(in *user.CreateInternalTradingUserReq) (*user.InternalTradingUserResp, error) {
	if in == nil || in.TenantId <= 0 || strings.TrimSpace(in.AccountKey) == "" {
		return nil, fmt.Errorf("tenant_id and account_key are required")
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(in.AccountKey)))
	username := fmt.Sprintf("mm_%d_%x", in.TenantId, sum[:8])
	existing, err := l.svcCtx.UserModel.FindOneByTenantIdUsername(l.ctx, in.TenantId, username)
	if err == nil {
		if existing.AccountType != int64(common.UserAccountType_USER_ACCOUNT_TYPE_INTERNAL_MARKET_MAKER) {
			return nil, fmt.Errorf("account key is occupied by a non-market-maker user")
		}
		return internalTradingUserResp(existing), nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	randomPassword := fmt.Sprintf("%d:%s", l.svcCtx.Node.Generate().Int64(), username)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	row := &models.TUser{
		TenantId: in.TenantId, UserNo: fmt.Sprintf("MM%d", l.svcCtx.Node.Generate().Int64()),
		Username:     username,
		Nickname:     sql.NullString{String: strings.TrimSpace(in.Nickname), Valid: strings.TrimSpace(in.Nickname) != ""},
		PasswordHash: string(passwordHash),
		RegisterType: int64(user.RegisterType_REGISTER_TYPE_USERNAME),
		AccountType:  int64(common.UserAccountType_USER_ACCOUNT_TYPE_INTERNAL_MARKET_MAKER),
		Status:       int64(user.UserStatus_USER_STATUS_NORMAL),
		Source:       sql.NullString{String: strings.TrimSpace(in.Source), Valid: strings.TrimSpace(in.Source) != ""},
		RegisterTime: now, IsGuest: int64(common.YesNo_YES_NO_NO),
		IsRecharge:  int64(common.YesNo_YES_NO_NO),
		Remark:      sql.NullString{String: strings.TrimSpace(in.Remark), Valid: strings.TrimSpace(in.Remark) != ""},
		CreateTimes: now, UpdateTimes: now,
	}
	result, err := l.svcCtx.UserModel.Insert(l.ctx, row)
	if err != nil {
		if existing, findErr := l.svcCtx.UserModel.FindOneByTenantIdUsername(l.ctx, in.TenantId, username); findErr == nil {
			return internalTradingUserResp(existing), nil
		}
		return nil, err
	}
	row.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return internalTradingUserResp(row), nil
}

func internalTradingUserResp(row *models.TUser) *user.InternalTradingUserResp {
	return &user.InternalTradingUserResp{
		Base: helper.OkResp(),
		Data: &user.InternalTradingUser{
			UserId: row.Id, TenantId: row.TenantId, Username: row.Username,
			AccountType: common.UserAccountType(row.AccountType), Status: user.UserStatus(row.Status),
		},
	}
}
