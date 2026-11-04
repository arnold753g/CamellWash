package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"starter-backend/internal/config"
	"starter-backend/internal/database"
	"starter-backend/internal/httpapi"
	"starter-backend/internal/models"
	"starter-backend/internal/security"
)

// This test deliberately requires a separately provisioned, disposable database.
// Verify its actual name before migrations or fixture writes, regardless of DSN spelling.
func TestAdminUserEmailChangeRevokesSessions(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN must point to a disposable camell_wash_security_test_* database")
	}
	db, err := database.Open(dsn, "test")
	if err != nil {
		t.Fatalf("open isolated test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("verify test database name: %v", err)
	}
	if !strings.HasPrefix(databaseName, "camell_wash_security_test_") {
		t.Fatalf("refusing writes: TEST_DATABASE_DSN database %q must start with camell_wash_security_test_", databaseName)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate isolated test database: %v", err)
	}

	token, err := security.RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	const password = "security regression password 2026!"
	hash, err := security.HashPassword(password, 10)
	if err != nil {
		t.Fatal(err)
	}
	admin := models.User{Email: "security-admin-" + token + "@example.test", PasswordHash: hash,
		FirstName: "Security", LastName: "Admin", Role: models.RoleAdmin, Status: models.StatusActive}
	user := models.User{Email: "security-user-" + token + "@example.test", PasswordHash: hash,
		FirstName: "Security", LastName: "User", Role: models.RoleManager, Status: models.StatusActive}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create administrator fixture: %v", err)
	}
	t.Cleanup(func() {
		// The users foreign key removes only these fixtures' sessions.
		if err := db.Where("id IN ?", []uint64{admin.ID, user.ID}).Delete(&models.User{}).Error; err != nil {
			t.Errorf("remove test fixtures: %v", err)
		}
	})
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user fixture: %v", err)
	}

	const origin = "https://camellwash.security.test"
	cfg := config.Config{AppEnv: "test", AllowedOrigins: []string{origin},
		SessionSecret: "security-integration-test-secret-32-characters",
		SessionTTL:    time.Hour, SessionIdleTimeout: 15 * time.Minute,
		CookieSecure: true, CookieSameSite: http.SameSiteLaxMode, PasswordBcryptCost: 10,
		LoginMaxAttempts: 5, LoginLockout: 15 * time.Minute,
		AuthRatePerMinute: 100, APIRatePerMinute: 1000}
	handler, err := httpapi.New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	newClient := func() *http.Client {
		client := server.Client()
		// server.Client returns the same pointer: copy it for independent cookie jars.
		copy := *client
		copy.Jar, err = cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		return &copy
	}
	adminClient, userClient := newClient(), newClient()

	request := func(t *testing.T, client *http.Client, method, path, csrf string, body any, want int) []byte {
		t.Helper()
		var payload bytes.Buffer
		if body != nil {
			if err := json.NewEncoder(&payload).Encode(body); err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequest(method, server.URL+path, &payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", origin)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if csrf != "" {
			req.Header.Set("X-CSRF-Token", csrf)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s status=%d want=%d body=%s", method, path, response.StatusCode, want, data)
		}
		if path == "/api/v1/auth/login" && response.StatusCode == http.StatusOK {
			var sessionCookie *http.Cookie
			for _, cookie := range response.Cookies() {
				if cookie.Name == security.CookieName {
					sessionCookie = cookie
				}
			}
			if sessionCookie == nil || !sessionCookie.Secure || !sessionCookie.HttpOnly || sessionCookie.Path != "/api/v1" {
				t.Fatal("login must issue a Secure HttpOnly session cookie scoped to /api/v1")
			}
		}
		return data
	}
	login := func(client *http.Client, email string) string {
		t.Helper()
		data := request(t, client, http.MethodPost, "/api/v1/auth/login", "",
			map[string]string{"email": email, "password": password}, http.StatusOK)
		var envelope struct {
			Data struct {
				CSRFToken string `json:"csrf_token"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil || envelope.Data.CSRFToken == "" {
			t.Fatalf("login did not return a CSRF token: %v", err)
		}
		return envelope.Data.CSRFToken
	}
	adminCSRF := login(adminClient, admin.Email)
	_ = login(userClient, user.Email)
	request(t, userClient, http.MethodGet, "/api/v1/auth/session", "", nil, http.StatusOK)

	t.Run("name_change_keeps_existing_session", func(t *testing.T) {
		request(t, adminClient, http.MethodPatch, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminCSRF,
			map[string]string{"email": user.Email, "first_name": "Updated", "last_name": user.LastName}, http.StatusOK)
		data := request(t, userClient, http.MethodGet, "/api/v1/auth/session", "", nil, http.StatusOK)
		var envelope struct {
			Data struct {
				User models.PublicUser `json:"user"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil || envelope.Data.User.FirstName != "Updated" {
			t.Fatalf("session must retain access and return updated profile: %v", err)
		}
	})
	t.Run("email_change_revokes_existing_session", func(t *testing.T) {
		request(t, adminClient, http.MethodPatch, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminCSRF,
			map[string]string{"email": "changed-" + token + "@example.test", "first_name": "Updated", "last_name": user.LastName}, http.StatusOK)
		data := request(t, userClient, http.MethodGet, "/api/v1/auth/session", "", nil, http.StatusUnauthorized)
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil || envelope.Error.Code != "INVALID_SESSION" {
			t.Fatalf("old cookie must be rejected as INVALID_SESSION: %v", err)
		}
	})
	t.Run("locked_account_rejects_both_passwords_without_changing_lock", func(t *testing.T) {
		var lockedUser models.User
		if err := db.First(&lockedUser, user.ID).Error; err != nil {
			t.Fatal(err)
		}
		lockedUntil := time.Now().UTC().Add(10 * time.Minute).Truncate(time.Microsecond)
		const failedAttempts = 3
		if err := db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
			"failed_login_attempts": failedAttempts, "locked_until": lockedUntil,
		}).Error; err != nil {
			t.Fatalf("lock user fixture: %v", err)
		}
		for _, candidate := range []string{"incorrect regression password", password} {
			data := request(t, userClient, http.MethodPost, "/api/v1/auth/login", "",
				map[string]string{"email": lockedUser.Email, "password": candidate}, http.StatusTooManyRequests)
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(data, &envelope); err != nil || envelope.Error.Code != "ACCOUNT_LOCKED" {
				t.Fatalf("locked login must return ACCOUNT_LOCKED for either password: %v", err)
			}
			var current models.User
			if err := db.First(&current, user.ID).Error; err != nil {
				t.Fatal(err)
			}
			if current.FailedLoginAttempts != failedAttempts || current.LockedUntil == nil || !current.LockedUntil.Equal(lockedUntil) {
				t.Fatalf("locked login changed lock state: attempts=%d locked_until=%v", current.FailedLoginAttempts, current.LockedUntil)
			}
		}
		if err := db.Model(&models.User{}).Where("id = ?", user.ID).
			Update("locked_until", time.Now().UTC().Add(-time.Minute)).Error; err != nil {
			t.Fatalf("expire fixture lock: %v", err)
		}
		request(t, userClient, http.MethodPost, "/api/v1/auth/login", "",
			map[string]string{"email": lockedUser.Email, "password": password}, http.StatusOK)
		request(t, userClient, http.MethodGet, "/api/v1/auth/session", "", nil, http.StatusOK)
		var unlocked models.User
		if err := db.First(&unlocked, user.ID).Error; err != nil {
			t.Fatal(err)
		}
		if unlocked.FailedLoginAttempts != 0 || unlocked.LockedUntil != nil {
			t.Fatalf("successful login did not clear expired lock: attempts=%d locked_until=%v", unlocked.FailedLoginAttempts, unlocked.LockedUntil)
		}
	})
}
