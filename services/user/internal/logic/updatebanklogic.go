package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBankLogic {
	return &UpdateBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新银行卡
func (l *UpdateBankLogic) UpdateBank(in *user.UpdateBankReq) (*user.UpdateBankResp, error) {
	// 获取银行卡信息
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if bank == nil {
		return &user.UpdateBankResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}

	// 验证银行卡是否属于该用户
	if bank.UserId != in.UserId {
		return &user.UpdateBankResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionModifyThisBankCard, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	isDefault := int64(0)
	if in.IsDefault {
		isDefault = 1

		// 如果设置为默认，需要取消其他卡的默认设置
		// TODO: 更新其他卡的默认状态
	}

	// 更新银行卡信息
	if in.BankName != "" {
		bank.BankName = in.BankName
	}
	if in.BankCode != "" {
		bank.BankCode = sql.NullString{String: in.BankCode, Valid: true}
	}
	if in.AccountName != "" {
		bank.AccountName = in.AccountName
	}
	if in.AccountNo != "" {
		bank.AccountNo = in.AccountNo
	}
	if in.BranchName != "" {
		bank.BranchName = sql.NullString{String: in.BranchName, Valid: true}
	}
	if in.CountryCode != "" {
		bank.CountryCode = sql.NullString{String: in.CountryCode, Valid: true}
	}
	if in.IsDefault {
		bank.IsDefault = isDefault
	}
	bank.UpdateTimes = now

	err = l.svcCtx.UserBankModel.Update(l.ctx, bank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 更新银行卡 %d 成功", in.UserId, in.Id)

	bankProto := &user.UserBank{
		Id:          bank.Id,
		TenantId:    bank.TenantId,
		UserId:      bank.UserId,
		BankName:    bank.BankName,
		BankCode:    bank.BankCode.String,
		AccountName: bank.AccountName,
		AccountNo:   bank.AccountNo,
		BranchName:  bank.BranchName.String,
		CountryCode: bank.CountryCode.String,
		IsDefault:   isDefault == 1,
		Status:      user.BankStatus(bank.Status),
		CreateTimes: bank.CreateTimes,
		UpdateTimes: bank.UpdateTimes,
	}

	return &user.UpdateBankResp{
		Base: helper.OkResp(),
		Bank: bankProto,
	}, nil
}
