package user

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOptionsLogic {
	return &GetUserOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserOptionsLogic) GetUserOptions() (resp *types.GetUserOptionsResp, err error) {
	return &types.GetUserOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("registerType", "注册方式", user.RegisterType_REGISTER_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("loginType", "登录方式", user.LoginType_LOGIN_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("userStatus", "用户状态", user.UserStatus_USER_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("riskLevel", "风控等级", user.RiskLevel_RISK_LEVEL_NORMAL.Descriptor()),
			logicutil.EnumGroup("gender", "性别", user.Gender_GENDER_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("idType", "证件类型", user.IdType_ID_TYPE_NONE.Descriptor()),
			logicutil.EnumGroup("kycLevel", "KYC等级", user.KycLevel_KYC_LEVEL_NONE.Descriptor()),
			logicutil.EnumGroup("verifyStatus", "实名状态", user.VerifyStatus_VERIFY_STATUS_NONE.Descriptor()),
			logicutil.EnumGroup("bankStatus", "银行卡状态", user.BankStatus_BANK_STATUS_UNKNOWN.Descriptor()),
		},
	}, nil
}
