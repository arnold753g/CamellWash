package services

import (
	"errors"
	"testing"
)

func TestNormalizeOptionalCI(t *testing.T) {
	t.Parallel()

	if got, err := NormalizeCI(nil); err != nil || got != nil {
		t.Fatalf("nil CI=(%v,%v) want=(nil,nil)", got, err)
	}
	empty := "  "
	if got, err := NormalizeCI(&empty); err != nil || got != nil {
		t.Fatalf("empty CI=(%v,%v) want=(nil,nil)", got, err)
	}
	value := "  12-34 lp "
	got, err := NormalizeCI(&value)
	if err != nil || got == nil || *got != "12-34 LP" {
		t.Fatalf("normalized CI=(%v,%v)", got, err)
	}
	placeholder := MissingCIDisplay
	if _, err := NormalizeCI(&placeholder); err == nil {
		t.Fatal("reserved CI placeholder was accepted")
	}
}

func TestMissingCIDisplayAndMask(t *testing.T) {
	t.Parallel()

	if value, provided := ClientCIForDisplay(nil); value != MissingCIDisplay || provided {
		t.Fatalf("missing display=(%q,%v)", value, provided)
	}
	if got := MaskPersonalValue("12345678"); got != "••••5678" {
		t.Fatalf("mask=%q", got)
	}
}

func TestNormalizePhoneAndPlate(t *testing.T) {
	t.Parallel()

	if got, err := NormalizePhone("+591 700-12345"); err != nil || got != "+59170012345" {
		t.Fatalf("phone=(%q,%v)", got, err)
	}
	plate := " 1234-abc "
	got, err := NormalizePlate(&plate)
	if err != nil || got == nil || *got != "1234ABC" {
		t.Fatalf("plate=(%v,%v)", got, err)
	}
	for _, invalid := range []string{"12", "ABC$123"} {
		value := invalid
		if _, err := NormalizePlate(&value); err == nil {
			t.Fatalf("invalid plate %q was accepted", invalid)
		}
	}
	if _, err := NormalizePhone("123"); err == nil {
		t.Fatal("short phone was accepted")
	}
	var fieldError FieldError
	if _, err := NormalizePhone("phone"); !errors.As(err, &fieldError) || fieldError.Field != "phone" {
		t.Fatalf("phone error=%v", err)
	}
}

func TestSearchNormalizationDoesNotTurnSymbolsIntoWildcards(t *testing.T) {
	t.Parallel()

	if got := escapeSearchLike("50%_!"); got != "50!%!_!!" {
		t.Fatalf("escaped LIKE=%q", got)
	}
	if got := normalizeSearchIdentifier("%_!"); got != "" {
		t.Fatalf("symbol-only identifier=%q", got)
	}
	if got := normalizeSearchPhone("%_!"); got != "" {
		t.Fatalf("symbol-only phone=%q", got)
	}
}
