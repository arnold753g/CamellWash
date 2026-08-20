package database

import (
	"strings"
	"testing"
)

func TestLavaderoRoleMigrationClassifiesOnlyApprovedLegacyUser(t *testing.T) {
	t.Parallel()

	body, err := migrations.ReadFile("migrations/000002_lavadero_roles.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"WHERE role = 'user'",
		"RAISE EXCEPTION",
		"LOWER(email) <> 'user@example.test'",
		"SET role = 'manager'",
		"LOWER(email) = 'user@example.test'",
		"'admin', 'manager', 'wash_staff'",
		"ALTER COLUMN role DROP DEFAULT",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration does not enforce %q", required)
		}
	}
}
