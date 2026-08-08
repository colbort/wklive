package middleware

import (
	"net/http"
	"regexp"
)

// auditRouteSpec 是敏感操作审计的唯一入口。只有在这里显式登记的路由
// 才写入 sys_op_log；新增普通接口不会被自动记录。
type auditRouteSpec struct {
	method  string
	pattern *regexp.Regexp
}

var sensitiveAuditRoutes = buildSensitiveAuditRoutes([][2]string{
	// 系统管理员、权限、配置、任务和租户。
	{http.MethodPost, "/admin/system/users/status"},
	{http.MethodPost, "/admin/system/users/resetPwd"},
	{http.MethodPost, "/admin/system/users/assignRoles"},
	{http.MethodPost, "/admin/system/users/google2fa/disable"},
	{http.MethodPost, "/admin/system/users/google2fa/reset"},
	{http.MethodDelete, "/admin/system/users/:id"},
	{http.MethodPost, "/admin/system/roles/grant"},
	{http.MethodPost, "/admin/system/menus"},
	{http.MethodPut, "/admin/system/menus"},
	{http.MethodDelete, "/admin/system/menus/:id"},
	{http.MethodPost, "/admin/system/configs"},
	{http.MethodPut, "/admin/system/configs"},
	{http.MethodDelete, "/admin/system/configs/:id"},
	{http.MethodPost, "/admin/system/jobs/:id/run"},
	{http.MethodPost, "/admin/system/jobs/:id/start"},
	{http.MethodPost, "/admin/system/jobs/:id/stop"},
	{http.MethodPost, "/admin/system/tenants"},
	{http.MethodPut, "/admin/system/tenants"},
	{http.MethodDelete, "/admin/system/tenants/:id"},
	{http.MethodPost, "/admin/system/tenant-domains"},
	{http.MethodPut, "/admin/system/tenant-domains"},
	{http.MethodDelete, "/admin/system/tenant-domains/:id"},

	// 会员身份、安全、风控和银行卡。
	{http.MethodPut, "/admin/member/users/:userId/status"},
	{http.MethodPut, "/admin/member/users/:userId/level"},
	{http.MethodPut, "/admin/member/users/:userId/reset-login-password"},
	{http.MethodPut, "/admin/member/users/:userId/reset-pay-password"},
	{http.MethodPut, "/admin/member/users/:userId/unlock"},
	{http.MethodPut, "/admin/member/users/:userId/risk-level"},
	{http.MethodPut, "/admin/member/users/:userId/reset2fa"},
	{http.MethodDelete, "/admin/member/users/:userId"},
	{http.MethodPut, "/admin/member/user-identities/:userId/review"},
	{http.MethodPost, "/admin/member/user-banks"},
	{http.MethodPut, "/admin/member/user-banks/:id"},
	{http.MethodDelete, "/admin/member/user-banks/:id"},
	{http.MethodPut, "/admin/member/user-banks/:id/status"},

	// 资产人工变更和平台资金。
	{http.MethodPost, "/admin/asset/add"},
	{http.MethodPost, "/admin/asset/sub"},
	{http.MethodPost, "/admin/asset/freeze"},
	{http.MethodPost, "/admin/asset/unfreeze"},
	{http.MethodPost, "/admin/asset/lock"},
	{http.MethodPost, "/admin/asset/unlock"},
	{http.MethodPost, "/admin/asset/platform-accounts"},
	{http.MethodPost, "/admin/asset/platform-accounts/adjust"},
	{http.MethodPost, "/admin/asset/platform-backstop-policies"},
	{http.MethodPost, "/admin/asset/platform-backstop-policies/:policyId/review"},

	// 支付订单人工干预和钱包配置。
	{http.MethodPost, "/admin/payment/recharge-order/:orderNo/close"},
	{http.MethodPost, "/admin/payment/recharge-order/:orderNo/manual-success"},
	{http.MethodPost, "/admin/payment/recharge-order/:orderNo/retry-notify"},
	{http.MethodPost, "/admin/payment/withdraw-order/:orderNo/audit"},
	{http.MethodPost, "/admin/payment/crypto-wallet-account"},
	{http.MethodPut, "/admin/payment/crypto-wallet-account"},

	// Trade 风控、人工重试和保险基金。
	{http.MethodPost, "/admin/trade/user-trade-limit"},
	{http.MethodPost, "/admin/trade/user-symbol-limit"},
	{http.MethodPost, "/admin/trade/user-trade-controls/disable"},
	{http.MethodPost, "/admin/trade/user-trade-config"},
	{http.MethodPost, "/admin/trade/contract-user-config"},
	{http.MethodPost, "/admin/trade/user-leverage-config"},
	{http.MethodPost, "/admin/trade/events/retry"},
	{http.MethodPost, "/admin/trade/risk-tiers"},
	{http.MethodPost, "/admin/trade/account-liquidations/retry"},
	{http.MethodPost, "/admin/trade/operations/settlement-instructions/retry"},
	{http.MethodPost, "/admin/trade/operations/reconciliation-issues/ignore"},
	{http.MethodPost, "/admin/trade/insurance-fund/accounts"},

	// Option 交易、结算、风控及公司行动。
	{http.MethodPost, "/admin/option/combo-orders/force-cancel"},
	{http.MethodPost, "/admin/option/exercises/retry"},
	{http.MethodPost, "/admin/option/contracts/force-cancel-orders"},
	{http.MethodPost, "/admin/option/trading-controls/release-kill-switch"},
	{http.MethodPost, "/admin/option/mmp/config"},
	{http.MethodPost, "/admin/option/mmp/reset"},
	{http.MethodPost, "/admin/option/settlement-prices/corrections"},
	{http.MethodPost, "/admin/option/settlement-prices/review"},
	{http.MethodPost, "/admin/option/trade-corrections"},
	{http.MethodPost, "/admin/option/trade-corrections/review"},
	{http.MethodPost, "/admin/option/recovery/asset-instructions/retry"},
	{http.MethodPost, "/admin/option/trading-calendars/review"},
	{http.MethodPost, "/admin/option/trading-halts"},
	{http.MethodPost, "/admin/option/trading-halts/resume"},
	{http.MethodPost, "/admin/option/corporate-actions"},
	{http.MethodPost, "/admin/option/corporate-actions/review"},

	// Staking 产品状态、人工派息/赎回和失败重试。
	{http.MethodPost, "/admin/staking/products/status"},
	{http.MethodPost, "/admin/staking/manual-reward"},
	{http.MethodPost, "/admin/staking/manual-redeem"},
	{http.MethodPost, "/admin/staking/operations/retry"},

	// Market 权威行情与快照人工干预。
	{http.MethodPost, "/admin/market/authorities"},
	{http.MethodPut, "/admin/market/price-formulas/:id/status"},
	{http.MethodPost, "/admin/market/snapshot-outbox/:id/retry"},
	{http.MethodPost, "/admin/market/authoritative-snapshots/revoke"},
})

func buildSensitiveAuditRoutes(routes [][2]string) []auditRouteSpec {
	result := make([]auditRouteSpec, 0, len(routes))
	for _, route := range routes {
		pattern, _, err := compilePathPattern(route[1])
		if err != nil {
			panic("invalid sensitive audit route: " + route[1])
		}
		result = append(result, auditRouteSpec{method: route[0], pattern: pattern})
	}
	return result
}
