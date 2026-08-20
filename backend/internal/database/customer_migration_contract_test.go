package database

import (
	"strings"
	"testing"
)

func TestCustomerMigrationContract(t *testing.T) {
	t.Parallel()

	body, err := migrations.ReadFile("migrations/000006_clients_vehicles.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE clients",
		"CREATE TABLE vehicles",
		"clients_normalized_ci_unique",
		"WHERE normalized_ci IS NOT NULL",
		"normalized_ci <> '00000000'",
		"vehicles_normalized_plate_unique",
		"WHERE normalized_plate IS NOT NULL",
		"REFERENCES clients(id) ON DELETE RESTRICT",
		"REFERENCES vehicle_types(id) ON DELETE RESTRICT",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("customer migration does not enforce %q", required)
		}
	}

	formatBody, err := migrations.ReadFile("migrations/000007_customer_identifier_format.sql")
	if err != nil {
		t.Fatal(err)
	}
	formatSQL := string(formatBody)
	for _, required := range []string{"clients_ci_format", "^[A-Z0-9]{4,20}$", "vehicles_plate_format", "^[A-Z0-9]{3,15}$"} {
		if !strings.Contains(formatSQL, required) {
			t.Fatalf("customer identifier migration does not enforce %q", required)
		}
	}
}
