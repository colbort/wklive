package utils

const (
	CtxKeyUid        string = "userId"
	CtxKeyUsername   string = "username"
	CtxKeyTenantId   string = "x-tenant-id"
	CtxKeyTenantCode string = "x-tenant-code"
	CtxKeyUserType   string = "x-user-type"
	CtxKeyClientIp   string = "x-client-ip"
)

const (
	SysUserTypeSystemAdmin int64 = 1
	SysUserTypeTenantOwner int64 = 2
	SysUserTypeTenantAdmin int64 = 3
)
