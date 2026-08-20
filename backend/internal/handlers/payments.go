package handlers

import (
	"errors"
	"net/http"

	"starter-backend/internal/respond"
	paymentservice "starter-backend/internal/services"
)

type PaymentHandler struct {
	payments *paymentservice.PaymentService
}

type registerPaymentRequest struct {
	Version           uint64 `json:"version"`
	Method            string `json:"method"`
	ExternalReference string `json:"external_reference"`
	IdempotencyKey    string `json:"idempotency_key"`
	Deliver           bool   `json:"deliver"`
}

type voidPaymentRequest struct {
	Version uint64 `json:"version"`
	Reason  string `json:"reason"`
}

type correctPaymentRequest struct {
	Version           uint64 `json:"version"`
	Method            string `json:"method"`
	ExternalReference string `json:"external_reference"`
	IdempotencyKey    string `json:"idempotency_key"`
	Reason            string `json:"reason"`
}

type reviewPaymentRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

func NewPaymentHandler(payments *paymentservice.PaymentService) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

func (h *PaymentHandler) Register(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	orderID, ok := pathID(w, r)
	if !ok {
		return
	}
	var request registerPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	result, created, err := h.payments.Register(r.Context(), actor, orderID, paymentservice.RegisterPaymentInput(request))
	if err != nil {
		handlePaymentError(w, r, err)
		return
	}
	status := http.StatusOK
	message := "Payment already registered"
	if created {
		status = http.StatusCreated
		message = "Payment registered"
	}
	respond.JSON(w, status, result, message)
}

func (h *PaymentHandler) ListForOrder(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	orderID, ok := pathID(w, r)
	if !ok {
		return
	}
	items, err := h.payments.ListForOrder(r.Context(), actor, orderID)
	if err != nil {
		handlePaymentError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"payments": items}, "Order payments loaded")
}

func (h *PaymentHandler) Void(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	paymentID, ok := pathID(w, r)
	if !ok {
		return
	}
	var request voidPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	result, err := h.payments.Void(r.Context(), actor, paymentID, paymentservice.VoidPaymentInput(request))
	if err != nil {
		handlePaymentError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, result, "Payment voided")
}

func (h *PaymentHandler) Correct(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	paymentID, ok := pathID(w, r)
	if !ok {
		return
	}
	var request correctPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	result, created, err := h.payments.Correct(r.Context(), actor, paymentID, paymentservice.CorrectPaymentInput(request))
	if err != nil {
		handlePaymentError(w, r, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	respond.JSON(w, status, result, "Payment corrected")
}

func (h *PaymentHandler) PendingReviews(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	page := queryInt(r, "page", 1, 1, 1000000)
	limit := queryInt(r, "limit", 20, 1, 100)
	items, pagination, err := h.payments.PendingReviews(r.Context(), actor, page, limit)
	if err != nil {
		handlePaymentError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"payments": items, "pagination": pagination}, "Pending payment reviews loaded")
}

func (h *PaymentHandler) Review(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	paymentID, ok := pathID(w, r)
	if !ok {
		return
	}
	var request reviewPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.payments.Review(r.Context(), actor, paymentID, paymentservice.ReviewPaymentInput(request))
	if err != nil {
		handlePaymentError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, item, "Payment review recorded")
}

func handlePaymentError(w http.ResponseWriter, r *http.Request, err error) {
	var fieldError paymentservice.FieldError
	switch {
	case errors.As(err, &fieldError):
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", map[string]string{fieldError.Field: fieldError.Message})
	case errors.Is(err, paymentservice.ErrOrderNotFound), errors.Is(err, paymentservice.ErrPaymentNotFound):
		respond.Error(w, r, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment or order was not found", nil)
	case errors.Is(err, paymentservice.ErrPaymentForbidden), errors.Is(err, paymentservice.ErrOrderForbidden):
		respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission for payment operations", nil)
	case errors.Is(err, paymentservice.ErrPaymentAlreadyExists):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_ALREADY_EXISTS", "This order already has a valid payment", nil)
	case errors.Is(err, paymentservice.ErrPaymentNotValid):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_NOT_VALID", "Only the current valid payment can be changed", nil)
	case errors.Is(err, paymentservice.ErrPaymentOrderUnavailable):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_ORDER_UNAVAILABLE", "Cancelled or delivered orders cannot receive a new standalone payment", nil)
	case errors.Is(err, paymentservice.ErrPaymentDeliveryState):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_DELIVERY_STATE", "Payment and delivery together require a finished order", nil)
	case errors.Is(err, paymentservice.ErrPaymentDeliveredCorrection):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_REQUIRES_CORRECTION", "A delivered order requires atomic payment correction", nil)
	case errors.Is(err, paymentservice.ErrPaymentConflict):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_CONFLICT", "Order or payment changed; reload before retrying", nil)
	case errors.Is(err, paymentservice.ErrPaymentReviewUnavailable):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_REVIEW_UNAVAILABLE", "This payment no longer requires review", nil)
	case errors.Is(err, paymentservice.ErrPaymentSelfReview):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_SELF_REVIEW", "The user who voided a payment cannot review it", nil)
	case errors.Is(err, paymentservice.ErrPaymentIdempotencyConflict):
		respond.Error(w, r, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "This operation key was already used with different data", nil)
	default:
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Payment operation could not be completed", nil)
	}
}
