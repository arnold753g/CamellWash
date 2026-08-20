package services

import (
	"errors"
	"testing"

	"starter-backend/internal/models"
)

func TestNormalizePaymentFields(t *testing.T) {
	t.Parallel()
	method, reference, key, err := normalizePaymentFields(" qr ", " Ref  123 ", "payment_key_123456789")
	if err != nil || method != models.PaymentMethodQR || reference == nil || *reference != "Ref 123" || key != "payment_key_123456789" {
		t.Fatalf("normalized=(%q,%v,%q,%v)", method, reference, key, err)
	}
	if _, _, _, err := normalizePaymentFields("crypto", "", "payment_key_123456789"); err == nil {
		t.Fatal("unknown method was accepted")
	}
	var fieldError FieldError
	_, _, _, err = normalizePaymentFields("cash", "", "short")
	if !errors.As(err, &fieldError) || fieldError.Field != "idempotency_key" {
		t.Fatalf("idempotency error=%v", err)
	}
}

func TestPaymentReasonAndMethods(t *testing.T) {
	t.Parallel()
	for _, method := range []string{models.PaymentMethodCash, models.PaymentMethodQR, models.PaymentMethodBankTransfer, models.PaymentMethodCard} {
		if !ValidPaymentMethod(method) {
			t.Fatalf("method %s is invalid", method)
		}
	}
	if _, err := normalizePaymentReason("reason", "no"); err == nil {
		t.Fatal("short reason was accepted")
	}
}
