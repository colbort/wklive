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

type UpdateUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankLogic {
	return &UpdateUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员更新用户银行卡
func (l *UpdateUserBankLogic) UpdateUserBank(in *user.UpdateUserBankReq) (*user.UpdateUserBankResp, error) {
	// 获取银行卡信息
	bank, err := l.svcCtx.UserBankModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if bank == nil {
		return &user.UpdateUserBankResp{
			Base: helper.ErrResp(i18n.BankCardNotFound, i18n.Translate(i18n.BankCardNotFound, l.ctx)),
		}, nil
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, bank.TenantId, i18n.NoPermissionModifyThisBankCard); err != nil {
		return nil, err
	} else if base != nil {
		return &user.UpdateUserBankResp{
			Base: base,
		}, nil
	}

	// 验证银行卡是否属于该用户
	if bank.UserId != in.UserId {
		return &user.UpdateUserBankResp{
			Base: helper.ErrResp(i18n.NoPermissionModifyThisBankCard, i18n.Translate(i18n.NoPermissionModifyThisBankCard, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	isDefault := common.YesNo(in.IsDefault)
	if isDefault == common.YesNo_YES_NO_YES {
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
	if isDefault != common.YesNo_YES_NO_UNKNOWN {
		bank.IsDefault = int64(in.IsDefault)
	}
	if in.Enabled != 0 {
		bank.Enabled = int64(in.Enabled)
	}
	bank.UpdateTimes = now

	err = l.svcCtx.UserBankModel.Update(l.ctx, bank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 银行卡 %d 成功", in.UserId, in.Id)

	return &user.UpdateUserBankResp{
		Base: helper.OkResp(),
		Data: toUserBankItemProto(bank),
	}, nil
}
