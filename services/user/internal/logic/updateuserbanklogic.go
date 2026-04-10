package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
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
			Base: &common.RespBase{
				Code: 404,
				Msg:  "银行卡不存在",
			},
		}, nil
	}

	// 验证银行卡是否属于该用户
	if bank.UserId != in.UserId {
		return &user.UpdateUserBankResp{
			Base: &common.RespBase{
				Code: 403,
				Msg:  "无权修改此银行卡",
			},
		}, nil
	}

	now := time.Now().UnixMilli()
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
	if in.Status != 0 {
		bank.Status = int64(in.Status)
	}
	bank.UpdateTimes = now

	err = l.svcCtx.UserBankModel.Update(l.ctx, bank)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 银行卡 %d 成功", in.UserId, in.Id)

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

	return &user.UpdateUserBankResp{
		Base: helper.OkResp(),
		Bank: bankProto,
	}, nil
}
