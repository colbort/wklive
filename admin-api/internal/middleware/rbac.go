package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"wklive/admin-api/internal/svc"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionRule struct {
	Method     string
	Path       string
	PermKey    string
	Pattern    *regexp.Regexp
	StaticSegs int
}

type RbacMiddleware struct {
	svcCtx *svc.ServiceContext
	rules  []PermissionRule
}

func NewRbacMiddleware(svcCtx *svc.ServiceContext) *RbacMiddleware {
	result, err := svcCtx.SystemCli.SysPermList(context.Background(), &system.Empty{})
	if err != nil {
		logx.Errorf("fetch system permissions failed: %v", err)
	}

	rules := make([]PermissionRule, 0)
	if result != nil {
		rules = make([]PermissionRule, 0, len(result.Data))
		for _, item := range result.Data {
			pattern, staticSegs, err := compilePathPattern(item.Path)
			if err != nil {
				logx.Errorf("compile path pattern failed: method=%s path=%s err=%v", item.Method, item.Path, err)
				continue
			}

			method := strings.TrimPrefix(item.Method.String(), "REQUEST_METHOD_")
			if item.Method == system.RequestMethod_REQUEST_METHOD_UNKNOWN {
				method = ""
			}

			rules = append(rules, PermissionRule{
				Method:     strings.ToUpper(strings.TrimSpace(method)),
				Path:       normalizePath(item.Path),
				PermKey:    item.PermKey,
				Pattern:    pattern,
				StaticSegs: staticSegs,
			})
		}
	}

	return &RbacMiddleware{
		svcCtx: svcCtx,
		rules:  rules,
	}
}

func (m *RbacMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		if isWhitePath(path) {
			next(w, r)
			return
		}

		uid, err := utils.GetUidFromCtx(r.Context())
		if err != nil {
			logx.Errorf("invalid token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		resp, err := m.svcCtx.SystemCli.LoginUserPerms(r.Context(), &system.LoginUserPermsReq{
			UserId: uid,
		})
		if err != nil {
			logx.Errorf("get profile failed, uid=%d err=%v", uid, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		requiredPerm := getRequiredPermission(m.rules, path, method)
		if requiredPerm == "" {
			logx.Errorf("permission route not found, method=%s path=%s", method, path)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		enforcer, err := newUserPermEnforcer(fmt.Sprintf("%d", uid), resp.Perms)
		if err != nil {
			logx.Errorf("create casbin enforcer failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		obj, act, ok := parsePerm(requiredPerm)
		if !ok {
			logx.Errorf("invalid required permission format: %s", requiredPerm)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		allowed, err := enforcer.Enforce(fmt.Sprintf("%d", uid), obj, act)
		if err != nil {
			logx.Errorf("casbin enforce failed, uid=%d perm=%s err=%v", uid, requiredPerm, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			logx.Errorf("forbidden, uid=%d method=%s path=%s required=%s userPerms=%v",
				uid, method, path, requiredPerm, resp.Perms)
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

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, err
	}

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

func isWhitePath(path string) bool {
	whiteList := map[string]struct{}{
		"/admin/system/core":         {},
		"/admin/system/auth/login":   {},
		"/admin/system/auth/profile": {},
		"/admin/system/auth/captcha": {},
		"/health":                    {},
	}

	_, ok := whiteList[path]
	return ok
}

func getRequiredPermission(rules []PermissionRule, path, method string) string {
	path = strings.TrimSpace(path)

	if strings.HasPrefix(path, "/admin/") {
		path = strings.TrimPrefix(path, "/admin")
	} else if path == "/admin" {
		path = "/"
	}

	path = normalizePath(path)
	method = strings.ToUpper(strings.TrimSpace(method))

	var matched *PermissionRule

	for i := range rules {
		rule := &rules[i]
		if rule.Method != method {
			continue
		}
		if !rule.Pattern.MatchString(path) {
			continue
		}

		if matched == nil || rule.StaticSegs > matched.StaticSegs {
			matched = rule
		}
	}

	if matched == nil {
		return ""
	}

	return matched.PermKey
}

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

// compilePathPattern
// /member/users/{id}            -> ^/member/users/[^/]+$
// /member/users/{id}/status     -> ^/member/users/[^/]+/status$
// /dept/{deptId}/users/{userId} -> ^/dept/[^/]+/users/[^/]+$
func compilePathPattern(route string) (*regexp.Regexp, int, error) {
	route = normalizePath(route)

	parts := strings.Split(route, "/")
	staticSegs := 0

	for i, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			parts[i] = `[^/]+`
		} else {
			staticSegs++
		}
	}

	pattern := "^" + strings.Join(parts, "/") + "$"
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, 0, err
	}

	return reg, staticSegs, nil
}
