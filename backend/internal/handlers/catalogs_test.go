package handlers

import "testing"

func TestNormalizeCatalogDisplayName(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "  Automóvil  ", want: "Automóvil"},
		{input: "Lavado\t Premium\ncon aspirado", want: "Lavado Premium con aspirado"},
		{input: "SUV/Camioneta", want: "SUV/Camioneta"},
	} {
		if got := normalizeCatalogDisplayName(test.input); got != test.want {
			t.Fatalf("normalizeCatalogDisplayName(%q)=%q want=%q", test.input, got, test.want)
		}
	}
}
