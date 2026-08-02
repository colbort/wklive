package payment

import "wklive/admin-api/internal/types"

const unsupportedPaymentDeleteCode int32 = 100003

// unsupportedDelete prevents an HTTP route without a backing payment RPC from
// returning the zero-value response (code=0), which callers can mistake for a
// successful destructive operation.
func unsupportedDelete(resource string) *types.RespBase {
	return &types.RespBase{
		Code: unsupportedPaymentDeleteCode,
		Msg:  resource + " deletion is not supported by payment RPC",
	}
}
