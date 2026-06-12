package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserBankLogic {
	return &AddUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员添加用户银行卡
func (l *AddUserBankLogic) AddUserBank(in *user.AddUserBankReq) (*user.AddUserBankResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AddUserBankResp{
			Base: helper.GetErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	isDefault := int64(in.IsDefault)
	if common.YesNo(in.IsDefault) == common.YesNo_YES_NO_UNKNOWN {
		isDefault = int64(common.YesNo_YES_NO_NO)
	}
	enabled := int64(in.Enabled)
	if common.Enable(in.Enabled) == common.Enable_ENABLE_UNKNOWN {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}

	// 如果设置为默认，需要取消其他卡的默认设置
	if common.YesNo(isDefault) == common.YesNo_YES_NO_YES {
		// TODO: 更新其他卡的默认状态
	}

	// 创建银行卡
	bank := &models.TUserBank{
		Id:          l.svcCtx.Node.Generate().Int64(),
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		BankName:    in.BankName,
		BankCode:    sql.NullString{String: in.BankCode, Valid: in.BankCode != ""},
		AccountName: in.AccountName,
		AccountNo:   in.AccountNo,
		BranchName:  sql.NullString{String: in.BranchName, Valid: in.BranchName != ""},
		CountryCode: sql.NullString{String: in.CountryCode, Valid: in.CountryCode != ""},
		IsDefault:   isDefault,
		Enabled:     enabled,
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err = l.svcCtx.UserBankModel.Insert(l.ctx, bank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员为用户 %d 添加银行卡成功，卡号后4位：%s", in.UserId, getLastFourDigits(in.AccountNo))

	return &user.AddUserBankResp{
		Base: helper.OkResp(),
		Data: toUserBankItemProto(bank),
	}, nil
}
