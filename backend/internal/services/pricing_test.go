package services

import (
	"errors"
	"testing"
)

func TestNormalizePrice(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "1", want: "1.00"},
		{input: "001.5", want: "1.50"},
		{input: "9999999999.99", want: "9999999999.99"},
		{input: " 12.30 ", want: "12.30"},
	} {
		got, err := NormalizePrice(test.input)
		if err != nil || got != test.want {
			t.Fatalf("NormalizePrice(%q)=(%q,%v) want=(%q,nil)", test.input, got, err, test.want)
		}
	}
}

func TestNormalizePriceRejectsInvalidAndZero(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "-1", "1.234", "1,20", "1e2", "10000000000", "abc"} {
		if _, err := NormalizePrice(input); !errors.Is(err, ErrInvalidPrice) {
			t.Fatalf("NormalizePrice(%q) error=%v want ErrInvalidPrice", input, err)
		}
	}
	for _, input := range []string{"0", "0.0", "000.00"} {
		if _, err := NormalizePrice(input); !errors.Is(err, ErrPriceMustBePositive) {
			t.Fatalf("NormalizePrice(%q) error=%v want ErrPriceMustBePositive", input, err)
		}
	}
}
