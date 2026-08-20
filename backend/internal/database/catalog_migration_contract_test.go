package database

import (
	"strings"
	"testing"
)

func TestCatalogMigrationContract(t *testing.T) {
	t.Parallel()

	body, err := migrations.ReadFile("migrations/000003_catalogs.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"normalize_catalog_name",
		"CREATE TABLE vehicle_types",
		"CREATE TABLE service_categories",
		"CREATE TABLE services",
		"GENERATED ALWAYS AS",
		"UNIQUE (category_id, normalized_name)",
		"ON DELETE RESTRICT",
		"active BOOLEAN NOT NULL DEFAULT TRUE",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("catalog migration does not enforce %q", required)
		}
	}
}
