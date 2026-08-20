package database

import (
	"strings"
	"testing"
)

func TestOrderMigrationContract(t *testing.T) {
	t.Parallel()
	body, err := migrations.ReadFile("migrations/000008_orders.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE orders", "NUMERIC(12,2)", "orders_estimated_delivery",
		"CREATE TABLE order_services", "service_name_snapshot", "price_snapshot",
		"UNIQUE (order_id, service_id)", "CREATE TABLE order_staff", "PRIMARY KEY (order_id, user_id)",
		"CREATE TABLE order_status_history", "order_status_history_append_only",
		"ON DELETE RESTRICT", "version BIGINT NOT NULL DEFAULT 1",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("order migration does not enforce %q", required)
		}
	}
}
