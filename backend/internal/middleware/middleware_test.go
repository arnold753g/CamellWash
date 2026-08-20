package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"starter-backend/internal/models"
)

func TestRoleMiddlewareRejectsCrossRoleAccess(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := RequireRoles(models.RoleAdmin)(next)
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request = WithPrincipal(request, Principal{User: models.User{Role: models.RoleManager}})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", response.Code, http.StatusForbidden)
	}
}

func TestAdminRoleMiddlewareMatrix(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		role string
		want int
	}{
		{role: models.RoleAdmin, want: http.StatusNoContent},
		{role: models.RoleManager, want: http.StatusForbidden},
		{role: models.RoleWashStaff, want: http.StatusForbidden},
	} {
		t.Run(test.role, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			handler := RequireRoles(models.RoleAdmin)(next)
			request := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
			request = WithPrincipal(request, Principal{User: models.User{Role: test.role}})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("role=%q status=%d want=%d", test.role, response.Code, test.want)
			}
		})
	}
}

func TestClientIPIgnoresForwardedHeaderWhenProxyIsUntrusted(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "127.0.0.1:1234"
	request.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := ClientIP(request, false); got != "127.0.0.1" {
		t.Fatalf("client IP=%q", got)
	}
}
