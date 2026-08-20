package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidRoleAcceptsOnlyLavaderoRoles(t *testing.T) {
	t.Parallel()

	for _, role := range []string{RoleAdmin, RoleManager, RoleWashStaff} {
		if !ValidRole(role) {
			t.Fatalf("expected role %q to be valid", role)
		}
	}
	for _, role := range []string{"", "user", "administrator", "manager "} {
		if ValidRole(role) {
			t.Fatalf("expected role %q to be invalid", role)
		}
	}
}

func TestPublicUserNeverSerializesPasswordHash(t *testing.T) {
	t.Parallel()

	user := User{ID: 7, Email: "persona@example.test", PasswordHash: "sensitive-hash", Role: RoleManager, Status: StatusActive}
	body, err := json.Marshal(user.Public())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "sensitive-hash") || strings.Contains(string(body), "password") {
		t.Fatalf("public user leaked password data: %s", body)
	}
}
