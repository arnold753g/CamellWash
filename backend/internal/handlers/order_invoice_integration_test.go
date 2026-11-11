package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	"starter-backend/internal/services"
)

func decodeInvoiceData[T any](t *testing.T, body []byte) T {
	t.Helper()
	var envelope struct {
		Data T `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data
}

func TestOrderInvoiceClientEmailAndAccountingAccess(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("requires disposable camell_wash_security_test_* database")
	}
	db, err := database.Open(dsn, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close(db)
	var name string
	if err := db.Raw("SELECT current_database()").Scan(&name).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(name, "camell_wash_security_test_") {
		t.Fatalf("refusing fixture writes to %s", name)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	const password = "invoice access local tests 2026!"
	hash, err := security.HashPassword(password, 10)
	if err != nil {
		t.Fatal(err)
	}
	users := map[string]models.User{}
	for _, role := range []string{models.RoleAdmin, models.RoleManager, models.RoleAccountant, models.RoleWashStaff} {
		user := models.User{Email: fmt.Sprintf("invoice-%s-%d@example.test", role, time.Now().UnixNano()), PasswordHash: hash,
			FirstName: "Invoice", LastName: role, Role: role, Status: models.StatusActive}
		if err := tx.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
		users[role] = user
	}
	const origin = "http://camell.invoice.test"
	cfg := config.Config{AppEnv: "test", AllowedOrigins: []string{origin}, SessionSecret: "invoice-integration-secret-32-characters",
		SessionTTL: time.Hour, SessionIdleTimeout: time.Hour, PasswordBcryptCost: 10,
		CookieSameSite: http.SameSiteLaxMode, LoginMaxAttempts: 5, LoginLockout: time.Minute,
		AuthRatePerMinute: 100, APIRatePerMinute: 10000}
	handler, err := httpapi.New(tx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	cookies := map[string][]*http.Cookie{}
	csrf := map[string]string{}
	request := func(role, method, path string, body any, want int) []byte {
		t.Helper()
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, origin+"/api/v1"+path, bytes.NewReader(payload))
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrf[role])
		for _, cookie := range cookies[role] {
			req.AddCookie(cookie)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != want {
			t.Fatalf("%s %s %s = %d want %d: %s", role, method, path, recorder.Code, want, recorder.Body)
		}
		if path == "/auth/login" {
			cookies[role] = recorder.Result().Cookies()
		}
		return recorder.Body.Bytes()
	}
	for role, user := range users {
		body := request(role, "POST", "/auth/login", map[string]any{"email": user.Email, "password": password}, 200)
		data := decodeInvoiceData[struct {
			CSRF string `json:"csrf_token"`
		}](t, body)
		csrf[role] = data.CSRF
	}
	manager := models.RoleManager
	accountant := models.RoleAccountant
	request(manager, "POST", "/clients", map[string]any{"full_name": "Correo inválido", "phone": "70000000", "email": "mal@"}, 422)
	client := decodeInvoiceData[services.ClientDetail](t, request(manager, "POST", "/clients", map[string]any{
		"full_name": "Cliente Factura", "phone": "70000000", "email": " CLIENTE@EXAMPLE.TEST "}, 201))
	if client.Email == nil || *client.Email != "cliente@example.test" {
		t.Fatalf("email not normalized: %+v", client)
	}
	withoutEmail := decodeInvoiceData[services.ClientDetail](t, request(manager, "POST", "/clients", map[string]any{"full_name": "Sin Correo", "phone": "70000001"}, 201))
	if withoutEmail.Email != nil {
		t.Fatal("omitted email must stay absent")
	}
	path := fmt.Sprintf("/clients/%d", client.ID)
	updated := decodeInvoiceData[services.ClientDetail](t, request(manager, "PATCH", path, map[string]any{"full_name": client.FullName, "phone": client.Phone}, 200))
	if updated.Email == nil || *updated.Email != *client.Email {
		t.Fatal("legacy update lost email")
	}
	request(accountant, "GET", fmt.Sprintf("/accountant/clients/%d", client.ID), nil, 200)
	request(accountant, "GET", "/accountant/clients", nil, 200)
	request(accountant, "GET", "/clients", nil, 403)
	request(manager, "GET", "/accountant/clients", nil, 403)
	request(models.RoleWashStaff, "GET", "/accountant/clients", nil, 403)
	request(accountant, "PATCH", path, map[string]any{"email": "changed@example.test"}, 403)
	request(accountant, "POST", "/clients", map[string]any{"full_name": "Forbidden", "phone": "70000001"}, 403)
	var vt models.VehicleType
	var catalog models.Service
	if err := tx.First(&vt).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.First(&catalog).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&models.ServicePrice{ServiceID: catalog.ID, VehicleTypeID: vt.ID, Price: "50.00"}).Error; err != nil {
		t.Fatal(err)
	}
	vehicle := decodeInvoiceData[models.Vehicle](t, request(manager, "POST", fmt.Sprintf("/clients/%d/vehicles", client.ID), map[string]any{"vehicle_type_id": vt.ID, "plate": "FCT2026"}, 201))
	input := map[string]any{"vehicle_id": vehicle.ID, "service_ids": []uint64{catalog.ID}, "staff_ids": []uint64{users[models.RoleWashStaff].ID}, "estimated_delivery_at": time.Now().Add(time.Hour)}
	input["invoice_requested"] = "yes"
	request(manager, "POST", "/orders", input, 400)
	input["invoice_requested"] = true
	order := decodeInvoiceData[models.Order](t, request(manager, "POST", "/orders", input, 201))
	if !order.InvoiceRequested {
		t.Fatal("order lost invoice preference")
	}
	delete(input, "invoice_requested")
	other := decodeInvoiceData[models.Order](t, request(manager, "POST", "/orders", input, 201))
	if other.InvoiceRequested {
		t.Fatal("new orders must default to no invoice request")
	}
	reception := decodeInvoiceData[services.DocumentView](t, request(manager, "POST", fmt.Sprintf("/reception/orders/%d/document", order.ID), nil, 200))
	var snapshot services.DocumentSnapshot
	if err := json.Unmarshal(reception.Snapshot, &snapshot); err != nil || !snapshot.InvoiceRequested {
		t.Fatal("reception lost invoice preference")
	}
	var documentCount int64
	if err := tx.Model(&models.Document{}).Where("order_id = ? AND kind = 'RECEPTION'", order.ID).Count(&documentCount).Error; err != nil || documentCount != 1 {
		t.Fatal("reception issuance must be idempotent")
	}
	payment := decodeInvoiceData[services.PaymentResult](t, request(manager, "POST", fmt.Sprintf("/orders/%d/payments", order.ID), map[string]any{"version": order.Version, "method": "CASH", "idempotency_key": "invoice_order_payment_2026"}, 201))
	if !payment.Payment.InvoiceRequested {
		t.Fatal("payment without invoice field must inherit the order")
	}
	corrected := decodeInvoiceData[services.PaymentResult](t, request(manager, "POST", fmt.Sprintf("/payments/%d/correct", payment.Payment.ID), map[string]any{"version": payment.Order.Version, "method": "QR", "reason": "El cliente cambia su solicitud", "invoice_requested": false, "idempotency_key": "invoice_order_correct_2026"}, 201))
	if corrected.Order.InvoiceRequested || corrected.Payment.InvoiceRequested {
		t.Fatal("correction must synchronize invoice preference")
	}
	oldPayment := models.Payment{}
	if err := tx.First(&oldPayment, payment.Payment.ID).Error; err != nil || !oldPayment.InvoiceRequested || oldPayment.Status != models.PaymentVoided {
		t.Fatal("correction must preserve original request")
	}
	period := time.Now().In(time.FixedZone("Bolivia", -4*60*60)).Format("2006-01-02")
	reportPath := "/reports/orders?date_from=" + period + "&date_to=" + period
	for _, role := range []string{manager, models.RoleWashStaff} {
		request(role, "GET", reportPath, nil, 403)
	}
	report := decodeInvoiceData[services.AccountingOrdersReport](t, request(accountant, "GET", reportPath, nil, 200))
	count := 0
	for _, row := range report.Orders {
		if row.ID == order.ID {
			count++
			if row.ClientEmail == nil || *row.ClientEmail != *client.Email || row.PaymentID == nil || *row.PaymentID != corrected.Payment.ID || row.InvoiceRequested {
				t.Fatalf("incorrect accounting row: %+v", row)
			}
		}
	}
	if count != 1 {
		t.Fatal("corrected payments must not duplicate accounting rows")
	}
	request(accountant, "GET", reportPath+"&invoice_requested=invalid", nil, 422)
	filtered := decodeInvoiceData[services.AccountingOrdersReport](t, request(accountant, "GET", reportPath+"&invoice_requested=true", nil, 200))
	for _, row := range filtered.Orders {
		if !row.InvoiceRequested {
			t.Fatal("invoice filter failed")
		}
	}
	request(accountant, "POST", fmt.Sprintf("/reception/orders/%d/document", other.ID), nil, 403)
	request(accountant, "POST", fmt.Sprintf("/orders/%d/payments", other.ID), map[string]any{}, 403)
	cleared := decodeInvoiceData[services.ClientDetail](t, request(manager, "PATCH", path, map[string]any{"full_name": client.FullName, "phone": client.Phone, "email": ""}, 200))
	if cleared.Email != nil {
		t.Fatal("blank email must clear the optional field")
	}
}
