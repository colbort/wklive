package i18n

// Language 定义支持的语言类型
type Language string

const (
	EN Language = "en" // English
	ZH Language = "zh" // Chinese
)

// 错误/消息代码
const (
	OK                     = 200
	ParamError             = 400
	Unauthorized           = 401
	Forbidden              = 403
	NotFound               = 404
	ConflictError          = 409
	InternalServerError    = 500
	ServiceUnavailable     = 503
	OperationNotAllowed    = 1001
	ResourceAlreadyExists  = 1002
	InvalidRequest         = 1003
	DataValidationError    = 1004
	AuthenticationRequired = 1005
	UserNotFound           = 1006
	InvalidCredentials     = 1007
	TokenExpired           = 1008
	PermissionDenied       = 1009
)

// MessageMap 定义所有支持的错误消息翻译
var MessageMap = map[int32]map[Language]string{
	OK: {
		EN: "OK",
		ZH: "成功",
	},
	ParamError: {
		EN: "Parameter error",
		ZH: "参数错误",
	},
	Unauthorized: {
		EN: "Unauthorized",
		ZH: "未授权",
	},
	Forbidden: {
		EN: "Forbidden",
		ZH: "禁止访问",
	},
	NotFound: {
		EN: "Not found",
		ZH: "资源不存在",
	},
	ConflictError: {
		EN: "Conflict",
		ZH: "冲突",
	},
	InternalServerError: {
		EN: "Internal server error",
		ZH: "服务器内部错误",
	},
	ServiceUnavailable: {
		EN: "Service unavailable",
		ZH: "服务不可用",
	},
	OperationNotAllowed: {
		EN: "Operation not allowed",
		ZH: "操作不被允许",
	},
	ResourceAlreadyExists: {
		EN: "Resource already exists",
		ZH: "资源已存在",
	},
	InvalidRequest: {
		EN: "Invalid request",
		ZH: "无效请求",
	},
	DataValidationError: {
		EN: "Data validation error",
		ZH: "数据验证错误",
	},
	AuthenticationRequired: {
		EN: "Authentication required",
		ZH: "需要验证",
	},
	UserNotFound: {
		EN: "User not found",
		ZH: "用户不存在",
	},
	InvalidCredentials: {
		EN: "Invalid credentials",
		ZH: "凭证无效",
	},
	TokenExpired: {
		EN: "Token expired",
		ZH: "令牌已过期",
	},
	PermissionDenied: {
		EN: "Permission denied",
		ZH: "权限被拒绝",
	},
}
