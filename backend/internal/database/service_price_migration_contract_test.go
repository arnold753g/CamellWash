package database

import (
	"strings"
	"testing"
)

func TestServicePriceMigrationContract(t *testing.T) {
	t.Parallel()

	body, err := migrations.ReadFile("migrations/000005_service_prices.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE service_prices",
		"price NUMERIC(12, 2) NOT NULL",
		"CHECK (price > 0)",
		"UNIQUE (service_id, vehicle_type_id)",
		"REFERENCES services(id) ON DELETE RESTRICT",
		"REFERENCES vehicle_types(id) ON DELETE RESTRICT",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("service price migration does not enforce %q", required)
		}
	}
}
