package alipay

type PaymentReq struct {
	OutTradeNo  string `json:"out_trade_no"`
	TotalAmount string `json:"total_amount"`
	Subject     string `json:"subject"`
	NotifyURL   string `json:"notify_url"`
}

type PaymentResp struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	TradeNo string `json:"trade_no"`
	PayURL  string `json:"pay_url"`
	QRCode  string `json:"qr_code"`
}

type PaymentQueryReq struct {
	OutTradeNo string `json:"out_trade_no"`
	TradeNo    string `json:"trade_no,omitempty"`
}

type PaymentQueryResp struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	TradeNo     string `json:"trade_no"`
	TradeStatus string `json:"trade_status"`
	TotalAmount string `json:"total_amount"`
}

type RefundReq struct {
	OutTradeNo   string `json:"out_trade_no"`
	TradeNo      string `json:"trade_no,omitempty"`
	OutRequestNo string `json:"out_request_no"`
	RefundAmount string `json:"refund_amount"`
	RefundReason string `json:"refund_reason,omitempty"`
}

type RefundResp struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	TradeNo    string `json:"trade_no"`
	OutTradeNo string `json:"out_trade_no"`
	RefundFee  string `json:"refund_fee"`
}

type PayoutReq struct {
	OutBizNo      string `json:"out_biz_no"`
	PayeeIdentity string `json:"payee_identity"`
	Amount        string `json:"amount"`
	Remark        string `json:"remark,omitempty"`
}

type PayoutResp struct {
	Code           string `json:"code"`
	Msg            string `json:"msg"`
	OrderID        string `json:"order_id"`
	PayFundOrderID string `json:"pay_fund_order_id"`
	Status         string `json:"status"`
}

type PayoutQueryReq struct {
	OutBizNo string `json:"out_biz_no"`
	OrderID  string `json:"order_id,omitempty"`
}

type PayoutQueryResp struct {
	Code           string `json:"code"`
	Msg            string `json:"msg"`
	OrderID        string `json:"order_id"`
	PayFundOrderID string `json:"pay_fund_order_id"`
	Status         string `json:"status"`
}

type PaymentNotifyReq struct {
	NotifyID    string `json:"notify_id"`
	OutTradeNo  string `json:"out_trade_no"`
	TradeNo     string `json:"trade_no"`
	TradeStatus string `json:"trade_status"`
	TotalAmount string `json:"total_amount"`
	Sign        string `json:"sign"`
	SignType    string `json:"sign_type"`
}

type PayoutNotifyReq struct {
	NotifyID string `json:"notify_id"`
	OutBizNo string `json:"out_biz_no"`
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
	Sign     string `json:"sign"`
	SignType string `json:"sign_type"`
}
