package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"wklive/admin-api/internal/config"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/ws"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

type notificationSystemClient struct {
	system.AdminClient
	user  *system.SysUserItem
	perms []string
}

func (c *notificationSystemClient) SysUserDetail(
	_ context.Context,
	_ *system.SysUserDetailReq,
	_ ...grpc.CallOption,
) (*system.SysUserDetailResp, error) {
	return &system.SysUserDetailResp{Data: c.user}, nil
}

func (c *notificationSystemClient) LoginUserPerms(
	_ context.Context,
	_ *system.LoginUserPermsReq,
	_ ...grpc.CallOption,
) (*system.LoginUserPermsResp, error) {
	return &system.LoginUserPermsResp{Perms: c.perms}, nil
}

func TestParseTokenRequiresFixedProtocolAndRejectsURLToken(t *testing.T) {
	token := notificationToken(t)

	request := httptest.NewRequest("GET", "/admin/ws/notifications", nil)
	request.Header.Set(
		"Sec-WebSocket-Protocol",
		notificationSubprotocol+", bearer."+token,
	)
	claims, err := parseToken(request, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserId != 1 {
		t.Fatalf("userId=%d", claims.UserId)
	}

	request = httptest.NewRequest("GET", "/admin/ws/notifications?token="+token, nil)
	request.Header.Set("Sec-WebSocket-Protocol", notificationSubprotocol)
	if _, err := parseToken(request, "secret"); err == nil {
		t.Fatal("URL token must be rejected")
	}

	request = httptest.NewRequest("GET", "/admin/ws/notifications", nil)
	request.Header.Set("Sec-WebSocket-Protocol", "bearer."+token)
	if _, err := parseToken(request, "secret"); err == nil {
		t.Fatal("fixed notification subprotocol must be required")
	}
}

func TestNotificationsHandlerAuthenticatesUserAndEchoesOnlyFixedProtocol(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	svcCtx := notificationServiceContext(hub, common.Enable_ENABLE_ENABLED)
	server := httptest.NewServer(NotificationsHandler(svcCtx))
	defer server.Close()

	dialer := websocket.Dialer{
		Subprotocols: []string{
			notificationSubprotocol,
			"bearer." + notificationToken(t),
		},
	}
	headers := http.Header{"Origin": []string{"https://admin.example.com"}}
	conn, response, err := dialer.Dial(
		"ws"+strings.TrimPrefix(server.URL, "http")+"/admin/ws/notifications",
		headers,
	)
	if err != nil {
		if response != nil {
			t.Fatalf("dial status=%d err=%v", response.StatusCode, err)
		}
		t.Fatal(err)
	}
	defer conn.Close()

	if conn.Subprotocol() != notificationSubprotocol {
		t.Fatalf("subprotocol=%q", conn.Subprotocol())
	}
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"id":"connected"`) {
		t.Fatalf("payload=%s", payload)
	}
}

func TestNotificationsHandlerRejectsMissingOriginAndDisabledUser(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	for name, testCase := range map[string]struct {
		svcCtx     *svc.ServiceContext
		origin     string
		wantStatus int
	}{
		"missing origin": {
			svcCtx:     notificationServiceContext(hub, common.Enable_ENABLE_ENABLED),
			wantStatus: http.StatusForbidden,
		},
		"disabled user": {
			svcCtx:     notificationServiceContext(hub, common.Enable_ENABLE_DISABLED),
			origin:     "https://admin.example.com",
			wantStatus: http.StatusForbidden,
		},
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(NotificationsHandler(testCase.svcCtx))
			defer server.Close()
			dialer := websocket.Dialer{
				Subprotocols: []string{
					notificationSubprotocol,
					"bearer." + notificationToken(t),
				},
			}
			headers := http.Header{}
			if testCase.origin != "" {
				headers.Set("Origin", testCase.origin)
			}
			conn, response, err := dialer.Dial(
				"ws"+strings.TrimPrefix(server.URL, "http")+"/admin/ws/notifications",
				headers,
			)
			if conn != nil {
				_ = conn.Close()
			}
			if err == nil {
				t.Fatal("expected websocket handshake rejection")
			}
			if response == nil || response.StatusCode != testCase.wantStatus {
				t.Fatalf("response=%v want=%d err=%v", response, testCase.wantStatus, err)
			}
		})
	}
}

func notificationServiceContext(
	hub *ws.Hub,
	enabled common.Enable,
) *svc.ServiceContext {
	var cfg config.Config
	cfg.Jwt.AccessSecret = "secret"
	cfg.NotificationWS.AllowedOrigins = []string{"https://admin.example.com"}
	return &svc.ServiceContext{
		Config: cfg,
		SystemCli: &notificationSystemClient{
			user: &system.SysUserItem{
				Id:       1,
				TenantId: 0,
				Enabled:  enabled,
				UserType: system.UserType_USER_TYPE_SYSTEM_ADMIN,
			},
		},
		NotificationHub: hub,
	}
}

func notificationToken(t *testing.T) string {
	t.Helper()
	token, err := utils.GenToken("secret", 1, "admin", "", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
