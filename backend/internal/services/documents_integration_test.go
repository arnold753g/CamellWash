package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"starter-backend/internal/database"
	"starter-backend/internal/models"
	"strings"
	"testing"
	"time"
)

func TestThermalDocumentsLifecycle(t *testing.T) {
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
		t.Fatalf("refusing writes to %s", name)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	actor := models.User{Email: fmt.Sprintf("thermal-%d@example.test", time.Now().UnixNano()), PasswordHash: "not-used-for-login", FirstName: "Caja", LastName: "Prueba", Role: models.RoleManager, Status: models.StatusActive}
	if err := tx.Create(&actor).Error; err != nil {
		t.Fatal(err)
	}
	client := models.Client{FullName: "Cliente Original", Phone: "+59170001234"}
	if err := tx.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	var vehicleType models.VehicleType
	if err := tx.First(&vehicleType).Error; err != nil {
		t.Fatal(err)
	}
	plate := "THERM123"
	vehicle := models.Vehicle{ClientID: client.ID, VehicleTypeID: vehicleType.ID, Brand: "Toyota", Model: "Corolla", Color: "Blanco", Plate: &plate}
	if err := tx.Create(&vehicle).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	order := models.Order{VehicleID: vehicle.ID, RegisteredByUserID: actor.ID, ReceivedAt: now, EstimatedDeliveryAt: now.Add(time.Hour), Status: models.OrderPending, Total: "50.00", Version: 1}
	if err := tx.Create(&order).Error; err != nil {
		t.Fatal(err)
	}
	var catalogService models.Service
	if err := tx.First(&catalogService).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&models.OrderService{OrderID: order.ID, ServiceID: catalogService.ID, ServiceNameSnapshot: "Lavado original", PriceSnapshot: "50.00"}).Error; err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payments, documents := NewPaymentService(tx), NewDocumentService(tx)
	paid, created, err := payments.Register(ctx, actor, order.ID, RegisterPaymentInput{Version: 1, Method: "CASH", IdempotencyKey: "thermal_payment_original"})
	if err != nil || !created {
		t.Fatalf("register: %v created=%v", err, created)
	}
	receipt, err := documents.Issue(ctx, actor, "RECEIPT", paid.Payment.ID)
	if err != nil {
		t.Fatal(err)
	}
	ficha, err := documents.Issue(ctx, actor, "RECEPTION", order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Number == ficha.Number || receipt.State != "VALID" || ficha.PaymentID != nil {
		t.Fatalf("documents: %+v %+v", receipt, ficha)
	}
	original := append([]byte{}, receipt.Snapshot...)
	if !bytes.Contains(original, []byte(plate)) || !bytes.Contains(original, []byte("Lavado original")) {
		t.Fatal("snapshot lacks plate or service")
	}
	if err := tx.Model(&client).Update("full_name", "Cliente Modificado").Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Model(&vehicle).Update("plate", "THERM456").Error; err != nil {
		t.Fatal(err)
	}
	again, err := documents.Issue(ctx, actor, "RECEIPT", paid.Payment.ID)
	if err != nil || again.ID != receipt.ID || !bytes.Equal(again.Snapshot, original) {
		t.Fatalf("original changed: %v", err)
	}
	a, err := documents.RequestPrint(ctx, actor, receipt.ID, "print_original_123456")
	if err != nil {
		t.Fatal(err)
	}
	retry, err := documents.RequestPrint(ctx, actor, receipt.ID, "print_original_123456")
	if err != nil {
		t.Fatal(err)
	}
	copy, err := documents.RequestPrint(ctx, actor, receipt.ID, "print_duplicate_123456")
	if err != nil {
		t.Fatal(err)
	}
	if a.Attempt.CopyNumber != 0 || retry.Attempt.ID != a.Attempt.ID || copy.Attempt.CopyNumber != 1 {
		t.Fatal("print idempotency/copy numbering failed")
	}
	if len(copy.Document.Attempts) != 2 || copy.Document.Attempts[1].ActorUserID != actor.ID {
		t.Fatal("audit missing")
	}
	var count int64
	tx.Model(&models.Payment{}).Where("order_id = ?", order.ID).Count(&count)
	if count != 1 {
		t.Fatal("printing created another payment")
	}
	// Savepoints let PostgreSQL reject mutations without aborting the fixture transaction.
	for _, query := range []string{
		fmt.Sprintf("UPDATE documents SET snapshot = '{}'::jsonb WHERE id = %d", receipt.ID),
		fmt.Sprintf("DELETE FROM documents WHERE id = %d", receipt.ID),
		fmt.Sprintf("UPDATE document_print_attempts SET copy_number = 20 WHERE id = %d", a.Attempt.ID),
	} {
		tx.SavePoint("immutable_check")
		if err := tx.Exec(query).Error; err == nil {
			t.Fatal("history mutation accepted")
		}
		tx.RollbackTo("immutable_check")
	}
	for _, role := range []string{models.RoleAccountant, models.RoleWashStaff, "unknown"} {
		denied := actor
		denied.Role = role
		if _, err := documents.Get(ctx, denied, receipt.ID); !errors.Is(err, ErrPaymentForbidden) {
			t.Fatalf("%s read allowed", role)
		}
		if _, err := documents.Issue(ctx, denied, "RECEPTION", order.ID); !errors.Is(err, ErrPaymentForbidden) {
			t.Fatalf("%s issue allowed", role)
		}
		if _, err := documents.RequestPrint(ctx, denied, receipt.ID, "print_forbidden_123456"); !errors.Is(err, ErrPaymentForbidden) {
			t.Fatalf("%s print allowed", role)
		}
	}
	corrected, created, err := payments.Correct(ctx, actor, paid.Payment.ID, CorrectPaymentInput{Version: paid.Order.Version, Method: "QR", Reason: "Método equivocado", IdempotencyKey: "thermal_payment_replacement"})
	if err != nil || !created {
		t.Fatalf("correct: %v", err)
	}
	old, err := documents.Get(ctx, actor, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := documents.Issue(ctx, actor, "RECEIPT", corrected.Payment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if old.State != "REPLACED" || old.ReplacementNumber != fresh.Number || fresh.ReplacesNumber != old.Number || !bytes.Equal(old.Snapshot, original) {
		t.Fatal("replacement lost original or linking")
	}
	if _, err := payments.Void(ctx, actor, corrected.Payment.ID, VoidPaymentInput{Version: corrected.Order.Version, Reason: "Anulación de prueba"}); err != nil {
		t.Fatal(err)
	}
	voided, err := documents.Get(ctx, actor, fresh.ID)
	if err != nil || voided.State != "VOIDED" {
		t.Fatalf("void state missing: %v %+v", err, voided)
	}
	printed, err := documents.RequestPrint(ctx, actor, fresh.ID, "print_voided_12345678")
	if err != nil || printed.Document.State != "VOIDED" {
		t.Fatal("print did not refresh void status")
	}
	admin := actor
	admin.Role = models.RoleAdmin
	if _, err := documents.Get(ctx, admin, ficha.ID); err != nil {
		t.Fatal(err)
	}
}
