package database

import (
	"strings"
	"testing"
)

func TestReportIndexesCoverDateAndStatusQueries(t *testing.T) {
	t.Parallel()
	body, err := migrations.ReadFile("migrations/000010_report_indexes.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"idx_orders_received_at", "received_at DESC",
		"idx_order_status_history_status_time", "to_status, changed_at DESC",
		"idx_payments_status_paid_at", "status, paid_at DESC",
		"idx_payments_voided_at", "WHERE status = 'VOIDED'",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("report migration does not contain %q", required)
		}
	}
}
