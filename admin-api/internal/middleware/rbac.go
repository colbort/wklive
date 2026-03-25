package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type RbacMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewRbacMiddleware(svcCtx *svc.ServiceContext) *RbacMiddleware {
	return &RbacMiddleware{
		svcCtx: svcCtx,
	}
}

func (m *RbacMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 放行登录接口
		if isWhitePath(r.URL.Path) {
			next(w, r)
			return
		}

		// 2. 获取并解析 JWT
		uid, err := utils.GetUidFromCtx(r.Context())
		if err != nil {
			logx.Errorf("invalid token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 3. 查询当前用户权限列表
		resp, err := m.svcCtx.SystemCli.LoginUserPerms(r.Context(), &system.LoginUserPermsReq{
			UserId: uid,
		})
		if err != nil {
			logx.Errorf("get profile failed, uid=%d err=%v", uid, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 4. 根据 path + method 映射出需要的权限
		requiredPerm := getRequiredPermission(r.URL.Path, r.Method)

		// 没配权限映射的接口，默认放过
		// 如果你想改成“没配置就拒绝”，可以改这里
		if requiredPerm == "" {
			next(w, r)
			return
		}

		// 5. 创建当前请求专属的 Casbin Enforcer
		enforcer, err := newUserPermEnforcer(fmt.Sprintf("%d", uid), resp.Perms)
		if err != nil {
			logx.Errorf("create casbin enforcer failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 6. 解析 requiredPerm
		obj, act, ok := parsePerm(requiredPerm)
		if !ok {
			logx.Errorf("invalid required permission format: %s", requiredPerm)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 7. 权限校验
		allowed, err := enforcer.Enforce(fmt.Sprintf("%d", uid), obj, act)
		if err != nil {
			logx.Errorf("casbin enforce failed, uid=%d perm=%s err=%v", uid, requiredPerm, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			logx.Errorf("forbidden, uid=%d method=%s path=%s required=%s userPerms=%v",
				uid, r.Method, r.URL.Path, requiredPerm, resp.Perms)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// newUserPermEnforcer 创建“用户直接权限”模式的 Enforcer
func newUserPermEnforcer(userID string, perms []string) (*casbin.Enforcer, error) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

	m, err := casbinModel.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, err
	}

	// 把当前用户最终权限直接挂到 userID 上
	for _, perm := range perms {
		obj, act, ok := parsePerm(perm)
		if !ok {
			logx.Errorf("skip invalid perm: %s", perm)
			continue
		}

		_, err = enforcer.AddPolicy(userID, obj, act)
		if err != nil {
			return nil, err
		}
	}

	return enforcer, nil
}

func parsePerm(perm string) (obj string, act string, ok bool) {
	perm = strings.TrimSpace(perm)
	if perm == "" {
		return "", "", false
	}

	parts := strings.Split(perm, ":")
	if len(parts) < 2 {
		return "", "", false
	}

	act = strings.TrimSpace(parts[len(parts)-1])
	obj = strings.TrimSpace(strings.Join(parts[:len(parts)-1], ":"))

	if obj == "" || act == "" {
		return "", "", false
	}

	return obj, act, true
}

// isWhitePath 白名单接口
func isWhitePath(path string) bool {
	whiteList := map[string]struct{}{
		"/admin/system/core":  {},
		"/admin/auth/login":   {},
		"/admin/auth/captcha": {},
		"/health":             {},
	}

	_, ok := whiteList[path]
	return ok
}

// getRequiredPermission 根据 path + method 映射权限
func getRequiredPermission(path, method string) string {
	// 去掉 query 参数影响（通常 URL.Path 本身不带 query，这里只是保险）
	path = strings.TrimSpace(path)

	// 去掉 /admin 前缀
	if strings.HasPrefix(path, "/admin/") {
		path = strings.TrimPrefix(path, "/admin")
	} else if path == "/admin" {
		path = "/"
	}

	// 标准化 path，移除末尾 /
	path = normalizePath(path)

	// 某些带 id 的路径可做归一化
	path = normalizeDynamicPath(path)

	permMap := map[string]string{
		// auth
		"GET /system/core":   "",
		"GET /auth/profile":  "",
		"POST /auth/profile": "", // 修改个人资料接口，暂不区分修改和查看权限
		// 白名单接口
		"GET /users":      "",
		"GET /roles":      "",
		"GET /menus":      "",
		"GET /logs/login": "",
		"GET /logs/op":    "",
		"GET /configs":    "",
		"GET /jobs":       "",
		"GET /jobs/logs":  "",

		// users
		"POST /users":                   "sys:user:add",
		"PUT /users":                    "sys:user:update",
		"DELETE /users":                 "sys:user:delete",
		"POST /users/status":            "sys:user:update",
		"POST /users/resetPwd":          "sys:user:resetPwd",
		"POST /users/assignRoles":       "sys:user:assignRoles",
		"POST /users/google2fa/init":    "sys:user:2fa:init",
		"POST /users/google2fa/bind":    "sys:user:2fa:bind",
		"POST /users/google2fa/reset":   "sys:user:2fa:reset",
		"POST /users/google2fa/enable":  "sys:user:2fa:enable",
		"POST /users/google2fa/disable": "sys:user:2fa:disable",

		// roles
		"POST /roles":       "sys:role:add",
		"PUT /roles":        "sys:role:update",
		"DELETE /roles":     "sys:role:delete",
		"POST /roles/grant": "sys:role:grant",

		// menus
		"POST /menus":   "sys:menu:add",
		"PUT /menus":    "sys:menu:update",
		"DELETE /menus": "sys:menu:delete",

		// logs
		"DELETE /logs/login": "sys:log:delete",
		"DELETE /logs/op":    "sys:log:delete",

		// configs
		"POST /configs":   "sys:config:add",
		"PUT /configs":    "sys:config:update",
		"DELETE /configs": "sys:config:delete",

		// cron jobs
		"POST /jobs":         "sys:job:add",
		"PUT /jobs":          "sys:job:update",
		"DELETE /jobs":       "sys:job:delete",
		"POST /jobs/run":     "sys:job:run",
		"POST /jobs/start":   "sys:job:start",
		"POST /jobs/stop":    "sys:job:stop",
		"GET /jobs/handlers": "sys:job:handlers",
	}

	key := strings.ToUpper(method) + " " + path
	return permMap[key]
}

// normalizePath 统一 path 格式
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}

	return path
}

// normalizeDynamicPath 把常见动态参数路径归一化
// 例如：
// /users/123         -> /users
// /roles/999         -> /roles
// /users/123/detail  -> /users/detail
func normalizeDynamicPath(path string) string {
	parts := strings.Split(path, "/")
	clean := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}

		// 纯数字 ID
		if isNumeric(part) {
			continue
		}

		clean = append(clean, part)
	}

	if len(clean) == 0 {
		return "/"
	}

	return "/" + strings.Join(clean, "/")
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
