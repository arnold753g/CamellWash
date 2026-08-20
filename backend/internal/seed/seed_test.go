package seed

import "testing"

func TestDevelopmentSeedRequiresExplicitConfirmation(t *testing.T) {
	t.Setenv("SEED_DEVELOPMENT", "false")
	if err := Development(nil, "development", 10); err == nil {
		t.Fatal("expected development seed without confirmation to fail")
	}
}

func TestDevelopmentSeedIsBlockedInProduction(t *testing.T) {
	t.Setenv("SEED_DEVELOPMENT", "true")
	if err := Development(nil, "production", 10); err == nil {
		t.Fatal("expected production seed to fail")
	}
}
