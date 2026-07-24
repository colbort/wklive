package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"wklive/common/utils"
	"wklive/liquidity-admin-api/internal/svc"
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
	mu     sync.RWMutex
	rules  []PermissionRule
}

func NewRbacMiddleware(svcCtx *svc.ServiceContext) *RbacMiddleware {
	m := &RbacMiddleware{svcCtx: svcCtx}
	m.refreshRules(context.Background())
	return m
}

func (m *RbacMiddleware) refreshRules(ctx context.Context) {
	rules, err := loadPermissionRules(ctx, m.svcCtx)
	if err != nil {
		logx.Errorf("fetch system permissions failed: %v", err)
		return
	}
	m.mu.Lock()
	m.rules = rules
	m.mu.Unlock()
	logx.Infof("loaded liquidity rbac permission rules: %d", len(rules))
}

func loadPermissionRules(ctx context.Context, svcCtx *svc.ServiceContext) ([]PermissionRule, error) {
	result, err := svcCtx.SystemCli.SysPermList(ctx, &system.Empty{})
	if err != nil {
		return nil, err
	}
	rules := make([]PermissionRule, 0, len(result.GetData()))
	for _, item := range result.GetData() {
		if item.AppScope != system.ApplicationScope_APPLICATION_SCOPE_LIQUIDITY {
			continue
		}
		pattern, staticSegs, err := compilePathPattern(item.Path)
		if err != nil {
			logx.Errorf("compile liquidity permission path failed: method=%s path=%s err=%v", item.Method, item.Path, err)
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
	return rules, nil
}

func (m *RbacMiddleware) requiredPermission(ctx context.Context, path, method string) string {
	m.mu.RLock()
	required := getRequiredPermission(m.rules, path, method)
	m.mu.RUnlock()
	if required != "" {
		return required
	}
	// 管理后台新建菜单权限后，无需重启 API；首次访问会自动刷新规则。
	m.refreshRules(ctx)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return getRequiredPermission(m.rules, path, method)
}

func (m *RbacMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isWhitePath(r.URL.Path) {
			next(w, r)
			return
		}
		tokenUserID, tokenUsername, ok := m.liquidityTokenIdentity(r)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		authCtx := context.WithValue(r.Context(), utils.CtxKeyUid, tokenUserID)
		authCtx = context.WithValue(authCtx, utils.CtxKeyUsername, tokenUsername)
		r = r.WithContext(authCtx)
		userID := tokenUserID
		perms, err := m.svcCtx.SystemCli.LoginUserPerms(r.Context(), &system.LoginUserPermsReq{UserId: userID})
		if err != nil {
			logx.Errorf("get liquidity user permissions failed, userId=%d err=%v", userID, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		user, err := m.svcCtx.SystemCli.SysUserDetail(r.Context(), &system.SysUserDetailReq{Id: userID})
		if err != nil || user == nil || user.Data == nil {
			logx.Errorf("get liquidity user detail failed, userId=%d err=%v", userID, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		required := m.requiredPermission(r.Context(), r.URL.Path, r.Method)
		if required == "" {
			logx.Errorf("liquidity permission route not found, method=%s path=%s", r.Method, r.URL.Path)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		enforcer, err := newUserPermEnforcer(fmt.Sprintf("%d", userID), perms.Perms)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		obj, act, ok := parsePerm(required)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		allowed, err := enforcer.Enforce(fmt.Sprintf("%d", userID), obj, act)
		if err != nil || !allowed {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), utils.CtxKeyTenantId, user.Data.TenantId)
		ctx = context.WithValue(ctx, utils.CtxKeyUserType, int64(user.Data.UserType))
		next(w, r.WithContext(ctx))
	}
}

func (m *RbacMiddleware) liquidityTokenIdentity(r *http.Request) (int64, string, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	claims, err := utils.ParseToken(m.svcCtx.Config.Jwt.AccessSecret, token)
	if err != nil {
		return 0, "", false
	}
	var expand struct {
		AppScope int32 `json:"appScope"`
	}
	ok := json.Unmarshal([]byte(claims.Expand), &expand) == nil &&
		expand.AppScope == int32(system.ApplicationScope_APPLICATION_SCOPE_LIQUIDITY)
	return claims.UserId, claims.Username, ok
}

func newUserPermEnforcer(userID string, perms []string) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`)
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
			continue
		}
		if _, err = enforcer.AddPolicy(userID, obj, act); err != nil {
			return nil, err
		}
	}
	return enforcer, nil
}

func parsePerm(perm string) (obj, act string, ok bool) {
	parts := strings.Split(strings.TrimSpace(perm), ":")
	if len(parts) < 2 {
		return "", "", false
	}
	act = strings.TrimSpace(parts[len(parts)-1])
	obj = strings.TrimSpace(strings.Join(parts[:len(parts)-1], ":"))
	return obj, act, obj != "" && act != ""
}

func isWhitePath(path string) bool {
	return path == "/liquidity/admin/auth/login" ||
		path == "/liquidity/admin/auth/profile" ||
		path == "/liquidity/admin/options" ||
		path == "/health"
}

func getRequiredPermission(rules []PermissionRule, path, method string) string {
	path = normalizePath(path)
	method = strings.ToUpper(strings.TrimSpace(method))
	var matched *PermissionRule
	for i := range rules {
		rule := &rules[i]
		if rule.Method != method || !rule.Pattern.MatchString(path) {
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

func compilePathPattern(route string) (*regexp.Regexp, int, error) {
	parts := strings.Split(normalizePath(route), "/")
	staticSegs := 0
	for i, part := range parts {
		switch {
		case part == "":
		case strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}"):
			parts[i] = `[^/]+`
		case strings.HasPrefix(part, ":"):
			parts[i] = `[^/]+`
		default:
			staticSegs++
		}
	}
	reg, err := regexp.Compile("^" + strings.Join(parts, "/") + "$")
	return reg, staticSegs, err
}
