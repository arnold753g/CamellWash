package services

import (
	"errors"
	"testing"

	"starter-backend/internal/models"
)

func TestOrderStateMachine(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		models.OrderPending: models.OrderInProgress, models.OrderInProgress: models.OrderQualityControl,
		models.OrderQualityControl: models.OrderFinished, models.OrderFinished: models.OrderDelivered,
		models.OrderDelivered: "", models.OrderCancelled: "",
	}
	for current, expected := range cases {
		if got := NextOrderStatus(current); got != expected {
			t.Fatalf("next(%s)=%s want=%s", current, got, expected)
		}
	}
}

func TestOrderIdentifiersAndMoney(t *testing.T) {
	t.Parallel()
	if _, err := uniqueOrderIDs("service_ids", []uint64{1, 1}, 50); err == nil {
		t.Fatal("duplicate service was accepted")
	}
	if got, err := priceToCents("123.45"); err != nil || got != 12345 {
		t.Fatalf("price cents=(%d,%v)", got, err)
	}
	if got := centsToPrice(12345); got != "123.45" {
		t.Fatalf("price=%s", got)
	}
	var fieldError FieldError
	_, err := uniqueOrderIDs("staff_ids", nil, 25)
	if !errors.As(err, &fieldError) || fieldError.Field != "staff_ids" {
		t.Fatalf("empty staff error=%v", err)
	}
}
