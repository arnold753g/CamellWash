package database

import (
	"strings"
	"testing"
)

func TestCatalogInitialValuesMigrationContract(t *testing.T) {
	t.Parallel()

	body, err := migrations.ReadFile("migrations/000004_catalog_initial_values.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"ON CONFLICT (normalized_name) DO NOTHING",
		"ON CONFLICT (category_id, normalized_name) DO NOTHING",
		"'Moto'",
		"'Torito/Mototaxi'",
		"'Automóvil'",
		"'SUV/Camioneta'",
		"'Furgoneta'",
		"'Camión'",
		"'Lavado'",
		"'Detailing'",
		"'Chasis'",
		"'Básico'",
		"'Premium con aspirado'",
		"'Raibanizada Americana'",
		"'Raibanizada China'",
		"'PPF'",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("initial catalog migration does not contain %q", required)
		}
	}
}
