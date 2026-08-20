package database

import (
	"strings"
	"testing"
)

func TestPaymentMigrationContract(t *testing.T) {
	t.Parallel()
	body, err := migrations.ReadFile("migrations/000009_payments.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE payments", "NUMERIC(12,2)", "CASH", "BANK_TRANSFER",
		"payments_one_valid_per_order", "WHERE status = 'VALID'", "idempotency_key",
		"PENDING_REVIEW", "replaces_payment_id", "CREATE TABLE payment_events",
		"payment_events_append_only", "ON DELETE RESTRICT",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("payment migration does not enforce %q", required)
		}
	}
}
