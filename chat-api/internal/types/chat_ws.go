package types

type ChatWSMessagesReq struct {
	SessionNo  string
	Token      string
	MerchantId int64
	UserId     int64
	Nickname   string
	AvatarUrl  string
	IsGuest    bool
}
