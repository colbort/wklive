package wechat

type PaymentReq struct {
	AppID       string    `json:"appid"`
	MerchantID  string    `json:"mchid"`
	Description string    `json:"description"`
	OutTradeNo  string    `json:"out_trade_no"`
	NotifyURL   string    `json:"notify_url"`
	Amount      AmountReq `json:"amount"`
}

type AmountReq struct {
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

type PaymentResp struct {
	PrepayID string `json:"prepay_id"`
	CodeURL  string `json:"code_url"`
	H5URL    string `json:"h5_url"`
}

type PaymentQueryResp struct {
	AppID         string    `json:"appid"`
	MerchantID    string    `json:"mchid"`
	OutTradeNo    string    `json:"out_trade_no"`
	TransactionID string    `json:"transaction_id"`
	TradeState    string    `json:"trade_state"`
	Amount        AmountReq `json:"amount"`
	SuccessTime   string    `json:"success_time"`
}

type RefundReq struct {
	OutTradeNo    string          `json:"out_trade_no"`
	TransactionID string          `json:"transaction_id,omitempty"`
	OutRefundNo   string          `json:"out_refund_no"`
	Reason        string          `json:"reason,omitempty"`
	NotifyURL     string          `json:"notify_url,omitempty"`
	Amount        RefundAmountReq `json:"amount"`
}

type RefundAmountReq struct {
	Refund   int64  `json:"refund"`
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

type RefundResp struct {
	RefundID    string `json:"refund_id"`
	OutRefundNo string `json:"out_refund_no"`
	Status      string `json:"status"`
	SuccessTime string `json:"success_time"`
}

type PayoutReq struct {
	AppID       string `json:"appid"`
	OutBatchNo  string `json:"out_batch_no"`
	BatchName   string `json:"batch_name"`
	BatchRemark string `json:"batch_remark"`
	TotalAmount int64  `json:"total_amount"`
	TotalNum    int64  `json:"total_num"`
}

type PayoutResp struct {
	BatchID    string `json:"batch_id"`
	CreateTime string `json:"create_time"`
}

type PayoutQueryResp struct {
	BatchID       string `json:"batch_id"`
	OutBatchNo    string `json:"out_batch_no"`
	BatchStatus   string `json:"batch_status"`
	SuccessAmount int64  `json:"success_amount"`
	SuccessNum    int64  `json:"success_num"`
}

type NotifyResource struct {
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

type NotifyReq struct {
	ID           string         `json:"id"`
	EventType    string         `json:"event_type"`
	ResourceType string         `json:"resource_type"`
	Resource     NotifyResource `json:"resource"`
	Summary      string         `json:"summary"`
}

type PaymentNotifyReq struct {
	AppID         string    `json:"appid"`
	MerchantID    string    `json:"mchid"`
	OutTradeNo    string    `json:"out_trade_no"`
	TransactionID string    `json:"transaction_id"`
	TradeState    string    `json:"trade_state"`
	Amount        AmountReq `json:"amount"`
	SuccessTime   string    `json:"success_time"`
}

type PayoutNotifyReq struct {
	OutBatchNo    string `json:"out_batch_no"`
	BatchID       string `json:"batch_id"`
	BatchStatus   string `json:"batch_status"`
	SuccessAmount int64  `json:"success_amount"`
}
