package types

type ChatAdminWSMessagesReq struct {
	UserId     int64
	MerchantId int64
	AgentId    int64
	SessionNo  string
}
