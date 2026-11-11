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

// Exercise the real router, database migrations, authentication and CSRF together.
// Never migrate or write fixtures until the actual database name is verified.
func TestAccountantReportAccessAndRoleIsolation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN must point to a disposable camell_wash_security_test_* database")
	}
	db, err := database.Open(dsn, "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close(db) })
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(databaseName, "camell_wash_security_test_") {
		t.Fatalf("refusing writes: database %q must start with camell_wash_security_test_", databaseName)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate isolated database: %v", err)
	}
	token, err := security.RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	const password = "accountant access regression password 2026!"
	hash, err := security.HashPassword(password, 10)
	if err != nil {
		t.Fatal(err)
	}
	var fixtureIDs []uint64
	t.Cleanup(func() {
		if len(fixtureIDs) > 0 {
			if err := db.Where("id IN ?", fixtureIDs).Delete(&models.User{}).Error; err != nil {
				t.Errorf("remove role-isolation fixtures: %v", err)
			}
		}
	})
	users := make(map[string]models.User)
	for _, role := range []string{models.RoleAdmin, models.RoleManager, models.RoleWashStaff} {
		user := models.User{Email: role + "-" + token + "@example.test", PasswordHash: hash,
			FirstName: "Access", LastName: "Regression", Role: role, Status: models.StatusActive}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create %s fixture: %v", role, err)
		}
		fixtureIDs = append(fixtureIDs, user.ID)
		users[role] = user
	}
	const origin = "https://camellwash.security.test"
	cfg := config.Config{AppEnv: "test", AllowedOrigins: []string{origin},
		SessionSecret: "accountant-isolation-secret-32-characters",
		SessionTTL:    time.Hour, SessionIdleTimeout: 15 * time.Minute,
		CookieSecure: true, CookieSameSite: http.SameSiteLaxMode, PasswordBcryptCost: 10,
		LoginMaxAttempts: 5, LoginLockout: 15 * time.Minute,
		AuthRatePerMinute: 100, APIRatePerMinute: 10000}
	handler, err := httpapi.New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	clients := make(map[string]*http.Client)
	csrfTokens := make(map[string]string)
	request := func(t *testing.T, role, method, path string, body any, want int) []byte {
		t.Helper()
		var payload bytes.Buffer
		if body != nil {
			if err := json.NewEncoder(&payload).Encode(body); err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequest(method, server.URL+"/api/v1"+path, &payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrfTokens[role])
		response, err := clients[role].Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("role=%s %s %s status=%d want=%d body=%s", role, method, path, response.StatusCode, want, data)
		}
		if want == http.StatusForbidden {
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(data, &envelope); err != nil || envelope.Error.Code != "FORBIDDEN" {
				t.Fatalf("role denial must be FORBIDDEN, not failed CSRF or origin validation: %s", data)
			}
		}
		return data
	}
	login := func(t *testing.T, role string) {
		t.Helper()
		client := *server.Client()
		client.Jar, err = cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		clients[role] = &client
		data := request(t, role, http.MethodPost, "/auth/login", map[string]string{
			"email": users[role].Email, "password": password,
		}, http.StatusOK)
		var envelope struct {
			Data struct {
				User      models.PublicUser `json:"user"`
				CSRFToken string            `json:"csrf_token"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil || envelope.Data.User.Role != users[role].Role || envelope.Data.CSRFToken == "" {
			t.Fatalf("login must return the assigned role and CSRF token: %s", data)
		}
		csrfTokens[role] = envelope.Data.CSRFToken
	}
	login(t, models.RoleAdmin)
	// Creating through the administrator API checks both role validation and the
	// database constraint, rather than bypassing either with a mocked principal.
	accountantEmail := "accountant-" + token + "@example.test"
	data := request(t, models.RoleAdmin, http.MethodPost, "/admin/users", map[string]string{
		"email": accountantEmail, "password": password, "first_name": "Access", "last_name": "Accountant",
		"role": models.RoleAccountant, "status": models.StatusActive,
	}, http.StatusCreated)
	var created struct {
		Data models.PublicUser `json:"data"`
	}
	if err := json.Unmarshal(data, &created); err != nil || created.Data.ID == 0 || created.Data.Role != models.RoleAccountant {
		t.Fatalf("administrator must be able to create an accountant: %s", data)
	}
	fixtureIDs = append(fixtureIDs, created.Data.ID)
	users[models.RoleAccountant] = models.User{ID: created.Data.ID, Email: accountantEmail, Role: models.RoleAccountant}
	for _, role := range []string{models.RoleManager, models.RoleWashStaff, models.RoleAccountant} {
		login(t, role)
	}

	t.Run("report_permission_matrix", func(t *testing.T) {
		for _, tc := range []struct {
			path    string
			allowed []string
		}{
			{"/reports/operations", []string{models.RoleAdmin, models.RoleManager, models.RoleAccountant}},
			{"/reports/dashboard", []string{models.RoleAdmin, models.RoleAccountant}},
			{"/reports/financial", []string{models.RoleAdmin, models.RoleAccountant}},
			{"/admin/dashboard", []string{models.RoleAdmin}},
			{"/admin/reports/financial", []string{models.RoleAdmin}},
		} {
			for _, role := range []string{models.RoleAdmin, models.RoleManager, models.RoleWashStaff, models.RoleAccountant} {
				t.Run(role+tc.path, func(t *testing.T) {
					want := http.StatusForbidden
					for _, allowed := range tc.allowed {
						if role == allowed {
							want = http.StatusOK
						}
					}
					request(t, role, http.MethodGet, tc.path+"?date_from=2026-01-01&date_to=2026-01-02", nil, want)
				})
			}
		}
	})
	t.Run("existing_roles_keep_order_and_user_permissions", func(t *testing.T) {
		for _, role := range []string{models.RoleAdmin, models.RoleManager, models.RoleWashStaff} {
			request(t, role, http.MethodGet, "/orders", nil, http.StatusOK)
			want := http.StatusForbidden
			if role == models.RoleAdmin {
				want = http.StatusOK
			}
			request(t, role, http.MethodGet, "/admin/users", nil, want)
		}
		request(t, models.RoleAdmin, http.MethodGet, "/admin/users?role=accountant", nil, http.StatusOK)
	})
	t.Run("wash_staff_sees_only_assigned_orders", func(t *testing.T) {
		vehicleType := models.VehicleType{Name: "Access " + token, Active: true}
		client := models.Client{FullName: "Access Regression", Phone: "70000000"}
		var vehicle models.Vehicle
		var orderIDs []uint64
		t.Cleanup(func() {
			// These fixtures have no history or payments. Remove only the IDs we created.
			if len(orderIDs) > 0 {
				if err := db.Where("order_id IN ?", orderIDs).Delete(&models.OrderStaff{}).Error; err != nil {
					t.Errorf("remove order assignments: %v", err)
				}
				if err := db.Where("id IN ?", orderIDs).Delete(&models.Order{}).Error; err != nil {
					t.Errorf("remove orders: %v", err)
				}
			}
			for _, fixture := range []struct {
				model any
				id    uint64
			}{{&models.Vehicle{}, vehicle.ID}, {&models.Client{}, client.ID}, {&models.VehicleType{}, vehicleType.ID}} {
				if fixture.id != 0 {
					if err := db.Where("id = ?", fixture.id).Delete(fixture.model).Error; err != nil {
						t.Errorf("remove operational fixture: %v", err)
					}
				}
			}
		})
		if err := db.Create(&vehicleType).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&client).Error; err != nil {
			t.Fatal(err)
		}
		vehicle = models.Vehicle{ClientID: client.ID, VehicleTypeID: vehicleType.ID}
		if err := db.Create(&vehicle).Error; err != nil {
			t.Fatal(err)
		}
		now := time.Now().UTC()
		for range 2 {
			order := models.Order{VehicleID: vehicle.ID, RegisteredByUserID: users[models.RoleManager].ID,
				ReceivedAt: now, EstimatedDeliveryAt: now.Add(time.Hour), Status: models.OrderPending, Total: "50.00", Version: 1}
			if err := db.Create(&order).Error; err != nil {
				t.Fatal(err)
			}
			orderIDs = append(orderIDs, order.ID)
		}
		if err := db.Create(&models.OrderStaff{OrderID: orderIDs[0], UserID: users[models.RoleWashStaff].ID, AssignedAt: now}).Error; err != nil {
			t.Fatal(err)
		}
		for _, role := range []string{models.RoleAdmin, models.RoleManager, models.RoleWashStaff} {
			data := request(t, role, http.MethodGet, fmt.Sprintf("/orders?vehicle_id=%d", vehicle.ID), nil, http.StatusOK)
			var envelope struct {
				Data struct {
					Orders []models.Order `json:"orders"`
				} `json:"data"`
			}
			if err := json.Unmarshal(data, &envelope); err != nil {
				t.Fatal(err)
			}
			want := 2
			if role == models.RoleWashStaff {
				want = 1
			}
			if len(envelope.Data.Orders) != want {
				t.Fatalf("role=%s returned %d orders, want %d", role, len(envelope.Data.Orders), want)
			}
			if role == models.RoleWashStaff && envelope.Data.Orders[0].ID != orderIDs[0] {
				t.Fatal("wash staff received an unassigned order")
			}
		}
		request(t, models.RoleWashStaff, http.MethodGet, fmt.Sprintf("/orders/%d", orderIDs[0]), nil, http.StatusOK)
		request(t, models.RoleWashStaff, http.MethodGet, fmt.Sprintf("/orders/%d", orderIDs[1]), nil, http.StatusNotFound)
		request(t, models.RoleAccountant, http.MethodGet, fmt.Sprintf("/orders/%d", orderIDs[0]), nil, http.StatusForbidden)
	})
	t.Run("accountant_cannot_read_operational_or_administration_routes", func(t *testing.T) {
		for _, path := range []string{
			"/documents/1", "/orders", "/orders/1", "/orders/1/payments", "/clients", "/clients/1", "/clients/1/vehicles",
			"/clients/1/orders", "/vehicles/1", "/vehicles/1/orders", "/pricing/current",
			"/operations/wash-staff", "/operations/service-options", "/catalogs/vehicle-types",
			"/catalogs/service-categories", "/catalogs/services", "/admin/users", "/admin/users/1",
			"/admin/vehicle-types", "/admin/vehicle-types/1", "/admin/service-categories", "/admin/service-categories/1",
			"/admin/services", "/admin/services/1", "/admin/service-prices/matrix", "/admin/payment-reviews",
		} {
			t.Run(path, func(t *testing.T) { request(t, models.RoleAccountant, http.MethodGet, path, nil, http.StatusForbidden) })
		}
	})
	t.Run("accountant_cannot_write_or_approve_even_with_valid_csrf", func(t *testing.T) {
		// Prove this session's CSRF token and origin work before testing denied writes.
		request(t, models.RoleAccountant, http.MethodPatch, "/me", map[string]string{
			"first_name": "Accountant", "last_name": "Regression",
		}, http.StatusOK)
		for _, tc := range []struct{ method, path string }{
			{http.MethodPost, "/payments/1/documents/receipt"}, {http.MethodPost, "/reception/orders/1/document"}, {http.MethodPost, "/documents/1/print-attempts"},
			{http.MethodPost, "/orders"}, {http.MethodPatch, "/orders/1"}, {http.MethodPost, "/orders/1/transitions"},
			{http.MethodPost, "/orders/1/cancel"}, {http.MethodPost, "/orders/1/payments"},
			{http.MethodPost, "/payments/1/void"}, {http.MethodPost, "/payments/1/correct"},
			{http.MethodPost, "/clients"}, {http.MethodPatch, "/clients/1"}, {http.MethodPost, "/clients/1/vehicles"},
			{http.MethodPatch, "/vehicles/1"}, {http.MethodPost, "/admin/users"}, {http.MethodPatch, "/admin/users/1"},
			{http.MethodPatch, "/admin/users/1/role"}, {http.MethodPatch, "/admin/users/1/status"},
			{http.MethodPost, "/admin/vehicle-types"}, {http.MethodPatch, "/admin/vehicle-types/1"},
			{http.MethodPatch, "/admin/vehicle-types/1/status"}, {http.MethodPost, "/admin/service-categories"},
			{http.MethodPatch, "/admin/service-categories/1"}, {http.MethodPatch, "/admin/service-categories/1/status"},
			{http.MethodPost, "/admin/services"}, {http.MethodPatch, "/admin/services/1"},
			{http.MethodPatch, "/admin/services/1/status"}, {http.MethodPost, "/admin/service-prices"},
			{http.MethodPatch, "/admin/service-prices/1/1"}, {http.MethodPost, "/admin/payment-reviews/1"},
		} {
			t.Run(tc.method+tc.path, func(t *testing.T) {
				request(t, models.RoleAccountant, tc.method, tc.path, map[string]any{}, http.StatusForbidden)
			})
		}
	})
	t.Run("unknown_roles_rejected_by_api_and_database", func(t *testing.T) {
		request(t, models.RoleAdmin, http.MethodPost, "/admin/users", map[string]string{
			"email": "unknown-" + token + "@example.test", "password": password, "first_name": "Access", "last_name": "Unknown",
			"role": "accountant_admin", "status": models.StatusActive,
		}, http.StatusUnprocessableEntity)
		invalid := models.User{Email: "invalid-db-" + token + "@example.test", PasswordHash: hash,
			FirstName: "Invalid", LastName: "Role", Role: "unknown", Status: models.StatusActive}
		if err := db.Create(&invalid).Error; err == nil {
			fixtureIDs = append(fixtureIDs, invalid.ID)
			t.Fatal("database role constraint accepted an unknown role")
		}
	})
	t.Run("changing_role_revokes_old_permissions", func(t *testing.T) {
		manager := users[models.RoleManager]
		path := fmt.Sprintf("/admin/users/%d/role", manager.ID)
		request(t, models.RoleAdmin, http.MethodPatch, path, map[string]string{"role": models.RoleAccountant}, http.StatusOK)
		request(t, models.RoleManager, http.MethodGet, "/orders", nil, http.StatusUnauthorized)
		manager.Role = models.RoleAccountant
		users[models.RoleManager] = manager
		login(t, models.RoleManager)
		request(t, models.RoleManager, http.MethodGet, "/reports/financial", nil, http.StatusOK)
		request(t, models.RoleManager, http.MethodGet, "/orders", nil, http.StatusForbidden)
		request(t, models.RoleAdmin, http.MethodPatch, path, map[string]string{"role": models.RoleManager}, http.StatusOK)
		request(t, models.RoleManager, http.MethodGet, "/reports/financial", nil, http.StatusUnauthorized)
		manager.Role = models.RoleManager
		users[models.RoleManager] = manager
		login(t, models.RoleManager)
		request(t, models.RoleManager, http.MethodGet, "/orders", nil, http.StatusOK)
		request(t, models.RoleManager, http.MethodGet, "/reports/financial", nil, http.StatusForbidden)
	})
}
